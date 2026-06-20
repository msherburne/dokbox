package app

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/msherburne/dokbox/internal/domain"
)

type ConnectionChecker interface {
	CheckConnection(context.Context) domain.ConnectionStatus
}

type Dependencies struct {
	ConnectionChecker ConnectionChecker
}

type connectionStatusMsg struct {
	connectionStatus *domain.ConnectionStatus
}

type startupCompleteMsg struct {
	connectionStatus *domain.ConnectionStatus
}

const connectionCheckTimeout = 2 * time.Second

type Model struct {
	ready            bool
	connectionStatus *domain.ConnectionStatus
	deps             Dependencies
}

func NewModel(deps Dependencies) *Model {
	return &Model{deps: deps}
}

func (m *Model) Init() tea.Cmd {
	return m.checkConnectionCmd()
}

func (m *Model) checkConnectionCmd() tea.Cmd {
	return func() tea.Msg {
		if m.deps.ConnectionChecker != nil {
			ctx, cancel := context.WithTimeout(context.Background(), connectionCheckTimeout)
			defer cancel()

			status := m.deps.ConnectionChecker.CheckConnection(ctx)
			return connectionStatusMsg{connectionStatus: &status}
		}
		return startupCompleteMsg{}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case startupCompleteMsg:
		m.ready = true
		return m, nil
	case connectionStatusMsg:
		m.ready = true
		m.connectionStatus = msg.connectionStatus
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", QuitKey:
			return m, tea.Quit
		case "r":
			if m.deps.ConnectionChecker != nil {
				return m, m.checkConnectionCmd()
			}
		}
	}

	return m, nil
}

func (m *Model) View() string {
	if !m.ready {
		return ShellStyle.Render("dokbox-go\n\nBootstrapping shell...")
	}

	lines := []string{
		"dokbox-go",
		"",
		"Starting rewrite shell...",
	}

	if m.connectionStatus != nil {
		if m.connectionStatus.OK {
			lines = append(lines, "", "Docker: "+m.connectionStatus.Message)
		} else {
			lines = append(lines, "", "Docker connection failed: "+m.connectionStatus.Message)
		}
	}

	lines = append(lines, "", "Press q to quit.")
	return ShellStyle.Render(strings.Join(lines, "\n"))
}
