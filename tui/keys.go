package tui

import "devserve/cli"

// renderHelp returns the bottom help bar string.
func renderHelp() string {
	return cli.Dim.Render("  ↑/↓ navigate • s stop • r refresh • q quit")
}
