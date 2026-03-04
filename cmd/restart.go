package cmd

import (
	"github.com/jaiir320/devserve/cli"
	"github.com/jaiir320/devserve/client"
	"fmt"

	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart [name]",
	Args:  cobra.ExactArgs(1),
	Short: "Restart a running process",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		var err error
		cli.Spin("Stopping process...", func() {
			err = client.Stop(name)
		})
		if err != nil {
			return fmt.Errorf("failed to stop: %w", err)
		}

		// Start it again from saved config
		return runStart(name)
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
