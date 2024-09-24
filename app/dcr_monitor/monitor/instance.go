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

package monitor

import (
	"context"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"

	"github.com/manatee-project/manatee/app/dcr_api/biz/model/job"
	"github.com/manatee-project/manatee/pkg/cloud"
)

func getCreator(token string) (string, error) {
	parts := strings.Split(token, "-")
	return parts[0], nil
}

func CheckTeeInstance(ctx context.Context) error {
	hlog.Info("[InstancesMonitor] start to monitor instances' status")
	provider := cloud.GetCloudProvider(ctx)
	list, err := provider.ListAllInstances()
	if err != nil {
		hlog.Errorf("[InstancesMonitor] failed to query instances' status %+v", err)
		return err
	}
	for _, instance := range list {
		hlog.Debugf("[InstancesMonitor] instance status: %v", instance)
		token := instance.Token
		status := instance.Status
		UUID := instance.UUID
		if UUID == "" {
			continue
		}
		creator, err := getCreator(instance.Name)
		if err != nil {
			hlog.Errorf("[InstancesMonitor] failed to get creator info %+v", err)
			continue
		}
		timeCreation := instance.CreationTime
		layout := "2006-01-02T15:04:05.999999999Z07:00"
		formattedTimeCreation, err := time.Parse(layout, timeCreation)
		jobStuck := time.Since(formattedTimeCreation) > 6*time.Hour

		if status == cloud.INSTANCE_TERMINATED {
			err = updateTeeInstanceStatus(creator, UUID, token, int64(job.JobStatus_VMFinished))
			if err != nil {
				return err
			}
			hlog.Info("[InstancesMonitor]Successfully updated the instance status to finished")
			err := provider.DeleteInstance(instance.Name)
			if err != nil {
				return err
			}

		}
		if jobStuck {
			hlog.Infof("[InstancesMonitor] job %v has benn stucked more than 6 hours", UUID)
			err = updateTeeInstanceStatus(creator, UUID, token, int64(job.JobStatus_VMFinished))
			if err != nil {
				return err
			}
			return provider.DeleteInstance(instance.Name)
		}
	}
	return nil
}

func convertTokenToStage1(stage2Token string) (string, error) {
	// TODO add authentication
	return stage2Token, nil
}

func updateTeeInstanceStatus(creator, UUID, token string, status int64) error {
	stage1Token, err := convertTokenToStage1(token)
	if err != nil {
		return err
	}
	err = updateJobStatus(creator, UUID, stage1Token, "", status)
	return err
}
