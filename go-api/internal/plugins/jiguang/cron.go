package jiguang

import (
	"context"
	"time"

	obslogger "go-api/internal/observability/logger"
)

func RunCron(ctx context.Context) {
	if !sleepCron(ctx, 2*time.Minute) {
		return
	}
	for {
		cfg, _ := Jiguang().loadConfig()
		interval := cfg.SyncInterval
		if interval <= 0 {
			interval = 300
		}
		if cfg.AutoSync {
			if updated, err := Jiguang().SyncOrders(ctx); err != nil {
				obslogger.L().Warn("Jiguang auto sync failed", "error", err)
			} else if updated > 0 {
				obslogger.L().Info("Jiguang auto sync completed", "updated", updated)
			}
		}
		if !sleepCron(ctx, time.Duration(interval)*time.Second) {
			return
		}
	}
}

func sleepCron(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
