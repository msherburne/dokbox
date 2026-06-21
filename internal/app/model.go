package app

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/msherburne/dokbox/internal/domain"
	"github.com/msherburne/dokbox/internal/ui/browser"
	"github.com/msherburne/dokbox/internal/ui/containerdetail"
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

type LogProvider interface {
	ContainerLogs(containerID string, tail int) ([]domain.LogLine, error)
}

type MetricsProvider interface {
	ContainerMetrics(containerID string) ([]domain.MetricSample, error)
}

type ShellProvider interface {
	OpenShell(containerID string) (*domain.ExecSession, error)
}

type ShellLauncher interface {
	LaunchShell(containerID string, session *domain.ExecSession, callback func(error) tea.Msg) tea.Cmd
}

type FileProvider interface {
	ListContainerPath(containerID string, path string) ([]domain.FileEntry, error)
}

type Dependencies struct {
	ConnectionChecker ConnectionChecker
	ActionRunner      ActionRunner
	LogProvider       LogProvider
	MetricsProvider   MetricsProvider
	ShellProvider     ShellProvider
	ShellLauncher     ShellLauncher
	FileProvider      FileProvider
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

type shellLaunchResultMsg struct {
	session *domain.ExecSession
	err     error
}

type detailLogsLoadedMsg struct {
	containerID   string
	containerName string
	metrics       []domain.MetricSample
	logs          []domain.LogLine
	shellSession  *domain.ExecSession
	shellErr      error
	files         []domain.FileEntry
	filesErr      error
	err           error
}

type containerPathLoadedMsg struct {
	path  string
	files []domain.FileEntry
	err   error
}

const connectionCheckTimeout = 2 * time.Second

type Model struct {
	ready            bool
	connectionStatus *domain.ConnectionStatus
	lastActionStatus string
	deps             Dependencies
	styles           Styles
	browser          *browser.Model
	detail           *containerdetail.Model
}

func NewModel(deps Dependencies, styles ...Styles) *Model {
	activeStyles := defaultStyles()
	if len(styles) > 0 {
		activeStyles = styles[0]
	}

	return &Model{
		deps:    deps,
		styles:  activeStyles,
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
	case browser.OpenContainerDetailRequest:
		return m, m.loadDetailLogsCmd(msg)
	case containerdetail.LaunchShellRequest:
		return m, m.launchShellCmd(msg)
	case containerdetail.NavigateContainerPathRequest:
		return m, m.loadContainerPathCmd(msg)
	case actionResultMsg:
		m.lastActionStatus = msg.status
		return m, nil
	case detailLogsLoadedMsg:
		m.detail = containerdetail.NewModel(
			msg.containerID,
			msg.containerName,
			msg.metrics,
			msg.logs,
			msg.shellSession,
			msg.files,
		)
		if msg.shellErr != nil {
			m.detail.ApplyShellLaunchResult(nil, msg.shellErr)
		}
		if msg.filesErr != nil {
			m.detail.ApplyFileNavigationResult("/", nil, msg.filesErr)
		}
		return m, nil
	case shellLaunchResultMsg:
		if m.detail != nil {
			m.detail.ApplyShellLaunchResult(msg.session, msg.err)
		}
		return m, nil
	case shellLaunchReadyMsg:
		return m, m.execShellCmd(msg.containerID, msg.session)
	case containerPathLoadedMsg:
		if m.detail != nil {
			m.detail.ApplyFileNavigationResult(msg.path, msg.files, msg.err)
		}
		return m, nil
	case tea.KeyMsg:
		if m.detail != nil {
			if msg.String() == QuitKey {
				m.detail = nil
				return m, nil
			}

			nextDetail, cmd := m.detail.Update(msg)
			if next, ok := nextDetail.(*containerdetail.Model); ok {
				m.detail = next
			}
			return m, cmd
		}

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
		return m.styles.Shell.Render("dokbox-go\n\nBootstrapping shell...")
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

	if m.detail != nil {
		lines = append(lines, "", m.detail.View())
		return m.styles.Shell.Render(strings.Join(lines, "\n"))
	}

	if m.browser != nil {
		lines = append(lines, "", m.browser.View())
	}

	if m.lastActionStatus != "" {
		lines = append(lines, "", "Last action: "+m.lastActionStatus)
	}

	return m.styles.Shell.Render(strings.Join(lines, "\n"))
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

func (m *Model) loadDetailLogsCmd(request browser.OpenContainerDetailRequest) tea.Cmd {
	return func() tea.Msg {
		var logs []domain.LogLine
		var err error
		if m.deps.LogProvider != nil {
			logs, err = m.deps.LogProvider.ContainerLogs(request.ResourceID, 200)
		}

		var metrics []domain.MetricSample
		if m.deps.MetricsProvider != nil {
			metrics, _ = m.deps.MetricsProvider.ContainerMetrics(request.ResourceID)
		}
		var shellSession *domain.ExecSession
		var shellErr error
		if m.deps.ShellProvider != nil {
			shellSession, shellErr = m.deps.ShellProvider.OpenShell(request.ResourceID)
		}
		var files []domain.FileEntry
		var filesErr error
		if m.deps.FileProvider != nil {
			files, filesErr = m.deps.FileProvider.ListContainerPath(request.ResourceID, "/")
		}
		return detailLogsLoadedMsg{
			containerID:   request.ResourceID,
			containerName: request.ResourceName,
			metrics:       metrics,
			logs:          logs,
			shellSession:  shellSession,
			shellErr:      shellErr,
			files:         files,
			filesErr:      filesErr,
			err:           err,
		}
	}
}

func (m *Model) launchShellCmd(request containerdetail.LaunchShellRequest) tea.Cmd {
	return func() tea.Msg {
		if m.deps.ShellProvider == nil {
			return shellLaunchResultMsg{err: fmt.Errorf("shell unavailable")}
		}

		session, err := m.deps.ShellProvider.OpenShell(request.ContainerID)
		if err != nil {
			return shellLaunchResultMsg{err: err}
		}

		return shellLaunchReadyMsg{
			containerID: request.ContainerID,
			session:     session,
		}
	}
}

type shellLaunchReadyMsg struct {
	containerID string
	session     *domain.ExecSession
}

func (m *Model) loadContainerPathCmd(request containerdetail.NavigateContainerPathRequest) tea.Cmd {
	return func() tea.Msg {
		if m.deps.FileProvider == nil {
			return containerPathLoadedMsg{path: request.Path}
		}

		files, err := m.deps.FileProvider.ListContainerPath(request.ContainerID, request.Path)
		return containerPathLoadedMsg{
			path:  request.Path,
			files: files,
			err:   err,
		}
	}
}

func (m *Model) execShellCmd(containerID string, session *domain.ExecSession) tea.Cmd {
	if session == nil || len(session.Command) == 0 {
		return func() tea.Msg {
			return shellLaunchResultMsg{err: fmt.Errorf("shell unavailable")}
		}
	}

	if m.deps.ShellLauncher != nil {
		return m.deps.ShellLauncher.LaunchShell(containerID, session, func(err error) tea.Msg {
			return shellLaunchResultMsg{
				session: session,
				err:     err,
			}
		})
	}

	args := append([]string{"exec", "-it", containerID}, session.Command...)
	command := exec.Command("docker", args...)
	return tea.ExecProcess(command, func(err error) tea.Msg {
		return shellLaunchResultMsg{
			session: session,
			err:     err,
		}
	})
}
