package testsgo

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/msherburne/dokbox/internal/app"
	"github.com/msherburne/dokbox/internal/domain"
)

func TestNewModelBootstrapsIntoReadyShell(t *testing.T) {
	model := app.NewModel(app.Dependencies{})
	if model == nil {
		t.Fatal("expected model")
	}

	initialView := model.View()
	if !strings.Contains(initialView, "Bootstrapping shell") {
		t.Fatalf("expected bootstrap view, got %q", initialView)
	}

	initCmd := model.Init()
	if initCmd == nil {
		t.Fatal("expected startup init command")
	}

	startupMsg := initCmd()
	if startupMsg == nil {
		t.Fatal("expected startup init command to emit a message")
	}

	nextModel, nextCmd := model.Update(startupMsg)
	if nextCmd != nil {
		t.Fatal("expected bootstrap update to finish without follow-up command")
	}

	readyView := nextModel.View()
	if !strings.Contains(readyView, "Containers") {
		t.Fatalf("expected browser shell view, got %q", readyView)
	}

	if strings.Contains(readyView, "Bootstrapping shell") {
		t.Fatalf("expected bootstrap copy to clear after startup, got %q", readyView)
	}
}

func TestNewModelShowsConnectedDockerStatusInReadyShell(t *testing.T) {
	model := app.NewModel(app.Dependencies{
		ConnectionChecker: &stubConnectionChecker{
			status: domain.ConnectionStatus{OK: true, Message: "Connected"},
		},
	})

	startupMsg := model.Init()()
	nextModel, nextCmd := model.Update(startupMsg)
	if nextCmd != nil {
		t.Fatal("expected startup update to finish without follow-up command")
	}

	readyView := nextModel.View()
	if !strings.Contains(readyView, "Docker: Connected") {
		t.Fatalf("expected connected docker status, got %q", readyView)
	}
}

func TestNewModelShowsDockerConnectionFailureWithoutExitingShell(t *testing.T) {
	model := app.NewModel(app.Dependencies{
		ConnectionChecker: &stubConnectionChecker{
			status: domain.ConnectionStatus{
				OK:      false,
				Message: "permission denied while trying to connect",
			},
		},
	})

	startupMsg := model.Init()()
	nextModel, nextCmd := model.Update(startupMsg)
	if nextCmd != nil {
		t.Fatal("expected failed connection startup to finish without follow-up command")
	}

	readyView := nextModel.View()
	if !strings.Contains(readyView, "Docker connection failed: permission denied while trying to connect") {
		t.Fatalf("expected connection failure copy, got %q", readyView)
	}
	if !strings.Contains(readyView, "q Quit") {
		t.Fatalf("expected shell to remain usable after connection failure, got %q", readyView)
	}
}

func TestNewModelShowsStartupFailureWithoutExitingShell(t *testing.T) {
	model := app.NewModel(app.Dependencies{
		ConnectionChecker: &stubConnectionChecker{
			status: domain.ConnectionStatus{
				OK:      false,
				Message: "invalid docker host",
			},
		},
	})

	startupMsg := model.Init()()
	nextModel, nextCmd := model.Update(startupMsg)
	if nextCmd != nil {
		t.Fatal("expected startup failure to finish without follow-up command")
	}

	readyView := nextModel.View()
	if !strings.Contains(readyView, "Docker connection failed: invalid docker host") {
		t.Fatalf("expected startup failure copy, got %q", readyView)
	}
	if !strings.Contains(readyView, "q Quit") {
		t.Fatalf("expected shell to remain usable after startup failure, got %q", readyView)
	}
}

