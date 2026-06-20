# Cross-Platform Release Design

## Goal

Ship Dokbox as a Go-first application with:

- Linux native packages and standalone archives
- macOS standalone archives
- Windows standalone archives
- one native installer entrypoint per platform
- GitHub Actions as the canonical cross-platform build and release system

The installer experience should feel like a one-liner on each platform while
still using the native install surface of that platform:

- Linux and macOS: `install.sh`
- Windows: `install.ps1`

## Scope

This design covers:

- release artifact formats for Linux, macOS, and Windows
- release metadata contract
- installer responsibilities
- GitHub Actions build and publish flow
- Homebrew and winget integration boundaries

This design does not include:

- macOS `.pkg` or `.app` packaging
- Windows `.msi` or `.exe` installer generation
- signing and notarization
- auto-update behavior inside the application

## Requirements

1. GitHub Actions must build release artifacts for Linux, macOS, and Windows.
2. Linux must continue to support native package outputs:
   - `.deb`
   - `.rpm`
   - `.pkg.tar.zst`
3. macOS and Windows must ship standalone binary archives.
4. Installers must resolve the correct artifact for the current OS and
   architecture from release metadata.
5. Linux installers must use the local system package manager to install the
   matching local package file.
6. Windows installation must use PowerShell and winget-compatible release
   metadata.
7. Homebrew and winget publication should be supported as a follow-on layer, not
   required for the first artifact/install implementation.

## Chosen Approach

Use native platform installers with a shared release metadata contract:

- `install.sh` for Linux and macOS
- `install.ps1` for Windows
- GitHub Actions matrix builds per OS
- GitHub Releases as the distribution source of truth
- machine-readable metadata consumed by installers and later by Homebrew and
  winget automation

This avoids forcing Bash to act as a Windows-native installer while still
keeping the install surface simple and consistent.

## Alternatives Considered

### 1. Single Bash installer for every platform

Pros:

- one entrypoint
- minimal conceptual surface

Cons:

- Windows is not a Bash-native install environment
- handoff into winget or PowerShell becomes awkward
- debugging platform-specific failures is harder

Rejected because the simplicity is superficial and reliability is weaker on
Windows.

### 2. Native installer per platform

Pros:

- standard user experience on each OS
- cleaner platform-specific logic
- easy future expansion into Homebrew and winget publication

Cons:

- two bootstrap scripts instead of one
- some shared logic must be represented through metadata rather than one shell
  script

Chosen because it is the most reliable and least surprising.

### 3. Direct binary placement everywhere

Pros:

- simplest implementation
- avoids package manager behavior differences

Cons:

- does not match the native package-manager install goal
- weaker uninstall and upgrade story

Rejected because it misses the intended installation model.

## Artifact Contract

### Linux

Continue publishing:

- standalone directory archive
- `.deb`
- `.rpm`
- `.pkg.tar.zst`

Existing naming conventions should remain stable where already established.

### macOS

Publish standalone archives:

- `dokbox_<version>_darwin_amd64.tar.gz`
- `dokbox_<version>_darwin_arm64.tar.gz`

Each archive contains a single `dokbox` executable plus any required supporting
files.

### Windows

Publish standalone archives:

- `dokbox_<version>_windows_amd64.zip`
- `dokbox_<version>_windows_arm64.zip`

Each archive contains `dokbox.exe` plus any required supporting files.

## Release Metadata Contract

Each release publishes a machine-readable metadata file, for example:

- `release.json`

The file must contain:

- version
- release date
- artifact list
- for each artifact:
  - OS
  - architecture
  - package type
  - filename
  - checksum
  - download URL
  - optional installer channel hint such as `apt`, `dnf`, `pacman`, `brew`, or
    `winget`

Example package types:

- `tar.gz`
- `zip`
- `deb`
- `rpm`
- `pkg.tar.zst`

This metadata is the single contract used by:

- `install.sh`
- `install.ps1`
- future Homebrew automation
- future winget automation

## Installer Design

### `install.sh`

Responsibilities:

- detect Linux vs macOS
- detect architecture
- fetch release metadata
- choose the correct artifact

