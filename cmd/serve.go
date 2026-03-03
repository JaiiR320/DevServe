package cmd

import (
	"devserve/cli"
	"devserve/client"
	"devserve/protocol"
	"fmt"
	"os"
	"strconv"

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

	port, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}

	var result *protocol.ServeResult
	cli.Spin("Starting process...", func() {
		result, err = client.Serve(args[0], port, args[2], cwd)
	})
	if err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	fmt.Println(cli.RenderServeResult(result))
	return nil
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
