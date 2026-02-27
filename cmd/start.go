package cmd

import (
	"devserve/cli"
	"devserve/config"
	"devserve/daemon"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [name]",
	Args:  cobra.ExactArgs(1),
	Short: "Start a process from saved configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStart(args[0])
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func runStart(name string) error {
	// Load the config
	cfg, err := config.GetConfig(config.ConfigFile, name)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Send serve request to daemon
	req := &daemon.Request{
		Action: "serve",
		Args: map[string]any{
			"name":    cfg.Name,
			"port":    cfg.Port,
			"command": cfg.Command,
			"cwd":     cfg.Directory,
		},
	}

	var resp *daemon.Response
	cli.Spin(fmt.Sprintf("Starting '%s'...", name), func() {
		resp, err = sendRequest(req)
	})

	if err != nil {
		return fmt.Errorf("failed to send start request: %w", err)
	}
	if !resp.OK {
		// Check if it's "already running" error
		if errors.Is(errors.New(resp.Error), errors.New("already")) {
			fmt.Println(cli.Info(fmt.Sprintf("process '%s' is already running", name)))
			return nil
		}
		return errors.New(resp.Error)
	}

	fmt.Println(cli.Success(resp.Data))
	return nil
}