Linux behavior:

- detect the native install tool available on the host
- prefer the matching artifact:
  - `apt` for `.deb`
  - `dnf` or `rpm` for `.rpm`
  - `pacman` for `.pkg.tar.zst`
- download the selected package into a temporary location
- invoke the native local-package install command

macOS behavior:

- prefer Homebrew-managed install flow once Homebrew publication exists
- require Homebrew for the native install path
- if Homebrew publication is not yet available for a release, fail with a clear
  message rather than silently installing outside the package-manager flow

The initial recommended implementation is:

- Linux direct local package install
- macOS Homebrew-first install path, with an explicit “not yet published” error
  until tap automation is in place

### `install.ps1`

Responsibilities:

- detect architecture
- fetch release metadata
- install through winget when the package is published there
- provide a controlled fallback if winget publication is not yet present

The first implementation should be structured for winget consumption even if the
winget publication automation lands later.

## Build System Design

GitHub Actions becomes the canonical release builder.

### Workflow responsibilities

1. Build all platform artifacts in a matrix:
   - `ubuntu-latest`
   - `macos-latest`
   - `windows-latest`
2. Run relevant tests
3. Collect checksums
4. Generate release metadata
5. Upload artifacts to GitHub Releases

### Build matrix

Recommended targets:

- Linux:
  - `amd64`
  - `arm64`
- macOS:
  - `amd64`
  - `arm64`
- Windows:
  - `amd64`
  - `arm64`

If a specific target proves unsupported by a dependency or runtime behavior, the
workflow should make that explicit in code and docs rather than silently omitting
it.

## Repository Layout

### `packaging/`

Owns artifact creation logic:

- Linux package builders
- macOS archive builder
- Windows archive builder
- shared metadata helpers

### `install/`

Owns installer entrypoints:

- `install.sh`
- `install.ps1`

### `.github/workflows/`

Owns CI release orchestration:

- build
- package
- metadata generation
- release publication

## Testing Strategy

### Automated tests

- packaging helper tests for artifact naming and structure
- metadata generation tests
- installer selection tests with mocked OS and architecture inputs
- regression tests for platform-specific packaging edge cases

### CI verification

- Linux package builders run in CI
- macOS archive builder runs in CI
- Windows archive builder runs in CI
- installers are smoke-tested against mocked metadata

### Out of scope for first slice

- live Homebrew publication
- live winget publication
- signing/notarization validation

## Rollout Plan

### Phase 1

Implement cross-platform artifact builders and release metadata generation.

Deliverables:

- macOS archive builder
- Windows archive builder
- metadata generation script or program
- tests for artifact and metadata selection

### Phase 2

Add GitHub Actions workflow for cross-platform builds and release publication.

Deliverables:

- workflow files
- release asset upload
- checksum publication
- release metadata publication

### Phase 3

Add platform-native installers.

Deliverables:

- `install.sh`
- `install.ps1`
- installer tests for platform detection and asset selection

At this phase, macOS installer behavior should remain Homebrew-first and must
not silently fall back to unmanaged binary placement.

### Phase 4

Add Homebrew and winget publication support using the same release metadata.

Deliverables:

- Homebrew update automation
- winget manifest update automation

## Risks

### Platform packaging drift

Artifact naming can diverge across scripts and workflows.

Mitigation:

- centralize naming rules
- test generated metadata and artifact filenames together

### Windows release complexity

Windows archive and winget publication can appear simple but have higher edge
case risk around paths, shells, and quoting.

Mitigation:

- keep bootstrap PowerShell-native
- isolate Windows-specific logic in dedicated scripts and tests

### macOS installation ambiguity

The long-term desired install path is Homebrew, but the first implementation may
need a direct archive fallback.

Mitigation:

- keep the metadata contract neutral
- make Homebrew a layer on top of published assets rather than a blocker for
  shipping assets

## Recommendation Summary

Build portable binaries on GitHub Actions for Linux, macOS, and Windows; publish
them and their metadata to GitHub Releases; install them with native bootstrap
scripts per platform; and add Homebrew and winget publication after the shared
artifact contract is stable.
