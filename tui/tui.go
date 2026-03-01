package tui

import (
	"devserve/cli"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// listItem represents a unified process entry (configured or ephemeral)
type listItem struct {
	Name       string
	Port       int
	Command    string
	Dir        string
	Running    bool
	Configured bool // true if saved to config
	LocalURL   string
	IPURL      string
	DNSURL     string
}

type model struct {
	items     []listItem
	cursor    int
	width     int
	statusMsg string
	statusErr bool
}

// Run launches the TUI. It fetches the initial data and starts the
// bubbletea program.
func Run() error {
	items, err := fetchItems()
	if err != nil {
		items = nil
	}

	m := model{
		items: items,
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
			m.moveCursor(-1)
			m.statusMsg = ""
			return m, nil

		case "down", "j":
			m.moveCursor(1)
			m.statusMsg = ""
			return m, nil

		case "enter":
			return m.toggleStartStop()

		case "s":
			return m.toggleSave()
		}
	}
	return m, nil
}

// -- cursor helpers --

func (m *model) moveCursor(delta int) {
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// -- layout --

// Pane border styles: rounded borders for both panes with minimal padding.
var (
	leftPaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1) // minimal: 0 vertical, 1 horizontal

	rightPaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1) // minimal: 0 vertical, 1 horizontal
)

func (m model) View() string {
	if m.width == 0 {
		return ""
	}

	// Fixed left pane width, give rest to right pane
	leftContentWidth := 22        // content width for left pane
	leftW := leftContentWidth + 6 // +6 for borders/padding
	rightW := m.width - leftW - 6 // -6 for right pane borders/padding only (no gap needed)
	if rightW < 30 {
		rightW = 30 // minimum right pane width
	}

	// Render pane contents first (without styling)
	leftContent := renderLeftPane(m)
	rightContent := renderRightPane(m)

	// Calculate content heights
	leftContentHeight := lipgloss.Height(leftContent)
	rightContentHeight := lipgloss.Height(rightContent)
	contentHeight := leftContentHeight
	if rightContentHeight > contentHeight {
		contentHeight = rightContentHeight
	}

	// Total height includes borders (2 rows) but padding is internal
	paneHeight := contentHeight + 2 // +2 for top/bottom borders only

	// Apply border styles with calculated dimensions
	leftPane := leftPaneStyle.
		Width(leftW).
		Height(paneHeight).
		Render(leftContent)

	rightPane := rightPaneStyle.
		Width(rightW).
		Height(paneHeight).
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

func (m model) toggleStartStop() (model, tea.Cmd) {
	if len(m.items) == 0 {
		return m, nil
	}

	item := m.items[m.cursor]

	if item.Running {
		// Stop the process
		err := stopProcess(item.Name)
		if err != nil {
			m.statusMsg = fmt.Sprintf("failed to stop '%s': %s", item.Name, err)
			m.statusErr = true
			return m, nil
		}
		m.statusMsg = fmt.Sprintf("process '%s' stopped", item.Name)
		m.statusErr = false
	} else {
		// Start the process (only configured items can be started)
		if !item.Configured {
			m.statusMsg = fmt.Sprintf("cannot start '%s': save to config first (press 's')", item.Name)
			m.statusErr = true
			return m, nil
		}
		err := startItem(item)
		if err != nil {
			m.statusMsg = fmt.Sprintf("failed to start '%s': %s", item.Name, err)
			m.statusErr = true
			return m, nil
		}
		m.statusMsg = fmt.Sprintf("process '%s' started", item.Name)
		m.statusErr = false
	}

	// Reload data to reflect changes
	return m.reload()
}

func (m model) toggleSave() (model, tea.Cmd) {
	if len(m.items) == 0 {
		return m, nil
	}

	item := m.items[m.cursor]

	if item.Configured {
		// Remove from config
		err := removeFromConfig(item.Name)
		if err != nil {
			m.statusMsg = fmt.Sprintf("failed to remove '%s' from config: %s", item.Name, err)
			m.statusErr = true
			return m, nil
		}
		m.statusMsg = fmt.Sprintf("'%s' removed from config", item.Name)
		m.statusErr = false
	} else {
		// Save to config
		err := saveToConfig(item)
		if err != nil {
			m.statusMsg = fmt.Sprintf("failed to save '%s' to config: %s", item.Name, err)
			m.statusErr = true
			return m, nil
		}
		m.statusMsg = fmt.Sprintf("'%s' saved to config", item.Name)
		m.statusErr = false
	}

	// Reload data to reflect changes
	return m.reload()
}

// reload refreshes the data from daemon and config.
func (m model) reload() (model, tea.Cmd) {
	items, err := fetchItems()
	if err != nil {
		m.statusMsg = fmt.Sprintf("failed to reload: %s", err)
		m.statusErr = true
		return m, nil
	}

	// Preserve cursor position if possible
	oldCursor := m.cursor
	m.items = items

	// Adjust cursor if out of bounds
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	// If cursor changed dramatically, try to find the same item by name
	if oldCursor < len(m.items) && oldCursor >= 0 {
		// Cursor is still valid, keep it
		m.cursor = oldCursor
	}

	return m, nil
}
