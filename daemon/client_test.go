package daemon_test

import (
	"devserve/daemon"
	"devserve/util"
	"errors"
	"testing"
)

// Task 7.1: Test Send returns ErrDaemonNotRunning when no socket exists
func TestSendDaemonNotRunning(t *testing.T) {
	original := util.Socket
	util.Socket = "/tmp/devserve-test-nonexistent.sock"
	t.Cleanup(func() { util.Socket = original })

	_, err := daemon.Send(&daemon.Request{Action: "ping"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, daemon.ErrDaemonNotRunning) {
		t.Errorf("expected error to wrap ErrDaemonNotRunning, got %v", err)
	}
}
