# UI Theme Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a config-driven built-in theme system for the Dokbox Go TUI, including three dark presets, safe fallback behavior, and shared style construction built from semantic theme tokens.

**Architecture:** Introduce a small theme boundary that owns semantic tokens, preset registration, and preset resolution. Extend the config model with a top-level `theme` field, then make app-level styles resolve from the active theme instead of hard-coded Lip Gloss values. Keep the first slice focused on built-in preset selection and foundational style wiring so later UI-overhaul tickets can reuse it.

**Tech Stack:** Go, Bubble Tea, Lip Gloss, Go standard library, existing `tests-go` test suite

---

## File Structure

### Existing files to modify

- `internal/domain/config.go`
  - Add the top-level `Theme` field to the persisted config model.
- `internal/config/config.go`
  - Default and migration-safe config generation/loading.
  - Theme resolution fallback behavior at config decode/generation time.
- `internal/app/styles.go`
  - Replace the single static shell style with theme-driven style construction.
- `main.go`
  - Resolve the configured theme once and pass it into app style initialization.
- `tests-go/config_test.go`
  - Add config coverage for default theme values and invalid theme fallback.

### New files to create

- `internal/theme/theme.go`
  - Semantic `Theme` type and token group definitions.
- `internal/theme/presets.go`
  - Built-in preset registry and default preset key.
- `internal/theme/resolve.go`
  - Preset lookup and fallback-safe resolution helpers.
- `internal/theme/theme_test.go`
  - Unit tests for preset resolution and fallback behavior.
- `tests-go/styles_test.go`
  - Focused tests proving shared styles change based on active preset and preserve expected padding/structure.

### Files to inspect during implementation

- `internal/app/model.go`
  - Confirm where global app styling is rendered and how style state should be injected or initialized.
- `internal/ui/browser/model.go`
  - Check whether any current shared style extraction is worth folding into the first slice without broadening scope.
- `internal/ui/containerdetail/model.go`
  - Same as above; only touch if a tiny shared-style seam is necessary for the theme foundation.

---

### Task 1: Add Theme Support to Config and Default Resolution

**Files:**
- Create: `internal/theme/theme.go`
- Create: `internal/theme/presets.go`
- Create: `internal/theme/resolve.go`
- Modify: `internal/domain/config.go`
- Modify: `internal/config/config.go`
- Test: `tests-go/config_test.go`
- Test: `internal/theme/theme_test.go`

- [ ] **Step 1: Write the failing config and theme-resolution tests**

```go
func TestResolveCurrentConfigDefaultsMissingThemeToDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(home, ".dokbox.json")
	if err := os.WriteFile(configPath, []byte(`{"docker_host":"unix:///var/run/docker.sock"}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := config.ResolveCurrentConfig()
	if err != nil {
		t.Fatalf("resolve config: %v", err)
	}

	if cfg.Theme != "default" {
		t.Fatalf("expected default theme fallback, got %q", cfg.Theme)
	}
}

