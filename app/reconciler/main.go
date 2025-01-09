package main

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/manatee-project/manatee/app/dcr_api/biz/dal/db"
)

func main() {

	ctx := context.Background()

	db.Init()

	reconciler := NewReconciler(ctx)

	for {
		hlog.Info("Reconciling...")
		reconciler.Reconcile(ctx)

		time.Sleep(10 * time.Second)
	}
}
