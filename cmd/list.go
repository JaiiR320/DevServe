package cmd

import (
	"github.com/jaiir320/devserve/cli"
	"github.com/jaiir320/devserve/client"
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Args:  cobra.NoArgs,
	Short: "List processes",
	RunE: func(cmd *cobra.Command, args []string) error {
		lr, err := client.List()
		if err != nil {
			return fmt.Errorf("failed to list: %w", err)
		}
		fmt.Println(cli.RenderTable(lr))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
