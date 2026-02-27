package cmd

import (
	"devserve/config"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigList(t *testing.T) {
	// Create a temp config file with test data
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Save test configs directly to file
	configs := []config.ProcessConfig{
		{Name: "app1", Port: 3000, Command: "npm start", Directory: "/app1"},
		{Name: "app2", Port: 4000, Command: "npm run dev", Directory: "/app2"},
	}
	data, _ := json.MarshalIndent(configs, "", "  ")
	os.WriteFile(configPath, data, 0644)

	// Call the function to list configs
	output, err := runConfigList(configPath)
	if err != nil {
		t.Fatalf("runConfigList failed: %v", err)
	}

	// Verify output contains both app names
	if !strings.Contains(output, "app1") {
		t.Errorf("expected output to contain 'app1', got: %s", output)
	}
	if !strings.Contains(output, "app2") {
		t.Errorf("expected output to contain 'app2', got: %s", output)
	}
	if !strings.Contains(output, "3000") {
		t.Errorf("expected output to contain port '3000', got: %s", output)
	}
	if !strings.Contains(output, "4000") {
		t.Errorf("expected output to contain port '4000', got: %s", output)
	}
}

func TestConfigListEmpty(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent.json")

	// Call the function with non-existent config
	output, err := runConfigList(configPath)
	if err != nil {
		t.Fatalf("runConfigList failed: %v", err)
	}

	// Should show empty or "no configs" message
	if !strings.Contains(output, "No saved configurations") && output != "" {
		t.Errorf("expected 'No saved configurations' or empty output, got: %s", output)
	}
}

func TestConfigPersistence(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Save a config directly using the config package
	cfg := config.ProcessConfig{
		Name:      "myapp",
		Port:      3000,
		Command:   "npm run dev",
		Directory: "/home/user/myapp",
	}

	err := config.SaveConfig(configPath, cfg)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify the config was saved
	configs, err := config.LoadConfigs(configPath)
	if err != nil {
		t.Fatalf("failed to load saved configs: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}

	if configs[0].Name != "myapp" {
		t.Errorf("expected name 'myapp', got %q", configs[0].Name)
	}
	if configs[0].Port != 3000 {
		t.Errorf("expected port 3000, got %d", configs[0].Port)
	}
	if configs[0].Command != "npm run dev" {
		t.Errorf("expected command 'npm run dev', got %q", configs[0].Command)
	}
	if configs[0].Directory != "/home/user/myapp" {
		t.Errorf("expected directory '/home/user/myapp', got %q", configs[0].Directory)
	}
}
