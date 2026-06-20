package testsgo

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/msherburne/dokbox/internal/domain"
	"github.com/msherburne/dokbox/internal/ui/containerdetail"
)

func TestContainerDetailDefaultsToOverviewAndCanShowLogs(t *testing.T) {
	model := containerdetail.NewModel("api", []domain.MetricSample{
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
	model := containerdetail.NewModel("api", nil, nil, nil, nil)
	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})

	logsView := nextModel.View()
	if !strings.Contains(logsView, "No logs.") {
		t.Fatalf("expected empty logs state, got %q", logsView)
	}
}

func TestContainerDetailShowsShellAndFiles(t *testing.T) {
	model := containerdetail.NewModel(
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
