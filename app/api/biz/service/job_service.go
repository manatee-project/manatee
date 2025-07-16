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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/google/uuid"
	"github.com/manatee-project/manatee/app/api/biz/dal/db"
	"github.com/manatee-project/manatee/app/api/biz/model/job"
	"github.com/manatee-project/manatee/app/api/biz/pkg/errno"
	"github.com/manatee-project/manatee/app/api/biz/pkg/storage"
	"github.com/pkg/errors"
)

type JobService struct {
	ctx     context.Context
	storage storage.Storage
}

// NewJobService create job service
func NewJobService(ctx context.Context) *JobService {
	storage, err := storage.GetStorage(ctx)
	if err != nil {
		panic(err)
	}
	return &JobService{
		ctx:     ctx,
		storage: storage,
	}
}

func (js *JobService) Drop() {
	js.storage.Close()
}

func (js *JobService) SubmitJob(req *job.SubmitJobRequest, userWorkspace io.Reader) (string, error) {
	creator := req.Creator

	jobInList, err := db.GetInProgressJobs(creator)
	if err != nil {
		return "", err
	} else if len(jobInList) > 2 {
		return "", errors.Wrap(fmt.Errorf("%s", errno.ReachJobLimitErrMsg), "")
	}

	var keys []string
	var extraEnvs = make(map[string]string)
	for _, v := range req.GetEnvs() {
		keys = append(keys, v.GetKey())
		extraEnvs[v.GetKey()] = v.GetValue()
	}

	// inject Dockerfile into the build context
	dockerFileContent := js.generateDockerfile(keys)
	buildctx, err := js.addDockerfileToTarGz(userWorkspace, string(dockerFileContent))
	if err != nil {
		return "", err
	}
	remotePath := fmt.Sprintf("%s/%s-workspace.tar.gz", creator, creator)
	err = js.storage.UploadFile(buildctx, remotePath, false)
	if err != nil {
		return "", err
	}
	buildctxpath := fmt.Sprintf("%s/%s/%s-workspace.tar.gz", js.storage.BucketPath(), creator, creator)

	uuidStr, err := uuid.NewUUID()
	if err != nil {
		return "", errors.Wrap(err, "failed to generate uuid")
	}

	// generate signed put url for the output and custom token
	outputPath := fmt.Sprintf("%s/output/out-%s-%s", creator, uuidStr.String(), req.JupyterFileName)
	outputPutSignedUrl, err := js.storage.IssueSignedUrl(outputPath, "PUT", time.Hour*6)
	if err != nil {
		return "", err
	}
	customTokenPath := fmt.Sprintf("%s/output/%s-token", creator, uuidStr.String())
	customTokenPathPutSignedUrl, err := js.storage.IssueSignedUrl(customTokenPath, "PUT", time.Hour*6)
	if err != nil {
		return "", err
	}
	t := db.Job{
		UUID:                    uuidStr.String(),
		Dockerfile:              dockerFileContent,
		Creator:                 req.Creator,
		JupyterFileName:         req.JupyterFileName,
		JobStatus:               int(job.JobStatus_Created),
		BuildContextPath:        buildctxpath,
		OutputPutSignedUrl:      outputPutSignedUrl,
		CustomTokenPutSignedUrl: customTokenPathPutSignedUrl,
		ExtraEnvs:               extraEnvs,
	}
	err = db.CreateJob(&t)

	if err != nil {
		return "", err
	}
	hlog.Infof("[JobService] inserted job. Job Status %+v", job.JobStatus_ImageBuilding)
	return uuidStr.String(), nil
}

var dockerFileTemplate string = `ARG BASE_IMAGE
ARG BASE_IMAGE
FROM $BASE_IMAGE
ARG OUTPUT_SIGNED_URL
ARG JUPYTER_FILENAME
ARG CUSTOMTOKEN_SIGNED_URL 

ENV OUTPUT_SIGNED_URL=$OUTPUT_SIGNED_URL
ENV JUPYTER_FILENAME=$JUPYTER_FILENAME
ENV CUSTOMTOKEN_SIGNED_URL=$CUSTOMTOKEN_SIGNED_URL

WORKDIR /home/jovyan
COPY $USER_WORKSPACE/* ./
%s

ENTRYPOINT rm -rf lm-evaluation-harness \
	&& git clone https://github.com/EleutherAI/lm-evaluation-harness \
	&& pushd lm-evaluation-harness \
	&& git checkout 3102a8e4a8f3a3163a52e943f14068680753356f \
	&& popd \
	&& pip install -e ./lm-evaluation-harness[wandb] \
	&& jupyter nbconvert --execute --to notebook --inplace $JUPYTER_FILENAME --ExecutePreprocessor.timeout=-1 --allow-errors \
    && hash=$(md5sum $JUPYTER_FILENAME | awk '{ print $1 }') \
    && curl -X PUT -T $JUPYTER_FILENAME $OUTPUT_SIGNED_URL \
    && ./gen_custom_token --nonce $hash \
    && curl -X PUT -T custom_token $CUSTOMTOKEN_SIGNED_URL
`