func TestNewModelRefreshesRuntimeConnectionStatus(t *testing.T) {
	checker := &stubConnectionChecker{
		statuses: []domain.ConnectionStatus{
			{OK: true, Message: "Connected"},
			{OK: false, Message: "daemon unreachable"},
		},
	}
	model := app.NewModel(app.Dependencies{
		ConnectionChecker: checker,
	})

	startupMsg := model.Init()()
	nextModel, nextCmd := model.Update(startupMsg)
	if nextCmd != nil {
		t.Fatal("expected startup update to finish without follow-up command")
	}

	readyView := nextModel.View()
	if !strings.Contains(readyView, "Docker: Connected") {
		t.Fatalf("expected connected startup status, got %q", readyView)
	}

	refreshedModel, refreshCmd := nextModel.Update(keyMsg("r"))
	if refreshCmd == nil {
		t.Fatal("expected refresh command after runtime refresh key")
	}

	refreshMsg := refreshCmd()
	if refreshMsg == nil {
		t.Fatal("expected refresh command to emit a message")
	}

	refreshedModel, nextCmd = refreshedModel.Update(refreshMsg)
	if nextCmd != nil {
		t.Fatal("expected refresh update to finish without follow-up command")
	}

	refreshedView := refreshedModel.View()
	if !strings.Contains(refreshedView, "Docker connection failed: daemon unreachable") {
		t.Fatalf("expected runtime failure status, got %q", refreshedView)
	}
}

func TestNewModelRefreshesFromStartupFailureToConnected(t *testing.T) {
	checker := &stubConnectionChecker{
		statuses: []domain.ConnectionStatus{
			{OK: false, Message: "invalid docker host"},
			{OK: true, Message: "Connected"},
		},
	}
	model := app.NewModel(app.Dependencies{
		ConnectionChecker: checker,
	})

	startupMsg := model.Init()()
	nextModel, nextCmd := model.Update(startupMsg)
	if nextCmd != nil {
		t.Fatal("expected startup update to finish without follow-up command")
	}

	readyView := nextModel.View()
	if !strings.Contains(readyView, "Docker connection failed: invalid docker host") {
		t.Fatalf("expected startup failure status, got %q", readyView)
	}

	refreshedModel, refreshCmd := nextModel.Update(keyMsg("r"))
	if refreshCmd == nil {
		t.Fatal("expected refresh command after startup failure")
	}

	refreshMsg := refreshCmd()
	if refreshMsg == nil {
		t.Fatal("expected refresh command to emit a message")
	}

	refreshedModel, nextCmd = refreshedModel.Update(refreshMsg)
	if nextCmd != nil {
		t.Fatal("expected refresh update to finish without follow-up command")
	}

	refreshedView := refreshedModel.View()
	if !strings.Contains(refreshedView, "Docker: Connected") {
		t.Fatalf("expected recovered docker status, got %q", refreshedView)
	}
}

func TestNewModelTimesOutHungConnectionChecks(t *testing.T) {
	model := app.NewModel(app.Dependencies{
		ConnectionChecker: blockingConnectionChecker{},
	})

	startupMsg := model.Init()()
	nextModel, nextCmd := model.Update(startupMsg)
	if nextCmd != nil {
		t.Fatal("expected timeout startup to finish without follow-up command")
	}

	readyView := nextModel.View()
	if !strings.Contains(readyView, "Docker connection failed: context deadline exceeded") {
		t.Fatalf("expected timeout failure copy, got %q", readyView)
	}
}

func TestNewModelLetsBrowserHandleQWhenTableIsFocused(t *testing.T) {
	model := app.NewModel(app.Dependencies{})

	startupMsg := model.Init()()
	nextModel, nextCmd := model.Update(startupMsg)
	if nextCmd != nil {
		t.Fatal("expected startup update to finish without follow-up command")
	}

	nextModel, nextCmd = nextModel.Update(keyMsg("enter"))
	if nextCmd != nil {
		t.Fatal("expected enter to focus browser table without quit")
	}

	nextModel, nextCmd = nextModel.Update(keyMsg("q"))
	if nextCmd != nil {
		t.Fatal("expected q to return browser focus to tabs instead of quitting")
	}

	view := nextModel.View()
	if !strings.Contains(view, "Enter Focus Table") {
		t.Fatalf("expected tab shortcuts after returning focus, got %q", view)
	}
}

