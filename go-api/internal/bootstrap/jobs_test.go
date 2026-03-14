package bootstrap

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-api/internal/app"
	"go-api/internal/autosync"
	"go-api/internal/license"
	"go-api/internal/ws"
)

func restoreBootstrapHooks() func() {
	oldInitSyncTicker := initSyncTickerFn
	oldStartLicense := startLicenseFn
	oldRunHub := runHubFn
	oldStartHZWSocket := startHZWSocketFn
	oldRunYDSJ := runYDSJCronFn
	oldRunW := runWCronFn
	oldRunXM := runXMCronFn
	oldRunYongye := runYongyeCronFn
	oldRunSDXY := runSDXYCronFn
	oldRunLonglong := runLonglongDaemonFn
	oldRunSimpleThread := runSimpleThreadSyncFn
	oldAutoShelfCron := autoShelfCronFn
	oldGetSyncConfig := getSyncConfigFn
	oldSetAutoSyncNextRun := setAutoSyncNextRunFn
	oldSleepContext := sleepContextFn
	oldArchiveOldMessages := archiveOldMessagesFn
	oldTrimSessionMessages := trimSessionMessagesFn

	return func() {
		initSyncTickerFn = oldInitSyncTicker
		startLicenseFn = oldStartLicense
		runHubFn = oldRunHub
		startHZWSocketFn = oldStartHZWSocket
		runYDSJCronFn = oldRunYDSJ
		runWCronFn = oldRunW
		runXMCronFn = oldRunXM
		runYongyeCronFn = oldRunYongye
		runSDXYCronFn = oldRunSDXY
		runLonglongDaemonFn = oldRunLonglong
		runSimpleThreadSyncFn = oldRunSimpleThread
		autoShelfCronFn = oldAutoShelfCron
		getSyncConfigFn = oldGetSyncConfig
		setAutoSyncNextRunFn = oldSetAutoSyncNextRun
		sleepContextFn = oldSleepContext
		archiveOldMessagesFn = oldArchiveOldMessages
		trimSessionMessagesFn = oldTrimSessionMessages
	}
}

func waitForCall(t *testing.T, ch <-chan string) string {
	t.Helper()

	select {
	case name := <-ch:
		return name
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for background call")
		return ""
	}
}

func TestStartCoreJobsStartsBackgroundWorkers(t *testing.T) {
	defer restoreBootstrapHooks()()

	calls := make(chan string, 8)
	initCalled := false

	initSyncTickerFn = func(d time.Duration) {
		initCalled = true
		if d != 2*time.Minute {
			t.Fatalf("unexpected sync ticker duration: %v", d)
		}
	}
	startHZWSocketFn = func() { calls <- "hzw" }
	runYDSJCronFn = func(context.Context) { calls <- "ydsj" }
	runWCronFn = func(context.Context) { calls <- "w" }
	runXMCronFn = func(context.Context) { calls <- "xm" }
	runYongyeCronFn = func(context.Context) { calls <- "yongye" }
	runSDXYCronFn = func(context.Context) { calls <- "sdxy" }
	runLonglongDaemonFn = func(context.Context) { calls <- "longlong" }
	runSimpleThreadSyncFn = func(context.Context) { calls <- "simple" }
	sleepContextFn = func(context.Context, time.Duration) bool { return false }
	autoShelfCronFn = func() { t.Fatal("autoShelfCron should not run when initial sleep is cancelled") }

	StartCoreJobs(context.Background(), nil)

	if !initCalled {
		t.Fatal("expected sync ticker to be initialized")
	}

	expected := map[string]bool{
		"hzw": true, "ydsj": true, "w": true, "xm": true,
		"yongye": true, "sdxy": true, "longlong": true, "simple": true,
	}
	seen := map[string]bool{}
	for len(seen) < len(expected) {
		seen[waitForCall(t, calls)] = true
	}
	for name := range expected {
		if !seen[name] {
			t.Fatalf("missing background worker call: %s", name)
		}
	}
}

func TestStartCoreJobsStartsLicenseAndHubWhenPresent(t *testing.T) {
	defer restoreBootstrapHooks()()

	calls := make(chan string, 2)
	sleepContextFn = func(context.Context, time.Duration) bool { return false }
	initSyncTickerFn = func(time.Duration) {}
	startHZWSocketFn = func() {}
	runYDSJCronFn = func(context.Context) {}
	runWCronFn = func(context.Context) {}
	runXMCronFn = func(context.Context) {}
	runYongyeCronFn = func(context.Context) {}
	runSDXYCronFn = func(context.Context) {}
	runLonglongDaemonFn = func(context.Context) {}
	runSimpleThreadSyncFn = func(context.Context) {}
	autoShelfCronFn = func() {}

	startLicenseFn = func(m *license.Manager) {
		if m == nil {
			t.Fatal("expected license manager")
		}
		calls <- "license"
	}
	runHubFn = func(h *ws.Hub) {
		if h == nil {
			t.Fatal("expected hub")
		}
		calls <- "hub"
	}

	hub := ws.NewHub()
	defer hub.Stop()

	StartCoreJobs(context.Background(), &app.App{
		License: &license.Manager{},
		Hub:     hub,
	})

	seen := map[string]bool{
		waitForCall(t, calls): true,
		waitForCall(t, calls): true,
	}
	if !seen["license"] || !seen["hub"] {
		t.Fatalf("expected license and hub startup, got %v", seen)
	}
}

