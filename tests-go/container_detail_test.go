package testsgo

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/msherburne/dokbox/internal/domain"
	"github.com/msherburne/dokbox/internal/theme"
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
	if !strings.Contains(initialView, "Container Detail") {
		t.Fatalf("expected framed detail title, got %q", initialView)
	}
	if !strings.Contains(initialView, "[Overview]") {
		t.Fatalf("expected overview tab active, got %q", initialView)
	}
	if !strings.Contains(initialView, "Container: api") {
		t.Fatalf("expected container header, got %q", initialView)
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
	if !strings.Contains(filesView, "[Files]") || !strings.Contains(filesView, "Type") || !strings.Contains(filesView, "Name") || !strings.Contains(filesView, "Mode") {
		t.Fatalf("expected structured files headers, got %q", filesView)
	}
	if !strings.Contains(filesView, "dir") || !strings.Contains(filesView, "etc") || !strings.Contains(filesView, "drwxr-xr-x") {
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
	if !strings.Contains(view, "Selected: etc") {
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

func TestContainerDetailTruncatesLongFileRowsWithinConfiguredWidth(t *testing.T) {
	model := containerdetail.NewModel(
		"container-1",
		"api",
		nil,
		nil,
		nil,
		[]domain.FileEntry{
			{
				Name: "extremely-long-config-filename-that-should-truncate.yaml",
				Path: "/etc/extremely-long-config-filename-that-should-truncate.yaml",
				Mode: "-rw-r--r--",
			},
		},
	)
	model.SetWidth(58)

	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})
	nextModel, _ = nextModel.Update(tea.KeyMsg{Type: tea.KeyRight})

	view := nextModel.View()
	if got := lipgloss.Width(view); got > 58 {
		t.Fatalf("expected detail view width <= 58, got %d with view %q", got, view)
	}
	if !strings.Contains(view, "…") {
		t.Fatalf("expected truncated file row to include ellipsis, got %q", view)
	}
}

func TestContainerDetailClipsTallLogsWithinConfiguredHeight(t *testing.T) {
	model := containerdetail.NewModel(
		"container-1",
		"api",
		nil,
		[]domain.LogLine{
			{Text: "line-1"},
			{Text: "line-2"},
			{Text: "line-3"},
			{Text: "line-4"},
			{Text: "line-5"},
			{Text: "line-6"},
		},
		nil,
		nil,
	)
	model.SetWidth(64)
	model.SetHeight(10)

	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	view := nextModel.View()
	if got := lipgloss.Height(view); got > 10 {
		t.Fatalf("expected detail view height <= 10, got %d with view %q", got, view)
	}
	if !strings.Contains(view, "More content below") {
		t.Fatalf("expected clipped detail view to indicate hidden content, got %q", view)
	}
}

func TestContainerDetailTabsStaySingleLineWhenSwitching(t *testing.T) {
	model := containerdetail.NewModel(
		"container-1",
		"api",
		[]domain.MetricSample{{Name: "CPU", Label: "25%"}},
		[]domain.LogLine{{Text: "ready"}},
		nil,
		nil,
	)
	model.SetWidth(90)

	initialView := model.View()
	if strings.Contains(initialView, "│ │ [Overview] │") || strings.Contains(initialView, "┌────────────┐") {
		t.Fatalf("expected detail tabs to render inline without boxed multi-line chrome, got %q", initialView)
	}
	if !strings.Contains(initialView, "[Overview]   Logs   Shell   Files") {
		t.Fatalf("expected detail tabs to share one inline row, got %q", initialView)
	}

	nextModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	switchedView := nextModel.View()
	if strings.Contains(switchedView, "│ │ [Logs] │") || strings.Contains(switchedView, "┌────────┐") {
		t.Fatalf("expected switched detail tabs to remain inline without boxed multi-line chrome, got %q", switchedView)
	}
	if !strings.Contains(switchedView, "Overview   [Logs]   Shell   Files") {
		t.Fatalf("expected switched detail tabs to share one inline row, got %q", switchedView)
	}
}

func TestContainerDetailStylesUseSemanticFeedbackColors(t *testing.T) {
	styles := containerdetail.NewStyles(theme.Theme{
		Name:       "test",
		Panel:      "#202020",
		Border:     "#303030",
		Text:       "#efefef",
		Muted:      "#9a9a9a",
		Error:      "#cc2222",
		Info:       "#2277cc",
		Focus:      "#55bbff",
	})

	if got := styles.Error.GetForeground(); got != lipgloss.Color("#cc2222") {
		t.Fatalf("expected detail error foreground %q, got %q", "#cc2222", got)
	}
	if got := styles.Info.GetForeground(); got != lipgloss.Color("#2277cc") {
		t.Fatalf("expected detail info foreground %q, got %q", "#2277cc", got)
	}
	if got := styles.Meta.GetForeground(); got != lipgloss.Color("#9a9a9a") {
		t.Fatalf("expected detail meta foreground %q, got %q", "#9a9a9a", got)
	}
}