// TODO: this actually needs to support different TEE backends.
// for now, we only support GCP confidential space.
// in the future, the build context should be completed by the ImageBuilder,
// which will also finalize the Dockerfile.
func (js *JobService) generateDockerfile(keys []string) string {
	if len(keys) == 0 {
		return fmt.Sprintf(dockerFileTemplate, "")
	} else {
		envOverride := fmt.Sprintf(`LABEL "tee.launch_policy.allow_env_override"="%s"`, strings.Join(keys, ","))
		return fmt.Sprintf(dockerFileTemplate, envOverride)
	}
}

func (js *JobService) addDockerfileToTarGz(input io.Reader, dockerfileContent string) (io.Reader, error) {
	gzReader, err := gzip.NewReader(input)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	var buffer bytes.Buffer
	gzWriter := gzip.NewWriter(&buffer)
	defer gzWriter.Close()

	tarReader := tar.NewReader(gzReader)
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Write the existing header and file content to the new tar archive
		if err := tarWriter.WriteHeader(header); err != nil {
			return nil, fmt.Errorf("failed to write tar header: %w", err)
		}

		if _, err := io.Copy(tarWriter, tarReader); err != nil {
			return nil, fmt.Errorf("failed to write tar entry: %w", err)
		}
	}
	// Add the Dockerfile as a new entry
	dockerfileHeader := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len(dockerfileContent)),
		Mode: 0600,
	}
	if err := tarWriter.WriteHeader(dockerfileHeader); err != nil {
		return nil, fmt.Errorf("failed to write Dockerfile header: %w", err)
	}
	if _, err := tarWriter.Write([]byte(dockerfileContent)); err != nil {
		return nil, fmt.Errorf("failed to write Dockerfile content: %w", err)
	}
	// Close the tar and gzip writers
	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar writer: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}
	return &buffer, nil
}

func convertEntityToModel(j *db.Job) *job.Job {
	return &job.Job{
		ID:              int64(j.ID),
		UUID:            j.UUID,
		Creator:         j.Creator,
		JobStatus:       job.JobStatus(j.JobStatus),
		JupyterFileName: j.JupyterFileName,
		CreatedAt:       j.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       j.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (js *JobService) QueryUsersJobs(req *job.QueryJobRequest) ([]*job.Job, int64, error) {
	jobs, total, err := db.QueryJobsByCreator(req.Creator, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}
	res := []*job.Job{}
	for _, j := range jobs {
		res = append(res, convertEntityToModel(j))
	}
	return res, total, nil
}

func (js *JobService) DownloadJobOutput(req *job.DownloadJobOutputRequest) (string, string, error) {
	j, err := db.QueryJobByIdAndCreator(req.ID, req.Creator)
	if err != nil {
		return "", "", err
	}
	outputPath := js.getJobOutputPath(j.Creator, j.UUID, j.JupyterFileName)
	filename := fmt.Sprintf("out-%v-%s", j.ID, j.JupyterFileName)
	signedUrl, err := js.storage.IssueSignedUrl(outputPath, "GET", time.Hour)
	if err != nil {
		return "", "", err
	}

	return signedUrl, filename, nil
}

func (js *JobService) DeleteJob(req *job.DeleteJobRequest) {
	db.DeleteJob(req.Creator, req.UUID)
}

func (js *JobService) GetJobAttestationReport(req *job.QueryJobAttestationRequest) (string, error) {
	j, err := db.QueryJobByIdAndCreator(req.ID, req.Creator)
	if err != nil {
		return "", err
	}
	attestationReportPath := fmt.Sprintf("%s/output/%s-token", j.Creator, j.UUID)
	signedUrl, err := js.storage.IssueSignedUrl(attestationReportPath, "GET", time.Hour)
	if err != nil {
		return "", nil
	}
	return signedUrl, nil
}

func (js *JobService) getJobOutputFilename(UUID string, originName string) string {
	return fmt.Sprintf("out-%s-%s", UUID, originName)
}

func (js *JobService) getJobOutputPath(creator string, UUID string, originName string) string {
	return fmt.Sprintf("%s/output/%s", creator, js.getJobOutputFilename(UUID, originName))
}
