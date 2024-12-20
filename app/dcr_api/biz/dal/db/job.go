// Copyright 2024 TikTok Pte. Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db

import (
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	ID                uint64            `gorm:"id" json:"id""`
	UUID              string            `gorm:"uuid" json:"uuid"`
	Creator           string            `gorm:"creator" json:"creator"`
	JupyterFileName   string            `gorm:"jupyter_file_name" json:"jupyter_file_name"`
	BuildContextPath  string            `gorm:"build_context_path" json:"build_context_path"`
	Dockerfile        string            `gorm:"dockerfile" json:"dockerfile"`
	DockerImage       string            `gorm:"docker_image" json:"docker_image"`
	DockerImageDigest string            `gorm:"docker_image_digest" json:"docker_image_digest"`
	AttestationReport string            `gorm:"attestation_report" json:"attestation_report"`
	JobStatus         int               `gorm:"job_status" json:"job_status"`
	InstanceName      string            `gorm:"instance_name" json:"instance_name"`
	ExtraEnvs         map[string]string `gorm:"serializer:json"`
	CPUCount          int64             `gorm:"cpu_count" json:"cpu_count"`
	DiskSize          int64             `gorm:"disk_size" json:"disk_size"`
}

func (Job) TableName() string {
	return "jobs"
}

func CreateJob(job *Job) error {
	timestamp := time.Now()
	job.UpdatedAt = timestamp
	job.CreatedAt = timestamp
	err := DB.Create(job).Error
	if err != nil {
		return errors.Wrap(err, "failed to insert job into job table ")
	}
	return nil
}

func UpdateJob(j *Job) error {
	result := DB.Model(j).Updates(
		Job{
			JobStatus:         j.JobStatus,
			BuildContextPath:  j.BuildContextPath,
			DockerImageDigest: j.DockerImageDigest,
			DockerImage:       j.DockerImage,
			AttestationReport: j.AttestationReport,
			InstanceName:      j.InstanceName,
			ExtraEnvs:         j.ExtraEnvs,
		})
	if result.Error != nil {
		return errors.Wrap(result.Error, "failed to update job %v")
	}
	return nil
}

func QueryJobsByCreator(creator string, page, pageSize int64) ([]*Job, int64, error) {
	db := DB.Model(Job{})
	if len(creator) != 0 {
		db = db.Where("creator = ?", creator)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(err, "failed to count jobs ")
	}
	var res []*Job
	if err := db.Limit(int(pageSize)).Offset(int(pageSize * (page - 1))).Find(&res).Order("id DESC").Error; err != nil {
		return nil, 0, errors.Wrap(err, "failed to query jobs ")
	}
	return res, total, nil
}

func QueryJobByIdAndCreator(jobId int64, creator string) (*Job, error) {
	db := DB.Model(Job{})
	var res Job
	db = db.Where("id = ?", jobId).Where("creator = ?", creator)
	if err := db.First(&res).Error; err != nil {
		return nil, errors.Wrap(err, "failed to query jobs or it doesn't exist")
	}
	return &res, nil
}

func DeleteJob(creator string, uuid string) {
	DB.Model(Job{}).Where("creator = ? AND uuid = ?", creator, uuid).Delete(&Job{})
}

func QueryJobByUUIDAndCreator(creator string, uuid string) (*Job, error) {
	db := DB.Model(Job{})
	var res Job
	if err := db.Where("creator = ? AND uuid = ?", creator, uuid).First(&res).Error; err != nil {
		return nil, errors.Wrap(err, "failed to query jobs ")
	}
	return &res, nil
}

func GetInProgressJobs(creator string) ([]*Job, error) {
	var res []*Job
	if err := DB.Model(Job{}).Where("creator = ? AND job_status in (0, 1, 3, 4)", creator).Find(&res).Error; err != nil {
		return nil, errors.Wrap(err, "failed to find finished job status")
	}
	return res, nil
}

func GetAllInProgressJobs() ([]*Job, error) {
	var res []*Job
	if err := DB.Model(Job{}).Where("job_status in (0, 1, 3, 4)").Find(&res).Error; err != nil {
		return nil, errors.Wrap(err, "failed to find finished job status")
	}
	return res, nil
}
