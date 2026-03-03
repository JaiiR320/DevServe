package cmd

import (
	"devserve/cli"
	"devserve/daemon"
	"devserve/protocol"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart [name]",
	Args:  cobra.ExactArgs(1),
	Short: "Restart a running process",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Stop the process
		req := &protocol.Request{
			Action: "stop",
			Args: map[string]any{
				"name": name,
			},
		}
		var (
			resp *protocol.Response
			err  error
		)
		cli.Spin("Stopping process...", func() {
			resp, err = daemon.Send(req)
		})
		if err != nil {
			return fmt.Errorf("failed to send stop request: %w", err)
		}
		if !resp.OK {
			return errors.New(resp.Error)
		}

		// Start it again from saved config
		return runStart(name)
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