func TestNewModelRunsContainerStartAction(t *testing.T) {
	runner := &stubActionRunner{}
	model := app.NewModel(app.Dependencies{
		ActionRunner: runner,
		InitialContainers: []domain.ResourceSummary{
			{
				Kind: domain.ResourceKindContainer,
				ID:   "container-1",
				Name: "api",
				Columns: map[string]string{
					"Image":  "nginx:latest",
					"State":  "Running",
					"Status": "Up 2 hours",
				},
			},
		},
	})

	startupMsg := model.Init()()
	nextModel, _ := model.Update(startupMsg)
	nextModel, _ = nextModel.Update(keyMsg("enter"))
	nextModel, cmd := nextModel.Update(keyMsg("o"))
	if cmd != nil {
		t.Fatal("expected opening actions menu to avoid side effects")
	}

	nextModel, cmd = nextModel.Update(keyMsg("s"))
	if cmd == nil {
		t.Fatal("expected start action to schedule execution")
	}

	actionMsg := cmd()
	nextModel, cmd = nextModel.Update(actionMsg)
	if cmd == nil {
		t.Fatal("expected action request to schedule mutation command")
	}

	resultMsg := cmd()
	nextModel, _ = nextModel.Update(resultMsg)

	if runner.startedID != "container-1" {
		t.Fatalf("expected start action for selected container, got %q", runner.startedID)
	}

	if !strings.Contains(nextModel.View(), "Last action: start api") {
		t.Fatalf("expected last action status in view, got %q", nextModel.View())
	}
}

func TestNewModelRequiresConfirmationForContainerRemove(t *testing.T) {
	runner := &stubActionRunner{}
	model := app.NewModel(app.Dependencies{
		ActionRunner: runner,
		InitialContainers: []domain.ResourceSummary{
			{
				Kind: domain.ResourceKindContainer,
				ID:   "container-1",
				Name: "api",
				Columns: map[string]string{
					"Image":  "nginx:latest",
					"State":  "Running",
					"Status": "Up 2 hours",
				},
			},
		},
	})

	startupMsg := model.Init()()
	nextModel, _ := model.Update(startupMsg)
	nextModel, _ = nextModel.Update(keyMsg("enter"))
	nextModel, _ = nextModel.Update(keyMsg("o"))

	nextModel, cmd := nextModel.Update(keyMsg("x"))
	if cmd != nil {
		t.Fatal("expected remove to open confirmation without dispatch")
	}

	view := nextModel.View()
	if !strings.Contains(view, "Confirm remove api?") {
		t.Fatalf("expected remove confirmation prompt, got %q", view)
	}

	if runner.removedID != "" {
		t.Fatalf("expected no mutation before confirmation, got %q", runner.removedID)
	}

	nextModel, cmd = nextModel.Update(keyMsg("y"))
	if cmd == nil {
		t.Fatal("expected confirmation to dispatch remove action")
	}

	actionMsg := cmd()
	nextModel, cmd = nextModel.Update(actionMsg)
	if cmd == nil {
		t.Fatal("expected action request to schedule mutation command")
	}

	resultMsg := cmd()
	nextModel, _ = nextModel.Update(resultMsg)

	if runner.removedID != "container-1" {
		t.Fatalf("expected remove action for selected container, got %q", runner.removedID)
	}
}

func TestNewModelRequiresConfirmationForContainerPrune(t *testing.T) {
	runner := &stubActionRunner{}
	model := app.NewModel(app.Dependencies{
		ActionRunner: runner,
		InitialContainers: []domain.ResourceSummary{
			{
				Kind: domain.ResourceKindContainer,
				ID:   "container-1",
				Name: "api",
				Columns: map[string]string{
					"Image":  "nginx:latest",
					"State":  "Running",
					"Status": "Up 2 hours",
				},
			},
		},
	})

	startupMsg := model.Init()()
	nextModel, _ := model.Update(startupMsg)
	nextModel, _ = nextModel.Update(keyMsg("enter"))
	nextModel, _ = nextModel.Update(keyMsg("o"))

	nextModel, cmd := nextModel.Update(keyMsg("p"))
	if cmd != nil {
		t.Fatal("expected prune to open confirmation without dispatch")
	}

	view := nextModel.View()
	if !strings.Contains(view, "Confirm prune containers?") {
		t.Fatalf("expected prune confirmation prompt, got %q", view)
	}

	if runner.prunedKind != "" {
		t.Fatalf("expected no prune before confirmation, got %q", runner.prunedKind)
	}

	nextModel, cmd = nextModel.Update(keyMsg("y"))
	if cmd == nil {
		t.Fatal("expected confirmation to dispatch prune action")
	}

	actionMsg := cmd()
	nextModel, cmd = nextModel.Update(actionMsg)
	if cmd == nil {
		t.Fatal("expected action request to schedule mutation command")
	}

	resultMsg := cmd()
	nextModel, _ = nextModel.Update(resultMsg)

	if runner.prunedKind != domain.ResourceKindContainer {
		t.Fatalf("expected prune action for containers, got %q", runner.prunedKind)
	}
}

