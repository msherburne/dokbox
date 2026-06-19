# Dokbox Go Rewrite Design

## Summary

Dokbox will be rewritten from Python to Go using Bubble Tea as the terminal application framework. The Go application is intended to replace the current Python runtime, not coexist with it as a permanent dual implementation. The existing Python codebase, now moved under `python/`, serves as the functional reference during the rewrite.

The rewrite will be managed as a legitimate backlog in Notion using a small set of epics and concrete implementation tickets. Work should proceed in vertical slices so the Go app becomes runnable and progressively useful early, while still driving toward complete replacement of the current runtime.

## Goals

- Replace the current Python/Textual runtime with a Go/Bubble Tea runtime.
- Preserve the current user-facing capabilities of Dokbox as closely as practical.
- Improve the long-term Linux packaging and distribution story by producing a native Go binary.
- Use Notion as the planning and ticketing surface for the rewrite.
- Keep the rewrite organized around meaningful engineering milestones rather than vague rewrite tasks.

## Non-Goals

- Maintaining both Python and Go runtimes as equal long-term products.
- Achieving pixel-perfect parity with Textual-specific rendering behavior.
- Re-architecting the product into a fundamentally different workflow during the first rewrite pass.
- Building unrelated new features before the Go rewrite reaches practical parity.

## Recommended Technical Direction

The Go stack should use:

- `bubbletea` for the main application runtime and state/event loop.
- `bubbles` for common UI primitives such as tables, lists, text inputs, help, viewport, and spinners.
- `lipgloss` for styling and layout composition.
- The Go Docker SDK for Docker daemon communication.
- Plain Go structs for domain models and UI-facing view models.

This choice favors long-term product quality and a strong native terminal experience while staying within a language the team already knows.

## Replacement Strategy

The rewrite is a replacement program with staged delivery:

1. Keep the Python implementation in `python/` as the source of truth for current behavior.
2. Make the repository root Go-first as new implementation work begins.
3. Build the Go app in vertical slices so it becomes runnable and progressively closer to parity.
4. Switch packaging, release, and runtime defaults to Go once the Go application reaches agreed practical parity.
5. Retire the Python runtime from active use after the Go path is release-ready.

The rewrite should not wait until every low-level internal detail is perfect before the Go app becomes useful. However, the final product goal remains full runtime replacement rather than indefinite staged coexistence.

## Functional Scope

The rewrite target is “everything we currently have, re-implemented in Go,” organized into the following capability areas:

- Application shell and startup flow
- Configuration/bootstrap behavior
- Docker connection setup and validation
- Resource browsing for containers, images, volumes, and networks
- Keyboard navigation and tab switching
- Container actions and confirmation flows
- Logs, details, metrics, shell/exec, and filesystem/path exploration where currently supported
- Packaging and Linux-native release workflows

If implementation uncovers unclear behavior in the Python app, the Go version should follow the Python runtime unless there is a compelling reason to intentionally simplify or improve the behavior.

## Proposed Go Architecture

### 1. App Shell

Create a top-level Bubble Tea model that owns:

- global app state
- current active screen/pane
- focus state
- global keybindings
- modal/dialog state
- shared error/status messaging

This top-level model acts as the coordinator rather than embedding Docker logic directly.

### 2. Domain Layer

Create Go domain models for Docker resources and runtime state:

- containers
- images
- volumes
- networks
- connection status
- metrics samples
- log lines
- shell/exec sessions

These should map closely to the current Python resource and model structures, but use idiomatic Go types.

### 3. Docker Integration Layer

Create a focused service layer responsible for:

- building/configuring the Docker client
- checking daemon connectivity
- listing resources
- retrieving details
- performing mutations like start/stop/restart/remove/prune
- streaming logs
- opening shell/exec sessions
- retrieving metrics and filesystem/path data

The Bubble Tea models should depend on this layer rather than the Docker SDK directly.

### 4. View-Model Layer

Create UI-facing adapters that transform raw Docker/domain data into:

- table rows
- tab labels
- command palette/search entries
- status labels
- metric cards or metric display values
- shortcut/help text

This mirrors the existing Python separation and keeps rendering decisions out of the service layer.

### 5. UI Layer

Use Bubble Tea with Bubbles/Lip Gloss to build:

- root app shell
- resource browser tabs
- resource tables/lists
- command palette
- action dialogs
- details/logs/metrics panes
- shortcut/help bar

Prefer composition of small models/components over one oversized monolithic model file.

## Initial Backlog Structure

The Notion board should use epics plus concrete tickets.

### Epics

- Go app foundation
- Docker integration layer
- Resource browser parity
- Container workflows
- Release and replacement

### Ticket Characteristics

Tickets should:

- describe shippable outcomes
- represent concrete engineering work
- avoid vague “rewrite X” phrasing
- prefer vertical slices where possible
- move the Go app toward becoming the default runtime

Examples:

- Initialize Go module and Bubble Tea app shell
- Implement Docker connection bootstrap and error state
- Render container resource table with keyboard navigation
- Add resource browser tabs for images, volumes, and networks
- Add container start/stop/restart/remove actions
- Port metrics and logs views
- Build Linux release binary and packaging flow for Go runtime

## Delivery Order

The recommended delivery order is:

1. Bootable Go app shell
2. Docker connection and container browser
3. Full resource browser parity
4. Container workflows and deeper operational views
5. Packaging, release, and runtime cutover

This order provides early evidence that Bubble Tea is the right fit while still marching toward full replacement.

## Error Handling Strategy

- Surface connection failures clearly in the UI instead of silently exiting.
- Distinguish between startup/config errors and runtime action failures.
- Preserve a usable UI shell whenever possible, even if Docker is unavailable.
- Treat destructive actions with explicit confirmation flows.
- Prefer deterministic status messages and reusable error display patterns over ad hoc logging.

## Testing Strategy

The Go rewrite should include:

- unit tests for service and transformation logic
- focused tests for Bubble Tea model behavior where practical
- smoke tests for app startup and key flows
- packaging/build verification once the Go binary becomes the primary release artifact

Tests should concentrate on behavior and core interactions, not fragile terminal snapshots unless a snapshot meaningfully protects a real workflow.

## Repo and Migration Expectations

- The Python implementation remains under `python/` during the rewrite.
- New Go code should live at the repository root in a conventional Go module layout.
- Documentation and packaging should shift to Go once the Go app is the intended runtime.
- Old Python-specific build/release paths can remain temporarily as reference material, but should be retired during the release-and-replacement phase.

## Risks

### 1. Scope Creep

The main project risk is expanding the rewrite into a redesign. The rewrite should prioritize parity and replacement first.

### 2. UI Complexity Growth

Bubble Tea is flexible, but a careless implementation can become too monolithic. The architecture should enforce smaller models and clear boundaries.

### 3. Parity Ambiguity

Some Python behaviors may be underspecified. Where ambiguity appears, the Python runtime should be treated as the behavioral reference unless intentionally superseded.

### 4. Packaging and Cutover Drift

If the Go runtime becomes functional but release infrastructure remains Python-oriented, the project can get stuck in a half-migrated state. Packaging and release work must be treated as first-class tickets.

## Recommendation

Proceed with the Go rewrite using Bubble Tea, managed as an epic-and-ticket program in Notion. The first implementation milestone should be a runnable Go app that can connect to Docker and display core resource browsing behavior, but the overall project should remain explicitly targeted at full runtime replacement.
