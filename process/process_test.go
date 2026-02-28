package process_test

import (
	"devserve/config"
	"devserve/internal/testutil"
	"devserve/process"
	"devserve/tunnel"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	tunnel.SetTunnel(testutil.NoopTunnel{})
	t.Cleanup(func() { tunnel.SetTunnel(original) })
}

func TestCreateProcess(t *testing.T) {
	dir := t.TempDir()
	p, err := process.CreateProcess("testapp", 3000, dir, "echo test")
	if err != nil {
		t.Fatalf("CreateProcess failed: %v", err)
	}
	defer p.Stdout.Close()
	defer p.Stderr.Close()

	logDir := filepath.Join(dir, config.ProcessLogDir)
	info, err := os.Stat(logDir)
	if err != nil {
		t.Fatalf("expected log directory %q to exist: %v", logDir, err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", logDir)
	}

	for _, name := range []string{config.ProcessStdoutLog, config.ProcessStderrLog} {
		path := filepath.Join(logDir, name)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected log file %q to exist: %v", path, err)
		}
	}
}

func TestCreateProcessInvalidDir(t *testing.T) {
	_, err := process.CreateProcess("testapp", 3000, "/dev/null/bad", "echo test")
	if err == nil {
		t.Fatal("expected error for invalid directory, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create log directory") {
		t.Errorf("expected error to contain %q, got %q", "failed to create log directory", err.Error())
	}
}

func TestProcessStart(t *testing.T) {
	testutil.RequireNC(t)
	swapTunnel(t)

	port := testutil.FreePort(t)
	dir := t.TempDir()
	p, err := process.CreateProcess("testapp", port, dir, "echo test")
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
	testutil.RequireNC(t)
	swapTunnel(t)

	port := testutil.FreePort(t)
	dir := t.TempDir()
	p, err := process.CreateProcess("testapp", port, dir, "echo test")
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
	testutil.RequireNC(t)
	swapTunnel(t)

	port := testutil.FreePort(t)
	dir := t.TempDir()
	p, err := process.CreateProcess("testapp", port, dir, "echo test")
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
	p, err := process.CreateProcess("testapp", 3000, dir, "echo test")
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
	testutil.RequireNC(t)

	mockTunnel := &failOnceStopTunnel{}
	original := tunnel.DefaultTunnel
	tunnel.SetTunnel(mockTunnel)
	t.Cleanup(func() { tunnel.SetTunnel(original) })

	port := testutil.FreePort(t)
	dir := t.TempDir()
	p, err := process.CreateProcess("testapp", port, dir, "echo test")
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

	if mockTunnel.stopCalls != 2 {
		t.Errorf("expected tunnel Stop to be called 2 times, got %d", mockTunnel.stopCalls)
	}
}
