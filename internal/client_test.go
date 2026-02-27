package internal

import (
	"devserve/util"
	"errors"
	"testing"
)

// Task 7.1: Test Send returns ErrDaemonNotRunning when no socket exists
func TestSendDaemonNotRunning(t *testing.T) {
	original := util.Socket
	util.Socket = "/tmp/devserve-test-nonexistent.sock"
	t.Cleanup(func() { util.Socket = original })

	_, err := Send(&Request{Action: "ping"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrDaemonNotRunning) {
		t.Errorf("expected error to wrap ErrDaemonNotRunning, got %v", err)
	}
}
