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

// Task 4.4: Test RenderTable with valid JSON array
func TestRenderTableValid(t *testing.T) {
	input := `[{"name":"app","port":3000}]`
	out := RenderTable(input)

	for _, want := range []string{"NAME", "PORT", "app", "3000"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got %q", want, out)
		}
	}
}

// Task 4.5: Test RenderTable with multiple entries
func TestRenderTableMultiple(t *testing.T) {
	input := `[{"name":"web","port":8080},{"name":"api","port":9090}]`
	out := RenderTable(input)

	for _, want := range []string{"NAME", "PORT", "web", "8080", "api", "9090"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got %q", want, out)
		}
	}
}

// Task 4.6: Test RenderTable with empty array
func TestRenderTableEmpty(t *testing.T) {
	out := RenderTable("[]")
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
