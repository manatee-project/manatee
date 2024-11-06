package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/manatee-project/manatee/app/dcr_api/biz/dal/db"
	"github.com/manatee-project/manatee/pkg/config"
)

func main() {

	err := config.InitConfig()
	if err != nil {
		fmt.Printf("ERROR: failed to init config %+v \n", err)
		panic(err)
	}

	ctx := context.Background()

	db.Init()

	reconciler := NewReconciler(ctx)

	for {
		hlog.Info("Reconciling...")
		reconciler.Reconcile(ctx)

		time.Sleep(10 * time.Second)
	}
}
