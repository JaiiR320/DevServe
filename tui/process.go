package tui

import (
	"devserve/cli"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Left pane styles.
var (
	selectedStyle = lipgloss.NewStyle().Bold(true)
	normalStyle   = lipgloss.NewStyle()
	runningDot    = cli.Green.Render("●")
	stoppedDot    = cli.Dim.Render("○")
)

// renderLeftPane renders the unified list view with sections.
func renderLeftPane(m model) string {
	var b strings.Builder

	if len(m.items) == 0 {
		b.WriteString("  " + cli.Dim.Render("No processes") + "\n")
		return b.String()
	}

	// Separate items into configured and ephemeral
	var configured, ephemeral []listItem
	for _, item := range m.items {
		if item.Configured {
			configured = append(configured, item)
		} else {
			ephemeral = append(ephemeral, item)
		}
	}

	// Calculate column widths
	nameW, portW := columnWidths(m.items)

	// Configured section
	if len(configured) > 0 {
		for i, item := range configured {
			b.WriteString(renderItemRow(m, item, i, nameW, portW))
		}
	}

	// Separator if both sections exist
	if len(configured) > 0 && len(ephemeral) > 0 {
		separator := cli.Dim.Render(strings.Repeat("─", nameW+portW+6))
		b.WriteString("  " + separator + "\n")
	}

	// Ephemeral section
	if len(ephemeral) > 0 {
		// Calculate offset for ephemeral items
		offset := len(configured)
		for i, item := range ephemeral {
			b.WriteString(renderItemRow(m, item, offset+i, nameW, portW))
		}
	}

	return b.String()
}

// renderItemRow renders a single item row.
func renderItemRow(m model, item listItem, index, nameW, portW int) string {
	cols := fmt.Sprintf("%-*s  %-*d", nameW, item.Name, portW, item.Port)

	dot := stoppedDot
	if item.Running {
		dot = runningDot
	}

	if index == m.cursor {
		// Selected row
		return "  " + dot + " " + selectedStyle.Render(cols) + "\n"
	}
	// Normal row
	return "  " + dot + " " + normalStyle.Render(cols) + "\n"
}

// columnWidths calculates NAME and PORT column widths.
func columnWidths(items []listItem) (nameW, portW int) {
	nameW = 4 // "NAME"
	portW = 4 // "PORT"
	for _, item := range items {
		if len(item.Name) > nameW {
			nameW = len(item.Name)
		}
		ps := fmt.Sprintf("%d", item.Port)
		if len(ps) > portW {
			portW = len(ps)
		}
	}
	return
}
