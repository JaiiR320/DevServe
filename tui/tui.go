package tui

import (
	"devserve/cli"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	processes []processRow
	cursor    int
	width     int
	statusMsg string
	statusErr bool
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

	p := tea.NewProgram(m)
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
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			m.statusMsg = ""
			return m, nil

		case "down", "j":
			if m.cursor < len(m.processes)-1 {
				m.cursor++
			}
			m.statusMsg = ""
			return m, nil

		case "s":
			return m.stopSelected()

		case "r":
			return m.refresh()
		}
	}
	return m, nil
}

// -- layout --

// Pane border style: subtle left border for the right pane.
var (
	rightBorderStyle = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("8"))
)

func (m model) View() string {
	if m.width == 0 {
		return ""
	}

	// Pane widths
	leftW := m.width * 35 / 100
	if leftW < 25 {
		leftW = 25
	}
	rightW := m.width - leftW - 1 // -1 for border character

	// Render pane contents
	leftContent := renderLeftPane(m)
	rightContent := renderRightPane(m)

	// Apply widths to panes (no fixed height — content determines height)
	leftPane := lipgloss.NewStyle().
		Width(leftW).
		Render(leftContent)

	rightPane := rightBorderStyle.
		Width(rightW).
		Render(rightContent)

	// Join panes side by side
	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// Status line
	statusLine := ""
	if m.statusMsg != "" {
		if m.statusErr {
			statusLine = "  " + cli.Error(m.statusMsg)
		} else {
			statusLine = "  " + cli.Success(m.statusMsg)
		}
	}

	// Stack: body + status + help
	var b strings.Builder
	b.WriteString(body + "\n")
	if statusLine != "" {
		b.WriteString(statusLine + "\n")
	}
	b.WriteString(renderHelp())

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
