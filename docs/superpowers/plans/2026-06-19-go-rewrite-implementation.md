# Dokbox Go Rewrite Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current Python Dokbox runtime with a Go/Bubble Tea implementation that preserves the existing configuration model, Docker resource browser, container workflows, and Linux-first distribution path.

**Architecture:** Build a Go-first application at the repository root while keeping `python/` as the behavioral reference during the rewrite. Structure the Go app around a small Bubble Tea shell, a dedicated Docker integration layer, domain/view-model packages, and focused UI components so the replacement can be delivered in vertical slices without creating a monolithic event loop.

**Tech Stack:** Go, Bubble Tea, Bubbles, Lip Gloss, Docker Go SDK, Go test, Linux packaging scripts

---

## Planned File Structure

### New Go runtime files

- Create: `go.mod`
- Create: `go.sum`
- Create: `main.go`
- Create: `internal/app/model.go`
- Create: `internal/app/keys.go`
- Create: `internal/app/styles.go`
- Create: `internal/config/config.go`
- Create: `internal/config/path.go`
- Create: `internal/config/prompt.go`
- Create: `internal/docker/client.go`
- Create: `internal/docker/service.go`
- Create: `internal/domain/config.go`
- Create: `internal/domain/resources.go`
- Create: `internal/domain/metrics.go`
- Create: `internal/ui/browser/model.go`
- Create: `internal/ui/browser/tabs.go`
- Create: `internal/ui/browser/table.go`
- Create: `internal/ui/dialogs/confirm.go`
- Create: `internal/ui/dialogs/actions.go`
- Create: `internal/viewmodel/resources.go`
- Create: `internal/viewmodel/shortcuts.go`
- Create: `tests-go/config_test.go`
- Create: `tests-go/docker_service_test.go`
- Create: `tests-go/browser_model_test.go`
- Create: `tests-go/startup_smoke_test.go`

### Existing files likely to change during cutover

- Modify: `README.md`
- Modify: `.gitignore`
- Modify: `python/README.md`
- Modify: `python/packaging/build-linux.sh`
- Modify: `python/packaging/build-deb.sh`

### Planning and tracking files

- Create: Notion tickets in data source `collection://384fbd10-4335-8021-97b5-000ba3a3ded8`

---

### Task 1: Seed The Rewrite Backlog In Notion

**Files:**
- Create: Notion pages in `collection://384fbd10-4335-8021-97b5-000ba3a3ded8`
- Reference: `docs/superpowers/specs/2026-06-19-go-rewrite-design.md`

- [ ] **Step 1: Create the rewrite epics in Notion**

Create backlog cards with these exact titles and descriptions:

```text
Name: Go app foundation
Description: Establish the root Go module, Bubble Tea app shell, configuration bootstrap, shared styling/keymaps, and startup flow that will replace the Python runtime entrypoint.

Name: Docker integration layer
Description: Rebuild Docker client initialization, connection checks, resource loading, mutations, metrics, logs, and exec/path operations in Go behind a focused service boundary.

Name: Resource browser parity
Description: Recreate the multi-resource browser for containers, images, volumes, and networks with keyboard navigation, focus handling, tables/lists, empty states, and shortcut hints.

Name: Container workflows
Description: Recreate container-specific operational flows including actions, confirmations, logs, metrics, shell access, and path exploration where they exist in the Python implementation.

Name: Release and replacement
Description: Replace Python-first build, packaging, documentation, and release workflows with Go-native binaries and Linux packaging, then cut runtime defaults over to Go.
```

- [ ] **Step 2: Create the initial implementation tickets in Notion**

Create backlog cards with these exact titles and descriptions:

