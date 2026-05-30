package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaiir320/devserve/protocol"
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
				restartedName string
			)

			setTestDeps(t,
				func() ([]listItem, error) { return tt.reloadItems, nil },
				func(name string) error { return nil },
				func(item listItem) error {
					return nil
				},
				func(name string) error {
					restartedName = name
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

			if restartedName != tt.item.Name {
				t.Fatalf("restart called with %q, want %q", restartedName, tt.item.Name)
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

func TestModelRestartSelectedStartsStoppedConfiguredItem(t *testing.T) {
	var (
		stopCalls   int
		startedItem listItem
	)

	item := listItem{
		Name:       "api",
		Port:       4000,
		Command:    "go run ./cmd/api",
		Dir:        "/srv/api",
		Configured: true,
	}

	reloadItems := []listItem{
		{Name: "web", Port: 3000, Running: true, Configured: true},
		{Name: "api", Port: 4000, Command: "go run ./cmd/api", Dir: "/srv/api", Running: true, Configured: true},
	}

	setTestDeps(t,
		func() ([]listItem, error) { return reloadItems, nil },
		func(name string) error {
			stopCalls++
			return nil
		},
		func(got listItem) error {
			startedItem = got
			return nil
		},
		func(name string) error { return nil },
	)

	m := model{
		items: []listItem{
			{Name: "web", Port: 3000, Running: true, Configured: true},
			item,
		},
		cursor: 1,
	}

	got, _ := m.restartSelected()

	if stopCalls != 0 {
		t.Fatalf("stop called %d times, want 0", stopCalls)
	}
	if startedItem != item {
		t.Fatalf("start called with %+v, want %+v", startedItem, item)
	}
	if got.statusMsg != "process 'api' started" {
		t.Fatalf("status message = %q, want start success", got.statusMsg)
	}
	if got.statusErr {
		t.Fatal("statusErr = true, want false")
	}
	if got.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", got.cursor)
	}
	if got.items[got.cursor].Name != item.Name {
		t.Fatalf("selected item after reload = %q, want %q", got.items[got.cursor].Name, item.Name)
	}
}

func TestModelRestartSelectedFailure(t *testing.T) {
	setTestDeps(t,
		func() ([]listItem, error) {
			return []listItem{{Name: "web", Running: false}}, nil
		},
		func(name string) error { return nil },
		func(item listItem) error {
			return nil
		},
		func(name string) error {
			return errors.New("failed to start: port 4096 is already in use")
		},
	)

	m := model{items: []listItem{{Name: "web", Running: true}}}

	got, _ := m.restartSelected()

	if got.statusMsg != "failed to restart 'web': failed to start: port 4096 is already in use" {
		t.Fatalf("status message = %q", got.statusMsg)
	}
	if !got.statusErr {
		t.Fatal("statusErr = false, want true")
	}
	if len(got.items) != 1 || got.items[0].Running {
		t.Fatalf("items after failed restart = %+v, want stopped item", got.items)
	}
}

func TestModelRestartSelectedStoppedConfiguredFailure(t *testing.T) {
	setTestDeps(t,
		func() ([]listItem, error) {
			t.Fatal("fetchItems should not be called on start failure")
			return nil, nil
		},
		func(name string) error {
			t.Fatal("stopProcess should not be called for stopped configured item")
			return nil
		},
		func(item listItem) error {
			return errors.New("daemon unavailable")
		},
		func(name string) error { return nil },
	)

	m := model{items: []listItem{{Name: "api", Running: false, Configured: true}}}

	got, _ := m.restartSelected()

	if got.statusMsg != "failed to start 'api': daemon unavailable" {
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
		func(name string) error { return nil },
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

func TestModelUpdateOpenKey(t *testing.T) {
	var openedURL string
	setOpenCopyDeps(t,
		func(url string) error {
			openedURL = url
			return nil
		},
		func(url string) error { return nil },
	)

	m := model{items: []listItem{{Name: "web", Running: true, LocalURL: "http://localhost:3000"}}}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	got, ok := updated.(model)
	if !ok {
		t.Fatalf("updated model type = %T, want tui.model", updated)
	}
	if openedURL != "http://localhost:3000" {
		t.Fatalf("opened URL = %q, want local URL", openedURL)
	}
	if got.statusMsg != "opened http://localhost:3000" {
		t.Fatalf("status message = %q, want open success", got.statusMsg)
	}
	if got.statusErr {
		t.Fatal("statusErr = true, want false")
	}
}

func TestModelOpenSelectedLocalURLFailure(t *testing.T) {
	tests := []struct {
		name       string
		item       listItem
		openErr    error
		wantStatus string
	}{
		{
			name:       "stopped process",
			item:       listItem{Name: "web", Running: false, LocalURL: "http://localhost:3000"},
			wantStatus: "cannot open 'web': process is not running",
		},
		{
			name:       "missing local URL",
			item:       listItem{Name: "web", Running: true},
			wantStatus: "cannot open 'web': no local URL available",
		},
		{
			name:       "opener failure",
			item:       listItem{Name: "web", Running: true, LocalURL: "http://localhost:3000"},
			openErr:    errors.New("browser unavailable"),
			wantStatus: "failed to open 'web': browser unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			openCalls := 0
			setOpenCopyDeps(t,
				func(url string) error {
					openCalls++
					return tt.openErr
				},
				func(url string) error { return nil },
			)

			got, _ := (model{items: []listItem{tt.item}}).openSelectedLocalURL()

			if got.statusMsg != tt.wantStatus {
				t.Fatalf("status message = %q, want %q", got.statusMsg, tt.wantStatus)
			}
			if !got.statusErr {
				t.Fatal("statusErr = false, want true")
			}
			if !tt.item.Running || tt.item.LocalURL == "" {
				if openCalls != 0 {
					t.Fatalf("open called %d times, want 0", openCalls)
				}
			} else if openCalls != 1 {
				t.Fatalf("open called %d times, want 1", openCalls)
			}
		})
	}
}

func TestModelUpdateCopyKey(t *testing.T) {
	var copiedURL string
	setOpenCopyDeps(t,
		func(url string) error { return nil },
		func(url string) error {
			copiedURL = url
			return nil
		},
	)

	m := model{items: []listItem{{
		Name:    "web",
		Running: true,
		IPURL:   "http://100.64.0.1:3000",
		DNSURL:  "https://host.tailnet.ts.net:3000",
	}}}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	got, ok := updated.(model)
	if !ok {
		t.Fatalf("updated model type = %T, want tui.model", updated)
	}
	if copiedURL != "https://host.tailnet.ts.net:3000" {
		t.Fatalf("copied URL = %q, want DNS URL", copiedURL)
	}
	if got.statusMsg != "copied https://host.tailnet.ts.net:3000" {
		t.Fatalf("status message = %q, want copy success", got.statusMsg)
	}
	if got.statusErr {
		t.Fatal("statusErr = true, want false")
	}
}

func TestModelCopySelectedTailscaleURLFallbackAndFailure(t *testing.T) {
	tests := []struct {
		name       string
		item       listItem
		copyErr    error
		wantCopied string
		wantStatus string
		wantErr    bool
	}{
		{
			name:       "falls back to IP URL",
			item:       listItem{Name: "web", Running: true, IPURL: "http://100.64.0.1:3000"},
			wantCopied: "http://100.64.0.1:3000",
			wantStatus: "copied http://100.64.0.1:3000",
		},
		{
			name:       "stopped process",
			item:       listItem{Name: "web", Running: false, DNSURL: "https://host.tailnet.ts.net:3000"},
			wantStatus: "cannot copy URL for 'web': process is not running",
			wantErr:    true,
		},
		{
			name:       "missing tailscale URL",
			item:       listItem{Name: "web", Running: true},
			wantStatus: "cannot copy URL for 'web': no Tailscale URL available",
			wantErr:    true,
		},
		{
			name:       "copy failure",
			item:       listItem{Name: "web", Running: true, DNSURL: "https://host.tailnet.ts.net:3000"},
			copyErr:    errors.New("clipboard unavailable"),
			wantCopied: "https://host.tailnet.ts.net:3000",
			wantStatus: "failed to copy URL for 'web': clipboard unavailable",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var copiedURL string
			setOpenCopyDeps(t,
				func(url string) error { return nil },
				func(url string) error {
					copiedURL = url
					return tt.copyErr
				},
			)

			got, _ := (model{items: []listItem{tt.item}}).copySelectedTailscaleURL()

			if copiedURL != tt.wantCopied {
				t.Fatalf("copied URL = %q, want %q", copiedURL, tt.wantCopied)
			}
			if got.statusMsg != tt.wantStatus {
				t.Fatalf("status message = %q, want %q", got.statusMsg, tt.wantStatus)
			}
			if got.statusErr != tt.wantErr {
				t.Fatalf("statusErr = %v, want %v", got.statusErr, tt.wantErr)
			}
		})
	}
}

func TestSelectedTailscaleURL(t *testing.T) {
	tests := []struct {
		name string
		item listItem
		want string
	}{
		{
			name: "prefers DNS URL",
			item: listItem{IPURL: "http://100.64.0.1:3000", DNSURL: "https://host.tailnet.ts.net:3000"},
			want: "https://host.tailnet.ts.net:3000",
		},
		{
			name: "falls back to IP URL",
			item: listItem{IPURL: "http://100.64.0.1:3000"},
			want: "http://100.64.0.1:3000",
		},
		{
			name: "empty without tailscale URL",
			item: listItem{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectedTailscaleURL(tt.item); got != tt.want {
				t.Fatalf("selected URL = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderHelpIncludesRestart(t *testing.T) {
	help := renderHelp()
	if !strings.Contains(help, "r restart") {
		t.Fatalf("help text = %q, want restart hotkey", help)
	}
}

func TestRenderHelpIncludesOpenAndCopy(t *testing.T) {
	help := renderHelp()
	if !strings.Contains(help, "o open") {
		t.Fatalf("help text = %q, want open hotkey", help)
	}
	if !strings.Contains(help, "c copy") {
		t.Fatalf("help text = %q, want copy hotkey", help)
	}
}

func setTestDeps(t *testing.T, fetch func() ([]listItem, error), stop func(string) error, start func(listItem) error, restart func(string) error) {
	t.Helper()

	oldFetch := fetchItemsFunc
	oldStop := stopProcessFunc
	oldStart := startItemFunc
	oldRestart := restartFunc

	fetchItemsFunc = fetch
	stopProcessFunc = stop
	startItemFunc = start
	restartFunc = func(name string) (*protocol.ServeResult, error) {
		if err := restart(name); err != nil {
			return nil, err
		}
		return &protocol.ServeResult{Name: name}, nil
	}

	t.Cleanup(func() {
		fetchItemsFunc = oldFetch
		stopProcessFunc = oldStop
		startItemFunc = oldStart
		restartFunc = oldRestart
	})
}

func setOpenCopyDeps(t *testing.T, open func(string) error, copy func(string) error) {
	t.Helper()

	oldOpen := openURLFunc
	oldCopy := copyURLFunc

	openURLFunc = open
	copyURLFunc = copy

	t.Cleanup(func() {
		openURLFunc = oldOpen
		copyURLFunc = oldCopy
	})
}
