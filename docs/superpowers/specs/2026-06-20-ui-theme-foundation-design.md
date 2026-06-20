# UI Theme Foundation Design

## Summary

This spec covers the first UI-overhaul implementation slice for Dokbox Go:

- `DOK-16`: define the shared visual system for the TUI refresh
- `DOK-24`: add built-in themes configurable from Dokbox settings

The goal of this slice is to create a reusable, configuration-driven theme foundation before deeper browser/detail polish begins. The outcome should be a stable semantic theme model, a built-in preset registry, config support for selecting a preset, and shared style construction that future UI tickets can build on.

## Goals

- Add a top-level `theme` config field for selecting a built-in preset
- Ship three built-in dark terminal themes
- Move shared styling toward semantic theme tokens instead of hard-coded colors
- Establish a small, reusable design-system foundation for later UI-overhaul tickets
- Preserve a sensible default theme and safe fallback behavior

## Non-Goals

- No user-defined custom theme objects in config
- No partial theme overrides
- No in-app interactive theme switcher
- No attempt to complete the entire visual overhaul in this slice
- No broad refactor of every view before there is a shared theme foundation

## User Experience

Dokbox should continue to work with no config changes. If the user does nothing, the application should load a default built-in dark theme.

If the user sets a supported theme name in config, Dokbox should apply that preset consistently across shared UI styling. If the user sets an unsupported theme name, Dokbox should fall back to the default preset rather than failing startup.

The first-pass preset catalog should include:

- `default`: neutral dark baseline
- `slate`: cooler, softer contrast
- `ember`: warmer accent and status palette

## Configuration Design

### Config Shape

Theme selection should live at the top level of the Dokbox config object:

```json
{
  "config_version": 1,
  "docker_host": "unix:///var/run/docker.sock",
  "theme": "default"
}
```

### Defaulting Rules

- Generated default config should include the default theme explicitly
- Existing configs without `theme` should continue to resolve successfully
- Missing or empty `theme` should resolve to `default`
- Unknown theme names should resolve to `default`

### Rationale

The top-level field keeps the config small and avoids premature nesting. This is appropriate for a terminal-first application with a small number of global settings.

## Theme System Design

### Theme Resolution Model

The application should resolve an active theme in one place, using:

1. the configured theme name
2. the built-in preset registry
3. fallback to the default preset when resolution fails

Callers should not need to know fallback rules. They should request the active theme and use it.

### Semantic Theme Model

Themes should be defined in terms of semantic visual roles rather than component-specific prebuilt Lip Gloss styles.

First-pass semantic roles should cover:

- app background
- panel/background surface
- border
- primary text
- muted text
- emphasis text
- accent
- focus/active selection
- success
- warning
- error
- info

This keeps the first version intentionally small while still giving enough range for later browser/detail polish.

### Preset Registry

Built-in presets should be registered by stable string key. The registry should be the source of truth for:

- supported preset names
- each preset’s semantic token values
- the default preset name

The registry should be easy to extend later without redesigning the config model.

## Style Construction

### Principle

Application styles should be built from semantic tokens, not hard-coded color literals embedded directly in views.

### First-pass Style Coverage

This slice does not need to theme every surface in the app immediately, but it should move the shared style path onto the new foundation. That includes:

- app shell/frame style
- shared text emphasis patterns
- shared panel/border treatment where practical
- any central style helpers already used across views

The key requirement is that later UI tickets can consume the same theme model instead of introducing new styling islands.

### Later UI Work

`DOK-17` through `DOK-20` should build on this by introducing component-level style roles for:

- tabs
- tables/list rows
- prompts and confirmations
- feedback/result messages
- detail sections
- shortcut hints

Those component roles do not need to be fully implemented in this first slice, but the semantic theme model should make them straightforward to add.

## Codebase Impact

### Configuration Layer

The config domain model and config loading/generation code should be updated to include theme selection and fallback-safe resolution behavior.

### App Style Layer

The current app-level style code is very small and hard-coded. This slice should turn it into a theme-driven entry point instead of a single static style declaration.

### Theme Boundary

A dedicated theme or shared-style package/module should own:

- the semantic theme type
- built-in preset definitions
- preset resolution
- style construction helpers, or the primitives consumed by them

This should be a clear boundary so browser/detail refresh work can consume it without re-implementing theme logic.

## Error Handling

Theme resolution should fail soft.

- Invalid theme config should not crash the application
- Invalid theme config should not block config migration
- Unknown names should fall back to `default`
- If helpful, Dokbox may later surface a friendly message about fallback behavior, but that is not required in this slice

## Testing Strategy

### Config Tests

Add or update tests for:

- generated default config includes `theme`
- existing config without `theme` still resolves
- valid theme names resolve correctly
- invalid theme names fall back to `default`

### Theme Resolution Tests

Add unit tests for:

- preset registry lookup
- default preset lookup
- fallback behavior for unknown names

### Style/Application Tests

Add targeted tests where practical to verify:

- the active theme influences rendered styling paths
- shared style construction changes when a different preset is selected

These tests do not need to snapshot the entire UI. They only need enough coverage to keep theme wiring from regressing.

## Implementation Boundaries

This slice should stop once Dokbox has:

- a top-level config-driven theme name
- three built-in dark presets
- a semantic theme model
- default/fallback-safe theme resolution
- shared style construction driven by the active theme
- tests covering config and resolution behavior

It should not attempt to fully complete the visual redesign of the browser/detail flows. Those belong to the later UI tickets once the foundation exists.

## Success Criteria

This work is successful when:

- Dokbox can be configured with `theme: "default"`, `theme: "slate"`, or `theme: "ember"`
- missing or invalid theme values safely fall back to `default`
- the active theme is resolved centrally
- shared styles consume semantic theme tokens rather than hard-coded values
- the foundation is ready for `DOK-17` through `DOK-20` to build on
