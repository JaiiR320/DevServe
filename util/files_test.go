package util_test

import (
	"devserve/util"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLastNLinesMoreThanN(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.log")
	lines := make([]string, 10)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d", i+1)
	}
	os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644)

	got, err := util.LastNLines(path, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(got))
	}
	for i, want := range []string{"line 6", "line 7", "line 8", "line 9", "line 10"} {
		if got[i] != want {
			t.Errorf("line %d: expected %q, got %q", i, want, got[i])
		}
	}
}

func TestLastNLinesFewerThanN(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.log")
	os.WriteFile(path, []byte("line 1\nline 2\nline 3\n"), 0644)

	got, err := util.LastNLines(path, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(got))
	}
	for i, want := range []string{"line 1", "line 2", "line 3"} {
		if got[i] != want {
			t.Errorf("line %d: expected %q, got %q", i, want, got[i])
		}
	}
}

func TestLastNLinesExactlyN(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.log")
	os.WriteFile(path, []byte("a\nb\nc\nd\ne\n"), 0644)

	got, err := util.LastNLines(path, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(got))
	}
	for i, want := range []string{"a", "b", "c", "d", "e"} {
		if got[i] != want {
			t.Errorf("line %d: expected %q, got %q", i, want, got[i])
		}
	}
}

func TestLastNLinesEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.log")
	os.WriteFile(path, []byte(""), 0644)

	got, err := util.LastNLines(path, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestLastNLinesNonexistentFile(t *testing.T) {
	got, err := util.LastNLines("/tmp/does-not-exist-ever-12345.log", 5)
	if err != nil {
		t.Fatalf("expected nil error for nonexistent file, got %v", err)
	}
	if got != nil {
		t.Errorf("expected nil result for nonexistent file, got %v", got)
	}
}

func TestLastNLinesTrailingNewlines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.log")
	os.WriteFile(path, []byte("line 1\nline 2\n\n\n"), 0644)

	got, err := util.LastNLines(path, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Trailing newlines are trimmed, so no empty trailing element
	last := got[len(got)-1]
	if last == "" {
		t.Errorf("expected no empty trailing line, but got one")
	}
}

func TestLastNLinesSingleLineNoNewline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.log")
	os.WriteFile(path, []byte("only line"), 0644)

	got, err := util.LastNLines(path, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 line, got %d", len(got))
	}
	if got[0] != "only line" {
		t.Errorf("expected %q, got %q", "only line", got[0])
	}
}
