package cmd

import (
	"devserve/cli"
	"devserve/daemon"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve [name] [port] [command]",
	Args:  cobra.ExactArgs(3),
	Short: "Serve your dev server with tailscale",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServe(args)
	},
}

func runServe(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	req := &daemon.Request{
		Action: "serve",
		Args: map[string]any{
			"name":    args[0],
			"port":    args[1],
			"command": args[2],
			"cwd":     cwd,
		},
	}
	var resp *daemon.Response
	cli.Spin("Starting process...", func() {
		resp, err = sendRequest(req)
	})
	if err != nil {
		return fmt.Errorf("failed to send serve request: %w", err)
	}
	if !resp.OK {
		return errors.New(resp.Error)
	}
	fmt.Println(cli.Success(resp.Data))
	return nil
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
