package daemon

import (
	"devserve/internal"
	"net"
	"strings"
	"testing"
	"time"
)

// Task 7.2: Test handleConn dispatches ping action
func TestHandleConnPing(t *testing.T) {
	ResetProcesses()
	t.Cleanup(func() { ResetProcesses() })

	client, server := net.Pipe()
	defer client.Close()

	stop := make(chan struct{}, 1)

	go handleConn(server, stop)

	if err := internal.SendRequest(client, &internal.Request{Action: "ping"}); err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	resp, err := internal.ReadResponse(client)
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
	ResetProcesses()
	t.Cleanup(func() { ResetProcesses() })

	client, server := net.Pipe()
	defer client.Close()

	stop := make(chan struct{}, 1)

	go handleConn(server, stop)

	if err := internal.SendRequest(client, &internal.Request{Action: "bogus"}); err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	resp, err := internal.ReadResponse(client)
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
	ResetProcesses()
	t.Cleanup(func() { ResetProcesses() })

	client, server := net.Pipe()
	defer client.Close()

	stop := make(chan struct{}, 1)

	go handleConn(server, stop)

	// Write garbage bytes followed by a newline (JSON decoder reads line-delimited)
	if _, err := client.Write([]byte("not valid json\n")); err != nil {
		t.Fatalf("failed to write garbage: %v", err)
	}

	resp, err := internal.ReadResponse(client)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	if resp.OK {
		t.Fatal("expected error response, got OK")
	}
}

// Task 7.5: Test handleConn dispatches shutdown — stop channel is signaled
func TestHandleConnShutdown(t *testing.T) {
	ResetProcesses()
	t.Cleanup(func() { ResetProcesses() })

	client, server := net.Pipe()
	defer client.Close()

	stop := make(chan struct{}, 1)

	go handleConn(server, stop)

	if err := internal.SendRequest(client, &internal.Request{Action: "shutdown"}); err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	resp, err := internal.ReadResponse(client)
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
	ResetProcesses()
	t.Cleanup(func() { ResetProcesses() })

	failed := StopAllProcesses(time.Second)
	if failed != nil {
		t.Errorf("expected nil for empty map, got %v", failed)
	}
}
