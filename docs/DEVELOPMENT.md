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

### Linux and macOS

Build a local binary into `dist/`:

```bash
mkdir -p dist
go build -o dist/dokbox .
```

Run it:

```bash
./dist/dokbox
```

### Windows

Build a local binary into `dist/`:

```powershell
New-Item -ItemType Directory -Force dist | Out-Null
go build -o dist/dokbox.exe .
```

Run it:

```powershell
.\dist\dokbox.exe
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

## Current packaging status

The current locally supported binary path is `go build`.

The older standalone packaging helpers in `python/packaging/` are legacy
Python-era tooling and are not the primary development path for the Go rewrite.

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
