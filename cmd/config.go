package cmd

import (
	"devserve/cli"
	"devserve/config"
	"devserve/daemon"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage saved process configurations",
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved configurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		configs, err := config.LoadConfigs(config.ConfigFile)
		if err != nil {
			return fmt.Errorf("failed to load configs: %w", err)
		}
		fmt.Println(cli.RenderConfigTable(configs))
		return nil
	},
}

var configSaveCmd = &cobra.Command{
	Use:   "save [name]",
	Args:  cobra.ExactArgs(1),
	Short: "Save a running process configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigSave(args[0])
	},
}

var configDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Args:  cobra.ExactArgs(1),
	Short: "Delete a saved configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		err := config.DeleteConfig(config.ConfigFile, name)
		if err != nil {
			return fmt.Errorf("failed to delete config: %w", err)
		}
		fmt.Println(cli.Success(fmt.Sprintf("config '%s' deleted", name)))
		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit [name]",
	Args:  cobra.ExactArgs(1),
	Short: "Edit a saved configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("edit not implemented")
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configSaveCmd)
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configEditCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigList(configPath string) (string, error) {
	configs, err := config.LoadConfigs(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to load configs: %w", err)
	}

	if len(configs) == 0 {
		return "No saved configurations", nil
	}

	// Format as JSON for rendering
	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal configs: %w", err)
	}

	return string(data), nil
}

func runConfigSave(name string) error {
	// Query daemon for process info
	req := &daemon.Request{
		Action: "get",
		Args: map[string]any{
			"name": name,
		},
	}

	resp, err := sendRequest(req)
	if err != nil {
		return fmt.Errorf("failed to query process: %w", err)
	}
	if !resp.OK {
		return errors.New(resp.Error)
	}

	// Parse the response
	var data map[string]any
	if err := json.Unmarshal([]byte(resp.Data), &data); err != nil {
		return fmt.Errorf("failed to parse process info: %w", err)
	}

	// Extract process details
	cfg := config.ProcessConfig{
		Name:      name,
		Directory: data["dir"].(string),
	}

	// Handle port type
	switch v := data["port"].(type) {
	case float64:
		cfg.Port = int(v)
	case int:
		cfg.Port = v
	default:
		return fmt.Errorf("invalid port type in process info")
	}

	// Command is stored in the Process struct but not currently tracked
	// For now, we'll need to get it another way or store it differently
	// This is a limitation of the current architecture
	cfg.Command = data["command"].(string)

	if err := config.SaveConfig(config.ConfigFile, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println(cli.Success(fmt.Sprintf("config '%s' saved", name)))
	return nil
}
