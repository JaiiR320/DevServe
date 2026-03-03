package testutil

import (
	"fmt"
	"net"
	"os/exec"
	"testing"
)

// FreePort binds to :0, captures the assigned port, then closes the listener.
func FreePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

// OccupiedPort starts a TCP listener and returns its port.
// The listener is closed when the test finishes.
func OccupiedPort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to get occupied port: %v", err)
	}
	t.Cleanup(func() { l.Close() })
	return l.Addr().(*net.TCPAddr).Port
}

// RequireNC skips the test if nc (netcat) is not available.
func RequireNC(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("nc"); err != nil {
		t.Skip("nc not available")
	}
}

// NoopTunnel implements tunnel.Tunnel with no-op Serve and Stop.
type NoopTunnel struct{}

func (NoopTunnel) Serve(port int) error { return nil }
func (NoopTunnel) Stop(port int) error  { return nil }

// FailOnceStopTunnel fails the first Stop() call per port, succeeds on retry.
// Serve() always succeeds.
type FailOnceStopTunnel struct {
	failed map[int]bool
}

func NewFailOnceStopTunnel() *FailOnceStopTunnel {
	return &FailOnceStopTunnel{failed: make(map[int]bool)}
}

func (f *FailOnceStopTunnel) Serve(port int) error { return nil }
func (f *FailOnceStopTunnel) Stop(port int) error {
	if !f.failed[port] {
		f.failed[port] = true
		return fmt.Errorf("tailscale serve stop failed for port %d", port)
	}
	return nil
}
