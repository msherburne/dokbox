package testsgo

import (
	"context"
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

func keyMsg(key string) tea.KeyMsg {
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
	startedID string
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

func (s *stubActionRunner) RemoveResource(domain.ResourceKind, string) error {
	return nil
}

func (s *stubActionRunner) Prune(domain.ResourceKind) error {
	return nil
}
