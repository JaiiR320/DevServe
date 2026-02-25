package cmd

import (
	"devserve/internal"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List processes",
	RunE: func(cmd *cobra.Command, args []string) error {
		return internal.Send("list")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
