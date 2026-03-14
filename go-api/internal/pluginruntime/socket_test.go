package pluginruntime

import (
	"testing"
	"time"
)

func TestHZWSocketClientSleepStopsImmediately(t *testing.T) {
	client := newHZWSocketClient("http://example.test")
	client.Stop()

	start := time.Now()
	if client.sleep(200 * time.Millisecond) {
		t.Fatal("expected sleep to stop early after client stop")
	}
	if time.Since(start) > 50*time.Millisecond {
		t.Fatal("sleep did not stop promptly after client stop")
	}
}
