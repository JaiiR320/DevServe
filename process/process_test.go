package process_test

import (
	"devserve/process"
	"devserve/tunnel"
	"devserve/util"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type noopTunnel struct{}

func (noopTunnel) Serve(port int) error { return nil }
func (noopTunnel) Stop(port int) error  { return nil }

// failOnceStopTunnel fails the first Stop() call and succeeds on subsequent calls.
// Serve() always succeeds. stopCalls tracks total Stop() invocations.
type failOnceStopTunnel struct {
	stopCalls int
}

func (f *failOnceStopTunnel) Serve(port int) error { return nil }
func (f *failOnceStopTunnel) Stop(port int) error {
	f.stopCalls++
	if f.stopCalls == 1 {
		return fmt.Errorf("tailscale serve stop failed")
	}
	return nil
}

func swapTunnel(t *testing.T) {
	t.Helper()
	original := tunnel.DefaultTunnel
	tunnel.DefaultTunnel = noopTunnel{}
	t.Cleanup(func() { tunnel.DefaultTunnel = original })
}

func requireNC(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("nc"); err != nil {
		t.Skip("nc not available")
	}
}

func TestCreateProcess(t *testing.T) {
	dir := t.TempDir()
	p, err := process.CreateProcess("testapp", 3000, dir)
	if err != nil {
		t.Fatalf("CreateProcess failed: %v", err)
	}
	defer p.Stdout.Close()
	defer p.Stderr.Close()

	logDir := filepath.Join(dir, util.ProcessLogDir)
	info, err := os.Stat(logDir)
	if err != nil {
		t.Fatalf("expected log directory %q to exist: %v", logDir, err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", logDir)
	}

	for _, name := range []string{util.ProcessStdoutLog, util.ProcessStderrLog} {
		path := filepath.Join(logDir, name)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected log file %q to exist: %v", path, err)
		}
	}
}

func TestCreateProcessInvalidDir(t *testing.T) {
	_, err := process.CreateProcess("testapp", 3000, "/dev/null/bad")
	if err == nil {
		t.Fatal("expected error for invalid directory, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create log directory") {
		t.Errorf("expected error to contain %q, got %q", "failed to create log directory", err.Error())
	}
}

func TestProcessStart(t *testing.T) {
	requireNC(t)
	swapTunnel(t)

	port := freePort(t)
	dir := t.TempDir()
	p, err := process.CreateProcess("testapp", port, dir)
	if err != nil {
		t.Fatalf("CreateProcess failed: %v", err)
	}

	cmd := fmt.Sprintf("nc -l %d", port)
	if err := p.Start(cmd); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer p.Stop()

	if p.Cmd.Process == nil {
		t.Fatal("expected process to be running")
	}

	// Port was verified reachable by Start() via WaitForPort.
	// nc -l exits after the first connection, so we just verify the process started.
}

func TestProcessStopAfterStart(t *testing.T) {
	requireNC(t)
	swapTunnel(t)

	port := freePort(t)
	dir := t.TempDir()
	p, err := process.CreateProcess("testapp", port, dir)
	if err != nil {
		t.Fatalf("CreateProcess failed: %v", err)
	}

	cmd := fmt.Sprintf("nc -l %d", port)
	if err := p.Start(cmd); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if err := p.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	_, err = p.Stdout.Write([]byte("test"))
	if err == nil {
		t.Error("expected write to closed stdout log to fail")
	}
	_, err = p.Stderr.Write([]byte("test"))
	if err == nil {
		t.Error("expected write to closed stderr log to fail")
	}
}

func TestProcessStopIdempotent(t *testing.T) {
	requireNC(t)
	swapTunnel(t)

	port := freePort(t)
	dir := t.TempDir()
	p, err := process.CreateProcess("testapp", port, dir)
	if err != nil {
		t.Fatalf("CreateProcess failed: %v", err)
	}

	cmd := fmt.Sprintf("nc -l %d", port)
	if err := p.Start(cmd); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if err := p.Stop(); err != nil {
		t.Fatalf("first Stop failed: %v", err)
	}

	err = p.Stop()
	if err == nil {
		t.Fatal("expected error on second Stop, got nil")
	}
	if !strings.Contains(err.Error(), "already stopped") {
		t.Errorf("expected error to contain %q, got %q", "already stopped", err.Error())
	}
}

func TestProcessStopBeforeStart(t *testing.T) {
	dir := t.TempDir()
	p, err := process.CreateProcess("testapp", 3000, dir)
	if err != nil {
		t.Fatalf("CreateProcess failed: %v", err)
	}
	defer p.Stdout.Close()
	defer p.Stderr.Close()

	err = p.Stop()
	if err == nil {
		t.Fatal("expected error on Stop before Start, got nil")
	}
	if !strings.Contains(err.Error(), "has not been started") {
		t.Errorf("expected error to contain %q, got %q", "has not been started", err.Error())
	}
}

// Bug #16: When tailscale serve stop fails, a retry of Stop() should attempt
// to close the tailscale serve again rather than returning "already stopped".
func TestProcessStopCanRetryAfterTailscaleFailure(t *testing.T) {
	requireNC(t)

	mock_tunnel := &failOnceStopTunnel{}
	original := tunnel.DefaultTunnel
	tunnel.DefaultTunnel = mock_tunnel
	t.Cleanup(func() { tunnel.DefaultTunnel = original })

	port := freePort(t)
	dir := t.TempDir()
	p, err := process.CreateProcess("testapp", port, dir)
	if err != nil {
		t.Fatalf("CreateProcess failed: %v", err)
	}

	cmd := fmt.Sprintf("nc -l %d", port)
	if err := p.Start(cmd); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// First Stop: tailscale fails, should return error
	err = p.Stop()
	if err == nil {
		t.Fatal("expected error from first Stop (tailscale failure), got nil")
	}
	if !strings.Contains(err.Error(), "tailscale") {
		t.Fatalf("expected tailscale-related error, got %q", err.Error())
	}

	// Second Stop: should retry tailscale stop, not return "already stopped"
	err = p.Stop()
	if err != nil {
		t.Fatalf("expected second Stop to succeed after tailscale retry, got %v", err)
	}

	if mock_tunnel.stopCalls != 2 {
		t.Errorf("expected tunnel Stop to be called 2 times, got %d", mock_tunnel.stopCalls)
	}
}
