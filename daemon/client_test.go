package daemon_test

import (
	"devserve/config"
	"devserve/daemon"
	"errors"
	"testing"
)

func TestSendDaemonNotRunning(t *testing.T) {
	original := config.Socket
	config.Socket = "/tmp/devserve-test-nonexistent.sock"
	t.Cleanup(func() { config.Socket = original })

	_, err := daemon.Send(&daemon.Request{Action: "ping"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, daemon.ErrDaemonNotRunning) {
		t.Errorf("expected error to wrap ErrDaemonNotRunning, got %v", err)
	}
}
