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

// configRow holds a saved configuration with its running status.
type configRow struct {
	Name    string
	Port    int
	Command string
	Dir     string
	Running bool
}

// Left pane styles.
var (
	headerStyle   = cli.Bold.Foreground(lipgloss.Color("6"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	normalStyle   = lipgloss.NewStyle()
	cursorStr     = cli.Cyan.Render("▸")
	runningDot    = cli.Green.Render("●")
	stoppedDot    = cli.Dim.Render("○")
)

// renderLeftPane renders the left pane: title, tab bar, and the active list.
func renderLeftPane(m model) string {
	var b strings.Builder

	// Title
	title := cli.Bold.Render("devserve")
	b.WriteString("\n  " + title + "\n")

	// Tab bar
	b.WriteString(renderTabBar(m.tab) + "\n\n")

	// Table based on active tab
	if m.tab == 0 {
		b.WriteString(renderProcessTable(m))
	} else {
		b.WriteString(renderConfigTable(m))
	}

	return b.String()
}

// renderTabBar renders the tab selector line.
func renderTabBar(tab int) string {
	active := "Active"
	config := "Config"

	if tab == 0 {
		active = cli.Cyan.Bold(true).Render(active)
		config = cli.Dim.Render(config)
	} else {
		active = cli.Dim.Render(active)
		config = cli.Cyan.Bold(true).Render(config)
	}

	return "  " + active + cli.Dim.Render(" • ") + config
}

// renderProcessTable renders the active processes table.
func renderProcessTable(m model) string {
	if len(m.processes) == 0 {
		return "  " + cli.Dim.Render("No running processes") + "\n"
	}

	nameW, portW := columnWidthsProcesses(m.processes)

	var b strings.Builder

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

// renderConfigTable renders the saved configurations table with status dots.
func renderConfigTable(m model) string {
	if len(m.configs) == 0 {
		return "  " + cli.Dim.Render("No saved configs") + "\n"
	}

	nameW, portW := columnWidthsConfigs(m.configs)

	var b strings.Builder

	// Header — extra 2 chars for the dot + space before NAME
	header := fmt.Sprintf("    %-*s  %-*s", nameW, "NAME", portW, "PORT")
	b.WriteString("  " + headerStyle.Render(header) + "\n")

	// Rows
	for i, c := range m.configs {
		prefix := "  "
		style := normalStyle
		if i == m.configCur {
			prefix = cursorStr + " "
			style = selectedStyle
		}

		dot := stoppedDot
		if c.Running {
			dot = runningDot
		}

		cols := fmt.Sprintf("%-*s  %-*d", nameW, c.Name, portW, c.Port)
		b.WriteString("  " + prefix + dot + " " + style.Render(cols) + "\n")
	}

	return b.String()
}

// columnWidthsProcesses calculates NAME and PORT column widths for processes.
func columnWidthsProcesses(processes []processRow) (nameW, portW int) {
	nameW = 4 // "NAME"
	portW = 4 // "PORT"
	for _, p := range processes {
		if len(p.Name) > nameW {
			nameW = len(p.Name)
		}
		ps := fmt.Sprintf("%d", p.Port)
		if len(ps) > portW {
			portW = len(ps)
		}
	}
	return
}

// columnWidthsConfigs calculates NAME and PORT column widths for configs.
func columnWidthsConfigs(configs []configRow) (nameW, portW int) {
	nameW = 4 // "NAME"
	portW = 4 // "PORT"
	for _, c := range configs {
		if len(c.Name) > nameW {
			nameW = len(c.Name)
		}
		ps := fmt.Sprintf("%d", c.Port)
		if len(ps) > portW {
			portW = len(ps)
		}
	}
	return
}
