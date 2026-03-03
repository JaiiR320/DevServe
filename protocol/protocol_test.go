package protocol

import (
	"fmt"
	"net"
	"testing"
)

func TestRequestResponseRoundTrip(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	req := &Request{Action: "ping"}

	// Send request from client
	go func() {
		if err := SendRequest(client, req); err != nil {
			t.Errorf("failed to send request: %v", err)
		}
	}()

	// Read request on server side
	received, err := ReadRequest(server)
	if err != nil {
		t.Fatalf("failed to read request: %v", err)
	}

	if received.Action != "ping" {
		t.Errorf("expected action 'ping', got %q", received.Action)
	}

	// Send response back
	resp := OkResponse("pong")
	go func() {
		if err := SendResponse(server, resp); err != nil {
			t.Errorf("failed to send response: %v", err)
		}
	}()

	// Read response on client side
	receivedResp, err := ReadResponse(client)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	if !receivedResp.OK {
		t.Errorf("expected OK response")
	}
	if receivedResp.Data != "pong" {
		t.Errorf("expected data 'pong', got %q", receivedResp.Data)
	}
}

func TestErrResponse(t *testing.T) {
	resp := ErrResponse(fmt.Errorf("test error"))
	if resp.OK {
		t.Error("expected OK to be false")
	}
	if resp.Error != "test error" {
		t.Errorf("expected error 'test error', got %q", resp.Error)
	}
}

func TestMalformedInput(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	// Send garbage bytes
	go func() {
		client.Write([]byte("not valid json\n"))
		client.Close()
	}()

	_, err := ReadRequest(server)
	if err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}

func TestOmitempty(t *testing.T) {
	req := Request{Action: "test"}
	resp := Response{OK: true}

	// Ensure Args and Data are omitted when empty
	if req.Args != nil {
		t.Error("expected Args to be nil (omitted when empty)")
	}
	if resp.Data != "" {
		t.Error("expected Data to be empty (omitted when empty)")
	}
}
