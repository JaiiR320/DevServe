package daemon

import (
	"devserve/config"
	"devserve/internal/testutil"
	"devserve/process"
	"devserve/tunnel"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func resetState(t *testing.T) {
	t.Helper()
	mu.Lock()
	processes = make(map[string]*process.Process)
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		processes = make(map[string]*process.Process)
		mu.Unlock()
	})
}

func TestHandlePing(t *testing.T) {
	resp := handlePing(nil)

	if !resp.OK {
		t.Errorf("expected OK to be true, got false")
	}
	if resp.Data != "pong" {
		t.Errorf("expected Data %q, got %q", "pong", resp.Data)
	}
}

func TestHandleServeMissingName(t *testing.T) {
	resetState(t)

	resp := handleServe(map[string]any{"port": float64(3000), "command": "echo hi"})

	if resp.OK {
		t.Fatal("expected error response, got OK")
	}
	if !strings.Contains(resp.Error, "missing or invalid 'name'") {
		t.Errorf("expected error to contain %q, got %q", "missing or invalid 'name'", resp.Error)
	}
}

func TestHandleServeMissingPort(t *testing.T) {
	resetState(t)

	resp := handleServe(map[string]any{"name": "app", "command": "echo hi"})

	if resp.OK {
		t.Fatal("expected error response, got OK")
	}
	if !strings.Contains(resp.Error, "missing or invalid 'port'") {
		t.Errorf("expected error to contain %q, got %q", "missing or invalid 'port'", resp.Error)
	}
}

func TestHandleServeMissingCommand(t *testing.T) {
	resetState(t)

	port := testutil.FreePort(t)

	resp := handleServe(map[string]any{"name": "app", "port": float64(port)})

	if resp.OK {
		t.Fatal("expected error response, got OK")
	}
	if !strings.Contains(resp.Error, "missing or invalid 'command'") {
		t.Errorf("expected error to contain %q, got %q", "missing or invalid 'command'", resp.Error)
	}
}

func TestHandleServeDuplicateName(t *testing.T) {
	resetState(t)

	mu.Lock()
	processes["myapp"] = &process.Process{Name: "myapp", Port: 3000}
	mu.Unlock()

	resp := handleServe(map[string]any{
		"name":    "myapp",
		"port":    float64(4000),
		"command": "echo hi",
	})

	if resp.OK {
		t.Fatal("expected error response, got OK")
	}
	if !strings.Contains(resp.Error, "already in use") {
		t.Errorf("expected error to contain %q, got %q", "already in use", resp.Error)
	}
}

func TestHandleServePortTypes(t *testing.T) {
	resetState(t)

	// Use an occupied port so handleServe fails fast at CheckPortInUse,
	// proving the port type was parsed successfully (no "invalid port type" error).
	port := testutil.OccupiedPort(t)

	t.Run("float64 port passes validation", func(t *testing.T) {
		resp := handleServe(map[string]any{
			"name":    "app-float",
			"port":    float64(port),
			"command": "echo hi",
		})
		if strings.Contains(resp.Error, "invalid port type") {
			t.Errorf("float64 port should not produce 'invalid port type' error, got %q", resp.Error)
		}
		if !strings.Contains(resp.Error, "already in use") {
			t.Errorf("expected error to contain %q (port parsed OK), got %q", "already in use", resp.Error)
		}
	})

	t.Run("string port passes validation", func(t *testing.T) {
		portStr := strconv.Itoa(port)
		resp := handleServe(map[string]any{
			"name":    "app-string",
			"port":    portStr,
			"command": "echo hi",
		})
		if strings.Contains(resp.Error, "invalid port type") {
			t.Errorf("string port should not produce 'invalid port type' error, got %q", resp.Error)
		}
		if !strings.Contains(resp.Error, "already in use") {
			t.Errorf("expected error to contain %q (port parsed OK), got %q", "already in use", resp.Error)
		}
	})
}

func TestHandleServeInvalidPortType(t *testing.T) {
	resetState(t)

	resp := handleServe(map[string]any{
		"name":    "app",
		"port":    true, // bool is not a valid port type
		"command": "echo hi",
	})

	if resp.OK {
		t.Fatal("expected error response, got OK")
	}
	if !strings.Contains(resp.Error, "invalid port type") {
		t.Errorf("expected error to contain %q, got %q", "invalid port type", resp.Error)
	}
}

func TestHandleStopMissingName(t *testing.T) {
	resetState(t)

	resp := handleStop(map[string]any{})

	if resp.OK {
		t.Fatal("expected error response, got OK")
	}
	if !strings.Contains(resp.Error, "missing or invalid 'name'") {
		t.Errorf("expected error to contain %q, got %q", "missing or invalid 'name'", resp.Error)
	}
}

func TestHandleStopNotFound(t *testing.T) {
	resetState(t)

	resp := handleStop(map[string]any{"name": "ghost"})

	if resp.OK {
		t.Fatal("expected error response, got OK")
	}
	if !strings.Contains(resp.Error, "not found") {
		t.Errorf("expected error to contain %q, got %q", "not found", resp.Error)
	}
}