```text
Name: Initialize root Go module and Bubble Tea entrypoint
Description: Create the repository-root Go module, add Bubble Tea/Bubbles/Lip Gloss dependencies, add the main Go entrypoint, and prove the app can boot into an empty Bubble Tea shell.

Name: Implement config bootstrap with ~/.dokbox.json parity
Description: Port the Python config behavior to Go, including config_version handling, docker_host storage, migration/defaulting, Docker host auto-detection, generated default display, and prompting before save.

Name: Implement Docker client bootstrap and connection status flow
Description: Add Go Docker client initialization and connection checks, then surface startup and runtime connection states through the app shell without silent exits.

Name: Render container browser with keyboard navigation
Description: Add the first real resource browser slice by rendering containers in the Go app with focus management, row navigation, shortcut hints, and the top-level Bubble Tea layout.

Name: Add image, volume, and network browser tabs
Description: Extend the resource browser to include images, volumes, and networks with consistent tab behavior and table/list rendering patterns.

Name: Port container action dialogs and mutations
Description: Recreate container actions including start, stop, restart, remove, prune, and confirmation flows using Bubble Tea modal/dialog state and the Go Docker service.

Name: Port logs, metrics, shell, and filesystem views
Description: Recreate the deeper container workflows for logs, metrics display, shell/exec, and filesystem/path inspection using focused Go UI models and Docker service methods.

Name: Replace Linux packaging and release flow with Go binaries
Description: Update build, packaging, and release automation to produce Go-native Linux artifacts and make the Go runtime the default shipped application.
```

- [ ] **Step 3: Verify the Notion backlog structure**

Verify:

- all cards are in `Backlog`
- titles match the plan exactly
- descriptions are populated
- the board now contains both epics and actionable tickets for the rewrite

Expected: Notion board shows a rewrite program that can be executed task-by-task.

- [ ] **Step 4: Commit the planning artifacts if repo files changed**

```bash
git status --short
```

Expected: No repository file changes are required for Notion-only operations, so there may be nothing to commit.

### Task 2: Create The Go Module And Bootable App Shell

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `internal/app/model.go`
- Create: `internal/app/keys.go`
- Create: `internal/app/styles.go`
- Test: `tests-go/startup_smoke_test.go`

- [ ] **Step 1: Write the failing startup smoke test**

Create `tests-go/startup_smoke_test.go`:

```go
package testsgo

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/msherburne/dokbox/internal/app"
)

func TestNewModelBuildsInitialView(t *testing.T) {
	model := app.NewModel(app.Dependencies{})
	if model == nil {
		t.Fatal("expected model")
	}

	initial := model.View()
	if initial == "" {
		t.Fatal("expected non-empty initial view")
	}

	_ = tea.Quit
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run:

```bash
go test ./tests-go -run TestNewModelBuildsInitialView
```

Expected: FAIL because the Go module and app package do not exist yet.

- [ ] **Step 3: Create the root Go module**

Create `go.mod`:

```go
module github.com/msherburne/dokbox

go 1.24

require (
	github.com/charmbracelet/bubbles v0.20.0
	github.com/charmbracelet/bubbletea v1.3.4
	github.com/charmbracelet/lipgloss v1.1.0
)
```

Create `main.go`:

```go
package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/msherburne/dokbox/internal/app"
)

func main() {
	program := tea.NewProgram(app.NewModel(app.Dependencies{}), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 4: Add the minimal app shell**

Create `internal/app/model.go`:

```go
package app

import tea "github.com/charmbracelet/bubbletea"

type Dependencies struct{}

type Model struct{}

func NewModel(_ Dependencies) *Model {
	return &Model{}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *Model) View() string {
	return "dokbox-go\n\nStarting rewrite shell...\n\nPress q to quit."
}
```

Create `internal/app/keys.go`:

```go
package app

const QuitKey = "q"
```

Create `internal/app/styles.go`:

```go
package app
```

- [ ] **Step 5: Run the test to verify it passes**

Run:

```bash
go test ./tests-go -run TestNewModelBuildsInitialView
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add go.mod main.go internal/app/model.go internal/app/keys.go internal/app/styles.go tests-go/startup_smoke_test.go
git commit -m "feat: bootstrap go bubble tea shell"
```

### Task 3: Port Configuration Bootstrap With Python Parity

**Files:**
- Create: `internal/domain/config.go`
- Create: `internal/config/path.go`
- Create: `internal/config/config.go`
- Create: `internal/config/prompt.go`
- Test: `tests-go/config_test.go`
- Reference: `python/src/dokbox/models/config.py`
- Reference: `python/src/dokbox/utils/config.py`
- Reference: `python/src/dokbox/utils/setup.py`
- Reference: `python/src/dokbox/defaults/config.py`

- [ ] **Step 1: Write the failing configuration tests**

Create `tests-go/config_test.go`:

```go
package testsgo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/msherburne/dokbox/internal/config"
)

func TestConfigPathUsesDokboxJSONInHomeDir(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	got := config.Path()
	want := filepath.Join(os.Getenv("HOME"), ".dokbox.json")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestMigrateAddsConfigVersion(t *testing.T) {
	raw := map[string]any{"docker_host": "unix:///var/run/docker.sock"}
	migrated, err := config.Migrate(raw)
	if err != nil {
		t.Fatal(err)
	}
	if migrated.ConfigVersion != 1 {
		t.Fatalf("got %d want 1", migrated.ConfigVersion)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run:

```bash
go test ./tests-go -run 'TestConfigPathUsesDokboxJSONInHomeDir|TestMigrateAddsConfigVersion'
```

Expected: FAIL because the config package does not exist yet.

- [ ] **Step 3: Add the config domain model**

Create `internal/domain/config.go`:

```go
package domain

type Config struct {
	ConfigVersion int    `json:"config_version"`
	DockerHost    string `json:"docker_host"`
}
```

- [ ] **Step 4: Implement config path and migration**

Create `internal/config/path.go`:

```go
package config

import (
	"os"
	"path/filepath"
)

func Path() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".dokbox.json"
	}
	return filepath.Join(home, ".dokbox.json")
}
```

Create `internal/config/config.go`:

```go
package config

