package daemon

import (
	"devserve/internal/testutil"
	"devserve/process"
	"devserve/tunnel"
	"fmt"
	"testing"
	"time"
)

func TestStopAllProcessesEmpty(t *testing.T) {
	resetState(t)

	failed := stopAllProcesses(time.Second)
	if failed != nil {
		t.Errorf("expected nil for empty map, got %v", failed)
	}
}

// Bug #16: When stopping multiple processes and tailscale stop fails on one,
// stopAllProcesses should retry and ultimately close all tailscale serves.
func TestStopAllProcessesRetriesTailscaleFailure(t *testing.T) {
	testutil.RequireNC(t)
	resetState(t)

	mockTunnel := testutil.NewFailOnceStopTunnel()
	original := tunnel.DefaultTunnel
	tunnel.SetTunnel(mockTunnel)
	t.Cleanup(func() { tunnel.SetTunnel(original) })

	// Start two real processes
	port1 := testutil.FreePort(t)
	port2 := testutil.FreePort(t)

	p1, err := process.CreateProcess("core", port1, t.TempDir(), "echo test")
	if err != nil {
		t.Fatalf("CreateProcess core failed: %v", err)
	}
	if err := p1.Start(fmt.Sprintf("nc -l %d", port1)); err != nil {
		t.Fatalf("Start core failed: %v", err)
	}

	p2, err := process.CreateProcess("ui", port2, t.TempDir(), "echo test")
	if err != nil {
		t.Fatalf("CreateProcess ui failed: %v", err)
	}
	if err := p2.Start(fmt.Sprintf("nc -l %d", port2)); err != nil {
		t.Fatalf("Start ui failed: %v", err)
	}

	mu.Lock()
	processes["core"] = p1
	processes["ui"] = p2
	mu.Unlock()

	failed := stopAllProcesses(10 * time.Second)

	if len(failed) > 0 {
		t.Errorf("expected all processes to stop successfully, but ports failed: %v", failed)
	}

	mu.RLock()
	remaining := len(processes)
	mu.RUnlock()
	if remaining != 0 {
		t.Errorf("expected process map to be empty, got %d remaining", remaining)
	}
}
