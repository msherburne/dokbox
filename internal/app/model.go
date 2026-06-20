package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/msherburne/dokbox/internal/domain"
	"github.com/msherburne/dokbox/internal/ui/browser"
)

type ConnectionChecker interface {
	CheckConnection(context.Context) domain.ConnectionStatus
}

type ActionRunner interface {
	StartContainer(containerID string) error
	StopContainer(containerID string) error
	RestartContainer(containerID string) error
	RemoveResource(kind domain.ResourceKind, resourceID string) error
	Prune(kind domain.ResourceKind) error
}

type Dependencies struct {
	ConnectionChecker ConnectionChecker
	ActionRunner      ActionRunner
	InitialContainers []domain.ResourceSummary
}

type connectionStatusMsg struct {
	connectionStatus *domain.ConnectionStatus
}

type startupCompleteMsg struct {
	connectionStatus *domain.ConnectionStatus
}

type actionResultMsg struct {
	status string
}

const connectionCheckTimeout = 2 * time.Second

type Model struct {
	ready            bool
	connectionStatus *domain.ConnectionStatus
	lastActionStatus string
	deps             Dependencies
	browser          *browser.Model
}

func NewModel(deps Dependencies) *Model {
	return &Model{
		deps:    deps,
		browser: browser.NewModel(deps.InitialContainers),
	}
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
	case browser.ActionRequest:
		if m.deps.ActionRunner == nil {
			return m, nil
		}
		return m, m.runActionCmd(msg)
	case actionResultMsg:
		m.lastActionStatus = msg.status
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "r":
			if m.deps.ConnectionChecker != nil {
				return m, m.checkConnectionCmd()
			}
		case QuitKey:
			if m.browser != nil && m.browser.TableFocused() {
				nextBrowser, cmd := m.browser.Update(msg)
				if next, ok := nextBrowser.(*browser.Model); ok {
					m.browser = next
				}
				return m, cmd
			}
			return m, tea.Quit
		}

		if m.ready && m.browser != nil {
			nextBrowser, cmd := m.browser.Update(msg)
			if next, ok := nextBrowser.(*browser.Model); ok {
				m.browser = next
			}
			return m, cmd
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
	}

	if m.connectionStatus != nil {
		if m.connectionStatus.OK {
			lines = append(lines, "", "Docker: "+m.connectionStatus.Message)
		} else {
			lines = append(lines, "", "Docker connection failed: "+m.connectionStatus.Message)
		}
	}

	if m.browser != nil {
		lines = append(lines, "", m.browser.View())
	}

	if m.lastActionStatus != "" {
		lines = append(lines, "", "Last action: "+m.lastActionStatus)
	}

	return ShellStyle.Render(strings.Join(lines, "\n"))
}

func (m *Model) runActionCmd(request browser.ActionRequest) tea.Cmd {
	return func() tea.Msg {
		var err error

		switch request.Action {
		case "start":
			err = m.deps.ActionRunner.StartContainer(request.ResourceID)
		case "stop":
			err = m.deps.ActionRunner.StopContainer(request.ResourceID)
		case "restart":
			err = m.deps.ActionRunner.RestartContainer(request.ResourceID)
		case "remove":
			err = m.deps.ActionRunner.RemoveResource(request.Kind, request.ResourceID)
		case "prune":
			err = m.deps.ActionRunner.Prune(request.Kind)
		}

		if err != nil {
			return actionResultMsg{
				status: fmt.Sprintf("%s %s failed: %v", request.Action, request.ResourceName, err),
			}
		}

		return actionResultMsg{status: request.Action + " " + request.ResourceName}
	}
}