func TestStartAutoShelfCronRunsOnceAndSchedulesNextRun(t *testing.T) {
	defer restoreBootstrapHooks()()

	var sleeps []time.Duration
	sleepContextFn = func(ctx context.Context, d time.Duration) bool {
		sleeps = append(sleeps, d)
		return len(sleeps) == 1
	}

	autoShelfRuns := 0
	autoShelfCronFn = func() { autoShelfRuns++ }
	getSyncConfigFn = func() (*autosync.SyncConfig, error) {
		return &autosync.SyncConfig{AutoSyncInterval: 45}, nil
	}

	start := time.Now()
	var nextRun time.Time
	setAutoSyncNextRunFn = func(t time.Time) { nextRun = t }

	startAutoShelfCron(context.Background())

	if autoShelfRuns != 1 {
		t.Fatalf("expected one auto shelf run, got %d", autoShelfRuns)
	}
	if len(sleeps) != 2 {
		t.Fatalf("expected two sleep calls, got %d", len(sleeps))
	}
	if sleeps[0] != 5*time.Minute {
		t.Fatalf("expected initial sleep of 5 minutes, got %v", sleeps[0])
	}
	if sleeps[1] != 45*time.Minute {
		t.Fatalf("expected interval sleep of 45 minutes, got %v", sleeps[1])
	}
	if nextRun.IsZero() {
		t.Fatal("expected next run time to be scheduled")
	}
	if nextRun.Before(start.Add(44*time.Minute)) || nextRun.After(start.Add(46*time.Minute)) {
		t.Fatalf("unexpected next run time: %v", nextRun)
	}
}

func TestStartAutoShelfCronFallsBackToDefaultIntervalWhenConfigUnavailable(t *testing.T) {
	defer restoreBootstrapHooks()()

	var sleeps []time.Duration
	sleepContextFn = func(ctx context.Context, d time.Duration) bool {
		sleeps = append(sleeps, d)
		return len(sleeps) == 1
	}

	autoShelfRuns := 0
	autoShelfCronFn = func() { autoShelfRuns++ }
	getSyncConfigFn = func() (*autosync.SyncConfig, error) {
		return nil, errors.New("config unavailable")
	}

	start := time.Now()
	var nextRun time.Time
	setAutoSyncNextRunFn = func(t time.Time) { nextRun = t }

	startAutoShelfCron(context.Background())

	if autoShelfRuns != 1 {
		t.Fatalf("expected one auto shelf run, got %d", autoShelfRuns)
	}
	if len(sleeps) != 2 {
		t.Fatalf("expected two sleep calls, got %d", len(sleeps))
	}
	if sleeps[0] != 5*time.Minute {
		t.Fatalf("expected initial sleep of 5 minutes, got %v", sleeps[0])
	}
	if sleeps[1] != 30*time.Minute {
		t.Fatalf("expected fallback interval sleep of 30 minutes, got %v", sleeps[1])
	}
	if nextRun.IsZero() {
		t.Fatal("expected next run time to be scheduled")
	}
	if nextRun.Before(start.Add(29*time.Minute)) || nextRun.After(start.Add(31*time.Minute)) {
		t.Fatalf("unexpected next run time: %v", nextRun)
	}
}

func TestStartChatCleanupRunsArchiveAndTrim(t *testing.T) {
	defer restoreBootstrapHooks()()

	sleepCalls := 0
	sleepContextFn = func(context.Context, time.Duration) bool {
		sleepCalls++
		return sleepCalls == 1
	}

	calls := make(chan string, 2)
	archiveOldMessagesFn = func() (int64, error) {
		calls <- "archive"
		return 2, nil
	}
	trimSessionMessagesFn = func() (int64, error) {
		calls <- "trim"
		return 3, nil
	}

	StartChatCleanup(context.Background())

	seen := map[string]bool{
		waitForCall(t, calls): true,
		waitForCall(t, calls): true,
	}
	if !seen["archive"] || !seen["trim"] {
		t.Fatalf("expected archive and trim to run, got %v", seen)
	}
}

func TestStartChatCleanupStillTrimsWhenArchiveFails(t *testing.T) {
	defer restoreBootstrapHooks()()

	sleepCalls := 0
	sleepContextFn = func(context.Context, time.Duration) bool {
		sleepCalls++
		return sleepCalls == 1
	}

	archiveOldMessagesFn = func() (int64, error) {
		return 0, errors.New("archive failed")
	}

	trimmed := make(chan struct{}, 1)
	trimSessionMessagesFn = func() (int64, error) {
		trimmed <- struct{}{}
		return 0, nil
	}

	StartChatCleanup(context.Background())

	select {
	case <-trimmed:
	case <-time.After(2 * time.Second):
		t.Fatal("expected trim to run even when archive fails")
	}
}
