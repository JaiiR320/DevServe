package cmd

import (
	"devserve/internal"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve [name] [port] [command]",
	Short: "Serve your dev server with tailscale",
	RunE: func(cmd *cobra.Command, args []string) error {
		req := &internal.Request{
			Action: "serve",
			Args: map[string]any{
				"name":    args[0],
				"port":    args[1],
				"command": args[2],
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
	rootCmd.AddCommand(serveCmd)
}
