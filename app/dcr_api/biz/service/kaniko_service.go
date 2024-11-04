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

package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/pkg/archive"
	"github.com/manatee-project/manatee/app/dcr_api/biz/dal/db"
	"github.com/manatee-project/manatee/pkg/cloud"
	"github.com/manatee-project/manatee/pkg/config"
	"github.com/manatee-project/manatee/pkg/utils"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubernetesBuildService struct {
	ctx       context.Context
	namespace string
}

func NewKanikoService(ctx context.Context) *KubernetesBuildService {
	buildService := &KubernetesBuildService{
		ctx:       ctx,
		namespace: utils.GetNamespace(),
	}

	return buildService
}

func (k *KubernetesBuildService) CreateBuildCtx(ctx context.Context, creator string) (io.ReadCloser, error) {
	provider := cloud.GetCloudProvider(k.ctx)
	workingDir, err := utils.GetWorkDirectory()
	if err != nil {
		return nil, err
	}
	// change working dir to parent
	err = os.Chdir(workingDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to change work directory")
	}
	// create directory for the user
	directory := filepath.Clean(creator)
	// remove old user directory
	err = os.RemoveAll(directory)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "failed to remove user directory")
		}
	}
	// create user directory
	err = os.MkdirAll(directory, 0700)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create directory for user")
	}
	userWorkspaceTar := filepath.Join(directory, config.GetUserWorkspaceFile(creator))
	// download user's workspace, it's tar.gz file
	if err = provider.DownloadFile(config.GetUserWorkSpacePath(creator), userWorkspaceTar); err != nil {
		return nil, err
	}
	// unzip the user's workspace
	if err = utils.UnTarGz(userWorkspaceTar, directory); err != nil {
		return nil, err
	}
	dockerFile := utils.GetDockerFile()
	// copy dockerfile to directory
	err = utils.CopyFile(dockerFile, fmt.Sprintf("%s/Dockerfile", directory))
	if err != nil {
		return nil, err
	}
	configPath := utils.GetConfigFile()
	if err != nil {
		return nil, err
	}
	// copy config file to directory
	err = utils.CopyFile(configPath, fmt.Sprintf("%s/config.yaml", directory))
	if err != nil {
		return nil, err
	}

	buildCtx, err := archive.TarWithOptions(directory, &archive.TarOptions{})
	if err != nil {
		return nil, err
	}
	return buildCtx, nil
}

func (k *KubernetesBuildService) BuildImage(j *db.Job, token string) error {
	UUID := j.UUID
	creator := j.Creator

	imageTag := fmt.Sprintf("us-docker.pkg.dev/%s/dcr-%s-user-images/%s-%s:latest", config.GetProject(), config.GetEnv(), creator, UUID)
	buildCtx, err := k.CreateBuildCtx(k.ctx, creator)
	if err != nil {
		return err
	}
	defer buildCtx.Close()
	provider := cloud.GetCloudProvider(k.ctx)
	// upload build context
	err = provider.UploadFile(buildCtx, config.GetBuildContextPath(creator, UUID), true)
	if err != nil {
		return err
	}

	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return errors.Wrap(err, "failed to init cluster config")
	}
	clientSet, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes client set")
	}
	kanikoJobName := fmt.Sprintf("kaniko-%s", UUID)
	trustedServiceAccountEmail, err := provider.GetServiceAccountEmail()
	if err != nil {
		return err
	}
	buildArgs := []string{
		fmt.Sprintf("--context=%s", config.GetCloudStoragePath(config.GetBuildContextPath(creator, UUID))),
		fmt.Sprintf("--destination=%s", imageTag),
		fmt.Sprintf("--build-arg=CREATOR=%s", creator),
		fmt.Sprintf("--build-arg=OUTPUTPATH=%s", config.GetCloudStoragePath(config.GetJobOutputPath(creator, UUID, j.JupyterFileName))),
		fmt.Sprintf("--build-arg=ENCRYPTED_FILENAME=%s", config.GetEncryptedJobOutputFilename(UUID, j.JupyterFileName)),
		fmt.Sprintf("--build-arg=ENCRYPTED_CLOUDSTORAGE_PATH=%s", config.GetCloudStoragePath(config.GetEncryptedJobOutputPath(creator, UUID, j.JupyterFileName))),
		fmt.Sprintf("--build-arg=JUPYTER_FILENAME=%s", j.JupyterFileName),
		fmt.Sprintf("--build-arg=USER_WORKSPACE=%s", config.GetUserWorkSpaceDir(creator)),
		fmt.Sprintf("--build-arg=BASE_IMAGE=%s", fmt.Sprintf("us-docker.pkg.dev/%s/dcr-%s-user-images/%s:latest", config.GetProject(), config.GetEnv(), "data-clean-room-base")),
		fmt.Sprintf("--build-arg=CUSTOMTOKEN_CLOUDSTORAGE_PATH=%s", config.GetCloudStoragePath(config.GetCustomTokenPath(creator, UUID))),
		fmt.Sprintf("--build-arg=IMPERSONATION_SERVICE_ACCOUNT=%s", trustedServiceAccountEmail),
	}
	annotations := map[string]string{
		"USER_TOKEN":  token,
		"JOB_UUID":    UUID,
		"JOB_CREATOR": creator,
	}
	err = k.createBuildJob(clientSet, kanikoJobName, buildArgs, annotations)
	if err != nil {
		return err
	}
	return nil
}

func (k *KubernetesBuildService) createBuildJob(clientSet *kubernetes.Clientset, jobName string, buildArgs []string, annotations map[string]string) error {
	memQuantity, err := resource.ParseQuantity("6000M")
	if err != nil {
		return errors.Wrap(err, "failed to parse mem quantity")
	}
	ttlSecondsAfterFinished := int32(3600 * 24)
	jobClient := clientSet.BatchV1().Jobs(k.namespace)
	kanikoJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        jobName,
			Namespace:   k.namespace,
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: config.GetK8sPodServiceAccount(),
					Containers: []corev1.Container{
						{
							Name:  "kaniko",
							Image: "gcr.io/kaniko-project/executor:latest",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: memQuantity,
								},
							},
							Args: append([]string{
								"--dockerfile=Dockerfile",
								"--reproducible",
								"--compressed-caching=false",
								"--cache=true",
								"--cache-ttl=72h",
							}, buildArgs...),
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}
	_, err = jobClient.Create(k.ctx, kanikoJob, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes job")
	}
	return nil
}
