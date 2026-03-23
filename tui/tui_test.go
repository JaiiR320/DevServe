package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelRestartSelectedRunningItems(t *testing.T) {
	tests := []struct {
		name        string
		item        listItem
		reloadItems []listItem
		wantCursor  int
	}{
		{
			name: "running saved process",
			item: listItem{
				Name:       "web",
				Port:       3000,
				Command:    "npm run dev",
				Dir:        "/srv/web",
				Running:    true,
				Configured: true,
			},
			reloadItems: []listItem{
				{Name: "api", Port: 4000, Running: true, Configured: true},
				{Name: "web", Port: 3000, Running: true, Configured: true},
			},
			wantCursor: 1,
		},
		{
			name: "running unsaved process",
			item: listItem{
				Name:       "scratch",
				Port:       5173,
				Command:    "bun dev",
				Dir:        "/tmp/scratch",
				Running:    true,
				Configured: false,
			},
			reloadItems: []listItem{
				{Name: "scratch", Port: 5173, Command: "bun dev", Dir: "/tmp/scratch", Running: true},
			},
			wantCursor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				stoppedName string
				startedItem listItem
			)

			setTestDeps(t,
				func() ([]listItem, error) { return tt.reloadItems, nil },
				func(name string) error {
					stoppedName = name
					return nil
				},
				func(item listItem) error {
					startedItem = item
					return nil
				},
			)

			m := model{
				items: []listItem{
					tt.item,
					{Name: "other", Port: 9000, Running: true, Configured: true},
				},
			}

			got, _ := m.restartSelected()

			if stoppedName != tt.item.Name {
				t.Fatalf("stop called with %q, want %q", stoppedName, tt.item.Name)
			}
			if startedItem != tt.item {
				t.Fatalf("start called with %+v, want %+v", startedItem, tt.item)
			}
			if got.statusMsg != "process '"+tt.item.Name+"' restarted" {
				t.Fatalf("status message = %q, want restart success", got.statusMsg)
			}
			if got.statusErr {
				t.Fatal("statusErr = true, want false")
			}
			if got.cursor != tt.wantCursor {
				t.Fatalf("cursor = %d, want %d", got.cursor, tt.wantCursor)
			}
			if len(got.items) != len(tt.reloadItems) {
				t.Fatalf("got %d reloaded items, want %d", len(got.items), len(tt.reloadItems))
			}
			if got.items[got.cursor].Name != tt.item.Name {
				t.Fatalf("selected item after reload = %q, want %q", got.items[got.cursor].Name, tt.item.Name)
			}
		})
	}
}

func TestModelRestartSelectedFailure(t *testing.T) {
	setTestDeps(t,
		func() ([]listItem, error) {
			t.Fatal("fetchItems should not be called on restart failure")
			return nil, nil
		},
		func(name string) error {
			return errors.New("daemon unavailable")
		},
		func(item listItem) error {
			t.Fatal("startItem should not be called when stop fails")
			return nil
		},
	)

	m := model{items: []listItem{{Name: "web", Running: true}}}

	got, _ := m.restartSelected()

	if got.statusMsg != "failed to restart 'web': stop: daemon unavailable" {
		t.Fatalf("status message = %q", got.statusMsg)
	}
	if !got.statusErr {
		t.Fatal("statusErr = false, want true")
	}
}

func TestModelUpdateRestartKey(t *testing.T) {
	setTestDeps(t,
		func() ([]listItem, error) {
			return []listItem{{Name: "web", Running: true}}, nil
		},
		func(name string) error { return nil },
		func(item listItem) error { return nil },
	)

	m := model{items: []listItem{{Name: "web", Running: true}}}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	got, ok := updated.(model)
	if !ok {
		t.Fatalf("updated model type = %T, want tui.model", updated)
	}
	if got.statusMsg != "process 'web' restarted" {
		t.Fatalf("status message = %q, want restart success", got.statusMsg)
	}
}

func TestRenderHelpIncludesRestart(t *testing.T) {
	help := renderHelp()
	if !strings.Contains(help, "r restart") {
		t.Fatalf("help text = %q, want restart hotkey", help)
	}
}

func setTestDeps(t *testing.T, fetch func() ([]listItem, error), stop func(string) error, start func(listItem) error) {
	t.Helper()

	oldFetch := fetchItemsFunc
	oldStop := stopProcessFunc
	oldStart := startItemFunc

	fetchItemsFunc = fetch
	stopProcessFunc = stop
	startItemFunc = start

	t.Cleanup(func() {
		fetchItemsFunc = oldFetch
		stopProcessFunc = oldStop
		startItemFunc = oldStart
	})
}
