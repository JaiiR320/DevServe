/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"devserve/internal"
	"fmt"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "server your dev server with tailscale",
	Long: `Serve your dev server with a port and hostname,
	and use tailscale to serve the port. Provides your MagicDNS url
	and logs.`,
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			fmt.Println(err)
		}
		err = internal.Serve(port)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	serveCmd.Flags().Int("port", 3000, "port to use")
	rootCmd.AddCommand(serveCmd)
}
