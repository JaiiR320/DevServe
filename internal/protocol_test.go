package internal

import (
	"encoding/json"
	"errors"
	"net"
	"strings"
	"testing"
)

// Task 1.1: Test OkResponse — returns Response{OK: true, Data: data}
func TestOkResponse(t *testing.T) {
	resp := OkResponse("hello world")

	if !resp.OK {
		t.Errorf("expected OK to be true, got false")
	}
	if resp.Data != "hello world" {
		t.Errorf("expected Data to be %q, got %q", "hello world", resp.Data)
	}
	if resp.Error != "" {
		t.Errorf("expected Error to be empty, got %q", resp.Error)
	}
}

// Task 1.2: Test ErrResponse — returns Response{OK: false, Error: err.Error()}
func TestErrResponse(t *testing.T) {
	err := errors.New("something went wrong")
	resp := ErrResponse(err)

	if resp.OK {
		t.Errorf("expected OK to be false, got true")
	}
	if resp.Error != "something went wrong" {
		t.Errorf("expected Error to be %q, got %q", "something went wrong", resp.Error)
	}
	if resp.Data != "" {
		t.Errorf("expected Data to be empty, got %q", resp.Data)
	}
}

// Task 1.3: Test SendRequest / ReadRequest round-trip
func TestSendReadRequestRoundTrip(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	req := &Request{
		Action: "serve",
		Args:   map[string]any{"name": "myapp", "port": float64(3000)},
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- SendRequest(client, req)
	}()

	got, err := ReadRequest(server)
	if err != nil {
		t.Fatalf("ReadRequest failed: %v", err)
	}

	if sendErr := <-errCh; sendErr != nil {
		t.Fatalf("SendRequest failed: %v", sendErr)
	}

	if got.Action != req.Action {
		t.Errorf("expected Action %q, got %q", req.Action, got.Action)
	}
	if got.Args["name"] != req.Args["name"] {
		t.Errorf("expected Args[name] %q, got %q", req.Args["name"], got.Args["name"])
	}
	if got.Args["port"] != req.Args["port"] {
		t.Errorf("expected Args[port] %v, got %v", req.Args["port"], got.Args["port"])
	}
}

// Task 1.4: Test SendResponse / ReadResponse round-trip
func TestSendReadResponseRoundTrip(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	resp := &Response{OK: true, Data: "process started"}

	errCh := make(chan error, 1)
	go func() {
		errCh <- SendResponse(server, resp)
	}()

	got, err := ReadResponse(client)
	if err != nil {
		t.Fatalf("ReadResponse failed: %v", err)
	}

	if sendErr := <-errCh; sendErr != nil {
		t.Fatalf("SendResponse failed: %v", sendErr)
	}

	if got.OK != resp.OK {
		t.Errorf("expected OK %v, got %v", resp.OK, got.OK)
	}
	if got.Data != resp.Data {
		t.Errorf("expected Data %q, got %q", resp.Data, got.Data)
	}
	if got.Error != resp.Error {
		t.Errorf("expected Error %q, got %q", resp.Error, got.Error)
	}
}

// Task 1.5: Test ReadRequest with malformed JSON
func TestReadRequestMalformedJSON(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()

	go func() {
		client.Write([]byte("not valid json\n"))
		client.Close()
	}()

	_, err := ReadRequest(server)
	if err == nil {
		t.Fatal("expected error from ReadRequest with malformed JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode request") {
		t.Errorf("expected error to contain %q, got %q", "failed to decode request", err.Error())
	}
}

// Task 1.6: Test ReadResponse with malformed JSON
func TestReadResponseMalformedJSON(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()

	go func() {
		client.Write([]byte("{invalid json}\n"))
		client.Close()
	}()

	_, err := ReadResponse(server)
	if err == nil {
		t.Fatal("expected error from ReadResponse with malformed JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode response") {
		t.Errorf("expected error to contain %q, got %q", "failed to decode response", err.Error())
	}
}

// Task 1.7: Test Request with nil/empty Args — omitempty behavior
func TestRequestOmitEmptyArgs(t *testing.T) {
	t.Run("nil args omitted", func(t *testing.T) {
		req := Request{Action: "ping", Args: nil}
		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		if strings.Contains(string(data), `"args"`) {
			t.Errorf("expected args to be omitted from JSON, got %s", string(data))
		}
	})

	t.Run("populated args included", func(t *testing.T) {
		req := Request{Action: "serve", Args: map[string]any{"name": "app"}}
		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		if !strings.Contains(string(data), `"args"`) {
			t.Errorf("expected args to be present in JSON, got %s", string(data))
		}
	})
}

// Task 1.8: Test Response with empty Data/Error — omitempty behavior
func TestResponseOmitEmptyFields(t *testing.T) {
	t.Run("ok response omits error", func(t *testing.T) {
		resp := Response{OK: true, Data: "result"}
		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		if strings.Contains(string(data), `"error"`) {
			t.Errorf("expected error to be omitted from JSON, got %s", string(data))
		}
		if !strings.Contains(string(data), `"data"`) {
			t.Errorf("expected data to be present in JSON, got %s", string(data))
		}
	})

	t.Run("error response omits data", func(t *testing.T) {
		resp := Response{OK: false, Error: "something failed"}
		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		if strings.Contains(string(data), `"data"`) {
			t.Errorf("expected data to be omitted from JSON, got %s", string(data))
		}
		if !strings.Contains(string(data), `"error"`) {
			t.Errorf("expected error to be present in JSON, got %s", string(data))
		}
	})

	t.Run("minimal ok response omits data and error", func(t *testing.T) {
		resp := Response{OK: true}
		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		if strings.Contains(string(data), `"data"`) {
			t.Errorf("expected data to be omitted from JSON, got %s", string(data))
		}
		if strings.Contains(string(data), `"error"`) {
			t.Errorf("expected error to be omitted from JSON, got %s", string(data))
		}
	})
}
