package testsgo

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	if !strings.Contains(view, "Resources") {
		t.Fatalf("expected framed browser title, got %q", view)
	}
	if !strings.Contains(view, "api") {
		t.Fatalf("expected api row, got %q", view)
	}
	if !strings.Contains(view, "Stack") || !strings.Contains(view, "Name") || !strings.Contains(view, "Status") {
		t.Fatalf("expected table headers, got %q", view)
	}
	if !strings.Contains(view, "Enter Focus Table") {
		t.Fatalf("expected tab shortcuts, got %q", view)
	}
	if !strings.Contains(view, "Images") || !strings.Contains(view, "Volumes") || !strings.Contains(view, "Networks") {
		t.Fatalf("expected all resource tabs, got %q", view)
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
	if !strings.Contains(focusedView, "api") {
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
	if !strings.Contains(navigatedView, "worker") {
		t.Fatalf("expected second row to be selected, got %q", navigatedView)
	}
	if !strings.Contains(navigatedView, "Selected") {
		t.Fatalf("expected explicit selection summary after navigation, got %q", navigatedView)
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
	if !strings.Contains(unfocusedView, "Enter Focus Table") {
		t.Fatalf("expected tab shortcuts after returning focus, got %q", unfocusedView)
	}
}

func TestBrowserModelSwitchesTabsAndRendersKindSpecificRows(t *testing.T) {
	model := browser.NewModel([]domain.ResourceSummary{
		containerSummary("1", "api", "Running", "Up 2 hours", "compose"),
		imageSummary("img-1", "dokbox", "latest"),
		volumeSummary("vol-1", "dokbox-data"),
		networkSummary("net-1", "dokbox_default"),
	})

	nextModel, cmd := model.Update(rightKeyMsg())
	if cmd != nil {
		t.Fatal("expected right to switch tabs without command")
	}

	browserModel, ok := nextModel.(*browser.Model)
	if !ok {
		t.Fatalf("expected browser model after tab switch, got %T", nextModel)
	}

	imageView := browserModel.View()
	if !strings.Contains(imageView, "[Images]") {
		t.Fatalf("expected images tab to be active, got %q", imageView)
	}
	if !strings.Contains(imageView, "Repository") || !strings.Contains(imageView, "Tag") {
		t.Fatalf("expected image columns, got %q", imageView)
	}
	if !strings.Contains(imageView, "dokbox") || !strings.Contains(imageView, "latest") {
		t.Fatalf("expected image row, got %q", imageView)
	}

	nextModel, cmd = browserModel.Update(rightKeyMsg())
	if cmd != nil {
		t.Fatal("expected right to switch to volumes without command")
	}
	browserModel = nextModel.(*browser.Model)
	volumeView := browserModel.View()
	if !strings.Contains(volumeView, "[Volumes]") {
		t.Fatalf("expected volumes tab to be active, got %q", volumeView)
	}
	if !strings.Contains(volumeView, "Mountpoint") || !strings.Contains(volumeView, "dokbox-data") {
		t.Fatalf("expected volume table content, got %q", volumeView)
	}

	nextModel, cmd = browserModel.Update(rightKeyMsg())
	if cmd != nil {
		t.Fatal("expected right to switch to networks without command")
	}
	browserModel = nextModel.(*browser.Model)
	networkView := browserModel.View()
	if !strings.Contains(networkView, "[Networks]") {
		t.Fatalf("expected networks tab to be active, got %q", networkView)
	}
	if !strings.Contains(networkView, "Flags") || !strings.Contains(networkView, "dokbox_default") {
		t.Fatalf("expected network table content, got %q", networkView)
	}
}

func TestBrowserModelOpensContainerActionsMenu(t *testing.T) {
	model := browser.NewModel([]domain.ResourceSummary{
		containerSummary("1", "api", "Running", "Up 2 hours", "compose"),
	})

	nextModel, cmd := model.Update(browserKeyMsg("enter"))
	if cmd != nil {
		t.Fatal("expected enter to focus table without command")
	}

	browserModel := nextModel.(*browser.Model)
	nextModel, cmd = browserModel.Update(browserKeyMsg("o"))
	if cmd != nil {
		t.Fatal("expected actions menu to open without immediate command")
	}

	browserModel = nextModel.(*browser.Model)
	view := browserModel.View()
	if !strings.Contains(view, "Actions: api") {
		t.Fatalf("expected actions menu for selected container, got %q", view)
	}
	if !strings.Contains(view, "s Start") || !strings.Contains(view, "x Remove") {
		t.Fatalf("expected action key hints in menu, got %q", view)
	}
}

func TestBrowserModelOpensSelectedContainerDetail(t *testing.T) {
	model := browser.NewModel([]domain.ResourceSummary{
		containerSummary("1", "api", "Running", "Up 2 hours", "compose"),
	})

	nextModel, _ := model.Update(browserKeyMsg("enter"))
	browserModel := nextModel.(*browser.Model)

	nextModel, cmd := browserModel.Update(browserKeyMsg("enter"))
	if cmd == nil {
		t.Fatal("expected enter on focused container to request detail open")
	}

	msg := cmd()
	request, ok := msg.(browser.OpenContainerDetailRequest)
	if !ok {
		t.Fatalf("expected open detail request, got %T", msg)
	}

	if request.ResourceID != "1" || request.ResourceName != "api" {
		t.Fatalf("unexpected detail request payload: %#v", request)
	}
}

func TestBrowserModelTruncatesLongRowsWithinConfiguredWidth(t *testing.T) {
	model := browser.NewModel([]domain.ResourceSummary{
		containerSummary(
			"1",
			"api-with-a-very-long-container-name-that-should-truncate",
			"Running",
			"Up for an extremely long amount of time with a verbose status message",
			"compose-stack-with-a-very-long-name",
		),
	})
	model.SetWidth(60)

	view := model.View()
	if got := lipgloss.Width(view); got > 60 {
		t.Fatalf("expected browser view width <= 60, got %d with view %q", got, view)
	}
	if !strings.Contains(view, "…") {
		t.Fatalf("expected truncated browser content to include ellipsis, got %q", view)
	}
}

func TestBrowserModelClipsTallContentWithinConfiguredHeight(t *testing.T) {
	model := browser.NewModel([]domain.ResourceSummary{
		containerSummary("1", "api", "Running", "Up 2 hours", "compose"),
		containerSummary("2", "worker", "Running", "Up 2 hours", "compose"),
		containerSummary("3", "cron", "Running", "Up 2 hours", "compose"),
		containerSummary("4", "db", "Running", "Up 2 hours", "compose"),
		containerSummary("5", "cache", "Running", "Up 2 hours", "compose"),
		containerSummary("6", "proxy", "Running", "Up 2 hours", "compose"),
	})
	model.SetWidth(72)
	model.SetHeight(10)

	view := model.View()
	if got := lipgloss.Height(view); got > 10 {
		t.Fatalf("expected browser view height <= 10, got %d with view %q", got, view)
	}
	if !strings.Contains(view, "More rows below") {
		t.Fatalf("expected clipped browser view to indicate hidden rows, got %q", view)
	}
}

func TestBrowserTabsStaySingleLineWhenSwitching(t *testing.T) {
	model := browser.NewModel([]domain.ResourceSummary{
		containerSummary("1", "api", "Running", "Up 2 hours", "compose"),
		imageSummary("img-1", "dokbox", "latest"),
	})
	model.SetWidth(90)

	initialView := model.View()
	if !strings.Contains(initialView, "Resources") {
		t.Fatalf("expected browser view, got %q", initialView)
	}
	if strings.Contains(initialView, "│ │ [Containers] │") || strings.Contains(initialView, "┌──────────────┐") {
		t.Fatalf("expected tabs to render inline without boxed multi-line chrome, got %q", initialView)
	}
	if !strings.Contains(initialView, "[Containers]   Images   Volumes   Networks") {
		t.Fatalf("expected tabs to share one inline row, got %q", initialView)
	}

	nextModel, _ := model.Update(rightKeyMsg())
	switchedView := nextModel.View()
	if strings.Contains(switchedView, "│ │ [Images] │") || strings.Contains(switchedView, "┌────────┐") {
		t.Fatalf("expected switched tabs to remain inline without boxed multi-line chrome, got %q", switchedView)
	}
	if !strings.Contains(switchedView, "Containers   [Images]   Volumes   Networks") {
		t.Fatalf("expected switched tabs to share one inline row, got %q", switchedView)
	}
}

func browserKeyMsg(key string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}

func rightKeyMsg() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRight}
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

func imageSummary(id string, repository string, tag string) domain.ResourceSummary {
	return domain.ResourceSummary{
		Kind: domain.ResourceKindImage,
		ID:   id,
		Name: repository,
		Columns: map[string]string{
			"Repository": repository,
			"Tag":        tag,
			"Image ID":   "sha256:abc123",
			"Size":       "123MB",
			"Created":    "3 days ago",
		},
	}
}

func volumeSummary(id string, name string) domain.ResourceSummary {
	return domain.ResourceSummary{
		Kind: domain.ResourceKindVolume,
		ID:   id,
		Name: name,
		Columns: map[string]string{
			"Driver":     "local",
			"Scope":      "local",
			"Mountpoint": "/var/lib/docker/volumes/dokbox-data/_data",
			"Created":    "1 day ago",
		},
	}
}

func networkSummary(id string, name string) domain.ResourceSummary {
	return domain.ResourceSummary{
		Kind: domain.ResourceKindNetwork,
		ID:   id,
		Name: name,
		Columns: map[string]string{
			"Driver":     "bridge",
			"Scope":      "local",
			"Flags":      "internal",
			"Containers": "2",
		},
	}
}
