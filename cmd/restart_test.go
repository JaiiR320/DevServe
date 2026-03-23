package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/jaiir320/devserve/config"
	"github.com/jaiir320/devserve/protocol"
)

func TestRestartWaitsForPortReleaseBeforeServe(t *testing.T) {
	t.Helper()
	port := freePort(t)

	oldSocket := config.Socket
	config.Socket = filepath.Join(t.TempDir(), "daemon.sock")
	t.Cleanup(func() {
		config.Socket = oldSocket
	})

	listener, stopServer := startRestartTestDaemon(t, config.Socket, port, 250*time.Millisecond)
	defer stopServer()
	defer listener.Close()

	start := time.Now()
	err := restartCmd.RunE(restartCmd, []string{"opencode"})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("expected restart to wait for port release and succeed, got %v", err)
	}
	if elapsed < 250*time.Millisecond {
		t.Fatalf("expected restart to wait for port release, completed in %s", elapsed)
	}
	if elapsed > 5*time.Second {
		t.Fatalf("expected restart to finish shortly after port release, took %s", elapsed)
	}
}

func startRestartTestDaemon(t *testing.T, socketPath string, port int, releaseDelay time.Duration) (net.Listener, func()) {
	t.Helper()

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to listen on test socket: %v", err)
	}

	var heldPort net.Listener
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			func() {
				defer conn.Close()

				req, err := protocol.ReadRequest(conn)
				if err != nil {
					t.Errorf("failed to read request: %v", err)
					return
				}

				resp := restartTestResponse(t, req, port, releaseDelay, &heldPort)
				if err := protocol.SendResponse(conn, resp); err != nil {
					t.Errorf("failed to send response: %v", err)
				}
			}()
		}
	}()

	return listener, func() {
		if heldPort != nil {
			_ = heldPort.Close()
		}
		_ = listener.Close()
		<-done
	}
}

func restartTestResponse(t *testing.T, req *protocol.Request, port int, releaseDelay time.Duration, heldPort *net.Listener) *protocol.Response {
	switch req.Action {
	case "get":
		data, _ := json.Marshal(protocol.ProcessInfo{
			Name:    "opencode",
			Port:    port,
			Command: "npm run dev",
			Dir:     "/tmp/opencode",
		})
		return protocol.OkResponse(string(data))
	case "stop":
		listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		if err != nil {
			t.Fatalf("failed to hold test port %d: %v", port, err)
		}
		*heldPort = listener
		go func() {
			time.Sleep(releaseDelay)
			_ = listener.Close()
		}()
		return protocol.OkResponse("process 'opencode' stopped")
	case "serve":
		data, _ := json.Marshal(protocol.ServeResult{
			Name: "opencode",
			Port: port,
		})
		return protocol.OkResponse(string(data))
	default:
		return protocol.ErrResponse(fmt.Errorf("unexpected action %q", req.Action))
	}
}

func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to allocate free port: %v", err)
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}
