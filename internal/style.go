package internal

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
)

var (
	green = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	red   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	cyan  = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	bold  = lipgloss.NewStyle().Bold(true)
	dim   = lipgloss.NewStyle().Faint(true)
)

// Success returns a styled success message with a green checkmark.
func Success(msg string) string {
	return green.Render("✓") + " " + msg
}

// Error returns a styled error message with a red cross.
func Error(msg string) string {
	return red.Render("✗") + " " + msg
}

// Info returns a styled info message with a cyan bullet.
func Info(msg string) string {
	return cyan.Render("•") + " " + msg
}

// RenderTable parses the JSON list response and returns a formatted table string.
func RenderTable(data string) string {
	type entry struct {
		Name string `json:"name"`
		Port int    `json:"port"`
	}

	var entries []entry
	if err := json.Unmarshal([]byte(data), &entries); err != nil {
		return data // fallback to raw output
	}

	if len(entries) == 0 {
		return dim.Render("No active processes")
	}

	// Calculate column widths
	nameWidth := 4 // "NAME"
	portWidth := 4 // "PORT"
	for _, e := range entries {
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
	header := fmt.Sprintf("%-*s  %-*s", nameWidth, "NAME", portWidth, "PORT")
	b.WriteString(bold.Render(header))
	for _, e := range entries {
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("%-*s  %-*d", nameWidth, e.Name, portWidth, e.Port))
	}

	return b.String()
}

// Spin runs fn while displaying a spinner with the given title.
func Spin(title string, fn func()) {
	spinner.New().Title(title).Action(fn).Run()
}

// HelpTemplate returns a styled cobra help template.
func HelpTemplate() string {
	title := bold.Render("{{.Name}}")
	return title + `{{if .Short}} - {{.Short}}{{end}}

` + cyan.Render("Usage:") + `{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

` + cyan.Render("Aliases:") + `
  {{.NameAndAliases}}{{end}}{{if .HasAvailableSubCommands}}

` + cyan.Render("Commands:") + `{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding}}  {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

` + cyan.Render("Flags:") + `
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

` + cyan.Render("Global Flags:") + `
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

` + cyan.Render("Additional help topics:") + `{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
}
