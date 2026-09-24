package services

import (
	"encoding/json"
	"testing"
	"time"

	"indexarr/internal/config"

	"github.com/rs/zerolog"
)

func waitForCondition(t *testing.T, timeout time.Duration, fn func() bool, msg string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for condition: %s", msg)
}

func TestBroadcasterRegisterAndUnregister(t *testing.T) {
	logger := zerolog.Nop()
	config.GlobalLogger = &logger

	b := NewBroadcaster()
	go b.Run()

	client := &Client{send: make(chan []byte, 1)}
	b.Register(client)

	waitForCondition(t, time.Second, func() bool {
		return b.ClientCount() == 1
	}, "client to be registered")

	b.Unregister(client)

	waitForCondition(t, time.Second, func() bool {
		return b.ClientCount() == 0
	}, "client to be unregistered")

	_, ok := <-client.send
	if ok {
		t.Fatalf("expected client send channel to be closed on unregister")
	}
}

func TestBroadcasterBroadcastScanProgress(t *testing.T) {
	logger := zerolog.Nop()
	config.GlobalLogger = &logger

	b := NewBroadcaster()
	go b.Run()

	client := &Client{send: make(chan []byte, 2)}
	b.Register(client)

	waitForCondition(t, time.Second, func() bool {
		return b.ClientCount() == 1
	}, "client to be registered before broadcast")

	b.BroadcastScanProgress(3, 10)

	select {
	case raw := <-client.send:
		var msg WSMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			t.Fatalf("failed to decode broadcast message: %v", err)
		}
		if msg.Type != "scan_progress" {
			t.Fatalf("unexpected message type: got %q want %q", msg.Type, "scan_progress")
		}
		if msg.FilesProcessed != 3 || msg.FilesFound != 10 {
			t.Fatalf("unexpected message payload: %+v", msg)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for broadcast message")
	}
}