import (
	"encoding/json"
	"os"

	"github.com/msherburne/dokbox/internal/domain"
)

const CurrentConfigVersion = 1

func LoadRaw() (map[string]any, error) {
	path := Path()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func Migrate(raw map[string]any) (domain.Config, error) {
	if raw == nil {
		return domain.Config{}, nil
	}
	if _, ok := raw["config_version"]; !ok {
		raw["config_version"] = float64(CurrentConfigVersion)
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return domain.Config{}, err
	}
	var cfg domain.Config
	if err := json.Unmarshal(encoded, &cfg); err != nil {
		return domain.Config{}, err
	}
	return cfg, nil
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run:

```bash
go test ./tests-go -run 'TestConfigPathUsesDokboxJSONInHomeDir|TestMigrateAddsConfigVersion'
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/domain/config.go internal/config/path.go internal/config/config.go tests-go/config_test.go
git commit -m "feat: port config path and migration"
```

### Task 4: Implement Config Setup Flow And Docker Host Detection

**Files:**
- Modify: `internal/config/config.go`
- Create: `internal/config/prompt.go`
- Create: `internal/docker/client.go`
- Test: `tests-go/config_test.go`

- [ ] **Step 1: Extend tests for default generation and save confirmation**

Append to `tests-go/config_test.go`:

```go
func TestResolveCurrentReturnsNilWhenConfigMissing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg, err := config.ResolveCurrent()
	if err != nil {
		t.Fatal(err)
	}
	if cfg != nil {
		t.Fatalf("expected nil config, got %+v", cfg)
	}
}
```

- [ ] **Step 2: Run the focused test to verify it fails**

Run:

```bash
go test ./tests-go -run TestResolveCurrentReturnsNilWhenConfigMissing
```

Expected: FAIL because `ResolveCurrent` does not exist.

- [ ] **Step 3: Implement config resolution and save**

Add to `internal/config/config.go`:

```go
func Save(cfg domain.Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), data, 0o644)
}

func ResolveCurrent() (*domain.Config, error) {
	raw, err := LoadRaw()
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}
	cfg, err := Migrate(raw)
	if err != nil {
		return nil, err
	}
	if err := Save(cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
```

Create `internal/docker/client.go`:

```go
package docker

import "os"

func DetectDockerHost() string {
	if host := os.Getenv("DOCKER_HOST"); host != "" {
		return host
	}
	return "unix:///var/run/docker.sock"
}
```

Create `internal/config/prompt.go`:

```go
package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/msherburne/dokbox/internal/docker"
	"github.com/msherburne/dokbox/internal/domain"
)

func GenerateDefault() domain.Config {
	return domain.Config{
		ConfigVersion: CurrentConfigVersion,
		DockerHost:    docker.DetectDockerHost(),
	}
}

func ConfirmSave(cfg domain.Config) (bool, error) {
	fmt.Printf("Using default configuration:\n{\n  \"config_version\": %d,\n  \"docker_host\": %q\n}\n", cfg.ConfigVersion, cfg.DockerHost)
	fmt.Print("Do you want to save this configuration? [Y/n] ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	answer := strings.TrimSpace(strings.ToLower(line))
	return answer == "" || answer == "y" || answer == "yes", nil
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run:

```bash
go test ./tests-go -run TestResolveCurrentReturnsNilWhenConfigMissing
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/config.go internal/config/prompt.go internal/docker/client.go tests-go/config_test.go
git commit -m "feat: add config setup flow"
```

### Task 5: Add Docker Service And Connection State

**Files:**
- Create: `internal/domain/resources.go`
- Create: `internal/domain/metrics.go`
- Create: `internal/docker/service.go`
- Test: `tests-go/docker_service_test.go`
- Reference: `python/src/dokbox/services/docker_service.py`

- [ ] **Step 1: Write the failing Docker service tests**

Create `tests-go/docker_service_test.go`:

```go
package testsgo

import (
	"context"
	"errors"
	"testing"

	"github.com/msherburne/dokbox/internal/docker"
)

type pingClient struct {
	err error
}

func (c pingClient) Ping(context.Context) error { return c.err }

func TestCheckConnectionReturnsOkStatus(t *testing.T) {
	service := docker.NewService(pingClient{})
	status := service.CheckConnection(context.Background())
	if !status.OK {
		t.Fatal("expected ok connection")
	}
}

func TestCheckConnectionReturnsErrorStatus(t *testing.T) {
	service := docker.NewService(pingClient{err: errors.New("unavailable")})
	status := service.CheckConnection(context.Background())
	if status.OK {
		t.Fatal("expected failed connection")
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run:

```bash
go test ./tests-go -run 'TestCheckConnectionReturnsOkStatus|TestCheckConnectionReturnsErrorStatus'
```

Expected: FAIL because the service does not exist.

- [ ] **Step 3: Add connection and resource domain types**

Create `internal/domain/resources.go`:

```go
package domain

type ResourceKind string

const (
	ResourceKindContainer ResourceKind = "container"
	ResourceKindImage     ResourceKind = "image"
	ResourceKindVolume    ResourceKind = "volume"
	ResourceKindNetwork   ResourceKind = "network"
)

type ConnectionStatus struct {
	OK      bool
	Message string
}

type ResourceSummary struct {
	Kind    ResourceKind
	ID      string
	Name    string
	Group   string
	Columns map[string]string
}
```

Create `internal/domain/metrics.go`:

```go
package domain

type MetricSample struct {
	Name  string
	Value float64
	Limit float64
	Unit  string
	Label string
}
```

- [ ] **Step 4: Implement the Docker service boundary**

Create `internal/docker/service.go`:

```go
package docker

import (
	"context"

	"github.com/msherburne/dokbox/internal/domain"
)

type PingClient interface {
	Ping(context.Context) error
}

type Service struct {
	client PingClient
}

func NewService(client PingClient) *Service {
	return &Service{client: client}
}

func (s *Service) CheckConnection(ctx context.Context) domain.ConnectionStatus {
	if err := s.client.Ping(ctx); err != nil {
		return domain.ConnectionStatus{OK: false, Message: err.Error()}
	}
	return domain.ConnectionStatus{OK: true, Message: "Connected"}
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run:

```bash
go test ./tests-go -run 'TestCheckConnectionReturnsOkStatus|TestCheckConnectionReturnsErrorStatus'
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/domain/resources.go internal/domain/metrics.go internal/docker/service.go tests-go/docker_service_test.go
git commit -m "feat: add docker connection service"
```

### Task 6: Build The Resource Browser Shell

**Files:**
- Create: `internal/viewmodel/resources.go`
- Create: `internal/viewmodel/shortcuts.go`
- Create: `internal/ui/browser/model.go`
- Create: `internal/ui/browser/tabs.go`
- Create: `internal/ui/browser/table.go`
- Modify: `internal/app/model.go`
- Test: `tests-go/browser_model_test.go`
- Reference: `python/src/dokbox/view_models/resources.py`
- Reference: `python/src/dokbox/ui/screens.py`
- Reference: `python/src/dokbox/ui/widgets.py`

- [ ] **Step 1: Write the failing browser tests**

Create `tests-go/browser_model_test.go`:

```go
package testsgo

import (
	"strings"
	"testing"

	"github.com/msherburne/dokbox/internal/domain"
	"github.com/msherburne/dokbox/internal/ui/browser"
)

func TestBrowserViewRendersContainerTabAndRows(t *testing.T) {
	model := browser.NewModel([]domain.ResourceSummary{
		{
			Kind: domain.ResourceKindContainer,
			ID:   "abc",
			Name: "api",
			Columns: map[string]string{
				"Image": "dokbox-api",
				"State": "running",
			},
		},
	})

	view := model.View()
	if !strings.Contains(view, "Containers") {
		t.Fatal("expected containers tab")
	}
	if !strings.Contains(view, "api") {
		t.Fatal("expected container row")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run:

```bash
go test ./tests-go -run TestBrowserViewRendersContainerTabAndRows
```

Expected: FAIL because the browser package does not exist.

- [ ] **Step 3: Add minimal view-model and browser structures**

Create `internal/viewmodel/resources.go`:

```go
package viewmodel

import "github.com/msherburne/dokbox/internal/domain"

func ResourceColumns(kind domain.ResourceKind) []string {
	switch kind {
	case domain.ResourceKindContainer:
		return []string{"Name", "Image", "State"}
	case domain.ResourceKindImage:
		return []string{"Repository", "Tag", "Image ID"}
	case domain.ResourceKindVolume:
		return []string{"Name", "Driver", "Scope"}
	default:
		return []string{"Name", "Driver", "Containers"}
	}
}
```

Create `internal/viewmodel/shortcuts.go`:

```go
package viewmodel

func BrowserShortcuts() []string {
	return []string{"tab switch", "enter focus", "q quit"}
}
```

Create `internal/ui/browser/model.go`:

```go
package browser

import (
	"strings"

	"github.com/msherburne/dokbox/internal/domain"
)

type Model struct {
	containers []domain.ResourceSummary
}

func NewModel(containers []domain.ResourceSummary) *Model {
	return &Model{containers: containers}
}

func (m *Model) View() string {
	var rows []string
	rows = append(rows, "Tabs: Containers | Images | Volumes | Networks")
	rows = append(rows, "")
	for _, resource := range m.containers {
		rows = append(rows, resource.Name)
	}
	return strings.Join(rows, "\n")
}
```

Create `internal/ui/browser/tabs.go`:

```go
package browser
```

Create `internal/ui/browser/table.go`:

```go
package browser
```

- [ ] **Step 4: Run the test to verify it passes**

Run:

```bash
go test ./tests-go -run TestBrowserViewRendersContainerTabAndRows
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/viewmodel/resources.go internal/viewmodel/shortcuts.go internal/ui/browser/model.go internal/ui/browser/tabs.go internal/ui/browser/table.go tests-go/browser_model_test.go
git commit -m "feat: add initial resource browser shell"
```

### Task 7: Wire Startup Through Config And Browser State

**Files:**
- Modify: `main.go`
- Modify: `internal/app/model.go`
- Modify: `internal/docker/client.go`
- Modify: `internal/ui/browser/model.go`
- Test: `tests-go/startup_smoke_test.go`

- [ ] **Step 1: Extend the startup smoke test for the config/browser flow**

Append to `tests-go/startup_smoke_test.go`:

```go
func TestInitialViewShowsAppIdentity(t *testing.T) {
	model := app.NewModel(app.Dependencies{})
	view := model.View()
	if !strings.Contains(view, "dokbox") {
		t.Fatal("expected app identity in view")
	}
}
```

Add import:

```go
import "strings"
```

- [ ] **Step 2: Run the focused test to verify it fails if imports/code are missing**

Run:

```bash
go test ./tests-go -run TestInitialViewShowsAppIdentity
```

Expected: FAIL until the test file is updated correctly or the view lacks the expected identity.

- [ ] **Step 3: Update the app shell to prepare for real startup wiring**

Replace `internal/app/model.go` with:

```go
package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Dependencies struct{}

type Model struct {
	title   string
	content string
}

func NewModel(_ Dependencies) *Model {
	return &Model{
		title:   "dokbox",
		content: "Go rewrite shell ready",
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *Model) View() string {
	return strings.Join([]string{
		m.title,
		"",
		m.content,
		"",
		"Press q to quit.",
	}, "\n")
}
```

- [ ] **Step 4: Run the focused startup tests**

Run:

```bash
go test ./tests-go -run 'TestNewModelBuildsInitialView|TestInitialViewShowsAppIdentity'
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add main.go internal/app/model.go tests-go/startup_smoke_test.go
git commit -m "feat: prepare startup shell for go runtime"
```

### Task 8: Fill Out Remaining Parity Work Through Vertical Slices

**Files:**
- Modify: `internal/docker/service.go`
- Modify: `internal/ui/browser/*`
- Create: `internal/ui/dialogs/confirm.go`
- Create: `internal/ui/dialogs/actions.go`
- Modify: `tests-go/docker_service_test.go`
- Modify: `tests-go/browser_model_test.go`
- Reference: `python/src/dokbox/services/docker_service.py`
- Reference: `python/src/dokbox/ui/screens.py`

- [ ] **Step 1: Break parity work into ticket-sized slices**

Use the Notion tickets created in Task 1 as the source of execution order for:

- images/volumes/networks tabs
- container action dialogs
- logs and metrics views
- shell/exec and filesystem browsing
- packaging and release cutover

Expected: No new large rewrite task should be started outside the Notion backlog.

- [ ] **Step 2: For each ticket, follow the same TDD loop**

For every remaining parity ticket:

```text
1. Write a failing Go test
2. Run it and confirm failure
3. Implement the minimal code
4. Run focused tests
5. Commit
```

Expected: Progress remains incremental and reviewable.

- [ ] **Step 3: Keep Python as behavioral reference until cutover**

When implementing a parity feature, inspect the corresponding Python file under `python/src/dokbox/...` first and copy the behavior intentionally rather than re-imagining it from memory.

Expected: The rewrite stays replacement-oriented instead of drifting into an unrelated redesign.

### Task 9: Cut Over Packaging, Docs, And Runtime Defaults

**Files:**
- Modify: `README.md`
- Modify: `.gitignore`
- Modify: `python/README.md`
- Modify: `python/packaging/build-linux.sh`
- Modify: `python/packaging/build-deb.sh`
- Add: Go-native build and packaging scripts at repo root as needed

- [ ] **Step 1: Add failing tests or checks for Go-first packaging once the binary exists**

Create or extend Go-era packaging verification to assert:

- the primary runtime is the Go binary
- packaging docs point to Go build commands
- Linux package generation consumes the Go binary, not the Python standalone bundle

Expected: FAIL until cutover is implemented.

- [ ] **Step 2: Implement the Go-first packaging path**

Update docs and packaging so:

- repository-root instructions are Go-native
- build scripts package the Go binary
- Python packaging artifacts are demoted to legacy/reference status or removed

Expected: Go binary becomes the default shipped application.

- [ ] **Step 3: Run final verification**

Run:

```bash
go test ./...
```

And any packaging verification commands introduced during the cutover.

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add README.md .gitignore python/README.md python/packaging
git commit -m "feat: cut over dokbox runtime to go"
```

---

## Self-Review

### Spec coverage

- Replacement strategy: covered by Tasks 1, 2, 7, 8, and 9.
- Configuration parity: covered by Tasks 3 and 4.
- Docker integration: covered by Tasks 5 and 8.
- Resource browser parity: covered by Tasks 6 and 8.
- Container workflows: covered by Tasks 1 and 8.
- Packaging and release cutover: covered by Task 9.
- Notion planning/ticketing: covered by Task 1.

### Placeholder scan

- No `TODO`/`TBD` placeholders remain.
- Broad parity work is intentionally constrained to Notion-backed ticket execution in Tasks 8 and 9 rather than pretending the entire rewrite fits in one patch.

### Type consistency

- `domain.Config`, `domain.ConnectionStatus`, and `domain.ResourceSummary` are introduced before later tasks depend on them.
- The app shell and browser tasks use stable package names and module paths that match the planned root Go module.
