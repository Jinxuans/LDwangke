package baitan

import (
	"context"
	"time"

	obslogger "go-api/internal/observability/logger"
)

func RunCron(ctx context.Context) {
	if !sleepBaitanCron(ctx, 2*time.Minute) {
		return
	}
	for {
		cfg, _ := Baitan().loadConfig()
		interval := cfg.SyncInterval
		if interval <= 0 {
			interval = 300
		}
		if cfg.AutoSync {
			if updated, err := Baitan().SyncOrders(ctx, 100); err != nil {
				obslogger.L().Warn("Baitan auto sync failed", "error", err)
			} else if updated > 0 {
				obslogger.L().Info("Baitan auto sync completed", "updated", updated)
			}
		}
		if !sleepBaitanCron(ctx, time.Duration(interval)*time.Second) {
			return
		}
	}
}

func sleepBaitanCron(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
