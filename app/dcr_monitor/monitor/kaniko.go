// Copyright 2024 TikTok Pte. Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package monitor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/manatee-project/manatee/app/dcr_api/biz/model/job"
	"github.com/manatee-project/manatee/app/dcr_monitor/client"
	"github.com/manatee-project/manatee/pkg/cloud"
	"github.com/manatee-project/manatee/pkg/config"
)

func CheckKanikoJobs(ctx context.Context, clientSet *kubernetes.Clientset) error {
	hlog.Info("start to monitor kaniko build jobs")
	jobs, err := clientSet.BatchV1().Jobs(client.RunningNameSpace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		hlog.Errorf("[KanikoJobMonitor]failed to get job: %v", err)
		return errors.Wrap(err, "failed to list jobs")
	}

	for _, j := range jobs.Items {
		if !strings.Contains(j.Name, "kaniko-") {
			continue
		}
		if len(j.Status.Conditions) == 0 {
			hlog.Infof("[KanikoJobMonitor]job %v is still running", j.Name)
			continue
		}

		UUID, ok := j.ObjectMeta.Annotations["JOB_UUID"]
		if !ok {
			hlog.Error("[KanikoJobMonitor]failed to get job's UUID")
			continue
		}
		token, ok := j.ObjectMeta.Annotations["USER_TOKEN"]
		if !ok {
			hlog.Error("[KanikoJobMonitor]failed to get user's token")
			continue
		}
		creator, ok := j.ObjectMeta.Annotations["JOB_CREATOR"]
		if !ok {
			hlog.Error("[KanikoJobMonitor]failed to get job's creator")
			continue
		}

		hlog.Infof("[KanikoJobMonitor]job name: %v, job status: %v", j.Name, j.Status.Conditions[0].Type)
		if j.Status.Conditions[0].Type == batchv1.JobComplete {
			digest, err := getImageDigest(ctx, clientSet, j.Name, client.RunningNameSpace)
			if err != nil {
				hlog.Errorf("[KanikoJobMonitor]failed to get image digest: %+v", err)
				return err
			}
			err = updateJobStatus(creator, UUID, token, digest, int64(job.JobStatus_VMWaiting))
			if err != nil {
				return err
			}

			err = deleteJob(ctx, clientSet, j.Name, client.RunningNameSpace)
			if err != nil {
				hlog.Errorf("[KanikoJobMonitor]failed to delete job: %+v", err)
				return err
			}
			// deleteBuildContext(ctx, creator, UUID)
		} else if j.Status.Conditions[0].Type == batchv1.JobFailed {
			err = updateJobStatus(creator, UUID, token, "", int64(job.JobStatus_ImageBuildingFailed))
			if err != nil {
				return err
			}
			// deleteBuildContext(ctx, creator, UUID)
		}
	}
	return nil
}

func deleteJob(ctx context.Context, clientSet *kubernetes.Clientset, jobName string, namespace string) error {
	hlog.Infof("[KanikoJobMonitor]delete job: %v", jobName)
	deletePolicy := metav1.DeletePropagationForeground
	if err := clientSet.BatchV1().Jobs(namespace).Delete(ctx, jobName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		return errors.Wrap(err, "failed to delete job")
	}
	return nil
}

func getImageDigest(ctx context.Context, clientSet *kubernetes.Clientset, jobName string, namespace string) (string, error) {
	pods, err := clientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "job-name=" + jobName,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to list pod")
	}
	hlog.Infof("[KanikoJobMonitor] pods num: %d", len(pods.Items))
	for _, pod := range pods.Items {
		req := clientSet.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})
		logs, err := req.Stream(ctx)
		if err != nil {
			return "", errors.Wrap(err, "failed to read log stream")
		}

		digest, err := getDigestFromLog(logs)
		if err != nil {
			return "", err
		}
		_ = logs.Close()
		if digest == "" {
			continue
		}
		hlog.Infof("[KanikoJobMonitor]got image digest %v", digest)
		return digest, nil
	}
	return "", errors.New("failed to read digest")
}

func getDigestFromLog(reader io.Reader) (string, error) {
	var lastLine string
	scanner := bufio.NewScanner(reader)
	// get the last line
	for scanner.Scan() {
		lastLine = scanner.Text()
	}
	if scanner.Err() != nil {
		return "", errors.Wrap(scanner.Err(), "failed to get last line")
	}
	r, _ := regexp.Compile(`sha256:[a-z0-9]+`)
	digest := r.FindString(lastLine)
	return digest, nil
}

func updateJobStatus(creator, UUID, token, digest string, status int64) error {
	ctx := context.Background()
	req := &protocol.Request{}
	res := &protocol.Response{}
	req.SetMethod(consts.MethodPost)
	req.Header.SetContentTypeBytes([]byte("application/json"))
	apiHost := os.Getenv("DATA_CLEAN_ROOM_HOST")
	if apiHost == "" {
		return errors.New("DATA_CLEAN_ROOM_HOST environment variable not set")
	}
	req.SetRequestURI(apiHost + client.UpdatePath)
	req.SetHeader("Authorization", token)

	attestationReport := ""
	var err error
	if job.JobStatus(status) == job.JobStatus_VMFinished {
		attestationReport, err = getJobAttestationReport(ctx, creator, UUID)
		if err != nil {
			return err
		}
	}

	request := &job.UpdateJobStatusRequest{
		UUID:              UUID,
		Status:            job.JobStatus(status),
		DockerImageDigest: digest,
		DockerImage:       config.GetJobDockerImageFull(creator, UUID),
		Creator:           creator,
		AttestationToken:  attestationReport,
	}
	jsonByte, _ := json.Marshal(request)
	req.SetBody(jsonByte)
	err = client.HTTPClient.Do(ctx, req, res)
	if err != nil {
		return errors.Wrap(err, "failed to update job status")
	}
	if res.StatusCode() != 200 {
		return errors.Wrap(err, "failed to update job status")
	}
	hlog.Infof("resp %v", res.Body())
	resp := &job.UpdateJobStatusResponse{}
	err = json.Unmarshal(res.Body(), resp)
	if err != nil {
		return errors.Wrap(err, "failed to update job status")
	}
	if resp.Code != 0 {
		return fmt.Errorf("fail to update job status, response: %s", resp.Msg)
	}
	return nil
}

func deleteBuildContext(ctx context.Context, creator string, UUID string) {
	provider := cloud.GetCloudProvider(ctx)
	err := provider.DeleteFile(config.GetBuildContextPath(creator, UUID))
	if err != nil {
		hlog.Errorf("[KubernetesBuildService]failed to delete build context %+v", err)
	}
}

func getJobAttestationReport(ctx context.Context, creator, UUID string) (string, error) {
	provider := cloud.GetCloudProvider(ctx)
	attestationReportPath := config.GetCustomTokenPath(creator, UUID)
	chunkSize := 1024 * 1024 * 3
	token, err := provider.GetFilebyChunk(attestationReportPath, 0, int64(chunkSize))
	if err != nil {
		return "", err
	}
	return string(token), nil
}
