package testsgo

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/msherburne/dokbox/internal/domain"
	"github.com/msherburne/dokbox/internal/ui/containerdetail"
)

func TestContainerDetailDefaultsToOverviewAndCanShowLogs(t *testing.T) {
	model := containerdetail.NewModel("container-1", "api", []domain.MetricSample{
		{Name: "CPU", Label: "25% of 4 cores"},
	}, []domain.LogLine{
		{Text: "booting"},
		{Text: "ready"},
	}, &domain.ExecSession{Command: []string{"/bin/bash"}}, []domain.FileEntry{
		{Name: "etc", IsDir: true, Mode: "drwxr-xr-x"},
	})

	initialView := model.View()
	if !strings.Contains(initialView, "[Overview]") {
		t.Fatalf("expected overview tab active, got %q", initialView)
	}
	if !strings.Contains(initialView, "CPU: 25% of 4 cores") {
		t.Fatalf("expected overview content, got %q", initialView)
	}

	nextModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	if cmd != nil {
		t.Fatal("expected right arrow to switch tab without command")
	}

	detailModel := nextModel.(*containerdetail.Model)
	logsView := detailModel.View()
	if !strings.Contains(logsView, "[Logs]") {
		t.Fatalf("expected logs tab active, got %q", logsView)
	}
	if !strings.Contains(logsView, "booting") || !strings.Contains(logsView, "ready") {
		t.Fatalf("expected log lines rendered, got %q", logsView)
	}
}

func TestContainerDetailShowsNoLogsState(t *testing.T) {
	model := containerdetail.NewModel("container-1", "api", nil, nil, nil, nil)
	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})

	logsView := nextModel.View()
	if !strings.Contains(logsView, "No logs.") {
		t.Fatalf("expected empty logs state, got %q", logsView)
	}
}

func TestContainerDetailShowsShellAndFiles(t *testing.T) {
	model := containerdetail.NewModel(
		"container-1",
		"api",
		nil,
		nil,
		&domain.ExecSession{Command: []string{"/bin/sh"}},
		[]domain.FileEntry{{Name: "etc", IsDir: true, Mode: "drwxr-xr-x"}},
	)

	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})

	shellView := nextModel.View()
	if !strings.Contains(shellView, "[Shell]") || !strings.Contains(shellView, "Shell command: /bin/sh") {
		t.Fatalf("expected shell view, got %q", shellView)
	}

	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	filesView := nextModel.View()
	if !strings.Contains(filesView, "[Files]") || !strings.Contains(filesView, "[d] etc drwxr-xr-x") {
		t.Fatalf("expected files view, got %q", filesView)
	}
}

func TestContainerDetailShellCanEnterLaunchingState(t *testing.T) {
	model := containerdetail.NewModel(
		"container-1",
		"api",
		nil,
		nil,
		&domain.ExecSession{Command: []string{"/bin/sh"}},
		nil,
	)

	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, cmd := nextModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("enter")})
	if cmd == nil {
		t.Fatal("expected shell launch request")
	}

	msg := cmd()
	request, ok := msg.(containerdetail.LaunchShellRequest)
	if !ok {
		t.Fatalf("expected shell launch request message, got %T", msg)
	}
	if request.ContainerID != "container-1" {
		t.Fatalf("expected shell launch request for selected container, got %q", request.ContainerID)
	}

	view := nextModel.View()
	if !strings.Contains(view, "Shell status: launching") {
		t.Fatalf("expected launching shell state, got %q", view)
	}
}

func TestContainerDetailFilesCanRequestDirectoryEntry(t *testing.T) {
	model := containerdetail.NewModel(
		"container-1",
		"api",
		nil,
		nil,
		nil,
		[]domain.FileEntry{
			{Name: "etc", Path: "/etc", IsDir: true, Mode: "drwxr-xr-x"},
			{Name: "hosts", Path: "/hosts", Mode: "-rw-r--r--"},
		},
	)

	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, cmd := nextModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected file navigation request")
	}

	msg := cmd()
	request, ok := msg.(containerdetail.NavigateContainerPathRequest)
	if !ok {
		t.Fatalf("expected file navigation request message, got %T", msg)
	}
	if request.ContainerID != "container-1" {
		t.Fatalf("expected navigation request for selected container, got %q", request.ContainerID)
	}
	if request.Path != "/etc" {
		t.Fatalf("expected directory navigation to /etc, got %q", request.Path)
	}

	view := nextModel.View()
	if !strings.Contains(view, "Path: /") {
		t.Fatalf("expected files view to show current path, got %q", view)
	}
	if !strings.Contains(view, "> [d] etc drwxr-xr-x") {
		t.Fatalf("expected selected directory to be highlighted, got %q", view)
	}
}

func TestContainerDetailFilesCanRequestParentNavigation(t *testing.T) {
	model := containerdetail.NewModel(
		"container-1",
		"api",
		nil,
		nil,
		nil,
		[]domain.FileEntry{
			{Name: "config.yaml", Path: "/etc/config.yaml", Mode: "-rw-r--r--"},
		},
	)
	model.ApplyFileNavigationResult("/etc", []domain.FileEntry{
		{Name: "config.yaml", Path: "/etc/config.yaml", Mode: "-rw-r--r--"},
	}, nil)

	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, cmd := nextModel.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if cmd == nil {
		t.Fatal("expected parent navigation request")
	}

	msg := cmd()
	request, ok := msg.(containerdetail.NavigateContainerPathRequest)
	if !ok {
		t.Fatalf("expected parent navigation request message, got %T", msg)
	}
	if request.Path != "/" {
		t.Fatalf("expected parent navigation to root, got %q", request.Path)
	}
}
