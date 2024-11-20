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
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"cloud.google.com/go/storage"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/google/uuid"
	"github.com/manatee-project/manatee/app/dcr_api/biz/dal/db"
	"github.com/manatee-project/manatee/app/dcr_api/biz/model/job"
	"github.com/manatee-project/manatee/pkg/errno"
	"github.com/pkg/errors"
)

type JobService struct {
	ctx context.Context
}

// NewJobService create job service
func NewJobService(ctx context.Context) *JobService {
	return &JobService{ctx: ctx}
}

func (js *JobService) SubmitJob(req *job.SubmitJobRequest, userWorkspace io.Reader) (string, error) {
	creator := req.Creator

	jobInList, err := db.GetInProgressJobs(creator)
	if err != nil {
		return "", err
	} else if len(jobInList) > 2 {
		return "", errors.Wrap(fmt.Errorf(errno.ReachJobLimitErrMsg), "")
	}

	// inject Dockerfile in /usr/local/dcr_conf/Dockerfile into the build context
	// FIXME: avoid reading the file system, use a string instead in the future.
	dockerFileContent, err := os.ReadFile("/usr/local/dcr_conf/Dockerfile")
	if err != nil {
		return "", err
	}
	buildctx, err := js.addDockerfileToTarGz(userWorkspace, string(dockerFileContent))
	if err != nil {
		return "", err
	}
	err = js.uploadFile(buildctx, fmt.Sprintf("%s/%s-workspace.tar.gz", creator, creator), false)
	if err != nil {
		return "", err
	}
	buildctxpath := fmt.Sprintf("gs://%s/%s/%s-workspace.tar.gz", getBucket(), creator, creator)

	uuidStr, err := uuid.NewUUID()
	if err != nil {
		return "", errors.Wrap(err, "failed to generate uuid")
	}
	t := db.Job{
		UUID:             uuidStr.String(),
		Creator:          req.Creator,
		JupyterFileName:  req.JupyterFileName,
		JobStatus:        int(job.JobStatus_Created),
		BuildContextPath: buildctxpath,
	}
	if err != nil {
		return "", err
	}
	err = db.CreateJob(&t)

	if err != nil {
		return "", err
	}
	hlog.Infof("[JobService] inserted job. Job Status %+v", job.JobStatus_ImageBuilding)
	return uuidStr.String(), nil
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

func (js *JobService) GetJobOutputAttrs(req *job.QueryJobOutputRequest) (string, int64, error) {
	j, err := db.QueryJobByIdAndCreator(req.ID, req.Creator)
	if err != nil {
		return "", 0, err
	}
	outputPath := js.getJobOutputPath(j.Creator, j.UUID, j.JupyterFileName)

	size, err := js.getFileSize(outputPath)
	if err != nil {
		return "", 0, err
	}
	return js.getJobOutputFilename(fmt.Sprintf("%v", j.ID), j.JupyterFileName), size, nil
}

func (js *JobService) DownloadJobOutput(req *job.DownloadJobOutputRequest) (string, error) {
	j, err := db.QueryJobByIdAndCreator(req.ID, req.Creator)
	if err != nil {
		return "", err
	}
	outputPath := js.getJobOutputPath(j.Creator, j.UUID, j.JupyterFileName)
	datg, err := js.getFilebyChunk(outputPath, req.Offset, req.Chunk)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(datg)
	return encoded, nil
}

func (js *JobService) DeleteJob(req *job.DeleteJobRequest) {
	db.DeleteJob(req.Creator, req.UUID)
}

func (js *JobService) GetJobAttestationReport(req *job.QueryJobAttestationRequest) (string, error) {
	j, err := db.QueryJobByIdAndCreator(req.ID, req.Creator)
	if err != nil {
		return "", err
	}
	if j.AttestationReport == "" {
		return "", errors.Wrap(fmt.Errorf("failed to query attestation for job %v", req.ID), "")
	}
	return j.AttestationReport, nil
}

func (js *JobService) getJobOutputFilename(UUID string, originName string) string {
	return fmt.Sprintf("out-%s-%s", UUID, originName)
}

func (js *JobService) getJobOutputPath(creator string, UUID string, originName string) string {
	return fmt.Sprintf("%s/output/%s", creator, js.getJobOutputFilename(UUID, originName))
}

func (g *JobService) getFileSize(remotePath string) (int64, error) {
	client, err := storage.NewClient(g.ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to create gcp storage client")
	}
	defer client.Close()
	bucket := getBucket()
	attr, err := client.Bucket(bucket).Object(remotePath).Attrs(g.ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get file attributes, or it doesn't exist")
	}
	return attr.Size, nil
}

func (g *JobService) getFilebyChunk(remotePath string, offset int64, chunkSize int64) ([]byte, error) {
	client, err := storage.NewClient(g.ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gcp storage client")
	}
	defer client.Close()
	bucket := getBucket()
	objectHandle := client.Bucket(bucket).Object(remotePath)
	objectReader, err := objectHandle.NewRangeReader(g.ctx, offset, chunkSize)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to create reader on %s", remotePath))
	}
	defer objectReader.Close()
	data := make([]byte, chunkSize)
	n, err := objectReader.Read(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read cloud storage object")
	}
	data = data[:n]
	return data, nil
}

func (g *JobService) uploadFile(reader io.Reader, remotePath string, compress bool) error {
	client, err := storage.NewClient(g.ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create storage client")
	}
	defer client.Close()
	bucket := getBucket()
	writer := client.Bucket(bucket).Object(remotePath).NewWriter(g.ctx)
	defer writer.Close()
	if compress {
		gzipWriter := gzip.NewWriter(writer)
		if _, err = io.Copy(gzipWriter, reader); err != nil {
			return errors.Wrap(err, "failed to copy content to gzip writer")
		}
		defer gzipWriter.Close()
	} else {
		if _, err = io.Copy(writer, reader); err != nil {
			return errors.Wrap(err, "failed to copy content to writer")
		}
	}
	return nil
}

func getBucket() string {
	env := os.Getenv("ENV")
	return fmt.Sprintf("dcr-%s-hub", env)
}
