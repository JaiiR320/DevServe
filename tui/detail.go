package tui

import (
	"devserve/cli"
	"fmt"
	"strings"
)

// Detail pane styles.
var (
	detailLabel = cli.Dim
	detailValue = cli.Bold
)

// renderRightPane renders the right pane: process metadata for the selected row.
func renderRightPane(m model) string {
	if len(m.processes) == 0 {
		return "\n  " + cli.Dim.Render("Select a process") + "\n"
	}

	p := m.processes[m.cursor]
	var b strings.Builder

	// Process name header
	b.WriteString("\n  " + cli.Bold.Render(p.Name) + "\n\n")

	// Key-value pairs
	rows := []struct {
		label string
		value string
	}{
		{"Port", fmt.Sprintf("%d", p.Port)},
		{"Command", p.Command},
		{"Directory", p.Dir},
		{"Local", p.LocalURL},
	}
	if p.IPURL != "" {
		rows = append(rows, struct {
			label string
			value string
		}{"IP", p.IPURL})
	}
	if p.DNSURL != "" {
		rows = append(rows, struct {
			label string
			value string
		}{"DNS", p.DNSURL})
	}

	// Find widest label for alignment
	labelW := 0
	for _, r := range rows {
		if len(r.label) > labelW {
			labelW = len(r.label)
		}
	}

	for _, r := range rows {
		label := detailLabel.Render(fmt.Sprintf("  %-*s", labelW, r.label))

		// Render URLs as clickable hyperlinks
		value := r.value
		if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
			value = cli.Hyperlink(value, value)
		}

		b.WriteString(label + "  " + value + "\n")
	}

	return b.String()
}
