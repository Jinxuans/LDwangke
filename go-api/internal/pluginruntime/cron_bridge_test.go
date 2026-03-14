package pluginruntime

import (
	"context"
	"testing"
)

func restoreCronBridgeHooks() func() {
	oldYDSJ := runYDSJServiceCronFn
	oldW := runWServiceCronFn
	oldXM := runXMServiceCronFn
	oldYongye := runYongyeServiceCronFn
	oldSDXY := runSDXYServiceCronFn

	return func() {
		runYDSJServiceCronFn = oldYDSJ
		runWServiceCronFn = oldW
		runXMServiceCronFn = oldXM
		runYongyeServiceCronFn = oldYongye
		runSDXYServiceCronFn = oldSDXY
	}
}

func TestCronBridgeDelegatesToServiceHooks(t *testing.T) {
	defer restoreCronBridgeHooks()()

	calls := make(chan string, 5)
	runYDSJServiceCronFn = func(context.Context) { calls <- "ydsj" }
	runWServiceCronFn = func(context.Context) { calls <- "w" }
	runXMServiceCronFn = func(context.Context) { calls <- "xm" }
	runYongyeServiceCronFn = func(context.Context) { calls <- "yongye" }
	runSDXYServiceCronFn = func(context.Context) { calls <- "sdxy" }

	ctx := context.Background()
	RunYDSJCron(ctx)
	RunWCron(ctx)
	RunXMCron(ctx)
	RunYongyeCron(ctx)
	RunSDXYCron(ctx)

	expected := map[string]bool{
		"ydsj":   true,
		"w":      true,
		"xm":     true,
		"yongye": true,
		"sdxy":   true,
	}
	seen := map[string]bool{}
	for len(seen) < len(expected) {
		name := <-calls
		seen[name] = true
	}
	for name := range expected {
		if !seen[name] {
			t.Fatalf("missing cron bridge call: %s", name)
		}
	}
}
