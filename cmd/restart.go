package cmd

import (
	"fmt"

	"github.com/jaiir320/devserve/cli"
	"github.com/jaiir320/devserve/client"
	"github.com/jaiir320/devserve/protocol"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart [name]",
	Args:  cobra.ExactArgs(1),
	Short: "Restart a running process",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		var (
			result any
			err    error
		)
		cli.Spin(fmt.Sprintf("Restarting '%s'...", name), func() {
			result, err = client.Restart(name)
		})
		if err != nil {
			return err
		}

		sr, ok := result.(*protocol.ServeResult)
		if !ok {
			return fmt.Errorf("failed to restart '%s': invalid restart result", name)
		}

		fmt.Println(cli.RenderServeResult(sr))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