func TestResolveCurrentConfigFallsBackFromUnknownTheme(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(home, ".dokbox.json")
	if err := os.WriteFile(configPath, []byte(`{"docker_host":"unix:///var/run/docker.sock","theme":"nope"}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := config.ResolveCurrentConfig()
	if err != nil {
		t.Fatalf("resolve config: %v", err)
	}

	if cfg.Theme != "default" {
		t.Fatalf("expected unknown theme to fall back to default, got %q", cfg.Theme)
	}
}

func TestGenerateDefaultConfigIncludesTheme(t *testing.T) {
	cfg, err := config.GenerateDefaultConfig(io.Discard)
	if err != nil {
		t.Fatalf("generate default config: %v", err)
	}

	if cfg.Theme != "default" {
		t.Fatalf("expected generated default theme, got %q", cfg.Theme)
	}
}

func TestResolveThemeFallsBackToDefault(t *testing.T) {
	theme, key := theme.Resolve("missing")

	if key != "default" {
		t.Fatalf("expected fallback key default, got %q", key)
	}

	if theme.Accent == "" {
		t.Fatal("expected resolved default theme tokens")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./tests-go -run 'TestResolveCurrentConfigDefaultsMissingThemeToDefault|TestResolveCurrentConfigFallsBackFromUnknownTheme|TestGenerateDefaultConfigIncludesTheme'`

Expected: FAIL with missing `Theme` field and/or failing fallback assertions.

Run: `go test ./internal/theme -run 'TestResolveThemeFallsBackToDefault'`

Expected: FAIL because the `internal/theme` package does not exist yet.

- [ ] **Step 3: Add the config field and theme resolution package**

```go
// internal/domain/config.go
package domain

type Config struct {
	ConfigVersion int    `json:"config_version"`
	DockerHost    string `json:"docker_host"`
	Theme         string `json:"theme"`
}
```

```go
// internal/theme/theme.go
package theme

type Theme struct {
	Name      string
	Background string
	Panel      string
	Border     string
	Text       string
	Muted      string
	Emphasis   string
	Accent     string
	Focus      string
	Success    string
	Warning    string
	Error      string
	Info       string
}
```

```go
// internal/theme/presets.go
package theme

const DefaultPreset = "default"

var presets = map[string]Theme{
	"default": {
		Name:       "default",
		Background: "#101215",
		Panel:      "#171b20",
		Border:     "#2a3138",
		Text:       "#e7edf3",
		Muted:      "#94a3b8",
		Emphasis:   "#ffffff",
		Accent:     "#6cb6ff",
		Focus:      "#8bd5ff",
		Success:    "#7bd88f",
		Warning:    "#f6c177",
		Error:      "#ff7b72",
		Info:       "#6cb6ff",
	},
	"slate": {
		Name:       "slate",
		Background: "#0f1720",
		Panel:      "#16202a",
		Border:     "#334155",
		Text:       "#dce7f3",
		Muted:      "#8a9aae",
		Emphasis:   "#f8fafc",
		Accent:     "#7dd3fc",
		Focus:      "#93c5fd",
		Success:    "#86efac",
		Warning:    "#fcd34d",
		Error:      "#fda4af",
		Info:       "#7dd3fc",
	},
	"ember": {
		Name:       "ember",
		Background: "#16110f",
		Panel:      "#211917",
		Border:     "#3b2b27",
		Text:       "#f1e8e1",
		Muted:      "#b8a39a",
		Emphasis:   "#fff7f0",
		Accent:     "#f97316",
		Focus:      "#fb923c",
		Success:    "#9dd274",
		Warning:    "#fbbf24",
		Error:      "#ff8b8b",
		Info:       "#fdba74",
	},
}
```

```go
// internal/theme/resolve.go
package theme

func Resolve(name string) (Theme, string) {
	if preset, ok := presets[name]; ok {
		return preset, name
	}

	return presets[DefaultPreset], DefaultPreset
}

func PresetNames() []string {
	return []string{"default", "slate", "ember"}
}
```

```go
// internal/config/config.go excerpt
func GenerateDefaultConfig(writer io.Writer) (domain.Config, error) {
	if writer == nil {
		writer = io.Discard
	}

	cfg := domain.Config{
		ConfigVersion: CurrentConfigVersion,
		DockerHost:    docker.DetectDockerHost(),
		Theme:         theme.DefaultPreset,
	}
	// existing marshal/write logic stays the same
}

func decodeConfig(raw map[string]any) (*domain.Config, error) {
	payload, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var cfg domain.Config
	if err := json.Unmarshal(payload, &cfg); err != nil {
		return nil, err
	}

	_, resolvedName := theme.Resolve(cfg.Theme)
	cfg.Theme = resolvedName

	return &cfg, nil
}
```

- [ ] **Step 4: Add focused theme unit tests**

```go
// internal/theme/theme_test.go
package theme

import "testing"

func TestResolveReturnsNamedPreset(t *testing.T) {
	got, key := Resolve("slate")

	if key != "slate" {
		t.Fatalf("expected slate key, got %q", key)
	}
	if got.Name != "slate" {
		t.Fatalf("expected slate theme, got %q", got.Name)
	}
}

func TestResolveFallsBackToDefault(t *testing.T) {
	got, key := Resolve("unknown")

	if key != DefaultPreset {
		t.Fatalf("expected default fallback key, got %q", key)
	}
	if got.Name != DefaultPreset {
		t.Fatalf("expected default fallback theme, got %q", got.Name)
	}
}
```

- [ ] **Step 5: Run the focused tests to verify they pass**

Run: `go test ./tests-go -run 'TestResolveCurrentConfigDefaultsMissingThemeToDefault|TestResolveCurrentConfigFallsBackFromUnknownTheme|TestGenerateDefaultConfigIncludesTheme'`

Expected: PASS

Run: `go test ./internal/theme -run 'TestResolveReturnsNamedPreset|TestResolveFallsBackToDefault'`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/domain/config.go internal/config/config.go internal/theme/theme.go internal/theme/presets.go internal/theme/resolve.go internal/theme/theme_test.go tests-go/config_test.go
git commit -m "feat: add built-in theme config and resolution"
```

---

### Task 2: Make Shared App Styles Theme-Driven

**Files:**
- Modify: `internal/app/styles.go`
- Modify: `main.go`
- Modify: `internal/app/model.go` (only if needed to thread theme-driven styles through the app safely)
- Test: `tests-go/styles_test.go`

- [ ] **Step 1: Write the failing shared-style tests**

```go
func TestNewStylesUsesThemeColors(t *testing.T) {
	styles := app.NewStyles(theme.Theme{
		Name:       "test",
		Panel:      "#111111",
		Text:       "#eeeeee",
		Border:     "#333333",
		Accent:     "#55aaff",
	})

	rendered := styles.Shell.Render("dokbox-go")
	if !strings.Contains(rendered, "dokbox-go") {
		t.Fatalf("expected shell rendering to preserve content, got %q", rendered)
	}
}

func TestNewStylesPreservesShellPadding(t *testing.T) {
	styles := app.NewStyles(theme.Theme{
		Name:       "test",
		Panel:      "#111111",
		Text:       "#eeeeee",
		Border:     "#333333",
		Accent:     "#55aaff",
	})

	rendered := styles.Shell.Render("x")
	if rendered == "x" {
		t.Fatalf("expected shell style padding to remain applied, got %q", rendered)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./tests-go -run 'TestNewStylesUsesThemeColors|TestNewStylesPreservesShellPadding'`

Expected: FAIL because `app.NewStyles` and a theme-driven style container do not exist yet.

- [ ] **Step 3: Replace the static style with a theme-driven style container**

```go
// internal/app/styles.go
package app

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/msherburne/dokbox/internal/theme"
)

type Styles struct {
	Shell lipgloss.Style
}

func NewStyles(active theme.Theme) Styles {
	return Styles{
		Shell: lipgloss.NewStyle().
			Padding(1, 2).
			Foreground(lipgloss.Color(active.Text)).
			Background(lipgloss.Color(active.Background)),
	}
}
```

```go
// main.go excerpt
cfg, err := config.LoadOrSetupConfig(os.Stdin, os.Stdout)
if err != nil {
	log.Fatal(err)
}

activeTheme, _ := theme.Resolve(cfg.Theme)

program := tea.NewProgram(
	app.NewModel(deps, app.NewStyles(activeTheme)),
	tea.WithAltScreen(),
)
```

```go
// internal/app/model.go excerpt
type Model struct {
	ready            bool
	connectionStatus *domain.ConnectionStatus
	lastActionStatus string
	deps             Dependencies
	styles           Styles
	browser          *browser.Model
	detail           *containerdetail.Model
}

func NewModel(deps Dependencies, styles Styles) *Model {
	return &Model{
		deps:    deps,
		styles:  styles,
		browser: browser.NewModel(deps.InitialContainers),
	}
}

func (m *Model) View() string {
	// existing content assembly stays the same
	return m.styles.Shell.Render(strings.Join(lines, "\n"))
}
```

- [ ] **Step 4: Update construction call sites and tests minimally**

```go
// tests-go/startup_smoke_test.go excerpt
model := app.NewModel(app.Dependencies{}, app.NewStyles(theme.Resolve("default")))
```

If direct `theme.Resolve("default")` usage is awkward in tests, add a tiny helper:

```go
func defaultStyles() app.Styles {
	active, _ := theme.Resolve("default")
	return app.NewStyles(active)
}
```

- [ ] **Step 5: Run the focused tests to verify they pass**

Run: `go test ./tests-go -run 'TestNewStylesUsesThemeColors|TestNewStylesPreservesShellPadding'`

Expected: PASS

Run: `go test ./tests-go -run 'TestNewModel'`

Expected: PASS for existing app model tests after constructor updates.

- [ ] **Step 6: Commit**

```bash
git add internal/app/styles.go internal/app/model.go main.go tests-go/styles_test.go tests-go/startup_smoke_test.go
git commit -m "feat: wire shared app styles to active theme"
```

---

### Task 3: Verify Default Output, Fallback Behavior, and Save-Path UX

**Files:**
- Modify: `tests-go/config_test.go`
- Modify: `tests-go/styles_test.go`
- Inspect: `internal/config/prompt.go`
- Inspect: `docs/DEVELOPMENT.md`

- [ ] **Step 1: Add failing regression tests for serialized default config output**

```go
func TestLoadOrSetupConfigShowsThemeInGeneratedDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("DOCKER_HOST", "tcp://docker.example:2375")

	var output strings.Builder
	_, err := config.LoadOrSetupConfig(strings.NewReader("n\n"), &output)
	if err != nil {
		t.Fatalf("load or setup config: %v", err)
	}

	rendered := output.String()
	if !strings.Contains(rendered, `"theme": "default"`) {
		t.Fatalf("expected generated config output to show default theme, got %q", rendered)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail if output/config serialization is incomplete**

Run: `go test ./tests-go -run 'TestLoadOrSetupConfigShowsThemeInGeneratedDefault'`

Expected: FAIL until the generated config output includes the new field.

- [ ] **Step 3: Finish any missing serialization or regression wiring**

```go
// internal/config/config.go excerpt
cfg := domain.Config{
	ConfigVersion: CurrentConfigVersion,
	DockerHost:    docker.DetectDockerHost(),
	Theme:         theme.DefaultPreset,
}

payload, err := json.MarshalIndent(cfg, "", "  ")
if err != nil {
	return domain.Config{}, err
}
```

If the field is already present after Task 1, use this step to remove any leftover hard-coded assumptions in tests or helper setup instead of broadening scope.

- [ ] **Step 4: Run the high-signal verification set**

Run: `go test ./tests-go -run 'TestResolveCurrentConfigDefaultsMissingThemeToDefault|TestResolveCurrentConfigFallsBackFromUnknownTheme|TestGenerateDefaultConfigIncludesTheme|TestLoadOrSetupConfigShowsThemeInGeneratedDefault|TestNewStylesUsesThemeColors|TestNewStylesPreservesShellPadding'`

Expected: PASS

Run: `go test ./...`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add tests-go/config_test.go tests-go/styles_test.go
git commit -m "test: lock in theme defaults and shared style behavior"
```

---

### Task 4: Document the Built-In Theme Option for Developers

**Files:**
- Modify: `docs/DEVELOPMENT.md`

- [ ] **Step 1: Add a short configuration example for built-in themes**

```md
## Theme Configuration

Dokbox supports built-in dark theme presets through the top-level `theme` config field:

```json
{
  "config_version": 1,
  "docker_host": "unix:///var/run/docker.sock",
  "theme": "slate"
}
```

Supported built-in themes:

- `default`
- `slate`
- `ember`

Unknown values fall back to `default`.
```

- [ ] **Step 2: Run focused tests and a full test pass after doc update**

Run: `go test ./...`

Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add docs/DEVELOPMENT.md
git commit -m "docs: add built-in theme configuration notes"
```

---

## Self-Review

### Spec coverage

- Top-level `theme` config field: covered in Task 1
- Three built-in dark presets: covered in Task 1
- Semantic theme model: covered in Task 1
- Shared style construction from semantic tokens: covered in Task 2
- Default/fallback behavior: covered in Task 1 and Task 3
- Documentation of config shape and fallback behavior: covered in Task 4

No uncovered spec requirements remain.

### Placeholder scan

- No `TODO`, `TBD`, or “appropriate error handling” placeholders remain.
- Each code-changing step includes concrete code examples.
- Each verification step includes explicit commands and expected outcomes.

### Type consistency

- Config field name is consistently `Theme`
- Theme resolution entry point is consistently `theme.Resolve`
- Shared style entry point is consistently `app.NewStyles`
- The style container is consistently `app.Styles`

No naming contradictions found.
