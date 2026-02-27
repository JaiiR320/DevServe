package cmd

import (
	"devserve/daemon"
	"devserve/internal"
	"fmt"
	"path/filepath"
	"strings"

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

var daemonCmdLogs = &cobra.Command{
	Use:   "logs",
	Args:  cobra.NoArgs,
	Short: "Show daemon logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		lines, _ := cmd.Flags().GetInt("lines")
		path := filepath.Join(internal.DaemonDir, internal.DaemonLogFile)

		logLines, err := internal.LastNLines(path, lines)
		if err != nil {
			return fmt.Errorf("failed to read daemon log: %w", err)
		}

		if len(logLines) == 0 {
			fmt.Println(internal.Info("no daemon logs found"))
			return nil
		}

		var b strings.Builder
		b.WriteString(internal.Cyan.Render("─── daemon ───"))
		b.WriteString("\n")
		for _, line := range logLines {
			b.WriteString(line)
			b.WriteString("\n")
		}
		fmt.Print(b.String())
		return nil
	},
}

func init() {
	daemonCmdStart.Flags().BoolVarP(&foreground, "foreground", "f", false, "Run daemon in foreground")
	daemonCmdLogs.Flags().IntP("lines", "n", 50, "number of lines to show")
	daemonCmd.AddCommand(daemonCmdStart, daemonCmdStop, daemonCmdLogs)
	rootCmd.AddCommand(daemonCmd)
}