func TestHandleListEmpty(t *testing.T) {
	resetState(t)

	origRunner := tunnel.DefaultRunner
	tunnel.SetRunner(func() ([]byte, error) {
		return []byte(`{"TailscaleIPs":["100.1.2.3"],"Self":{"DNSName":"host.example.ts.net."}}`), nil
	})
	t.Cleanup(func() { tunnel.SetRunner(origRunner) })

	resp := handleList(nil)

	if !resp.OK {
		t.Fatalf("expected OK response, got error: %s", resp.Error)
	}

	var lr listResponse
	if err := json.Unmarshal([]byte(resp.Data), &lr); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(lr.Processes) != 0 {
		t.Errorf("expected 0 processes, got %d", len(lr.Processes))
	}
	if lr.Hostname != "host.example.ts.net" {
		t.Errorf("expected hostname %q, got %q", "host.example.ts.net", lr.Hostname)
	}
	if lr.IP != "100.1.2.3" {
		t.Errorf("expected IP %q, got %q", "100.1.2.3", lr.IP)
	}
}

func TestHandleListPopulated(t *testing.T) {
	resetState(t)

	origRunner := tunnel.DefaultRunner
	tunnel.SetRunner(func() ([]byte, error) {
		return []byte(`{"TailscaleIPs":["100.1.2.3"],"Self":{"DNSName":"host.example.ts.net."}}`), nil
	})
	t.Cleanup(func() { tunnel.SetRunner(origRunner) })

	mu.Lock()
	processes["web"] = &process.Process{Name: "web", Port: 3000}
	processes["api"] = &process.Process{Name: "api", Port: 4000}
	mu.Unlock()

	resp := handleList(nil)

	if !resp.OK {
		t.Fatalf("expected OK response, got error: %s", resp.Error)
	}

	var lr listResponse
	if err := json.Unmarshal([]byte(resp.Data), &lr); err != nil {
		t.Fatalf("failed to parse response data as JSON: %v", err)
	}

	if len(lr.Processes) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(lr.Processes))
	}

	found := map[string]int{}
	for _, e := range lr.Processes {
		found[e.Name] = e.Port
	}
	if found["web"] != 3000 {
		t.Errorf("expected web on port 3000, got %d", found["web"])
	}
	if found["api"] != 4000 {
		t.Errorf("expected api on port 4000, got %d", found["api"])
	}
	if lr.Hostname != "host.example.ts.net" {
		t.Errorf("expected hostname %q, got %q", "host.example.ts.net", lr.Hostname)
	}
}

func TestHandleListTailscaleUnavailable(t *testing.T) {
	resetState(t)

	origRunner := tunnel.DefaultRunner
	tunnel.SetRunner(func() ([]byte, error) {
		return nil, fmt.Errorf("tailscale not running")
	})
	t.Cleanup(func() { tunnel.SetRunner(origRunner) })

	resp := handleList(nil)

	if resp.OK {
		t.Fatal("expected error response when tailscale unavailable, got OK")
	}
	if !strings.Contains(resp.Error, "tailscale") {
		t.Errorf("expected error to mention tailscale, got %q", resp.Error)
	}
}

func TestHandleLogsMissingName(t *testing.T) {
	resetState(t)

	resp := handleLogs(map[string]any{})

	if resp.OK {
		t.Fatal("expected error response, got OK")
	}
	if !strings.Contains(resp.Error, "missing or invalid 'name'") {
		t.Errorf("expected error to contain %q, got %q", "missing or invalid 'name'", resp.Error)
	}
}

func TestHandleLogsNotFound(t *testing.T) {
	resetState(t)

	resp := handleLogs(map[string]any{"name": "ghost"})

	if resp.OK {
		t.Fatal("expected error response, got OK")
	}
	if !strings.Contains(resp.Error, "not found") {
		t.Errorf("expected error to contain %q, got %q", "not found", resp.Error)
	}
}

func TestHandleLogsValid(t *testing.T) {
	resetState(t)

	dir := t.TempDir()
	logDir := filepath.Join(dir, config.ProcessLogDir)
	if err := os.MkdirAll(logDir, config.DirPermissions); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}

	stdoutContent := "line one\nline two\nline three\n"
	stderrContent := "error alpha\nerror beta\n"

	if err := os.WriteFile(filepath.Join(logDir, config.ProcessStdoutLog), []byte(stdoutContent), 0644); err != nil {
		t.Fatalf("failed to write stdout log: %v", err)
	}
	if err := os.WriteFile(filepath.Join(logDir, config.ProcessStderrLog), []byte(stderrContent), 0644); err != nil {
		t.Fatalf("failed to write stderr log: %v", err)
	}

	mu.Lock()
	processes["myapp"] = &process.Process{Name: "myapp", Port: 3000, Dir: dir}
	mu.Unlock()

	resp := handleLogs(map[string]any{"name": "myapp"})

	if !resp.OK {
		t.Fatalf("expected OK response, got error: %s", resp.Error)
	}
	if !strings.Contains(resp.Data, "stdout") {
		t.Errorf("expected output to contain stdout header, got %q", resp.Data)
	}
	if !strings.Contains(resp.Data, "stderr") {
		t.Errorf("expected output to contain stderr header, got %q", resp.Data)
	}
	if !strings.Contains(resp.Data, "line one") {
		t.Errorf("expected output to contain stdout content 'line one', got %q", resp.Data)
	}
	if !strings.Contains(resp.Data, "line three") {
		t.Errorf("expected output to contain stdout content 'line three', got %q", resp.Data)
	}
	if !strings.Contains(resp.Data, "error alpha") {
		t.Errorf("expected output to contain stderr content 'error alpha', got %q", resp.Data)
	}
}
