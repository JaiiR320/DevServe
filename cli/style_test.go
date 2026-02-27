package cli_test

import (
	"devserve/cli"
	"strings"
	"testing"
)

func TestSuccess(t *testing.T) {
	out := cli.Success("process started")
	if !strings.Contains(out, "✓") {
		t.Errorf("expected output to contain '✓', got %q", out)
	}
	if !strings.Contains(out, "process started") {
		t.Errorf("expected output to contain message, got %q", out)
	}
}

func TestError(t *testing.T) {
	out := cli.Error("connection failed")
	if !strings.Contains(out, "✗") {
		t.Errorf("expected output to contain '✗', got %q", out)
	}
	if !strings.Contains(out, "connection failed") {
		t.Errorf("expected output to contain message, got %q", out)
	}
}

func TestInfo(t *testing.T) {
	out := cli.Info("checking status")
	if !strings.Contains(out, "•") {
		t.Errorf("expected output to contain '•', got %q", out)
	}
	if !strings.Contains(out, "checking status") {
		t.Errorf("expected output to contain message, got %q", out)
	}
}

func TestRenderTableValid(t *testing.T) {
	input := `{"processes":[{"name":"app","port":3000}],"hostname":"host.example.ts.net","ip":"100.1.2.3"}`
	out := cli.RenderTable(input)

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

func TestRenderTableMultiple(t *testing.T) {
	input := `{"processes":[{"name":"web","port":8080},{"name":"api","port":9090}],"hostname":"host.example.ts.net","ip":"100.1.2.3"}`
	out := cli.RenderTable(input)

	for _, want := range []string{"NAME", "PORT", "web", "8080", "api", "9090"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got %q", want, out)
		}
	}
}

func TestRenderTableEmpty(t *testing.T) {
	out := cli.RenderTable(`{"processes":[],"hostname":"host.example.ts.net","ip":"100.1.2.3"}`)
	if !strings.Contains(out, "No active processes") {
		t.Errorf("expected output to contain %q, got %q", "No active processes", out)
	}
}

func TestRenderTableInvalidJSON(t *testing.T) {
	input := "this is not json"
	out := cli.RenderTable(input)
	if out != input {
		t.Errorf("expected raw input %q returned as fallback, got %q", input, out)
	}
}

func TestRenderTableHyperlinks(t *testing.T) {
	input := `{"processes":[{"name":"app","port":3000}],"hostname":"host.example.ts.net","ip":"100.1.2.3"}`
	out := cli.RenderTable(input)

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

func TestHyperlink(t *testing.T) {
	url := "https://example.com"
	label := "click"
	out := cli.Hyperlink(url, label)

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

func TestHelpTemplate(t *testing.T) {
	tmpl := cli.HelpTemplate()
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
