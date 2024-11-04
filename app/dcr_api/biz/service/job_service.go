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
	"encoding/base64"
	"fmt"
	"io"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/google/uuid"
	"github.com/manatee-project/manatee/app/dcr_api/biz/dal/db"
	"github.com/manatee-project/manatee/app/dcr_api/biz/model/job"
	"github.com/manatee-project/manatee/pkg/cloud"
	"github.com/manatee-project/manatee/pkg/config"
	"github.com/manatee-project/manatee/pkg/errno"
	"github.com/manatee-project/manatee/pkg/utils"
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

	provider := cloud.GetCloudProvider(js.ctx)
	err = provider.UploadFile(userWorkspace, config.GetUserWorkSpacePath(creator), false)
	if err != nil {
		return "", err
	}
	err = provider.PrepareResourcesForUser(creator)
	if err != nil {
		return "", err
	}

	uuidStr, err := uuid.NewUUID()
	if err != nil {
		return "", errors.Wrap(err, "failed to generate uuid")
	}
	t := db.Job{
		UUID:            uuidStr.String(),
		Creator:         req.Creator,
		JupyterFileName: req.JupyterFileName,
		JobStatus:       int(job.JobStatus_ImageBuilding),
	}
	err = BuildImage(js.ctx, t, req.AccessToken)
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

func convertEntityToModel(j *db.Job) *job.Job {
	return &job.Job{
		ID:              int64(j.ID),
		UUID:            j.UUID,
		Creator:         j.Creator,
		JobStatus:       job.JobStatus(j.JobStatus),
		JupyterFileName: j.JupyterFileName,
		CreatedAt:       j.CreatedAt.Format(utils.Layout),
		UpdatedAt:       j.UpdatedAt.Format(utils.Layout),
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
	outputPath := config.GetJobOutputPath(j.Creator, j.UUID, j.JupyterFileName)

	provider := cloud.GetCloudProvider(js.ctx)
	size, err := provider.GetFileSize(outputPath)
	if err != nil {
		return "", 0, err
	}
	return config.GetJobOutputFilename(fmt.Sprintf("%v", j.ID), j.JupyterFileName), size, nil
}

func (js *JobService) DownloadJobOutput(req *job.DownloadJobOutputRequest) (string, error) {
	j, err := db.QueryJobByIdAndCreator(req.ID, req.Creator)
	if err != nil {
		return "", err
	}
	outputPath := config.GetJobOutputPath(j.Creator, j.UUID, j.JupyterFileName)
	provider := cloud.GetCloudProvider(js.ctx)
	datg, err := provider.GetFilebyChunk(outputPath, req.Offset, req.Chunk)
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
