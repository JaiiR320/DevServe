package cmd

import (
	"devserve/daemon"
	"devserve/internal"
	"fmt"

	"github.com/spf13/cobra"
)

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Interact with the daemon",
}

var foreground bool

var daemonCmdStart = &cobra.Command{
	Use:   "start",
	Args:  cobra.NoArgs,
	Short: "Start the daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.Start(!foreground)
	},
}

var daemonCmdStop = &cobra.Command{
	Use:   "stop",
	Args:  cobra.NoArgs,
	Short: "Stop the daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			msg string
			err error
		)
		internal.Spin("Stopping daemon...", func() {
			msg, err = daemon.Stop()
		})
		if err != nil {
			return err
		}
		fmt.Println(internal.Success(msg))
		return nil
	},
}

func init() {
	daemonCmdStart.Flags().BoolVarP(&foreground, "foreground", "f", false, "Run daemon in foreground")
	daemonCmd.AddCommand(daemonCmdStart, daemonCmdStop)
	rootCmd.AddCommand(daemonCmd)
}
