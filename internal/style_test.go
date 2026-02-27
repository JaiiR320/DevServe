package internal

import (
	"strings"
	"testing"
)

// Task 4.1: Test Success — assert output contains "✓" and the message string
func TestSuccess(t *testing.T) {
	out := Success("process started")
	if !strings.Contains(out, "✓") {
		t.Errorf("expected output to contain '✓', got %q", out)
	}
	if !strings.Contains(out, "process started") {
		t.Errorf("expected output to contain message, got %q", out)
	}
}

// Task 4.2: Test Error — assert output contains "✗" and the message string
func TestError(t *testing.T) {
	out := Error("connection failed")
	if !strings.Contains(out, "✗") {
		t.Errorf("expected output to contain '✗', got %q", out)
	}
	if !strings.Contains(out, "connection failed") {
		t.Errorf("expected output to contain message, got %q", out)
	}
}

// Task 4.3: Test Info — assert output contains "•" and the message string
func TestInfo(t *testing.T) {
	out := Info("checking status")
	if !strings.Contains(out, "•") {
		t.Errorf("expected output to contain '•', got %q", out)
	}
	if !strings.Contains(out, "checking status") {
		t.Errorf("expected output to contain message, got %q", out)
	}
}

// Task 4.4: Test RenderTable with valid JSON — new format with processes, hostname, IP
func TestRenderTableValid(t *testing.T) {
	input := `{"processes":[{"name":"app","port":3000}],"hostname":"host.example.ts.net","ip":"100.1.2.3"}`
	out := RenderTable(input)

	for _, want := range []string{"NAME", "PORT", "LOCAL", "IP", "DNS", "app", "3000"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got %q", want, out)
		}
	}
	// Check that link labels are present
	if !strings.Contains(out, "local") {
		t.Errorf("expected output to contain link label 'local', got %q", out)
	}
	if !strings.Contains(out, "ip") {
		t.Errorf("expected output to contain link label 'ip', got %q", out)
	}
	if !strings.Contains(out, "dns") {
		t.Errorf("expected output to contain link label 'dns', got %q", out)
	}
}

// Task 4.5: Test RenderTable with multiple entries
func TestRenderTableMultiple(t *testing.T) {
	input := `{"processes":[{"name":"web","port":8080},{"name":"api","port":9090}],"hostname":"host.example.ts.net","ip":"100.1.2.3"}`
	out := RenderTable(input)

	for _, want := range []string{"NAME", "PORT", "web", "8080", "api", "9090"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got %q", want, out)
		}
	}
}

// Task 4.6: Test RenderTable with empty processes array
func TestRenderTableEmpty(t *testing.T) {
	out := RenderTable(`{"processes":[],"hostname":"host.example.ts.net","ip":"100.1.2.3"}`)
	if !strings.Contains(out, "No active processes") {
		t.Errorf("expected output to contain %q, got %q", "No active processes", out)
	}
}

// Task 4.7: Test RenderTable with invalid JSON — assert raw input is returned as fallback
func TestRenderTableInvalidJSON(t *testing.T) {
	input := "this is not json"
	out := RenderTable(input)
	if out != input {
		t.Errorf("expected raw input %q returned as fallback, got %q", input, out)
	}
}

// Test RenderTable hyperlinks contain correct URLs
func TestRenderTableHyperlinks(t *testing.T) {
	input := `{"processes":[{"name":"app","port":3000}],"hostname":"host.example.ts.net","ip":"100.1.2.3"}`
	out := RenderTable(input)

	// Check OSC 8 sequences are present for each URL type
	if !strings.Contains(out, "http://localhost:3000") {
		t.Errorf("expected output to contain localhost URL, got %q", out)
	}
	if !strings.Contains(out, "http://100.1.2.3:3000") {
		t.Errorf("expected output to contain IP URL, got %q", out)
	}
	if !strings.Contains(out, "https://host.example.ts.net:3000") {
		t.Errorf("expected output to contain DNS URL, got %q", out)
	}
}

// Test Hyperlink — assert OSC 8 escape sequences wrap the label with the URL
func TestHyperlink(t *testing.T) {
	url := "https://example.com"
	label := "click"
	out := Hyperlink(url, label)

	// OSC 8 format: \x1b]8;;URL\x1b\\LABEL\x1b]8;;\x1b\\
	if !strings.Contains(out, url) {
		t.Errorf("expected output to contain URL %q, got %q", url, out)
	}
	if !strings.Contains(out, label) {
		t.Errorf("expected output to contain label %q, got %q", label, out)
	}
	if !strings.Contains(out, "\x1b]8;;") {
		t.Errorf("expected output to contain OSC 8 opener, got %q", out)
	}
	// Check the closing sequence
	if !strings.HasSuffix(out, "\x1b]8;;\x1b\\") {
		t.Errorf("expected output to end with OSC 8 closer, got %q", out)
	}
}

// Task 4.8: Test HelpTemplate — assert returns non-empty string containing "Usage:" and "Commands:"
func TestHelpTemplate(t *testing.T) {
	tmpl := HelpTemplate()
	if tmpl == "" {
		t.Fatal("expected non-empty help template")
	}
	if !strings.Contains(tmpl, "Usage:") {
		t.Errorf("expected help template to contain %q", "Usage:")
	}
	if !strings.Contains(tmpl, "Commands:") {
		t.Errorf("expected help template to contain %q", "Commands:")
	}
}
