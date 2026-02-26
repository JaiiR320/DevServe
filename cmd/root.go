package cmd

import (
	"devserve/internal"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "devserve",
	Short: "Serve JavaScript projects with Tailscale",
	Long:  `Serve your local dev servers across your Tailscale network with devserve.`,
}

func Execute() {
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	rootCmd.SetHelpTemplate(internal.HelpTemplate())

	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, internal.Error(err.Error()))
		os.Exit(1)
	}
}
