package cli

import (
	"devserve/config"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
)

var (
	Green = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	Red   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	Cyan  = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	Bold  = lipgloss.NewStyle().Bold(true)
	Dim   = lipgloss.NewStyle().Faint(true)
)

// Success returns a styled success message with a green checkmark.
func Success(msg string) string {
	return Green.Render("✓") + " " + msg
}

// Error returns a styled error message with a red cross.
func Error(msg string) string {
	return Red.Render("✗") + " " + msg
}

// Info returns a styled info message with a cyan bullet.
func Info(msg string) string {
	return Cyan.Render("•") + " " + msg
}

// RenderTable parses the JSON list response and returns a formatted table string.
func RenderTable(data string) string {
	type entry struct {
		Name string `json:"name"`
		Port int    `json:"port"`
	}

	type listResp struct {
		Processes []entry `json:"processes"`
		Hostname  string  `json:"hostname"`
		IP        string  `json:"ip"`
	}

	var lr listResp
	if err := json.Unmarshal([]byte(data), &lr); err != nil {
		return data // fallback to raw output
	}

	if len(lr.Processes) == 0 {
		return Dim.Render("No active processes")
	}

	// Calculate column widths
	nameWidth := 4 // "NAME"
	portWidth := 4 // "PORT"
	for _, e := range lr.Processes {
		if len(e.Name) > nameWidth {
			nameWidth = len(e.Name)
		}
		p := fmt.Sprintf("%d", e.Port)
		if len(p) > portWidth {
			portWidth = len(p)
		}
	}

	// Build table
	var b strings.Builder
	header := fmt.Sprintf("%-*s  %-*s  %-5s  %-5s  %-3s",
		nameWidth, "NAME", portWidth, "PORT", "LOCAL", "IP", "DNS")
	b.WriteString(Bold.Render(header))
	for _, e := range lr.Processes {
		localURL := fmt.Sprintf("http://localhost:%d", e.Port)
		ipURL := fmt.Sprintf("http://%s:%d", lr.IP, e.Port)
		dnsURL := fmt.Sprintf("https://%s:%d", lr.Hostname, e.Port)

		localLink := Hyperlink(localURL, "local")
		ipLink := Hyperlink(ipURL, "ip")
		dnsLink := Hyperlink(dnsURL, "dns")

		// OSC 8 escape sequences are zero-width in terminals but count in
		// Go's %-*s padding. Compensate by adding the invisible byte overhead.
		localPad := 5 + len(localLink) - len("local")
		ipPad := 5 + len(ipLink) - len("ip")

		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("%-*s  %-*d  %-*s  %-*s  %s",
			nameWidth, e.Name, portWidth, e.Port,
			localPad, localLink,
			ipPad, ipLink,
			dnsLink))
	}

	return b.String()
}

// RenderConfigTable renders a list of ProcessConfig entries as a formatted table.
func RenderConfigTable(configs []config.ProcessConfig) string {
	if len(configs) == 0 {
		return Dim.Render("No saved configurations")
	}

	// Calculate column widths
	nameWidth := 4 // "NAME"
	portWidth := 4 // "PORT"
	cmdWidth := 7  // "COMMAND"
	dirWidth := 9  // "DIRECTORY"

	for _, c := range configs {
		if len(c.Name) > nameWidth {
			nameWidth = len(c.Name)
		}
		p := fmt.Sprintf("%d", c.Port)
		if len(p) > portWidth {
			portWidth = len(p)
		}
		if len(c.Command) > cmdWidth {
			cmdWidth = len(c.Command)
		}
		if len(c.Directory) > dirWidth {
			dirWidth = len(c.Directory)
		}
	}

	// Truncate long values
	maxCmdWidth := 40
	maxDirWidth := 50
	if cmdWidth > maxCmdWidth {
		cmdWidth = maxCmdWidth
	}
	if dirWidth > maxDirWidth {
		dirWidth = maxDirWidth
	}

	// Build table
	var b strings.Builder
	header := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s",
		nameWidth, "NAME", portWidth, "PORT", cmdWidth, "COMMAND", dirWidth, "DIRECTORY")
	b.WriteString(Bold.Render(header))

	for _, c := range configs {
		cmd := c.Command
		if len(cmd) > maxCmdWidth {
			cmd = cmd[:maxCmdWidth-3] + "..."
		}
		dir := c.Directory
		if len(dir) > maxDirWidth {
			dir = dir[:maxDirWidth-3] + "..."
		}

		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("%-*s  %-*d  %-*s  %s",
			nameWidth, c.Name, portWidth, c.Port, cmdWidth, cmd, dir))
	}

	return b.String()
}

// Hyperlink returns an OSC 8 hyperlink that renders as a clickable label in
// supported terminals.
func Hyperlink(url, label string) string {
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, label)
}

// Spin runs fn while displaying a spinner with the given title.
func Spin(title string, fn func()) {
	spinner.New().Title(title).Action(fn).Run()
}

// HelpTemplate returns a styled cobra help template.
func HelpTemplate() string {
	title := Bold.Render("{{.Name}}")
	return title + `{{if .Short}} - {{.Short}}{{end}}

` + Cyan.Render("Usage:") + `{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

` + Cyan.Render("Aliases:") + `
  {{.NameAndAliases}}{{end}}{{if .HasAvailableSubCommands}}

` + Cyan.Render("Commands:") + `{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding}}  {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

` + Cyan.Render("Flags:") + `
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

` + Cyan.Render("Global Flags:") + `
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

` + Cyan.Render("Additional help topics:") + `{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
}
