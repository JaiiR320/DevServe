package cmd

import (
	"devserve/internal"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve [name] [port] [command]",
	Short: "Serve your dev server with tailscale",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		port := args[1]
		command := args[2]
		return internal.Send("serve|" + name + "|" + port + "|" + command)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
