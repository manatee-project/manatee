package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/manatee-project/manatee/app/dcr_api/biz/dal/db"
	"github.com/manatee-project/manatee/app/dcr_api/biz/model/job"
	"gorm.io/gorm"
)

type FakeTEEProvider struct {
	instances map[string]string
}

func (f *FakeTEEProvider) GetInstanceStatus(instanceName string) (string, error) {
	status, ok := f.instances[instanceName]
	if !ok {
		return "", fmt.Errorf("instance not found")
	}
	return status, nil
}

func (f *FakeTEEProvider) CleanUpInstance(instanceName string) error {
	delete(f.instances, instanceName)
	return nil
}

type testCase struct {
	job               *db.Job
	expectedJobStatus int
	cleanUpInstance   bool
}

func TestReconcileJob(t *testing.T) {
	tee := &FakeTEEProvider{
		instances: map[string]string{
			"instance1": "RUNNING",
			"instance2": "RUNNING",
			"instance3": "TERMINATED",
			"instance4": "RUNNING",
		},
	}

	tc := []testCase{
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
			cleanUpInstance:   false,
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
			cleanUpInstance:   false,
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
			cleanUpInstance:   true,
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
			cleanUpInstance:   true,
		},
	}

	reconciler := &ReconcilerImpl{
		ctx: context.Background(),
		tee: tee,
	}

	for _, tc := range tc {
		err := reconciler.reconcileJob(tc.job)
		assert.Nil(t, err)
		assert.DeepEqual(t, tc.expectedJobStatus, tc.job.JobStatus)

		// check if instance cleaned up after reconcile
		_, ok := tee.instances[tc.job.InstanceName]
		assert.NotEqual(t, tc.cleanUpInstance, ok)
	}
}
