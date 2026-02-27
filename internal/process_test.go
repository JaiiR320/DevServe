package internal

import (
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

func swapTunnel(t *testing.T) {
	t.Helper()
	original := DefaultTunnel
	DefaultTunnel = noopTunnel{}
	t.Cleanup(func() { DefaultTunnel = original })
}

func requireNC(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("nc"); err != nil {
		t.Skip("nc not available")
	}
}

// Task 5.3: Test CreateProcess — assert .devserve/ directory and log files created
func TestCreateProcess(t *testing.T) {
	dir := t.TempDir()
	p, err := CreateProcess("testapp", 3000, dir)
	if err != nil {
		t.Fatalf("CreateProcess failed: %v", err)
	}
	defer p.Stdout.Close()
	defer p.Stderr.Close()

	logDir := filepath.Join(dir, ProcessLogDir)
	info, err := os.Stat(logDir)
	if err != nil {
		t.Fatalf("expected log directory %q to exist: %v", logDir, err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", logDir)
	}

	for _, name := range []string{ProcessStdoutLog, ProcessStderrLog} {
		path := filepath.Join(logDir, name)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected log file %q to exist: %v", path, err)
		}
	}
}

// Task 5.4: Test CreateProcess with invalid directory
func TestCreateProcessInvalidDir(t *testing.T) {
	_, err := CreateProcess("testapp", 3000, "/dev/null/bad")
	if err == nil {
		t.Fatal("expected error for invalid directory, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create log directory") {
		t.Errorf("expected error to contain %q, got %q", "failed to create log directory", err.Error())
	}
}

// Task 5.5: Test Start with a simple command
func TestProcessStart(t *testing.T) {
	requireNC(t)
	swapTunnel(t)

	port := freePort(t)
	dir := t.TempDir()
	p, err := CreateProcess("testapp", port, dir)
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

// Task 5.6: Test Stop after Start — start a process, stop it, assert logs closed
func TestProcessStopAfterStart(t *testing.T) {
	requireNC(t)
	swapTunnel(t)

	port := freePort(t)
	dir := t.TempDir()
	p, err := CreateProcess("testapp", port, dir)
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

// Task 5.7: Test Stop idempotency — call Stop() twice, assert "already stopped"
func TestProcessStopIdempotent(t *testing.T) {
	requireNC(t)
	swapTunnel(t)

	port := freePort(t)
	dir := t.TempDir()
	p, err := CreateProcess("testapp", port, dir)
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

// Task 5.8: Test Stop before Start — assert "has not been started"
func TestProcessStopBeforeStart(t *testing.T) {
	dir := t.TempDir()
	p, err := CreateProcess("testapp", 3000, dir)
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
