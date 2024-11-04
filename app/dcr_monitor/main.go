package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/manatee-project/manatee/app/dcr_api/biz/dal/db"
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

	db.Init()

	reconciler := NewReconciler(ctx)

	for {
		hlog.Info("Reconciling...")
		reconciler.Reconcile(ctx)

		// TODO: remove these and replace with logic in reconciler
		err = monitor.CheckKanikoJobs(ctx, client.K8sClientSet)
		if err != nil {
			hlog.Errorf("[CronJob]failed to check kaniko jobs %+v", err)
		}

		time.Sleep(10 * time.Second)
	}
}
