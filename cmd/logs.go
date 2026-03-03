package cmd

import (
	"devserve/client"
	"devserve/protocol"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs [name]",
	Args:  cobra.ExactArgs(1),
	Short: "Show process logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		lines, _ := cmd.Flags().GetInt("lines")
		req := &protocol.Request{
			Action: "logs",
			Args: map[string]any{
				"name":  args[0],
				"lines": fmt.Sprintf("%d", lines),
			},
		}
		resp, err := client.Send(req)
		if err != nil {
			return fmt.Errorf("failed to send logs request: %w", err)
		}
		if !resp.OK {
			return errors.New(resp.Error)
		}
		fmt.Println(resp.Data)
		return nil
	},
}

func init() {
	logsCmd.Flags().IntP("lines", "n", 50, "number of lines to show")
	rootCmd.AddCommand(logsCmd)
}
