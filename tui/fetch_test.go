package tui

import (
	"devserve/config"
	"devserve/protocol"
	"testing"
)

func TestBuildItemsEmpty(t *testing.T) {
	items := buildItems(nil, "", "", nil)
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestBuildItemsOnlyConfigured(t *testing.T) {
	configs := []config.ProcessConfig{
		{Name: "web", Port: 3000, Command: "npm start", Directory: "/projects/web"},
		{Name: "api", Port: 4000, Command: "go run .", Directory: "/projects/api"},
	}

	items := buildItems(nil, "", "", configs)

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// Should be sorted by port
	if items[0].Name != "web" || items[0].Port != 3000 {
		t.Errorf("expected first item web:3000, got %s:%d", items[0].Name, items[0].Port)
	}
	if items[1].Name != "api" || items[1].Port != 4000 {
		t.Errorf("expected second item api:4000, got %s:%d", items[1].Name, items[1].Port)
	}

	// Should be marked as configured but not running
	for _, item := range items {
		if !item.Configured {
			t.Errorf("expected %s to be configured", item.Name)
		}
		if item.Running {
			t.Errorf("expected %s to not be running", item.Name)
		}
	}
}

func TestBuildItemsOnlyEphemeral(t *testing.T) {
	processes := []protocol.ListEntry{
		{Name: "web", Port: 3000, Command: "npm start", Dir: "/projects/web"},
		{Name: "api", Port: 4000, Command: "go run .", Dir: "/projects/api"},
	}

	items := buildItems(processes, "host.ts.net", "100.1.2.3", nil)

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// Should be sorted by port (ephemeral section)
	if items[0].Name != "web" || items[0].Port != 3000 {
		t.Errorf("expected first item web:3000, got %s:%d", items[0].Name, items[0].Port)
	}
	if items[1].Name != "api" || items[1].Port != 4000 {
		t.Errorf("expected second item api:4000, got %s:%d", items[1].Name, items[1].Port)
	}

	// Should be marked as running but not configured
	for _, item := range items {
		if item.Configured {
			t.Errorf("expected %s to not be configured", item.Name)
		}
		if !item.Running {
			t.Errorf("expected %s to be running", item.Name)
		}
		// URLs should be populated
		if item.LocalURL == "" {
			t.Errorf("expected %s to have LocalURL", item.Name)
		}
		if item.IPURL == "" {
			t.Errorf("expected %s to have IPURL", item.Name)
		}
		if item.DNSURL == "" {
			t.Errorf("expected %s to have DNSURL", item.Name)
		}
	}
}

func TestBuildItemsMixed(t *testing.T) {
	processes := []protocol.ListEntry{
		{Name: "web", Port: 3000, Command: "npm start", Dir: "/projects/web"},
		{Name: "api", Port: 4000, Command: "go run .", Dir: "/projects/api"},
	}
	configs := []config.ProcessConfig{
		{Name: "web", Port: 3000, Command: "npm start", Directory: "/projects/web"},              // Running and configured
		{Name: "worker", Port: 5000, Command: "python worker.py", Directory: "/projects/worker"}, // Not running
	}

	items := buildItems(processes, "host.ts.net", "100.1.2.3", configs)

	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	// Configured items come first, sorted by port
	if items[0].Name != "web" || items[0].Port != 3000 {
		t.Errorf("expected first item web:3000, got %s:%d", items[0].Name, items[0].Port)
	}
	if !items[0].Configured {
		t.Error("expected web to be configured")
	}
	if !items[0].Running {
		t.Error("expected web to be running (both configured and ephemeral)")
	}

	if items[1].Name != "worker" || items[1].Port != 5000 {
		t.Errorf("expected second item worker:5000, got %s:%d", items[1].Name, items[1].Port)
	}
	if !items[1].Configured {
		t.Error("expected worker to be configured")
	}
	if items[1].Running {
		t.Error("expected worker to not be running")
	}

	// Ephemeral items come last
	if items[2].Name != "api" || items[2].Port != 4000 {
		t.Errorf("expected third item api:4000, got %s:%d", items[2].Name, items[2].Port)
	}
	if items[2].Configured {
		t.Error("expected api to not be configured (ephemeral)")
	}
	if !items[2].Running {
		t.Error("expected api to be running")
	}
}

func TestBuildItemsCommandAndDirOverride(t *testing.T) {
	// Config has different command/dir than running process
	processes := []protocol.ListEntry{
		{Name: "app", Port: 3000, Command: "npm run dev", Dir: "/live/app"},
	}
	configs := []config.ProcessConfig{
		{Name: "app", Port: 3000, Command: "npm start", Directory: "/config/app"},
	}

	items := buildItems(processes, "", "", configs)

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	item := items[0]
	// Running process data should override config data
	if item.Command != "npm run dev" {
		t.Errorf("expected Command from running process 'npm run dev', got %q", item.Command)
	}
	if item.Dir != "/live/app" {
		t.Errorf("expected Dir from running process '/live/app', got %q", item.Dir)
	}
}

func TestBuildItemsNoHostname(t *testing.T) {
	processes := []protocol.ListEntry{
		{Name: "app", Port: 3000, Command: "npm start", Dir: "/projects/app"},
	}

	// No hostname or IP provided
	items := buildItems(processes, "", "", nil)

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	item := items[0]
	if item.LocalURL == "" {
		t.Error("expected LocalURL to be set")
	}
	if item.IPURL != "" {
		t.Error("expected IPURL to be empty when no IP provided")
	}
	if item.DNSURL != "" {
		t.Error("expected DNSURL to be empty when no hostname provided")
	}
}
