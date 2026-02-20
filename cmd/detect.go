/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"devserve/internal"
	"fmt"

	"github.com/spf13/cobra"
)

// detectCmd represents the detect command
var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect your package manager",
	Long: `Find which package lock file you use and infer the package manager.
	For example, if you have bun.lock, we assume you use bun as your package manager. `,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := internal.DetectPackageManager()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(c + " detected")
	},
}

func init() {
	rootCmd.AddCommand(detectCmd)
}
