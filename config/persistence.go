package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ProcessConfig represents a saved process configuration
type ProcessConfig struct {
	Name      string `json:"name"`
	Port      int    `json:"port"`
	Command   string `json:"command"`
	Directory string `json:"directory"`
}

// LoadConfigs loads all saved process configurations from the config file
func LoadConfigs(configPath string) ([]ProcessConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []ProcessConfig{}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var configs []ProcessConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return configs, nil
}

// SaveConfig saves a process configuration to the config file
func SaveConfig(configPath string, config ProcessConfig) error {
	configs, err := LoadConfigs(configPath)
	if err != nil {
		return err
	}

	// Check if config already exists, replace it
	found := false
	for i, c := range configs {
		if c.Name == config.Name {
			configs[i] = config
			found = true
			break
		}
	}

	if !found {
		configs = append(configs, config)
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, DirPermissions); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, DirPermissions); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfig retrieves a single process configuration by name
func GetConfig(configPath string, name string) (*ProcessConfig, error) {
	configs, err := LoadConfigs(configPath)
	if err != nil {
		return nil, err
	}

	for _, c := range configs {
		if c.Name == name {
			return &c, nil
		}
	}

	return nil, fmt.Errorf("config '%s' not found", name)
}

// DeleteConfig removes a process configuration from the config file
func DeleteConfig(configPath string, name string) error {
	configs, err := LoadConfigs(configPath)
	if err != nil {
		return err
	}

	// Find and remove the config with the given name
	found := false
	newConfigs := make([]ProcessConfig, 0, len(configs))
	for _, c := range configs {
		if c.Name != name {
			newConfigs = append(newConfigs, c)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("config '%s' not found", name)
	}

	// If no configs remain, delete the file
	if len(newConfigs) == 0 {
		if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove config file: %w", err)
		}
		return nil
	}

	data, err := json.MarshalIndent(newConfigs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, DirPermissions); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
