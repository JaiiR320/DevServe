package cmd

import (
	"devserve/cli"
	"devserve/internal"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Args:  cobra.NoArgs,
	Short: "List processes",
	RunE: func(cmd *cobra.Command, args []string) error {
		req := &internal.Request{
			Action: "list",
		}
		resp, err := internal.Send(req)
		if err != nil {
			return fmt.Errorf("failed to send list request: %w", err)
		}
		if !resp.OK {
			return errors.New(resp.Error)
		}
		fmt.Println(cli.RenderTable(resp.Data))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