func TestNewModelCancelsDestructiveContainerAction(t *testing.T) {
	for _, cancelKey := range []string{"n", "q", "esc"} {
		t.Run(cancelKey, func(t *testing.T) {
			runner := &stubActionRunner{}
			model := app.NewModel(app.Dependencies{
				ActionRunner: runner,
				InitialContainers: []domain.ResourceSummary{
					{
						Kind: domain.ResourceKindContainer,
						ID:   "container-1",
						Name: "api",
						Columns: map[string]string{
							"Image":  "nginx:latest",
							"State":  "Running",
							"Status": "Up 2 hours",
						},
					},
				},
			})

			startupMsg := model.Init()()
			nextModel, _ := model.Update(startupMsg)
			nextModel, _ = nextModel.Update(keyMsg("enter"))
			nextModel, _ = nextModel.Update(keyMsg("o"))
			nextModel, _ = nextModel.Update(keyMsg("x"))

			cancelMsg := keyMsg(cancelKey)
			if cancelKey == "esc" && cancelMsg.Type != tea.KeyEsc {
				t.Fatalf("expected esc helper to emit KeyEsc, got %v", cancelMsg.Type)
			}

			nextModel, cmd := nextModel.Update(cancelMsg)
			if cmd != nil {
				t.Fatal("expected cancel confirmation to avoid dispatch")
			}

			if runner.removedID != "" {
				t.Fatalf("expected no remove action after cancel, got %q", runner.removedID)
			}

			view := nextModel.View()
			if !strings.Contains(view, "Actions: api") {
				t.Fatalf("expected action menu after cancel, got %q", view)
			}
			if strings.Contains(view, "Confirm remove api?") {
				t.Fatalf("expected confirmation prompt to close after cancel, got %q", view)
			}
		})
	}
}

