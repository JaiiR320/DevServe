package cmd

import (
	"devserve/internal"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "devserve [name] [port] [command]",
	Short: "Serve your local projects with Tailscale",
	Long:  `Serve your local dev servers across your Tailscale network with devserve.`,
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		if len(args) != 3 {
			return fmt.Errorf("accepts 3 arg(s), received %d", len(args))
		}
		return runServe(args)
	},
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
