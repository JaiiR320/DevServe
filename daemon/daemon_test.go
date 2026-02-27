package daemon_test

import (
	"devserve/daemon"
	"devserve/internal"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// Task 7.2: Test handleConn dispatches ping action
func TestHandleConnPing(t *testing.T) {
	daemon.ResetProcesses()
	t.Cleanup(func() { daemon.ResetProcesses() })

	client, server := net.Pipe()
	defer client.Close()

	stop := make(chan struct{}, 1)

	go daemon.HandleConn(server, stop)

	if err := daemon.SendRequest(client, &daemon.Request{Action: "ping"}); err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	resp, err := daemon.ReadResponse(client)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected OK response, got error: %s", resp.Error)
	}
	if resp.Data != "pong" {
		t.Errorf("expected data %q, got %q", "pong", resp.Data)
	}
}

// Task 7.3: Test handleConn with unknown action
func TestHandleConnUnknownAction(t *testing.T) {
	daemon.ResetProcesses()
	t.Cleanup(func() { daemon.ResetProcesses() })

	client, server := net.Pipe()
	defer client.Close()

	stop := make(chan struct{}, 1)

	go daemon.HandleConn(server, stop)

	if err := daemon.SendRequest(client, &daemon.Request{Action: "bogus"}); err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	resp, err := daemon.ReadResponse(client)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	if resp.OK {
		t.Fatal("expected error response, got OK")
	}
	if !strings.Contains(resp.Error, "unknown action") {
		t.Errorf("expected error to contain %q, got %q", "unknown action", resp.Error)
	}
}

// Task 7.4: Test handleConn with malformed request
func TestHandleConnMalformed(t *testing.T) {
	daemon.ResetProcesses()
	t.Cleanup(func() { daemon.ResetProcesses() })

	client, server := net.Pipe()
	defer client.Close()

	stop := make(chan struct{}, 1)

	go daemon.HandleConn(server, stop)

	// Write garbage bytes followed by a newline (JSON decoder reads line-delimited)
	if _, err := client.Write([]byte("not valid json\n")); err != nil {
		t.Fatalf("failed to write garbage: %v", err)
	}

	resp, err := daemon.ReadResponse(client)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	if resp.OK {
		t.Fatal("expected error response, got OK")
	}
}

// Task 7.5: Test handleConn dispatches shutdown — stop channel is signaled
func TestHandleConnShutdown(t *testing.T) {
	daemon.ResetProcesses()
	t.Cleanup(func() { daemon.ResetProcesses() })

	client, server := net.Pipe()
	defer client.Close()

	stop := make(chan struct{}, 1)

	go daemon.HandleConn(server, stop)

	if err := daemon.SendRequest(client, &daemon.Request{Action: "shutdown"}); err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	resp, err := daemon.ReadResponse(client)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected OK response, got error: %s", resp.Error)
	}
	if !strings.Contains(resp.Data, "daemon stopped") {
		t.Errorf("expected data to contain %q, got %q", "daemon stopped", resp.Data)
	}

	// Verify stop channel was signaled
	select {
	case <-stop:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("expected stop channel to be signaled, timed out")
	}
}

// Task 7.6: Test StopAllProcesses with empty map — returns nil
func TestStopAllProcessesEmpty(t *testing.T) {
	daemon.ResetProcesses()
	t.Cleanup(func() { daemon.ResetProcesses() })

	failed := daemon.StopAllProcesses(time.Second)
	if failed != nil {
		t.Errorf("expected nil for empty map, got %v", failed)
	}
}

func requireNC(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("nc"); err != nil {
		t.Skip("nc not available")
	}
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

// failOnceStopTunnel fails the first Stop() call per port, succeeds on retry.
type failOnceStopTunnel struct {
	failed map[int]bool
}

func (f *failOnceStopTunnel) Serve(port int) error { return nil }
func (f *failOnceStopTunnel) Stop(port int) error {
	if !f.failed[port] {
		f.failed[port] = true
		return fmt.Errorf("tailscale serve stop failed for port %d", port)
	}
	return nil
}

// Bug #16: When stopping multiple processes and tailscale stop fails on one,
// stopAllProcesses should retry and ultimately close all tailscale serves.
func TestStopAllProcessesRetriesTailscaleFailure(t *testing.T) {
	requireNC(t)
	daemon.ResetProcesses()
	t.Cleanup(func() { daemon.ResetProcesses() })

	tunnel := &failOnceStopTunnel{failed: make(map[int]bool)}
	original := internal.DefaultTunnel
	internal.DefaultTunnel = tunnel
	t.Cleanup(func() { internal.DefaultTunnel = original })

	// Start two real processes (simulating "core" on 5173 and "ui" on 5174)
	port1 := freePort(t)
	port2 := freePort(t)

	p1, err := internal.CreateProcess("core", port1, t.TempDir())
	if err != nil {
		t.Fatalf("CreateProcess core failed: %v", err)
	}
	if err := p1.Start(fmt.Sprintf("nc -l %d", port1)); err != nil {
		t.Fatalf("Start core failed: %v", err)
	}

	p2, err := internal.CreateProcess("ui", port2, t.TempDir())
	if err != nil {
		t.Fatalf("CreateProcess ui failed: %v", err)
	}
	if err := p2.Start(fmt.Sprintf("nc -l %d", port2)); err != nil {
		t.Fatalf("Start ui failed: %v", err)
	}

	daemon.SetProcess("core", p1)
	daemon.SetProcess("ui", p2)

	failed := daemon.StopAllProcesses(10 * time.Second)

	if len(failed) > 0 {
		t.Errorf("expected all processes to stop successfully, but ports failed: %v", failed)
	}

	remaining := daemon.GetProcesses()
	if len(remaining) != 0 {
		t.Errorf("expected process map to be empty, got %d remaining", len(remaining))
	}
}
