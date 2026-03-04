package tui

import (
	"github.com/jaiir320/devserve/cli"
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
		b.WriteString(" " + cli.Dim.Render("No processes") + "\n")
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
	nameW, portW := 0, 0 // kept for compatibility, but not used with right-aligned layout

	// Configured section
	if len(configured) > 0 {
		for i, item := range configured {
			b.WriteString(renderItemRow(m, item, i, nameW, portW))
		}
	}

	// Separator if both sections exist
	if len(configured) > 0 && len(ephemeral) > 0 {
		separator := cli.Dim.Render(strings.Repeat("─", 22)) // match content width
		b.WriteString(" " + separator + "\n")
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

// renderItemRow renders a single item row with right-aligned port.
func renderItemRow(m model, item listItem, index, nameW, portW int) string {
	// Fixed content width for left pane
	const contentWidth = 22

	// Calculate spacing between name and port
	nameLen := len(item.Name)
	portStr := fmt.Sprintf("%d", item.Port)
	portLen := len(portStr)

	// Calculate needed spacing
	spacing := contentWidth - nameLen - portLen - 3 // 3 = dot + 2 spaces before dot
	if spacing < 1 {
		spacing = 1
	}

	// Build row with right-aligned port
	row := item.Name + strings.Repeat(" ", spacing) + portStr

	dot := stoppedDot
	if item.Running {
		dot = runningDot
	}

	if index == m.cursor {
		// Selected row
		return " " + dot + " " + selectedStyle.Render(row) + "\n"
	}
	// Normal row
	return " " + dot + " " + normalStyle.Render(row) + "\n"
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
