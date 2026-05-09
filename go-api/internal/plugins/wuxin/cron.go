package wuxin

import (
	"context"
	"time"

	obslogger "go-api/internal/observability/logger"
)

func RunCron(ctx context.Context) {
	timer := time.NewTimer(2 * time.Minute)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return
	case <-timer.C:
	}

	for {
		cfg, _ := Wuxin().loadConfig()
		interval := cfg.SyncInterval
		if interval <= 0 {
			interval = 300
		}
		if cfg.AutoSync {
			if updated, err := Wuxin().SyncOrders(ctx); err != nil {
				obslogger.L().Warn("Wuxin auto sync failed", "error", err)
			} else if updated > 0 {
				obslogger.L().Info("Wuxin auto sync completed", "updated", updated)
			}
		}
		if !sleepWuxinCron(ctx, time.Duration(interval)*time.Second) {
			return
		}
	}
}

func sleepWuxinCron(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
