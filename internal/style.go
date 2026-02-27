package internal

import (
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

	var entries []entry
	if err := json.Unmarshal([]byte(data), &entries); err != nil {
		return data // fallback to raw output
	}

	if len(entries) == 0 {
		return Dim.Render("No active processes")
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
	b.WriteString(Bold.Render(header))
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