func keyMsg(key string) tea.KeyMsg {
	if key == "esc" {
		return tea.KeyMsg{Type: tea.KeyEsc}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}

type stubConnectionChecker struct {
	status   domain.ConnectionStatus
	statuses []domain.ConnectionStatus
	calls    int
}

func (s *stubConnectionChecker) CheckConnection(context.Context) domain.ConnectionStatus {
	if len(s.statuses) > 0 {
		status := s.statuses[s.calls]
		if s.calls < len(s.statuses)-1 {
			s.calls++
		}
		return status
	}
	return s.status
}

type blockingConnectionChecker struct{}

func (blockingConnectionChecker) CheckConnection(ctx context.Context) domain.ConnectionStatus {
	<-ctx.Done()
	return domain.ConnectionStatus{
		OK:      false,
		Message: ctx.Err().Error(),
	}
}

type stubActionRunner struct {
	startedID  string
	removedID  string
	prunedKind domain.ResourceKind
}

func (s *stubActionRunner) StartContainer(containerID string) error {
	s.startedID = containerID
	return nil
}

func (s *stubActionRunner) StopContainer(string) error {
	return nil
}

func (s *stubActionRunner) RestartContainer(string) error {
	return nil
}

func (s *stubActionRunner) RemoveResource(_ domain.ResourceKind, resourceID string) error {
	s.removedID = resourceID
	return nil
}

func (s *stubActionRunner) Prune(kind domain.ResourceKind) error {
	s.prunedKind = kind
	return nil
}

func TestNewModelOpensContainerDetailAndShowsLogs(t *testing.T) {
	model := app.NewModel(app.Dependencies{
		LogProvider: &stubLogProvider{
			lines: []domain.LogLine{
				{Text: "booting"},
				{Text: "ready"},
			},
		},
		MetricsProvider: &stubMetricsProvider{
			metrics: []domain.MetricSample{
				{Name: "CPU", Label: "25% of 4 cores"},
			},
		},
		ShellProvider: &stubShellProvider{
			session: &domain.ExecSession{Command: []string{"/bin/bash"}},
		},
		FileProvider: &stubFileProvider{
			entries: []domain.FileEntry{
				{Name: "etc", IsDir: true, Mode: "drwxr-xr-x"},
			},
		},
		InitialContainers: []domain.ResourceSummary{
			{
				Kind: domain.ResourceKindContainer,
				ID:   "container-1",
				Name: "api",
				Columns: map[string]string{
					"Image":  "nginx:latest",
					"State":  "Running",
					"Status": "Up 2 hours",
				},
			},
		},
	})

	startupMsg := model.Init()()
	nextModel, _ := model.Update(startupMsg)
	nextModel, _ = nextModel.Update(keyMsg("enter"))
	nextModel, cmd := nextModel.Update(keyMsg("enter"))
	if cmd == nil {
		t.Fatal("expected opening detail view command")
	}

	detailRequest := cmd()
	nextModel, cmd = nextModel.Update(detailRequest)
	if cmd == nil {
		t.Fatal("expected detail request to schedule detail loading")
	}

	detailLoadedMsg := cmd()
	nextModel, _ = nextModel.Update(detailLoadedMsg)

	view := nextModel.View()
	if !strings.Contains(view, "CPU: 25% of 4 cores") {
		t.Fatalf("expected overview metrics in detail view, got %q", view)
	}

	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	view = nextModel.View()
	if !strings.Contains(view, "[Logs]") {
		t.Fatalf("expected logs tab to become active, got %q", view)
	}
	if !strings.Contains(view, "booting") || !strings.Contains(view, "ready") {
		t.Fatalf("expected log lines in detail view, got %q", view)
	}

	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	view = nextModel.View()
	if !strings.Contains(view, "[Shell]") || !strings.Contains(view, "Shell command: /bin/bash") {
		t.Fatalf("expected shell detail view, got %q", view)
	}

	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	view = nextModel.View()
	if !strings.Contains(view, "[Files]") || !strings.Contains(view, "Type") || !strings.Contains(view, "Name") || !strings.Contains(view, "Mode") {
		t.Fatalf("expected structured files detail headers, got %q", view)
	}
	if !strings.Contains(view, "dir") || !strings.Contains(view, "etc") || !strings.Contains(view, "Selected: etc") {
		t.Fatalf("expected files detail view, got %q", view)
	}
}

func TestNewModelLaunchesShellFromContainerDetail(t *testing.T) {
	shellProvider := &stubShellProvider{
		sessions: []*domain.ExecSession{
			{Command: []string{"/bin/bash"}},
			{Command: []string{"/bin/bash"}},
		},
	}
	shellLauncher := &stubShellLauncher{}
	model := app.NewModel(app.Dependencies{
		LogProvider:   &stubLogProvider{},
		ShellProvider: shellProvider,
		ShellLauncher: shellLauncher,
		FileProvider: &stubFileProvider{
			entries: []domain.FileEntry{{Name: "etc", IsDir: true, Mode: "drwxr-xr-x"}},
		},
		InitialContainers: []domain.ResourceSummary{
			{
				Kind: domain.ResourceKindContainer,
				ID:   "container-1",
				Name: "api",
				Columns: map[string]string{
					"Image":  "nginx:latest",
					"State":  "Running",
					"Status": "Up 2 hours",
				},
			},
		},
	})

	startupMsg := model.Init()()
	nextModel, _ := model.Update(startupMsg)
	nextModel, _ = nextModel.Update(keyMsg("enter"))
	nextModel, cmd := nextModel.Update(keyMsg("enter"))
	if cmd == nil {
		t.Fatal("expected opening detail view command")
	}

	nextModel, cmd = nextModel.Update(cmd())
	if cmd == nil {
		t.Fatal("expected detail request to schedule detail loading")
	}

	nextModel, _ = nextModel.Update(cmd())
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, cmd = nextModel.Update(keyMsg("enter"))
	if cmd == nil {
		t.Fatal("expected shell launch command")
	}

	nextModel, launchCmd := nextModel.Update(cmd())
	if launchCmd == nil {
		t.Fatal("expected shell launch request to schedule provider call")
	}

	nextModel, launchExecCmd := nextModel.Update(launchCmd())
	if launchExecCmd == nil {
		t.Fatal("expected shell session resolution to schedule interactive launcher")
	}

	nextModel, _ = nextModel.Update(launchExecCmd())
	if !strings.Contains(nextModel.View(), "Shell status: ready") {
		t.Fatalf("expected shell ready state, got %q", nextModel.View())
	}
	if shellProvider.calls != 2 {
		t.Fatalf("expected launch-time shell provider call, got %d calls", shellProvider.calls)
	}
	if shellProvider.containerID != "container-1" {
		t.Fatalf("expected shell launch to use container ID, got %q", shellProvider.containerID)
	}
	if shellLauncher.calls != 1 {
		t.Fatalf("expected interactive shell launcher to run once, got %d calls", shellLauncher.calls)
	}
	if shellLauncher.containerID != "container-1" {
		t.Fatalf("expected launcher to use container ID, got %q", shellLauncher.containerID)
	}
	if shellLauncher.session == nil || len(shellLauncher.session.Command) != 1 || shellLauncher.session.Command[0] != "/bin/bash" {
		t.Fatalf("expected launcher to receive resolved shell session, got %#v", shellLauncher.session)
	}
}

func TestNewModelShowsFailedShellLaunchState(t *testing.T) {
	shellProvider := &stubShellProvider{
		sessions: []*domain.ExecSession{
			{Command: []string{"/bin/bash"}},
			{Command: []string{"/bin/bash"}},
		},
	}
	shellLauncher := &stubShellLauncher{err: errors.New("shell launch denied")}
	model := app.NewModel(app.Dependencies{
		LogProvider:   &stubLogProvider{},
		ShellProvider: shellProvider,
		ShellLauncher: shellLauncher,
		FileProvider: &stubFileProvider{
			entries: []domain.FileEntry{{Name: "etc", IsDir: true, Mode: "drwxr-xr-x"}},
		},
		InitialContainers: []domain.ResourceSummary{
			{
				Kind: domain.ResourceKindContainer,
				ID:   "container-1",
				Name: "api",
				Columns: map[string]string{
					"Image":  "nginx:latest",
					"State":  "Running",
					"Status": "Up 2 hours",
				},
			},
		},
	})

	startupMsg := model.Init()()
	nextModel, _ := model.Update(startupMsg)
	nextModel, _ = nextModel.Update(keyMsg("enter"))
	nextModel, cmd := nextModel.Update(keyMsg("enter"))
	if cmd == nil {
		t.Fatal("expected opening detail view command")
	}

	nextModel, cmd = nextModel.Update(cmd())
	if cmd == nil {
		t.Fatal("expected detail request to schedule detail loading")
	}

	nextModel, _ = nextModel.Update(cmd())
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, cmd = nextModel.Update(keyMsg("enter"))
	if cmd == nil {
		t.Fatal("expected shell launch command")
	}

	nextModel, launchCmd := nextModel.Update(cmd())
	if launchCmd == nil {
		t.Fatal("expected shell launch request to schedule provider call")
	}

	nextModel, launchExecCmd := nextModel.Update(launchCmd())
	if launchExecCmd == nil {
		t.Fatal("expected shell session resolution to schedule interactive launcher")
	}

	nextModel, _ = nextModel.Update(launchExecCmd())
	view := nextModel.View()
	if !strings.Contains(view, "Shell status: failed") {
		t.Fatalf("expected failed shell state, got %q", view)
	}
	if !strings.Contains(view, "Shell error: shell launch denied") {
		t.Fatalf("expected shell launch error rendered, got %q", view)
	}
	if shellLauncher.calls != 1 {
		t.Fatalf("expected interactive shell launcher to run once, got %d calls", shellLauncher.calls)
	}
}

func TestNewModelNavigatesContainerFilesIntoDirectoryAndBack(t *testing.T) {
	fileProvider := &stubFileProvider{
		entriesByPath: map[string][]domain.FileEntry{
			"/": {
				{Name: "etc", Path: "/etc", IsDir: true, Mode: "drwxr-xr-x"},
				{Name: "hosts", Path: "/hosts", Mode: "-rw-r--r--"},
			},
			"/etc": {
				{Name: "config.yaml", Path: "/etc/config.yaml", Mode: "-rw-r--r--"},
			},
		},
	}
	model := app.NewModel(app.Dependencies{
		LogProvider:  &stubLogProvider{},
		FileProvider: fileProvider,
		InitialContainers: []domain.ResourceSummary{
			{
				Kind: domain.ResourceKindContainer,
				ID:   "container-1",
				Name: "api",
				Columns: map[string]string{
					"Image":  "nginx:latest",
					"State":  "Running",
					"Status": "Up 2 hours",
				},
			},
		},
	})

	startupMsg := model.Init()()
	nextModel, _ := model.Update(startupMsg)
	nextModel, _ = nextModel.Update(keyMsg("enter"))
	nextModel, cmd := nextModel.Update(keyMsg("enter"))
	if cmd == nil {
		t.Fatal("expected opening detail view command")
	}

	nextModel, cmd = nextModel.Update(cmd())
	if cmd == nil {
		t.Fatal("expected detail request to schedule detail loading")
	}

	nextModel, _ = nextModel.Update(cmd())
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})

	view := nextModel.View()
	if !strings.Contains(view, "Path: /") || !strings.Contains(view, "Selected: etc") {
		t.Fatalf("expected root files view with selected directory, got %q", view)
	}
	if len(fileProvider.paths) != 1 || fileProvider.paths[0] != "/" {
		t.Fatalf("expected initial root path lookup, got %#v", fileProvider.paths)
	}

	nextModel, cmd = nextModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected file navigation request")
	}

	nextModel, cmd = nextModel.Update(cmd())
	if cmd == nil {
		t.Fatal("expected file navigation to schedule provider call")
	}

	nextModel, _ = nextModel.Update(cmd())
	view = nextModel.View()
	if !strings.Contains(view, "Path: /etc") || !strings.Contains(view, "config.yaml") || !strings.Contains(view, "Selected: config.yaml") {
		t.Fatalf("expected child directory contents, got %q", view)
	}
	if got := fileProvider.paths[len(fileProvider.paths)-1]; got != "/etc" {
		t.Fatalf("expected directory navigation lookup for /etc, got %q", got)
	}

	nextModel, cmd = nextModel.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if cmd == nil {
		t.Fatal("expected parent navigation request")
	}

	nextModel, cmd = nextModel.Update(cmd())
	if cmd == nil {
		t.Fatal("expected parent navigation to schedule provider call")
	}

	nextModel, _ = nextModel.Update(cmd())
	view = nextModel.View()
	if !strings.Contains(view, "Path: /") || !strings.Contains(view, "Selected: etc") {
		t.Fatalf("expected root directory contents after navigating back, got %q", view)
	}
	if got := fileProvider.paths[len(fileProvider.paths)-1]; got != "/" {
		t.Fatalf("expected parent navigation lookup for /, got %q", got)
	}
}

