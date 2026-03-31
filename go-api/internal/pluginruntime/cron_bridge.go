package pluginruntime

import (
	"context"

	xmmodule "go-api/internal/plugins/xm"
	"go-api/internal/service"
)

var (
	runYDSJServiceCronFn   = service.RunYDSJCron
	runWServiceCronFn      = service.RunWCron
	runYongyeServiceCronFn = service.RunYongyeCron
	runSDXYServiceCronFn   = service.RunSDXYCron
)

// These plugin cron loops still live in service because they remain tightly
// coupled to legacy plugin implementations. Keep this bridge narrow.
func RunYDSJCron(ctx context.Context) {
	runYDSJServiceCronFn(ctx)
}

func RunWCron(ctx context.Context) {
	runWServiceCronFn(ctx)
}

func RunXMCron(ctx context.Context) {
	xmmodule.RunXMCron(ctx)
}

func RunYongyeCron(ctx context.Context) {
	runYongyeServiceCronFn(ctx)
}

func RunSDXYCron(ctx context.Context) {
	runSDXYServiceCronFn(ctx)
}
