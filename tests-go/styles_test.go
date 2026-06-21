package testsgo

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/msherburne/dokbox/internal/app"
	"github.com/msherburne/dokbox/internal/theme"
)

func TestNewStylesUsesThemeColors(t *testing.T) {
	styles := app.NewStyles(theme.Theme{
		Name:       "test",
		Background: "#0f0f0f",
		Panel:      "#0f0f0f",
		Text:       "#eeeeee",
	})

	if got := styles.Shell.GetForeground(); got != lipgloss.Color("#eeeeee") {
		t.Fatalf("expected shell foreground %q, got %q", "#eeeeee", got)
	}

	if got := styles.Shell.GetBackground(); got != lipgloss.Color("#0f0f0f") {
		t.Fatalf("expected shell background %q, got %q", "#0f0f0f", got)
	}
}

func TestNewStylesUsesSemanticFeedbackColors(t *testing.T) {
	styles := app.NewStyles(theme.Theme{
		Name:       "test",
		Background: "#101010",
		Panel:      "#202020",
		Text:       "#f0f0f0",
		Muted:      "#9a9a9a",
		Success:    "#11aa11",
		Warning:    "#ffaa22",
		Error:      "#dd2222",
		Info:       "#2288dd",
	})

	if got := styles.Success.GetForeground(); got != lipgloss.Color("#11aa11") {
		t.Fatalf("expected success foreground %q, got %q", "#11aa11", got)
	}
	if got := styles.Warning.GetForeground(); got != lipgloss.Color("#ffaa22") {
		t.Fatalf("expected warning foreground %q, got %q", "#ffaa22", got)
	}
	if got := styles.Error.GetForeground(); got != lipgloss.Color("#dd2222") {
		t.Fatalf("expected error foreground %q, got %q", "#dd2222", got)
	}
	if got := styles.Info.GetForeground(); got != lipgloss.Color("#2288dd") {
		t.Fatalf("expected info foreground %q, got %q", "#2288dd", got)
	}
	if got := styles.Muted.GetForeground(); got != lipgloss.Color("#9a9a9a") {
		t.Fatalf("expected muted foreground %q, got %q", "#9a9a9a", got)
	}
}

func TestNewStylesAppliesShellPadding(t *testing.T) {
	styles := app.NewStyles(theme.Theme{
		Name:       "test",
		Background: "#111111",
		Text:       "#eeeeee",
	})

	if got := styles.Shell.GetPaddingTop(); got != 1 {
		t.Fatalf("expected top padding 1, got %d", got)
	}

	if got := styles.Shell.GetPaddingBottom(); got != 1 {
		t.Fatalf("expected bottom padding 1, got %d", got)
	}

	if got := styles.Shell.GetPaddingLeft(); got != 1 {
		t.Fatalf("expected left padding 1, got %d", got)
	}

	if got := styles.Shell.GetPaddingRight(); got != 1 {
		t.Fatalf("expected right padding 1, got %d", got)
	}
}

func TestNewModelUsesInjectedStylesInBootstrapView(t *testing.T) {
	styles := app.Styles{
		Shell: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			Padding(0, 0).
			Foreground(lipgloss.Color("#f5f5f5")),
	}

	model := app.NewModel(app.Dependencies{}, styles)
	view := model.View()

	if !strings.Contains(view, "dokbox-go") {
		t.Fatalf("expected bootstrap content to be preserved, got %q", view)
	}

	if !strings.Contains(view, "┃") {
		t.Fatalf("expected bootstrap view to use injected shell border, got %q", view)
	}
}

func TestNewModelUsesInjectedStylesInReadyShellView(t *testing.T) {
	styles := app.Styles{
		Shell: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			Padding(0, 0).
			Foreground(lipgloss.Color("#f5f5f5")),
	}

	model := app.NewModel(app.Dependencies{}, styles)
	startupMsg := model.Init()()
	nextModel, nextCmd := model.Update(startupMsg)
	if nextCmd != nil {
		t.Fatal("expected startup update to finish without follow-up command")
	}

	view := nextModel.View()
	if !strings.Contains(view, "Resources") {
		t.Fatalf("expected ready browser shell content, got %q", view)
	}

	if !strings.Contains(view, "[Containers]") || !strings.Contains(view, "Images") || !strings.Contains(view, "Volumes") || !strings.Contains(view, "Networks") {
		t.Fatalf("expected resource tabs in ready browser shell content, got %q", view)
	}

	if !strings.Contains(view, "No resources found.") {
		t.Fatalf("expected ready empty-state content, got %q", view)
	}

	if !strings.Contains(view, "┃") {
		t.Fatalf("expected ready view to use injected shell border, got %q", view)
	}
}

func TestNewModelUsesDefaultStylesWhenNoneProvided(t *testing.T) {
	model := app.NewModel(app.Dependencies{})
	view := model.View()

	if view == "" {
		t.Fatal("expected default shell view to be non-empty")
	}

	if !strings.Contains(view, "dokbox-go") {
		t.Fatalf("expected default shell view to preserve content, got %q", view)
	}

	if view == "dokbox-go\n\nBootstrapping shell..." {
		t.Fatalf("expected default shell view to apply internal styling, got raw content %q", view)
	}

	if !strings.Contains(view, "╭") {
		t.Fatalf("expected default shell view to include framed shell border, got %q", view)
	}
}

func TestNewModelFillsViewportAfterWindowSizeInBootstrapView(t *testing.T) {
	model := app.NewModel(app.Dependencies{})

	nextModel, cmd := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd != nil {
		t.Fatal("expected window size update without follow-up command")
	}

	view := nextModel.View()
	if got := lipgloss.Width(view); got != 80 {
		t.Fatalf("expected bootstrap view width 80, got %d", got)
	}

	if got := lipgloss.Height(view); got != 24 {
		t.Fatalf("expected bootstrap view height 24, got %d", got)
	}
}

func TestNewModelFillsViewportAfterWindowSizeInReadyView(t *testing.T) {
	model := app.NewModel(app.Dependencies{})

	nextModel, cmd := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd != nil {
		t.Fatal("expected window size update without follow-up command")
	}

	startupMsg := nextModel.Init()()
	nextModel, cmd = nextModel.Update(startupMsg)
	if cmd != nil {
		t.Fatal("expected startup update to finish without follow-up command")
	}

	view := nextModel.View()
	if got := lipgloss.Width(view); got != 80 {
		t.Fatalf("expected ready view width 80, got %d", got)
	}

	if got := lipgloss.Height(view); got != 24 {
		t.Fatalf("expected ready view height 24, got %d", got)
	}
}
