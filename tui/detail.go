package tui

import (
	"devserve/cli"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Detail pane styles.
var (
	detailLabel = cli.Dim
	urlStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)

// renderRightPane renders the right pane based on the active tab.
func renderRightPane(m model) string {
	if m.tab == 0 {
		return renderProcessDetail(m)
	}
	return renderConfigDetail(m)
}

// renderProcessDetail renders metadata for the selected running process.
func renderProcessDetail(m model) string {
	if len(m.processes) == 0 {
		return ""
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

	b.WriteString(renderKeyValuePairs(rows))
	return b.String()
}

// renderConfigDetail renders metadata for the selected saved config.
func renderConfigDetail(m model) string {
	if len(m.configs) == 0 {
		return ""
	}

	c := m.configs[m.configCur]
	var b strings.Builder

	// Config name header
	b.WriteString("\n  " + cli.Bold.Render(c.Name) + "\n\n")

	// Status line
	if c.Running {
		b.WriteString("  " + cli.Green.Render("● Running") + "\n\n")
	} else {
		b.WriteString("  " + cli.Dim.Render("○ Stopped") + "\n\n")
	}

	// Key-value pairs
	rows := []struct {
		label string
		value string
	}{
		{"Port", fmt.Sprintf("%d", c.Port)},
		{"Command", c.Command},
		{"Directory", c.Dir},
	}

	b.WriteString(renderKeyValuePairs(rows))
	return b.String()
}

// renderKeyValuePairs renders aligned label-value rows.
func renderKeyValuePairs(rows []struct {
	label string
	value string
}) string {
	// Find widest label for alignment
	labelW := 0
	for _, r := range rows {
		if len(r.label) > labelW {
			labelW = len(r.label)
		}
	}

	var b strings.Builder
	for _, r := range rows {
		label := detailLabel.Render(fmt.Sprintf("  %-*s", labelW, r.label))

		// Render URLs with cyan underline styling (no OSC 8 hyperlinks in TUI)
		value := r.value
		if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
			value = urlStyle.Render(value)
		}

		b.WriteString(label + "  " + value + "\n")
	}

	return b.String()
}
