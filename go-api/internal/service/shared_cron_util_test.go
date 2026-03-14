package service

import (
	"context"
	"testing"
	"time"
)

func TestSleepWithContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if sleepWithContext(ctx, 100*time.Millisecond) {
		t.Fatal("expected sleepWithContext to stop immediately when context is canceled")
	}
}

func TestSleepWithContextWaitsForTimer(t *testing.T) {
	start := time.Now()
	if !sleepWithContext(context.Background(), 20*time.Millisecond) {
		t.Fatal("expected sleepWithContext to wait for timer")
	}
	if time.Since(start) < 15*time.Millisecond {
		t.Fatal("sleepWithContext returned too early")
	}
}
