package tui

import "devserve/cli"

// renderHelp returns the bottom help bar string based on the active tab.
func renderHelp(tab int) string {
	if tab == 0 {
		return cli.Dim.Render("  ↑/↓ navigate • tab switch • s stop • r refresh • q quit")
	}
	return cli.Dim.Render("  ↑/↓ navigate • tab switch • enter start • r refresh • q quit")
}
