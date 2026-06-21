# TUI Style Guide

This document defines the visual roles used by the refreshed Dokbox TUI so
future UI work can extend the same system instead of inventing one-off styles.

## Theme semantics

Dokbox themes expose a small set of semantic color roles in
`internal/theme/theme.go`.

### Surface roles

- `Background`: outer shell backdrop when a component needs the deepest base.
- `Panel`: primary panel fill for framed work surfaces inside the app shell.
- `Border`: panel and shell border color.

### Text roles

- `Text`: default readable body text.
- `Muted`: secondary copy such as metadata, helper text, and continuation hints.
- `Emphasis`: stronger text when the default `Text` treatment is not enough.

### Interactive roles

- `Accent`: non-destructive highlights that should stand out from the base UI.
- `Focus`: focused or selected interactive emphasis when a stronger cue is useful.

### State roles

- `Success`: successful or healthy runtime states.
- `Warning`: cautionary or degraded states that are not hard failures.
- `Error`: failure states and error feedback.
- `Info`: neutral informational feedback that still deserves emphasis.

## Component roles

### Shell

- The app shell is the outer framed workspace.
- Use `Panel` as the shell fill and `Border` for the shell frame.
- Shell-level text should default to `Text`.

### Section titles

- Top-level view titles such as `Resources` and `Container Detail` should use a
  bold, high-clarity treatment.
- Prefer `Emphasis` or a bright `Text`-equivalent treatment over decorative color.

### Tabs

- Tabs should remain single-line to avoid layout jumps during selection changes.
- Active tabs should use a compact highlighted treatment.
- Inactive tabs should use `Muted` or low-emphasis text.
- Do not switch between boxed multi-line and inline one-line tab states.

### Tables

- Tables are the primary scanning surface for browser and file-list content.
- Headers should be visually distinct through weight and framing, not noise.
- Selected rows should be visibly highlighted without changing table geometry.
- Long values should truncate with ellipses instead of breaking the overall layout.

### Metadata and helper text

- Secondary status lines such as `Selected: ...`, path labels, and continuation
  hints should use `Muted`.
- Shortcut rows should read as helpers, not primary content.

### Feedback surfaces

- Action menus, confirmations, and other nested framed blocks should reuse the
  same border language as the shell and panels.
- Feedback styling should use semantic state roles rather than ad-hoc colors.

## Runtime feedback mapping

These mappings should be followed whenever runtime feedback styling is added or
refined:

- Healthy Docker connectivity or successful actions: `Success`
- Recoverable issues or caution states: `Warning`
- Hard failures such as shell launch/file errors: `Error`
- Neutral explanatory updates: `Info`

## Layout conventions

- Favor framed workbench layouts over loose text blobs.
- Keep scan-heavy content in structured tables where possible.
- Width-constrained views should truncate instead of overflowing horizontally.
- Height-constrained views should clip gracefully and show a continuation hint.

## Current built-in presets

Dokbox currently ships these presets:

- `default`
- `slate`
- `ember`

All presets should honor the same semantic roles above, even when the palette
changes.
