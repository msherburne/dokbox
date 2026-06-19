# Dokbox TUI Design

## Summary

Dokbox v1 will be a keyboard-first Textual TUI for managing Docker resources. It will use top-level tabs for Docker object types, dense resource tables, full-screen detail views, per-screen shortcut help, and focused management actions. Containers get the richest experience first: monitoring, logs, shell access, and a read-only filesystem explorer.

## Goals

- Show all primary Docker object types in one TUI: containers, images, volumes, and networks.
- Keep the main interface fast and readable, closer to a file manager than a dashboard.
- Provide management actions where users are already working, including per-tab prune actions.
- Make container details visually useful with compact resource graphs or bars instead of text-heavy metric dumps.
- Support container shell access and read-only container filesystem browsing similar to Docker Desktop.

## Non-Goals

- Editing, deleting, renaming, or uploading files through the container file explorer.
- A global Docker activity/events tab in v1.
- Compose-specific project management in v1.
- Replacing Docker's full inspect output with a complete JSON editor or viewer in v1.

## Navigation

Dokbox opens directly into a Textual app with top-level tabs:

```text
Containers | Images | Volumes | Networks
```

Each tab shows a dense, keyboard-first table. Left and right move between tabs. Up and down move through rows. `/` filters the current table. `Enter` opens a full-screen detail view for the selected resource. `q` backs out of nested screens or quits from the root screen.

The main screen should not include a permanent side inspector. The selected tab stays focused on scanning and selecting resources. Detail views appear only when requested.

Every screen includes a dedicated shortcut bar, likely pinned to the bottom. The bar changes with context and shows only actions that apply to the current screen. For example, the Containers tab may show `Enter Details`, `s Start/Stop`, `r Restart`, `p Prune`, `/ Filter`, while the Files subtab may show `Enter Open`, `Backspace Up`, `/ Find`, `q Back`.

## Resource Tabs

### Containers

The Containers tab lists name, image, state, status, ports, created time, and concise health cues when available.

Actions:

- Open details
- Start
- Stop
- Restart
- Remove
- Prune stopped containers
- Filter

### Images

The Images tab lists repository/tag, image ID, size, created time, and usage hints when available.

Actions:

- Open details
- Remove
- Prune unused images
- Filter

### Volumes

The Volumes tab lists name, driver, mountpoint or scope summary, created time when available, and usage hints when Docker exposes them.

Actions:

- Open details
- Remove
- Prune unused volumes
- Filter

### Networks

The Networks tab lists name, driver, scope, attachable/internal flags, and connected container count when available.

Actions:

- Open details
- Remove
- Prune unused networks
- Filter

All destructive actions require confirmation. Prune is surfaced inside each resource tab rather than hidden in a global menu.

## Container Details

Pressing `Enter` on a container opens a full-screen detail view with subtabs:

```text
Overview | Logs | Shell | Files
```

### Overview

The Overview subtab is a polished monitoring surface rather than a wall of inspect text. It shows CPU, memory, network, and disk I/O as compact visual widgets. Bar percentages are the default terminal-friendly visualization. Pie or donut-style indicators may be used if they render clearly in Textual.

CPU usage should be normalized against configured container limits when Docker exposes those limits, such as CPU quota, cpuset, or nano CPUs. When no limit exists, CPU falls back to host-relative usage and labels that clearly. Memory uses limit-aware percentages when a memory limit exists.

The Overview subtab also includes high-value identity and runtime details, such as image, command, ports, mounts, restart policy, labels summary, and created time.

### Logs

The Logs subtab streams or tails container logs. It supports follow, pause/resume, and simple filtering. Leaving the Logs subtab cancels or pauses active streaming cleanly.

### Shell

The Shell subtab opens an interactive exec session in the selected container when possible. Dokbox should try `/bin/bash` first and fall back to `/bin/sh` if bash is unavailable. The UI should clearly report when a container is not running or cannot start an exec session.

### Files

The Files subtab is a read-only Docker Desktop-style filesystem explorer for the selected container. It supports directory navigation, opening file contents, showing metadata, and going up/back. It does not edit, delete, rename, or upload files in v1.

The implementation should prefer Docker APIs that can read container filesystem paths without requiring extra binaries inside the container. If a path cannot be read, Dokbox should show a clear error in context and keep the explorer usable.

## Detail Views for Other Resources

Images, volumes, and networks also have full-screen detail views, but they are simpler than container details in v1. They show formatted metadata, usage hints where available, and relevant actions such as remove and prune. These screens should avoid dumping raw inspect JSON as the primary experience.

## Architecture

The Textual UI should stay thin and delegate Docker work to service-style modules.

- `Dokbox` owns the app lifecycle, loaded configuration, and shared Docker client.
- UI modules own screens and widgets: main resource browser, resource tables, shortcut bar, confirmation dialogs, container detail screen, logs pane, shell pane, and file explorer pane.
- A Docker service layer exposes typed operations: list resources, inspect resources, start/stop/restart/remove containers, prune by resource type, stream logs, collect stats samples, open exec sessions, and read container filesystem entries.
- A view-model layer translates raw Docker SDK objects into stable table rows and detail models. This layer owns formatting for CPU and memory percentages, status labels, limit-aware metrics, and action availability.

This separation keeps Textual widgets from becoming Docker SDK wrappers and makes formatting and Docker behavior easier to test.

## Error Handling

Docker connection errors should fail gracefully with a clear full-screen state. The state shows the configured Docker host, the connection error, and a retry or setup path.

Per-action failures should appear as contextual notifications or dialogs without crashing the TUI. Actions that change Docker state refresh the current tab after completion. Long-running streams such as logs, stats, and shell sessions must be cancellable when leaving the detail screen.

## Testing

Testing should focus first on service and view-model behavior:

- Resource row formatting for containers, images, volumes, and networks
- Limit-aware CPU and memory calculations
- Fallback behavior when limits are unavailable
- Prune action routing by resource tab
- Confirmation requirements for destructive actions
- File explorer path behavior and read-only guarantees
- Docker connection error handling

UI tests can start with app construction and screen/widget composition smoke tests. Broader interaction tests should be added as the TUI stabilizes.

## Open Decisions

- Exact shortcut letters may change during implementation if conflicts appear.
- Pie or donut indicators are optional and depend on whether they remain readable in the terminal.
- Image, volume, and network usage hints depend on what can be derived reliably from Docker data without expensive scans.
