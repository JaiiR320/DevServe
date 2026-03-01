package tui

import (
	"devserve/cli"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// processRow holds all data for a running process.
type processRow struct {
	Name     string
	Port     int
	Command  string
	Dir      string
	LocalURL string
	IPURL    string
	DNSURL   string
}

// Left pane styles.
var (
	headerStyle   = cli.Bold.Foreground(lipgloss.Color("6"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	normalStyle   = lipgloss.NewStyle()
	cursorStr     = cli.Cyan.Render("▸")
)

// renderLeftPane renders the left pane: title + process table with cursor.
func renderLeftPane(m model) string {
	var b strings.Builder

	// Title
	title := cli.Bold.Render("devserve")
	b.WriteString("\n  " + title + "\n\n")

	if len(m.processes) == 0 {
		empty := cli.Dim.Render("No running processes")
		b.WriteString("  " + empty + "\n")
		return b.String()
	}

	// Calculate column widths
	nameW := 4 // "NAME"
	portW := 4 // "PORT"
	for _, p := range m.processes {
		if len(p.Name) > nameW {
			nameW = len(p.Name)
		}
		ps := fmt.Sprintf("%d", p.Port)
		if len(ps) > portW {
			portW = len(ps)
		}
	}

	// Header
	header := fmt.Sprintf("  %-*s  %-*s", nameW, "NAME", portW, "PORT")
	b.WriteString("  " + headerStyle.Render(header) + "\n")

	// Rows
	for i, p := range m.processes {
		prefix := "  "
		style := normalStyle
		if i == m.cursor {
			prefix = cursorStr + " "
			style = selectedStyle
		}

		cols := fmt.Sprintf("%-*s  %-*d", nameW, p.Name, portW, p.Port)
		b.WriteString("  " + prefix + style.Render(cols) + "\n")
	}

	return b.String()
}
