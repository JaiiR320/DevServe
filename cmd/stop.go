/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"devserve/internal"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a process based on a port",
	Long:  `Stop a process based on a port`,
	Run: func(cmd *cobra.Command, args []string) {
		portStr := args[0]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			fmt.Println("Invalid port number")
			return
		}
		p, err := internal.GetProcessByPort(port)
		if err != nil {
			fmt.Printf("Could not find a process with that port, %s", err)
			return
		}
		p.Stop()

		// remove files
		err = internal.RemoveProcess(port)
		if err != nil {
			fmt.Printf("Could not remove process: %s", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
