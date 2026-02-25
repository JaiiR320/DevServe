package cmd

import (
	"devserve/internal"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop your dev server and tailscale",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		return internal.Send("stop|" + name)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
