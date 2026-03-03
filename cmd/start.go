package cmd

import (
	"devserve/cli"
	"devserve/client"
	"devserve/config"
	"devserve/protocol"
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

	var result *protocol.ServeResult
	cli.Spin(fmt.Sprintf("Starting '%s'...", name), func() {
		result, err = client.Serve(cfg.Name, cfg.Port, cfg.Command, cfg.Directory)
	})

	if err != nil {
		// Check if it's "already running" error
		if errors.Is(err, errors.New("already")) {
			fmt.Println(cli.Info(fmt.Sprintf("process '%s' is already running", name)))
			return nil
		}
		return fmt.Errorf("failed to start: %w", err)
	}

	fmt.Println(cli.RenderServeResult(result))
	return nil
}
