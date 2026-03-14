package ws

import (
	"testing"
	"time"
)

func TestHubStopPreventsFurtherRegistration(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{UID: 1, Send: make(chan []byte, 1)}
	hub.Register(client)

	deadline := time.Now().Add(200 * time.Millisecond)
	for hub.OnlineCount() != 1 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if hub.OnlineCount() != 1 {
		t.Fatalf("expected one online client before stop, got %d", hub.OnlineCount())
	}

	hub.Stop()

	deadline = time.Now().Add(200 * time.Millisecond)
	for hub.OnlineCount() != 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if hub.OnlineCount() != 0 {
		t.Fatalf("expected no clients after stop, got %d", hub.OnlineCount())
	}

	done := make(chan struct{})
	go func() {
		hub.Register(&Client{UID: 2, Send: make(chan []byte, 1)})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("register should return immediately after hub stop")
	}
}
