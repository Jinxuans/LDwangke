package bootstrap

import (
	"context"
	"time"

	"go-api/internal/app"
	"go-api/internal/autosync"
	"go-api/internal/pluginruntime"
	shashouplugin "go-api/internal/plugins/shashou"
	wuxinplugin "go-api/internal/plugins/wuxin"
	"go-api/internal/runtimeops"
	"go-api/internal/ws"
)

var (
	initSyncTickerFn      = runtimeops.InitSyncTicker
	runHubFn              = func(h *ws.Hub) { h.Run() }
	startHZWSocketFn      = pluginruntime.StartHZWSocket
	runYDSJCronFn         = pluginruntime.RunYDSJCron
	runWCronFn            = pluginruntime.RunWCron
	runXMCronFn           = pluginruntime.RunXMCron
	runYongyeCronFn       = pluginruntime.RunYongyeCron
	runSDXYCronFn         = pluginruntime.RunSDXYCron
	runShashouCronFn      = shashouplugin.RunCron
	runWuxinCronFn        = wuxinplugin.RunCron
	runLonglongDaemonFn   = pluginruntime.RunLonglongDaemon
	runSimpleThreadSyncFn = pluginruntime.RunSimpleThreadSync
	autoShelfCronFn       = autosync.AutoShelfCron
	getSyncConfigFn       = autosync.GetSyncConfig
	setAutoSyncNextRunFn  = autosync.SetAutoSyncNextRun
	sleepContextFn        = sleepContext
)

// StartCoreJobs 启动 WebSocket 和各类历史后台任务。
func StartCoreJobs(ctx context.Context, a *app.App) {
	if a != nil && a.Hub != nil {
		go runHubFn(a.Hub)
	}

	initSyncTickerFn(2 * time.Minute)
	go startHZWSocketFn()
	go runYDSJCronFn(ctx)
	go runWCronFn(ctx)
	go runXMCronFn(ctx)
	go runYongyeCronFn(ctx)
	go runSDXYCronFn(ctx)
	go runShashouCronFn(ctx)
	go runWuxinCronFn(ctx)
	go runLonglongDaemonFn(ctx)
	go runSimpleThreadSyncFn(ctx)
	go startAutoShelfCron(ctx)
}

func startAutoShelfCron(ctx context.Context) {
	if !sleepContextFn(ctx, 5*time.Minute) {
		return
	}
	autoShelfCronFn()

	for {
		cfg, _ := getSyncConfigFn()
		interval := 30
		if cfg != nil && cfg.AutoSyncInterval > 0 {
			interval = cfg.AutoSyncInterval
		}
		dur := time.Duration(interval) * time.Minute
		setAutoSyncNextRunFn(time.Now().Add(dur))
		if !sleepContextFn(ctx, dur) {
			return
		}
		autoShelfCronFn()
	}
}

func sleepContext(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
