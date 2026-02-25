package cmd

import (
	"devserve/internal"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [name]",
	Args:  cobra.ExactArgs(1),
	Short: "Stop your dev server and tailscale",
	RunE: func(cmd *cobra.Command, args []string) error {
		req := &internal.Request{
			Action: "stop",
			Args: map[string]any{
				"name": args[0],
			},
		}
		resp, err := internal.Send(req)
		if err != nil {
			return err
		}
		if !resp.OK {
			return errors.New(resp.Error)
		}
		fmt.Println(resp.Data)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
