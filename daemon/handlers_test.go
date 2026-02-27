package daemon

import (
	"devserve/internal"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// occupiedPort starts a TCP listener and returns its port. The listener is
// closed when the test finishes.
func occupiedPort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to get occupied port: %v", err)
	}
	t.Cleanup(func() { l.Close() })
	return l.Addr().(*net.TCPAddr).Port
}

func resetState(t *testing.T) {
	t.Helper()
	ResetProcesses()
	t.Cleanup(func() { ResetProcesses() })
}

// Task 6.2: Test handlePing — assert returns OkResponse with data "pong"
func TestHandlePing(t *testing.T) {
	resp := handlePing(nil)

	if !resp.OK {
		t.Errorf("expected OK to be true, got false")
	}
	if resp.Data != "pong" {
		t.Errorf("expected Data %q, got %q", "pong", resp.Data)
	}
}

// Task 6.3: Test handleServe with missing name arg
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

// Task 6.4: Test handleServe with missing port arg
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

// Task 6.5: Test handleServe with missing command arg
func TestHandleServeMissingCommand(t *testing.T) {
	resetState(t)

	// Use a port that passes CheckPortInUse so we reach the command validation.
	// Get a free port by binding then immediately closing.
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	resp := handleServe(map[string]any{"name": "app", "port": float64(port)})

	if resp.OK {
		t.Fatal("expected error response, got OK")
	}
	if !strings.Contains(resp.Error, "missing or invalid 'command'") {
		t.Errorf("expected error to contain %q, got %q", "missing or invalid 'command'", resp.Error)
	}
}

// Task 6.6: Test handleServe with duplicate name
func TestHandleServeDuplicateName(t *testing.T) {
	resetState(t)

	SetProcess("myapp", &internal.Process{Name: "myapp", Port: 3000})

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

// Task 6.7: Test handleServe port type handling — float64 (JSON number) and string
func TestHandleServePortTypes(t *testing.T) {
	resetState(t)

	// Use an occupied port so handleServe fails fast at CheckPortInUse,
	// proving the port type was parsed successfully (no "invalid port type" error).
	port := occupiedPort(t)

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

// Task 6.8: Test handleServe with invalid port type
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

// Task 6.9: Test handleStop with missing name
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

// Task 6.10: Test handleStop with nonexistent process
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

// Task 6.11: Test handleList with empty map
func TestHandleListEmpty(t *testing.T) {
	resetState(t)

	origRunner := tsRunner
	tsRunner = func() ([]byte, error) {
		return []byte(`{"TailscaleIPs":["100.1.2.3"],"Self":{"DNSName":"host.example.ts.net."}}`), nil
	}
	t.Cleanup(func() { tsRunner = origRunner })

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

// Task 6.12: Test handleList with populated map
func TestHandleListPopulated(t *testing.T) {
	resetState(t)

	origRunner := tsRunner
	tsRunner = func() ([]byte, error) {
		return []byte(`{"TailscaleIPs":["100.1.2.3"],"Self":{"DNSName":"host.example.ts.net."}}`), nil
	}
	t.Cleanup(func() { tsRunner = origRunner })

	SetProcess("web", &internal.Process{Name: "web", Port: 3000})
	SetProcess("api", &internal.Process{Name: "api", Port: 4000})

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

// Test handleList when tailscale is unavailable — returns error response
func TestHandleListTailscaleUnavailable(t *testing.T) {
	resetState(t)

	origRunner := tsRunner
	tsRunner = func() ([]byte, error) {
		return nil, fmt.Errorf("tailscale not running")
	}
	t.Cleanup(func() { tsRunner = origRunner })

	resp := handleList(nil)

	if resp.OK {
		t.Fatal("expected error response when tailscale unavailable, got OK")
	}
	if !strings.Contains(resp.Error, "tailscale") {
		t.Errorf("expected error to mention tailscale, got %q", resp.Error)
	}
}

// Task 6.13: Test handleLogs with missing name
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

// Task 6.14: Test handleLogs with nonexistent process
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

// Task 6.15: Test handleLogs with valid process — log files with known content
func TestHandleLogsValid(t *testing.T) {
	resetState(t)

	dir := t.TempDir()
	logDir := filepath.Join(dir, internal.ProcessLogDir)
	if err := os.MkdirAll(logDir, internal.DirPermissions); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}

	stdoutContent := "line one\nline two\nline three\n"
	stderrContent := "error alpha\nerror beta\n"

	if err := os.WriteFile(filepath.Join(logDir, internal.ProcessStdoutLog), []byte(stdoutContent), 0644); err != nil {
		t.Fatalf("failed to write stdout log: %v", err)
	}
	if err := os.WriteFile(filepath.Join(logDir, internal.ProcessStderrLog), []byte(stderrContent), 0644); err != nil {
		t.Fatalf("failed to write stderr log: %v", err)
	}

	SetProcess("myapp", &internal.Process{Name: "myapp", Port: 3000, Dir: dir})

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
