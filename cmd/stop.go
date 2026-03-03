package cmd

import (
	"devserve/cli"
	"devserve/client"
	"fmt"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [name]",
	Args:  cobra.ExactArgs(1),
	Short: "Stop your dev server and tailscale",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cli.Spin("Stopping process...", func() {
			err = client.Stop(args[0])
		})
		if err != nil {
			return fmt.Errorf("failed to stop: %w", err)
		}
		fmt.Println(cli.Success(fmt.Sprintf("process '%s' stopped", args[0])))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
