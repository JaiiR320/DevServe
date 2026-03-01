package tui

import "devserve/cli"

// renderHelp returns the bottom help bar string based on the active tab.
func renderHelp(tab int) string {
	if tab == 0 {
		return cli.Dim.Render("  ↑/↓ navigate • tab switch • s stop • q quit")
	}
	return cli.Dim.Render("  ↑/↓ navigate • tab switch • enter start • s stop • q quit")
}
