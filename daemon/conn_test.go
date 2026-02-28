package daemon

import (
	"devserve/process"
	"net"
	"strings"
	"testing"
	"time"
)

func TestHandleConnPing(t *testing.T) {
	resetState(t)

	client, server := net.Pipe()
	defer client.Close()

	stop := make(chan struct{}, 1)

	go HandleConn(server, stop)

	if err := SendRequest(client, &Request{Action: "ping"}); err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	resp, err := ReadResponse(client)
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

func TestHandleConnUnknownAction(t *testing.T) {
	resetState(t)

	client, server := net.Pipe()
	defer client.Close()

	stop := make(chan struct{}, 1)

	go HandleConn(server, stop)

	if err := SendRequest(client, &Request{Action: "bogus"}); err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	resp, err := ReadResponse(client)
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

func TestHandleConnMalformed(t *testing.T) {
	resetState(t)

	client, server := net.Pipe()
	defer client.Close()

	stop := make(chan struct{}, 1)

	go HandleConn(server, stop)

	// Write garbage bytes followed by a newline (JSON decoder reads line-delimited)
	if _, err := client.Write([]byte("not valid json\n")); err != nil {
		t.Fatalf("failed to write garbage: %v", err)
	}

	resp, err := ReadResponse(client)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	if resp.OK {
		t.Fatal("expected error response, got OK")
	}
}

func TestHandleConnShutdown(t *testing.T) {
	mu.Lock()
	processes = make(map[string]*process.Process)
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		processes = make(map[string]*process.Process)
		mu.Unlock()
	})

	client, server := net.Pipe()
	defer client.Close()

	stop := make(chan struct{}, 1)

	go HandleConn(server, stop)

	if err := SendRequest(client, &Request{Action: "shutdown"}); err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	resp, err := ReadResponse(client)
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