func TestNewModelShowsInitialContainerFilesLoadError(t *testing.T) {
	fileProvider := &stubFileProvider{
		errByPath: map[string]error{
			"/": errors.New("permission denied"),
		},
	}
	model := app.NewModel(app.Dependencies{
		LogProvider:  &stubLogProvider{},
		FileProvider: fileProvider,
		InitialContainers: []domain.ResourceSummary{
			{
				Kind: domain.ResourceKindContainer,
				ID:   "container-1",
				Name: "api",
				Columns: map[string]string{
					"Image":  "nginx:latest",
					"State":  "Running",
					"Status": "Up 2 hours",
				},
			},
		},
	})

	startupMsg := model.Init()()
	nextModel, _ := model.Update(startupMsg)
	nextModel, _ = nextModel.Update(keyMsg("enter"))
	nextModel, cmd := nextModel.Update(keyMsg("enter"))
	if cmd == nil {
		t.Fatal("expected opening detail view command")
	}

	nextModel, cmd = nextModel.Update(cmd())
	if cmd == nil {
		t.Fatal("expected detail request to schedule detail loading")
	}

	nextModel, _ = nextModel.Update(cmd())
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})

	view := nextModel.View()
	if !strings.Contains(view, "Files error: permission denied") {
		t.Fatalf("expected file load error in files tab, got %q", view)
	}
}

