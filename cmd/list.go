/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"devserve/internal"
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list opened ports and their processes",
	Long:  `List opened ports nad their processes`,
	Run: func(cmd *cobra.Command, args []string) {
		ports, err := internal.ListProcesses()
		if err != nil {
			fmt.Println(err)
		}
		for _, p := range ports {
			fmt.Println(p)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
