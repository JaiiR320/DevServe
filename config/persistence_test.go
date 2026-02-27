package config

import (
	"path/filepath"
	"testing"
)

func TestSaveAndLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Save a config
	config := ProcessConfig{
		Name:      "myapp",
		Port:      3000,
		Command:   "npm run dev",
		Directory: "/home/user/myapp",
	}

	err := SaveConfig(configPath, config)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Load it back
	configs, err := LoadConfigs(configPath)
	if err != nil {
		t.Fatalf("LoadConfigs failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}

	loaded := configs[0]
	if loaded.Name != config.Name {
		t.Errorf("expected name %q, got %q", config.Name, loaded.Name)
	}
	if loaded.Port != config.Port {
		t.Errorf("expected port %d, got %d", config.Port, loaded.Port)
	}
	if loaded.Command != config.Command {
		t.Errorf("expected command %q, got %q", config.Command, loaded.Command)
	}
	if loaded.Directory != config.Directory {
		t.Errorf("expected directory %q, got %q", config.Directory, loaded.Directory)
	}
}

func TestLoadEmptyConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent.json")

	configs, err := LoadConfigs(configPath)
	if err != nil {
		t.Fatalf("LoadConfigs failed for non-existent file: %v", err)
	}

	if configs == nil {
		t.Error("expected empty slice, got nil")
	}

	if len(configs) != 0 {
		t.Errorf("expected 0 configs, got %d", len(configs))
	}
}

func TestDeleteConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Save two configs
	config1 := ProcessConfig{Name: "app1", Port: 3000, Command: "npm start", Directory: "/app1"}
	config2 := ProcessConfig{Name: "app2", Port: 4000, Command: "npm run dev", Directory: "/app2"}

	if err := SaveConfig(configPath, config1); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}
	if err := SaveConfig(configPath, config2); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Delete the first one
	if err := DeleteConfig(configPath, "app1"); err != nil {
		t.Fatalf("DeleteConfig failed: %v", err)
	}

	// Load and verify only one remains
	configs, err := LoadConfigs(configPath)
	if err != nil {
		t.Fatalf("LoadConfigs failed: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("expected 1 config after delete, got %d", len(configs))
	}

	if configs[0].Name != "app2" {
		t.Errorf("expected remaining config to be 'app2', got %q", configs[0].Name)
	}
}

func TestGetConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Save two configs
	config1 := ProcessConfig{Name: "app1", Port: 3000, Command: "npm start", Directory: "/app1"}
	config2 := ProcessConfig{Name: "app2", Port: 4000, Command: "npm run dev", Directory: "/app2"}

	if err := SaveConfig(configPath, config1); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}
	if err := SaveConfig(configPath, config2); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Get the first one
	loaded, err := GetConfig(configPath, "app1")
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	if loaded == nil {
		t.Fatal("expected config, got nil")
	}

	if loaded.Name != config1.Name {
		t.Errorf("expected name %q, got %q", config1.Name, loaded.Name)
	}
	if loaded.Port != config1.Port {
		t.Errorf("expected port %d, got %d", config1.Port, loaded.Port)
	}

	// Try to get non-existent
	notFound, err := GetConfig(configPath, "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent config, got nil")
	}
	if notFound != nil {
		t.Error("expected nil for non-existent config")
	}
}
