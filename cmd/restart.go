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

		// Capture live process info before stopping
		info, err := client.Get(name)
		if err != nil {
			return fmt.Errorf("failed to query process '%s': %w", name, err)
		}

		cli.Spin("Stopping process...", func() {
			err = client.Stop(name)
		})
		if err != nil {
			return fmt.Errorf("failed to stop: %w", err)
		}

		// Re-serve using captured live state
		var result *protocol.ServeResult
		cli.Spin(fmt.Sprintf("Starting '%s'...", name), func() {
			result, err = client.Serve(info.Name, info.Port, info.Command, info.Dir)
		})
		if err != nil {
			return fmt.Errorf("failed to start: %w", err)
		}

		fmt.Println(cli.RenderServeResult(result))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
