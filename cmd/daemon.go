/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"devserve/daemon"
	"devserve/internal"

	"github.com/spf13/cobra"
)

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Interact with the daemon",
}

var daemonCmdStart = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := internal.InitLogger(); err != nil {
			return err
		}
		return daemon.Start()
	},
}

var daemonCmdStop = &cobra.Command{
	Use:   "stop",
	Short: "Stop the daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.Stop()
	},
}

func init() {
	daemonCmd.AddCommand(daemonCmdStart, daemonCmdStop)
	rootCmd.AddCommand(daemonCmd)
}
