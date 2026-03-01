package cmd

import (
	"devserve/cli"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "devserve",
	Short: "Serve your local projects with Tailscale",
	Long:  `Serve your local dev servers across your Tailscale network with devserve.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("TUI coming soon...")
		return nil
	},
}

func Execute() {
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	rootCmd.SetHelpTemplate(cli.HelpTemplate())

	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, cli.Error(err.Error()))
		os.Exit(1)
	}
}
