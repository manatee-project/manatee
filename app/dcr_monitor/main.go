package main

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/manatee-project/manatee/app/dcr_monitor/client"
	"github.com/manatee-project/manatee/app/dcr_monitor/monitor"
	"github.com/manatee-project/manatee/pkg/config"
)

func main() {
	err := config.InitConfig()
	if err != nil {
		fmt.Printf("ERROR: failed to init config %+v \n", err)
		panic(err)
	}
	client.InitK8sClient()
	client.InitHTTPClient()
	ctx := context.Background()
	err = monitor.CheckKanikoJobs(ctx, client.K8sClientSet)
	if err != nil {
		hlog.Errorf("[CronJob]failed to check kaniko jobs %+v", err)
	}
	err = monitor.CheckTeeInstance(ctx)
	if err != nil {
		hlog.Errorf("[CronJob]failed to check TEE instances %+v", err)
	}
}
