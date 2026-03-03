package cmd

import (
	"devserve/cli"
	"devserve/client"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs [name]",
	Args:  cobra.ExactArgs(1),
	Short: "Show process logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		lines, _ := cmd.Flags().GetInt("lines")
		logsResult, err := client.Logs(args[0], lines)
		if err != nil {
			return fmt.Errorf("failed to get logs: %w", err)
		}

		// Render with styling
		var b strings.Builder
		b.WriteString(cli.Cyan.Render("─── stdout ───"))
		b.WriteString("\n")
		for _, line := range logsResult.Stdout {
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(cli.Cyan.Render("─── stderr ───"))
		b.WriteString("\n")
		for _, line := range logsResult.Stderr {
			b.WriteString(cli.Red.Render(line))
			b.WriteString("\n")
		}

		fmt.Print(strings.TrimRight(b.String(), "\n"))
		return nil
	},
}

func init() {
	logsCmd.Flags().IntP("lines", "n", 50, "number of lines to show")
	rootCmd.AddCommand(logsCmd)
}
