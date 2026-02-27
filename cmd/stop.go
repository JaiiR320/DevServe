package cmd

import (
	"devserve/cli"
	"devserve/daemon"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [name]",
	Args:  cobra.ExactArgs(1),
	Short: "Stop your dev server and tailscale",
	RunE: func(cmd *cobra.Command, args []string) error {
		req := &daemon.Request{
			Action: "stop",
			Args: map[string]any{
				"name": args[0],
			},
		}
		var (
			resp *daemon.Response
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
		fmt.Println(cli.Success(resp.Data))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
