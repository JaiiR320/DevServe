package process_test

import (
	"devserve/process"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

// freePort starts a TCP listener on :0 to get a free port, then closes it.
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

func TestCheckPortInUseFree(t *testing.T) {
	port := freePort(t)

	err := process.CheckPortInUse(port)
	if err != nil {
		t.Errorf("expected no error for free port %d, got %v", port, err)
	}
}

func TestCheckPortInUseOccupied(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer l.Close()

	port := l.Addr().(*net.TCPAddr).Port
	err = process.CheckPortInUse(port)
	if err == nil {
		t.Fatalf("expected error for occupied port %d, got nil", port)
	}
	if !strings.Contains(err.Error(), "already in use") {
		t.Errorf("expected error to contain %q, got %q", "already in use", err.Error())
	}
}

func TestWaitForPortImmediate(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer l.Close()

	port := l.Addr().(*net.TCPAddr).Port
	start := time.Now()
	err = process.WaitForPort(port, 2*time.Second)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if elapsed > 1*time.Second {
		t.Errorf("expected quick return, took %s", elapsed)
	}
}

func TestWaitForPortDelayed(t *testing.T) {
	port := freePort(t)

	go func() {
		time.Sleep(200 * time.Millisecond)
		l, err := net.Listen("tcp", "localhost:0")
		if err != nil {
			return
		}
		// We need to listen on the specific port, not :0
		l.Close()
		l, _ = net.Listen("tcp", net.JoinHostPort("localhost", fmt.Sprintf("%d", port)))
		if l != nil {
			// Keep alive until test completes
			time.Sleep(5 * time.Second)
			l.Close()
		}
	}()

	err := process.WaitForPort(port, 3*time.Second)
	if err != nil {
		t.Fatalf("expected port to become available, got %v", err)
	}
}

func TestWaitForPortTimeout(t *testing.T) {
	port := freePort(t)

	start := time.Now()
	err := process.WaitForPort(port, 300*time.Millisecond)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "not ready after") {
		t.Errorf("expected error to contain %q, got %q", "not ready after", err.Error())
	}
	// Should not take much longer than the timeout
	if elapsed > 2*time.Second {
		t.Errorf("timeout took too long: %s", elapsed)
	}
}
