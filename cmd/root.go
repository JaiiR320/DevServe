package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "devserve",
	Short: "Serve JavaScript projects with Tailscale",
	Long:  `Serve your local dev servers across your Tailscale network with devserve.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