func TestNewModelShowsInitialShellLoadError(t *testing.T) {
	shellProvider := &stubShellProvider{
		errs: []error{errors.New("shell probe failed")},
	}
	model := app.NewModel(app.Dependencies{
		LogProvider:   &stubLogProvider{},
		ShellProvider: shellProvider,
		InitialContainers: []domain.ResourceSummary{
			{
				Kind: domain.ResourceKindContainer,
				ID:   "container-1",
				Name: "api",
				Columns: map[string]string{
					"Image":  "nginx:latest",
					"State":  "Running",
					"Status": "Up 2 hours",
				},
			},
		},
	})

	startupMsg := model.Init()()
	nextModel, _ := model.Update(startupMsg)
	nextModel, _ = nextModel.Update(keyMsg("enter"))
	nextModel, cmd := nextModel.Update(keyMsg("enter"))
	if cmd == nil {
		t.Fatal("expected opening detail view command")
	}

	nextModel, cmd = nextModel.Update(cmd())
	if cmd == nil {
		t.Fatal("expected detail request to schedule detail loading")
	}

	nextModel, _ = nextModel.Update(cmd())
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})

	view := nextModel.View()
	if !strings.Contains(view, "Shell unavailable.") {
		t.Fatalf("expected shell unavailable state, got %q", view)
	}
	if !strings.Contains(view, "Shell status: failed") {
		t.Fatalf("expected initial shell failure status, got %q", view)
	}
	if !strings.Contains(view, "Shell error: shell probe failed") {
		t.Fatalf("expected initial shell failure error, got %q", view)
	}
}

