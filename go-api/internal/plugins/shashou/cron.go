package shashou

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
		projects, _ := ShaShou().ListProjects(true)
		enabled := false
		interval := 300
		for _, p := range projects {
			if p.AutoSync == 1 && p.Status == 1 {
				enabled = true
				if p.SyncInterval > 0 && p.SyncInterval < interval {
					interval = p.SyncInterval
				}
			}
		}
		if enabled {
			if updated, err := ShaShou().SyncPending(ctx, 100); err != nil {
				obslogger.L().Warn("Shashou auto sync failed", "error", err)
			} else if updated > 0 {
				obslogger.L().Info("Shashou auto sync completed", "updated", updated)
			}
		}
		if interval <= 0 {
			interval = 300
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
