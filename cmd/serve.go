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
			return
		}

		bg, err := cmd.Flags().GetBool("bg")
		if err != nil {
			fmt.Println(err)
			return
		}

		err = internal.Serve(port, bg)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {
	serveCmd.Flags().IntP("port", "p", 3000, "port to use")
	serveCmd.Flags().Bool("bg", false, "run in background")
	rootCmd.AddCommand(serveCmd)
}
