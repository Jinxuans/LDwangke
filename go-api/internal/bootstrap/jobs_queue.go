package bootstrap

import (
	"context"
	"time"

	"go-api/internal/dockscheduler"
)

// StartPendingDockScheduler 启动待对接订单的定时调度器。
func StartPendingDockScheduler(ctx context.Context) {
	dockscheduler.Start(ctx, 30*time.Second, 100)
}
