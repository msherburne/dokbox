# Development

This repository is now Go-first.

The active application lives at the repository root. The `python/` directory is
kept as a legacy reference during the rewrite and should not be used for normal
day-to-day development unless you are intentionally comparing behavior.

## Prerequisites

### All environments

1. Install Go `1.25.x`.
2. Install Docker Desktop or Docker Engine.
3. Make sure the Docker daemon is running if you want live Docker data.

Dokbox can still start when Docker is unavailable, but it will show a
connection error in the shell and most runtime features will not work.

### Linux

Install dependencies with your package manager, then verify the tools:

```bash
go version
docker version
```

### macOS

Install Go and Docker Desktop, then verify the tools:

```bash
go version
docker version
```

### Windows

Install Go and Docker Desktop, then verify the tools in PowerShell:

```powershell
go version
docker version
```

## Project setup

From the repository root:

1. Download Go module dependencies:

```bash
go mod tidy
```

2. Run the full test suite:

```bash
go test ./...
```

## Run the application in development

From the repository root:

```bash
go run .
```

What happens on startup:

1. Dokbox loads or creates `~/.dokbox.json`.
2. Dokbox resolves the Docker host from config.
3. Dokbox starts the Bubble Tea shell.
4. Dokbox checks Docker connectivity and shows the current status in the UI.

### Docker environment variables

The Go Docker bootstrap respects standard Docker environment configuration,
including:

- `DOCKER_HOST`
- `DOCKER_TLS_VERIFY`
- `DOCKER_CERT_PATH`
- `DOCKER_API_VERSION`

Example:

```bash
DOCKER_HOST=tcp://127.0.0.1:2375 go run .
```

## Create binaries locally

### Linux standalone archive

Build the Linux payload directory and `.tar.gz` archive:

```bash
./packaging/build-linux.sh
```

This writes:

- `dist/dokbox-linux-x86_64/`
- `dist/dokbox-linux-x86_64.tar.gz`

Run the standalone binary directly:

```bash
./dist/dokbox-linux-x86_64/bin/dokbox
```

### macOS standalone archive

Build the macOS payload directory and `.tar.gz` archive:

```bash
./packaging/build-darwin.sh
```

This writes:

- `dist/dokbox-darwin-<arch>/`
- `dist/dokbox-darwin-<arch>.tar.gz`

Run the standalone binary directly:

```bash
./dist/dokbox-darwin-arm64/bin/dokbox
```

To target Intel macOS locally:

```bash
GOARCH=amd64 ./packaging/build-darwin.sh
```

### Windows standalone archive

Build the Windows payload directory and `.zip` archive:

```bash
./packaging/build-windows.sh
```

This writes:

- `dist/dokbox-windows-<arch>/`
- `dist/dokbox-windows-<arch>.zip`

Run the standalone binary directly:

```bash
./dist/dokbox-windows-x86_64/bin/dokbox.exe
```

### Debian package

Install the Debian packaging tool, then build:

```bash
sudo apt-get install dpkg-dev
./packaging/build-deb.sh
```

Install the resulting package system-wide:

```bash
sudo apt install ./dist/dokbox_*.deb
```

### RPM package

Install the RPM tooling for your distro, then build:

```bash
sudo dnf install rpm-build
./packaging/build-rpm.sh
```

Install the resulting package system-wide:

```bash
sudo dnf install ./dist/dokbox-*.rpm
```

### Pacman package

Install Arch packaging tools, then build:

```bash
sudo pacman -S --needed base-devel
./packaging/build-pacman.sh
```

On Debian or Ubuntu environments where `makepkg` is available but `bsdtar` is
not installed system-wide, the helper will try to bootstrap `bsdtar` locally
from the distro package metadata into `.cache/packaging-tools/`.

Install the resulting package system-wide:

```bash
sudo pacman -U ./dist/dokbox-*.pkg.tar.zst
```

### Raw local binary

If you only want a fast local binary without packaging:

```bash
mkdir -p dist
go build -o dist/dokbox .
```

Run it:

```bash
./dist/dokbox
```

## Useful development commands

Run all tests:

```bash
go test ./...
```

Run only the Go UI and startup tests:

```bash
go test ./tests-go
```

Format Go code:

```bash
gofmt -w .
```

Generate release metadata for cross-platform installers:

```bash
./packaging/release-manifest.sh
./packaging/generate-homebrew-formula.sh
./packaging/generate-winget-manifest.sh
```

These helpers expect `python3` plus a SHA-256 tool such as `sha256sum` or
`shasum`.

### Installer smoke checks

Linux or macOS:

```bash
./install/install.sh --help
```

Release-style one-liner:

```bash
curl -fsSL https://github.com/msherburne/dokbox/releases/latest/download/install.sh | bash
```

Current Linux arm64 note:

- Linux release artifacts are currently published for `amd64` only

Windows PowerShell:

```powershell
./install/install.ps1 -Help
```

Release-style one-liner:

```powershell
irm https://github.com/msherburne/dokbox/releases/latest/download/install.ps1 | iex
```

Current Windows note:

- Windows release artifacts are currently published for `amd64` only

## Current packaging status

The supported Linux packaging flow now lives in the repository root under
`packaging/`.

Cross-platform release automation also lives there, including builders for:

- Linux standalone plus native distro packages
- macOS standalone archives
- Windows standalone archives
- Homebrew and winget release metadata generation

The older helpers in `python/packaging/` are legacy Python-era tooling kept
only as rewrite reference material.

## Troubleshooting

### Docker connection errors on startup

Check:

1. The Docker daemon is running.
2. Your user can access the Docker socket or Docker Desktop daemon.
3. `DOCKER_HOST` points at a valid daemon endpoint.
4. TLS-related Docker environment variables are set correctly for remote hosts.

### Module or dependency issues

Refresh module metadata and retry:

```bash
go mod tidy
go test ./...
```
