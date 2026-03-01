package tui

import (
	"devserve/cli"
	"devserve/daemon"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// processRow holds the display data for a single running process.
type processRow struct {
	Name     string
	Port     int
	LocalURL string
}

type model struct {
	processes []processRow
	cursor    int
	width     int
	height    int
	statusMsg string
	statusErr bool
	err       error
}

// Run launches the TUI. It fetches the initial process list and starts
// the bubbletea program.
func Run() error {
	processes, err := fetchProcesses()
	if err != nil {
		processes = nil
	}

	m := model{
		processes: processes,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// -- bubbletea interface --

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			if m.cursor < len(m.processes)-1 {
				m.cursor++
			}
			return m, nil

		case "s":
			return m.stopSelected()

		case "r":
			return m.refresh()
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	// Title
	title := cli.Bold.Render("devserve")
	b.WriteString("\n  " + title + "\n\n")

	if len(m.processes) == 0 {
		empty := cli.Dim.Render("No running processes")
		b.WriteString("  " + empty + "\n")
	} else {
		b.WriteString(m.renderTable())
	}

	// Status message — positioned above the help line
	statusLine := ""
	if m.statusMsg != "" {
		if m.statusErr {
			statusLine = "  " + cli.Error(m.statusMsg)
		} else {
			statusLine = "  " + cli.Success(m.statusMsg)
		}
	}

	// Help line
	help := cli.Dim.Render("  ↑/↓ navigate • s stop • r refresh • q quit")

	// Fill remaining vertical space so help sits at the bottom
	contentLines := strings.Count(b.String(), "\n")
	// 2 more lines: status + help
	pad := m.height - contentLines - 2
	if pad < 1 {
		pad = 1
	}
	b.WriteString(strings.Repeat("\n", pad))
	if statusLine != "" {
		b.WriteString(statusLine + "\n")
	} else {
		b.WriteString("\n")
	}
	b.WriteString(help)

	return b.String()
}

// -- table rendering --

var (
	headerStyle   = cli.Bold.Foreground(lipgloss.Color("6"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	normalStyle   = lipgloss.NewStyle()
	cursorStr     = cli.Cyan.Render("▸")
)

func (m model) renderTable() string {
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

	var b strings.Builder

	// Header
	header := fmt.Sprintf("    %-*s  %-*s  %s", nameW, "NAME", portW, "PORT", "URL")
	b.WriteString("  " + headerStyle.Render(header) + "\n")

	// Rows
	for i, p := range m.processes {
		cursor := " "
		style := normalStyle
		if i == m.cursor {
			cursor = cursorStr
			style = selectedStyle
		}

		row := fmt.Sprintf("  %s %-*s  %-*d  %s",
			cursor,
			nameW, p.Name,
			portW, p.Port,
			p.LocalURL,
		)
		b.WriteString(style.Render(row) + "\n")
	}

	return b.String()
}

// -- actions --

func (m model) stopSelected() (model, tea.Cmd) {
	if len(m.processes) == 0 {
		return m, nil
	}

	proc := m.processes[m.cursor]
	err := stopProcess(proc.Name)
	if err != nil {
		m.statusMsg = fmt.Sprintf("failed to stop '%s': %s", proc.Name, err)
		m.statusErr = true
		return m, nil
	}

	m.statusMsg = fmt.Sprintf("process '%s' stopped", proc.Name)
	m.statusErr = false

	// Refresh after stop
	m, _ = m.refresh()
	return m, nil
}

func (m model) refresh() (model, tea.Cmd) {
	processes, err := fetchProcesses()
	if err != nil {
		m.statusMsg = fmt.Sprintf("failed to refresh: %s", err)
		m.statusErr = true
		return m, nil
	}

	m.processes = processes

	// Clamp cursor
	if m.cursor >= len(m.processes) {
		m.cursor = len(m.processes) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	return m, nil
}

// -- daemon communication --

// sendRequest sends a request to the daemon, auto-starting it if needed.
func sendRequest(req *daemon.Request) (*daemon.Response, error) {
	resp, err := daemon.Send(req)
	if err == nil {
		return resp, nil
	}

	if !errors.Is(err, daemon.ErrDaemonNotRunning) {
		return resp, err
	}

	// Auto-start the daemon
	if startErr := daemon.Start(true); startErr != nil {
		return nil, fmt.Errorf("failed to auto-start daemon: %w", startErr)
	}

	return daemon.Send(req)
}

// fetchProcesses queries the daemon for the current process list.
func fetchProcesses() ([]processRow, error) {
	resp, err := sendRequest(&daemon.Request{Action: "list"})
	if err != nil {
		return nil, fmt.Errorf("failed to send list request: %w", err)
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}

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
	if err := json.Unmarshal([]byte(resp.Data), &lr); err != nil {
		return nil, fmt.Errorf("failed to parse list response: %w", err)
	}

	rows := make([]processRow, len(lr.Processes))
	for i, e := range lr.Processes {
		rows[i] = processRow{
			Name:     e.Name,
			Port:     e.Port,
			LocalURL: fmt.Sprintf("http://localhost:%d", e.Port),
		}
	}

	return rows, nil
}

// stopProcess sends a stop request to the daemon for the named process.
func stopProcess(name string) error {
	resp, err := sendRequest(&daemon.Request{
		Action: "stop",
		Args:   map[string]any{"name": name},
	})
	if err != nil {
		return fmt.Errorf("failed to send stop request: %w", err)
	}
	if !resp.OK {
		return errors.New(resp.Error)
	}
	return nil
}
