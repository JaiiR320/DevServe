package cmd

import (
	"devserve/cli"
	"devserve/client"
	"devserve/config"
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

func runConfigSave(name string) error {
	// Query daemon for process info
	info, err := client.Get(name)
	if err != nil {
		return fmt.Errorf("failed to query process: %w", err)
	}

	// Extract process details
	cfg := config.ProcessConfig{
		Name:      name,
		Port:      info.Port,
		Command:   info.Command,
		Directory: info.Dir,
	}

	if err := config.SaveConfig(config.ConfigFile, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println(cli.Success(fmt.Sprintf("config '%s' saved", name)))
	return nil
}
