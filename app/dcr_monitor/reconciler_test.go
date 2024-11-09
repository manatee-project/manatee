package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/manatee-project/manatee/app/dcr_api/biz/dal/db"
	"github.com/manatee-project/manatee/app/dcr_api/biz/model/job"
	"github.com/manatee-project/manatee/app/dcr_monitor/imagebuilder"
	"gorm.io/gorm"
)

type FakeTEEProvider struct {
	instances map[string]string
}

type ImageBuildStatus struct {
	done bool
	info *imagebuilder.ImageInfo
}
type FakeImageBuilder struct {
	buildjobs map[string]ImageBuildStatus
}

func (f *FakeTEEProvider) GetInstanceStatus(instanceName string) (string, error) {
	status, ok := f.instances[instanceName]
	if !ok {
		return "", fmt.Errorf("instance not found")
	}
	return status, nil
}

func (f *FakeTEEProvider) LaunchInstance(instanceName string, image string, digest string) error {
	f.instances[instanceName] = "RUNNING"
	return nil
}

func (f *FakeTEEProvider) CleanUpInstance(instanceName string) error {
	delete(f.instances, instanceName)
	return nil
}

func (f *FakeImageBuilder) BuildImage(j *db.Job, bucket string, base string, image string) error {
	f.buildjobs[j.UUID] = ImageBuildStatus{
		done: false,
		info: nil,
	}
	return nil
}

func (f *FakeImageBuilder) CheckImageBuilderStatusAndGetInfo(uuid string) (bool, *imagebuilder.ImageInfo, error) {
	status, ok := f.buildjobs[uuid]
	if !ok {
		return false, nil, fmt.Errorf("instance not found")
	}
	return status.done, status.info, nil
}

type testCase struct {
	job               *db.Job
	expectedJobStatus int
}

func TestUpdateJobStatusInstances(t *testing.T) {
	tee := &FakeTEEProvider{
		instances: map[string]string{
			"instance1": "RUNNING",
			"instance2": "RUNNING",
			"instance3": "TERMINATED",
			"instance4": "RUNNING",
		},
	}

	tcs := []testCase{
		{
			job: &db.Job{
				UUID:         "job1",
				JobStatus:    int(job.JobStatus_VMWaiting),
				InstanceName: "instance1",
				Model: gorm.Model{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			expectedJobStatus: int(job.JobStatus_VMRunning),
		},
		{
			job: &db.Job{
				UUID:         "job2",
				JobStatus:    int(job.JobStatus_VMRunning),
				InstanceName: "instance2",
				Model: gorm.Model{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			expectedJobStatus: int(job.JobStatus_VMRunning),
		},
		{
			job: &db.Job{
				UUID:         "job3",
				JobStatus:    int(job.JobStatus_VMRunning),
				InstanceName: "instance3",
				Model: gorm.Model{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			expectedJobStatus: int(job.JobStatus_VMFinished),
		},
		{
			job: &db.Job{
				UUID:         "job4",
				JobStatus:    int(job.JobStatus_VMRunning),
				InstanceName: "instance4",
				Model: gorm.Model{
					CreatedAt: time.Now().Add(-7 * time.Hour),
					UpdatedAt: time.Now().Add(-7 * time.Hour),
				},
			},
			expectedJobStatus: int(job.JobStatus_VMFailed),
		},
	}

	reconciler := &ReconcilerImpl{
		ctx: context.Background(),
		tee: tee,
	}

	for _, tc := range tcs {
		err := reconciler.updateJobStatus(tc.job)
		assert.Nil(t, err)
		assert.DeepEqual(t, tc.expectedJobStatus, tc.job.JobStatus)
	}
}

func TestUpdateJobStatusImageBuilder(t *testing.T) {
	builder := &FakeImageBuilder{
		buildjobs: map[string]ImageBuildStatus{
			"job1": ImageBuildStatus{false, nil},
			"job2": ImageBuildStatus{true, nil},
			"job3": ImageBuildStatus{true, &imagebuilder.ImageInfo{Image: "my.image.registry/image", Digest: "deadbeef"}},
		},
	}
	tee := &FakeTEEProvider{
		instances: map[string]string{},
	}
	reconciler := &ReconcilerImpl{
		ctx:     context.Background(),
		builder: builder,
		tee:     tee,
	}

	tcs := []testCase{
		{
			job: &db.Job{
				UUID:      "job1",
				JobStatus: int(job.JobStatus_ImageBuilding),
				Creator:   "user1",
				Model: gorm.Model{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			expectedJobStatus: int(job.JobStatus_ImageBuilding),
		},
		{
			job: &db.Job{
				UUID:      "job2",
				JobStatus: int(job.JobStatus_ImageBuilding),
				Creator:   "user1",
				Model: gorm.Model{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			expectedJobStatus: int(job.JobStatus_ImageBuildingFailed),
		},
		{
			job: &db.Job{
				UUID:      "job3",
				JobStatus: int(job.JobStatus_ImageBuilding),
				Creator:   "user1",
				Model: gorm.Model{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			expectedJobStatus: int(job.JobStatus_VMWaiting),
		},
	}

	for _, tc := range tcs {
		err := reconciler.updateJobStatus(tc.job)
		if err != nil {
			t.Errorf("updating %s returns error: %v", tc.job.UUID, err)
		}
		if tc.expectedJobStatus != tc.job.JobStatus {
			t.Errorf("%s status does not match: expected %v, got %v", tc.job.UUID, tc.expectedJobStatus, tc.job.JobStatus)
		}
		if tc.expectedJobStatus == int(job.JobStatus_VMWaiting) {
			_, ok := tee.instances[fmt.Sprintf("%s-%s", tc.job.Creator, tc.job.UUID)]
			if !ok {
				t.Errorf("instance not found for %s", tc.job.UUID)
			}
		}
	}
}
