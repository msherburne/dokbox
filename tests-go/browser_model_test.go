package testsgo

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/msherburne/dokbox/internal/domain"
	"github.com/msherburne/dokbox/internal/ui/browser"
)

func TestBrowserModelRendersContainersTabAndRows(t *testing.T) {
	model := browser.NewModel([]domain.ResourceSummary{
		containerSummary("1", "api", "Running", "Up 2 hours", "compose"),
		containerSummary("2", "worker", "Exited", "Exited (0) 1 hour ago", "compose"),
	})

	view := model.View()
	if !strings.Contains(view, "Containers") {
		t.Fatalf("expected containers tab, got %q", view)
	}
	if !strings.Contains(view, "api") {
		t.Fatalf("expected api row, got %q", view)
	}
	if !strings.Contains(view, "Enter Focus Table") {
		t.Fatalf("expected tab shortcuts, got %q", view)
	}
}

func TestBrowserModelSwitchesFocusAndNavigatesRows(t *testing.T) {
	model := browser.NewModel([]domain.ResourceSummary{
		containerSummary("1", "api", "Running", "Up 2 hours", "compose"),
		containerSummary("2", "worker", "Exited", "Exited (0) 1 hour ago", "compose"),
	})

	nextModel, cmd := model.Update(browserKeyMsg("enter"))
	if cmd != nil {
		t.Fatal("expected enter to switch focus without command")
	}

	browserModel, ok := nextModel.(*browser.Model)
	if !ok {
		t.Fatalf("expected browser model, got %T", nextModel)
	}

	focusedView := browserModel.View()
	if !strings.Contains(focusedView, "> compose  api") {
		t.Fatalf("expected first row to be selected, got %q", focusedView)
	}
	if !strings.Contains(focusedView, "Up/Down Rows") {
		t.Fatalf("expected table shortcuts after focus, got %q", focusedView)
	}

	nextModel, cmd = browserModel.Update(browserKeyMsg("down"))
	if cmd != nil {
		t.Fatal("expected down to navigate without command")
	}

	browserModel, ok = nextModel.(*browser.Model)
	if !ok {
		t.Fatalf("expected browser model after navigation, got %T", nextModel)
	}

	navigatedView := browserModel.View()
	if !strings.Contains(navigatedView, "> compose  worker") {
		t.Fatalf("expected second row to be selected, got %q", navigatedView)
	}

	nextModel, cmd = browserModel.Update(browserKeyMsg("q"))
	if cmd != nil {
		t.Fatal("expected q to return focus to tabs without command")
	}

	browserModel, ok = nextModel.(*browser.Model)
	if !ok {
		t.Fatalf("expected browser model after unfocus, got %T", nextModel)
	}

	unfocusedView := browserModel.View()
	if strings.Contains(unfocusedView, "> compose  worker") {
		t.Fatalf("expected table selection marker to clear when returning to tabs, got %q", unfocusedView)
	}
	if !strings.Contains(unfocusedView, "Enter Focus Table") {
		t.Fatalf("expected tab shortcuts after returning focus, got %q", unfocusedView)
	}
}

func browserKeyMsg(key string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}

func containerSummary(id string, name string, state string, status string, group string) domain.ResourceSummary {
	return domain.ResourceSummary{
		Kind:  domain.ResourceKindContainer,
		ID:    id,
		Name:  name,
		Group: group,
		Columns: map[string]string{
			"Image":   "nginx:latest",
			"State":   state,
			"Status":  status,
			"Ports":   "80/tcp",
			"Created": "2 hours ago",
		},
	}
}
