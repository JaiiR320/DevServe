/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"devserve/internal"
	"fmt"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start your dev server with tailscale",
	Long: `Start your dev server with a port and hostname,
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

		err = internal.Start(port, bg)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {
	startCmd.Flags().IntP("port", "p", 3000, "port to use")
	startCmd.Flags().Bool("bg", false, "run in background")
	rootCmd.AddCommand(startCmd)
}
