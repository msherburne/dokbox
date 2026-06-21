package testsgo

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/msherburne/dokbox/internal/app"
	"github.com/msherburne/dokbox/internal/theme"
)

func TestNewStylesUsesThemeColors(t *testing.T) {
	styles := app.NewStyles(theme.Theme{
		Name:       "test",
		Background: "#0f0f0f",
		Text:       "#eeeeee",
	})

	if got := styles.Shell.GetForeground(); got != lipgloss.Color("#eeeeee") {
		t.Fatalf("expected shell foreground %q, got %q", "#eeeeee", got)
	}

	if got := styles.Shell.GetBackground(); got != lipgloss.Color("#0f0f0f") {
		t.Fatalf("expected shell background %q, got %q", "#0f0f0f", got)
	}
}

func TestNewStylesPreservesShellPadding(t *testing.T) {
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

	if got := styles.Shell.GetPaddingLeft(); got != 2 {
		t.Fatalf("expected left padding 2, got %d", got)
	}

	if got := styles.Shell.GetPaddingRight(); got != 2 {
		t.Fatalf("expected right padding 2, got %d", got)
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
	if !strings.Contains(view, "[Containers] Images Volumes Networks") {
		t.Fatalf("expected ready browser shell content, got %q", view)
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
}
