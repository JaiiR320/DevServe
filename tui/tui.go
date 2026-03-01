package tui

import (
	"devserve/cli"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	tab       int // 0 = active, 1 = config
	processes []processRow
	cursor    int // active tab cursor
	configs   []configRow
	configCur int // config tab cursor
	width     int
	statusMsg string
	statusErr bool
}

// Run launches the TUI. It fetches the initial data and starts the
// bubbletea program.
func Run() error {
	processes, err := fetchProcesses()
	if err != nil {
		processes = nil
	}

	configs, err := fetchConfigs(processes)
	if err != nil {
		configs = nil
	}

	m := model{
		processes: processes,
		configs:   configs,
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

		case "tab":
			m.tab = 1 - m.tab // toggle 0↔1
			m.statusMsg = ""
			return m, nil

		case "up", "k":
			m.moveCursor(-1)
			m.statusMsg = ""
			return m, nil

		case "down", "j":
			m.moveCursor(1)
			m.statusMsg = ""
			return m, nil

		case "s":
			if m.tab == 0 {
				return m.stopSelected()
			}
			return m, nil

		case "enter":
			if m.tab == 1 {
				return m.startSelected()
			}
			return m, nil

		case "r":
			return m.refresh()
		}
	}
	return m, nil
}

// -- cursor helpers --

func (m *model) moveCursor(delta int) {
	if m.tab == 0 {
		m.cursor += delta
		if m.cursor < 0 {
			m.cursor = 0
		}
		if m.cursor >= len(m.processes) {
			m.cursor = len(m.processes) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
	} else {
		m.configCur += delta
		if m.configCur < 0 {
			m.configCur = 0
		}
		if m.configCur >= len(m.configs) {
			m.configCur = len(m.configs) - 1
		}
		if m.configCur < 0 {
			m.configCur = 0
		}
	}
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

	// Apply widths to panes
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
	b.WriteString(renderHelp(m.tab))

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

	m, _ = m.refresh()
	return m, nil
}

func (m model) startSelected() (model, tea.Cmd) {
	if len(m.configs) == 0 {
		return m, nil
	}

	cfg := m.configs[m.configCur]
	if cfg.Running {
		m.statusMsg = fmt.Sprintf("'%s' is already running", cfg.Name)
		m.statusErr = true
		return m, nil
	}

	err := startProcess(cfg)
	if err != nil {
		m.statusMsg = fmt.Sprintf("failed to start '%s': %s", cfg.Name, err)
		m.statusErr = true
		return m, nil
	}

	m.statusMsg = fmt.Sprintf("process '%s' started", cfg.Name)
	m.statusErr = false

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

	configs, err := fetchConfigs(processes)
	if err != nil {
		m.statusMsg = fmt.Sprintf("failed to refresh configs: %s", err)
		m.statusErr = true
		return m, nil
	}
	m.configs = configs

	// Clamp cursors
	if m.cursor >= len(m.processes) {
		m.cursor = len(m.processes) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.configCur >= len(m.configs) {
		m.configCur = len(m.configs) - 1
	}
	if m.configCur < 0 {
		m.configCur = 0
	}

	return m, nil
}
