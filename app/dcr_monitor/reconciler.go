package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/manatee-project/manatee/app/dcr_api/biz/dal/db"
	"github.com/manatee-project/manatee/app/dcr_api/biz/model/job"
	"github.com/manatee-project/manatee/app/dcr_monitor/tee_backend"
	"github.com/manatee-project/manatee/pkg/config"
)

type Reconciler interface {
	Reconcile(ctx context.Context)
}

type ReconcilerImpl struct {
	tee tee_backend.TEEProvider
	ctx context.Context
}

func NewReconciler(ctx context.Context) *ReconcilerImpl {
	// FIXME: get config to determine which TEE provider to use.
	// for now, we only support GCP confidential space.
	tee, err := tee_backend.NewTEEProviderGCPConfidentialSpace(ctx, config.GetProject(), config.GetZone())
	if err != nil {
		hlog.Errorf("failed to init TEE provider %+v", err)
	}

	return &ReconcilerImpl{
		tee: tee,
		ctx: ctx,
	}
}

func (r *ReconcilerImpl) Reconcile(ctx context.Context) {
	jobs, err := db.GetAllInProgressJobs()
	if err != nil {
		hlog.Errorf("[Reconciler] failed to get all in progress jobs: %v", err)
		return
	}

	// debug log
	hlog.Infof("[Reconciler] found %d jobs in progress", len(jobs))
	for _, j := range jobs {
		// debug log
		hlog.Infof("[Reconciler] job %s in status %d", j.UUID, j.JobStatus)
		err := r.reconcileJob(j)
		if err != nil {
			hlog.Errorf("[Reconciler] failed to reconcile job %s: %v", j.UUID, err)
			continue
		}
		db.UpdateJob(j)
	}
}

func (r *ReconcilerImpl) reconcileJob(j *db.Job) error {

	// if job was not finished for more than the timeout, mark it as error
	if time.Since(j.CreatedAt) > 6*time.Hour {
		hlog.Infof("[Reconciler] job %s is not finished for more than 6 hours. cleaning up...", j.UUID)
		r.tee.CleanUpInstance(j.InstanceName)
		j.JobStatus = int(job.JobStatus_VMFailed)
	} else {
		switch j.JobStatus {
		case int(job.JobStatus_VMWaiting):
			instanceStatus, err := r.tee.GetInstanceStatus(j.InstanceName)
			if err != nil {
				return fmt.Errorf("failed to get instance status: %w", err)
			}
			if instanceStatus == "RUNNING" {
				j.JobStatus = int(job.JobStatus_VMRunning)
			} else if instanceStatus == "TERMINATED" {
				j.JobStatus = int(job.JobStatus_VMFinished)
			}
		case int(job.JobStatus_VMRunning):
			instanceStatus, err := r.tee.GetInstanceStatus(j.InstanceName)
			if err != nil {
				return fmt.Errorf("failed to get instance status: %w", err)
			}
			if instanceStatus == "TERMINATED" {
				j.JobStatus = int(job.JobStatus_VMFinished)
			}
		}
	}

	// clean up instance if necessary
	if j.JobStatus == int(job.JobStatus_VMFinished) || j.JobStatus == int(job.JobStatus_VMFailed) {
		r.tee.CleanUpInstance(j.InstanceName)
	}
	return nil
}