type stubLogProvider struct {
	lines       []domain.LogLine
	containerID string
}

func (s *stubLogProvider) ContainerLogs(containerID string, tail int) ([]domain.LogLine, error) {
	s.containerID = containerID
	return s.lines, nil
}

type stubMetricsProvider struct {
	metrics     []domain.MetricSample
	containerID string
}

func (s *stubMetricsProvider) ContainerMetrics(containerID string) ([]domain.MetricSample, error) {
	s.containerID = containerID
	return s.metrics, nil
}

type stubShellProvider struct {
	session     *domain.ExecSession
	sessions    []*domain.ExecSession
	errs        []error
	containerID string
	calls       int
}

func (s *stubShellProvider) OpenShell(containerID string) (*domain.ExecSession, error) {
	s.containerID = containerID
	index := s.calls
	s.calls++

	if len(s.sessions) > 0 {
		session := s.sessions[len(s.sessions)-1]
		if index < len(s.sessions) {
			session = s.sessions[index]
		}
		var err error
		if index < len(s.errs) {
			err = s.errs[index]
		}
		return session, err
	}
	if index < len(s.errs) {
		return s.session, s.errs[index]
	}
	return s.session, nil
}

type stubShellLauncher struct {
	containerID string
	session     *domain.ExecSession
	err         error
	calls       int
}

func (s *stubShellLauncher) LaunchShell(containerID string, session *domain.ExecSession, callback func(error) tea.Msg) tea.Cmd {
	s.containerID = containerID
	s.session = session
	s.calls++
	return func() tea.Msg {
		return callback(s.err)
	}
}

type stubFileProvider struct {
	entries       []domain.FileEntry
	entriesByPath map[string][]domain.FileEntry
	errByPath     map[string]error
	containerID   string
	path          string
	paths         []string
}

func (s *stubFileProvider) ListContainerPath(containerID string, path string) ([]domain.FileEntry, error) {
	s.containerID = containerID
	s.path = path
	s.paths = append(s.paths, path)
	if err, ok := s.errByPath[path]; ok {
		return nil, err
	}
	if len(s.entriesByPath) > 0 {
		if entries, ok := s.entriesByPath[path]; ok {
			return entries, nil
		}
		return nil, nil
	}
	return s.entries, nil
}
