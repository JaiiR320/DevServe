/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"devserve/internal"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a process based on a port",
	Long:  `Stop a process based on a port`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		portStr := args[0]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			fmt.Println("Invalid port number")
			return
		}
		p, err := internal.GetProcessByPort(port)
		if err != nil {
			fmt.Printf("could not find a process with that port: %v\n", err)
			return
		}
		fmt.Println("Stopping dev server on port ", portStr)
		p.Stop()

		// remove files
		err = internal.RemoveProcess(port)
		if err != nil {
			fmt.Printf("could not remove process: %v\n", err)
		}

		tm := internal.NewTailscaleManager(os.Stdout, os.Stderr)
		err = tm.Stop(port)
		if err != nil {
			fmt.Printf("could not stop tailscale: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
