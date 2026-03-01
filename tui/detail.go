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

// renderRightPane renders the right pane based on the selected item.
func renderRightPane(m model) string {
	if len(m.items) == 0 {
		return ""
	}

	item := m.items[m.cursor]
	var b strings.Builder

	// Name header
	b.WriteString("  " + cli.Bold.Render(item.Name) + "\n\n")

	// Key-value pairs
	rows := []struct {
		label string
		value string
	}{
		{"Port", fmt.Sprintf("%d", item.Port)},
		{"Command", item.Command},
		{"Directory", item.Dir},
	}

	if item.Running {
		rows = append(rows, struct {
			label string
			value string
		}{"Local", item.LocalURL})
		if item.IPURL != "" {
			rows = append(rows, struct {
				label string
				value string
			}{"IP", item.IPURL})
		}
		if item.DNSURL != "" {
			rows = append(rows, struct {
				label string
				value string
			}{"DNS", item.DNSURL})
		}
	}

	// Add configured status
	if item.Configured {
		rows = append(rows, struct {
			label string
			value string
		}{"Config", "Saved"})
	} else {
		rows = append(rows, struct {
			label string
			value string
		}{"Config", "Unsaved (press 's' to save)"})
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

		// Render URLs with cyan styling
		value := r.value
		if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
			value = urlStyle.Render(value)
		}

		b.WriteString(label + "  " + value + "\n")
	}

	return b.String()
}
