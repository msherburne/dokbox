# Packaging

Dokbox now uses Go-native packaging and release helpers from the repository root.

## Supported outputs

- Standalone Linux directory plus `.tar.gz`
- Standalone macOS directory plus `.tar.gz`
- Standalone Windows directory plus `.zip`
- Debian `.deb`
- RPM `.rpm`
- Pacman `.pkg.tar.zst`
- Release manifest metadata
- Homebrew formula output
- winget manifest output

All artifacts are written to `dist/`.

## Prerequisites

Install Go first, then install the packaging tool you need:

- `build-linux.sh`: `go`, `tar`
- `build-darwin.sh`: `go`, `tar`
- `build-windows.sh`: `go`, `zip`
- `build-deb.sh`: `dpkg-deb`
- `build-rpm.sh`: `rpmbuild`
- `build-pacman.sh`: `makepkg`
- `release-manifest.sh`: `python3`, `sha256sum` or `shasum`
- `generate-homebrew-formula.sh`: `python3`
- `generate-winget-manifest.sh`: `python3`

## Build commands

Build the standalone Linux payload and archive:

```bash
./packaging/build-linux.sh
```

Build the standalone macOS payload and archive:

```bash
./packaging/build-darwin.sh
```

Build the standalone Windows payload and archive:

```bash
./packaging/build-windows.sh
```

Build a Debian package:

```bash
./packaging/build-deb.sh
```

Build an RPM package:

```bash
./packaging/build-rpm.sh
```

Build a pacman package:

```bash
./packaging/build-pacman.sh
```

Generate cross-platform release metadata:

```bash
./packaging/release-manifest.sh
./packaging/generate-homebrew-formula.sh
./packaging/generate-winget-manifest.sh
```

## Installer entrypoints

Release automation also publishes installer entrypoints for end users:

- Linux and macOS: `install.sh`
- Windows PowerShell: `install.ps1`

Examples:

```bash
curl -fsSL https://github.com/msherburne/dokbox/releases/latest/download/install.sh | bash
```

```powershell
irm https://github.com/msherburne/dokbox/releases/latest/download/install.ps1 | iex
```

Current Linux arm64 note:

- Linux release artifacts are currently published for `amd64` only

Current Windows note:

- Windows release artifacts are currently published for `amd64` only

If `bsdtar` is unavailable on a Debian or Ubuntu development machine, the
pacman helper will try to bootstrap it locally into `.cache/packaging-tools/`
using `apt download` plus `dpkg-deb`.

## Local install checks

After building a package, test installation with the native tool for that
environment:

```bash
sudo apt install ./dist/dokbox_*.deb
sudo dnf install ./dist/dokbox-*.rpm
sudo pacman -U ./dist/dokbox-*.pkg.tar.zst
```

Standalone smoke checks:

```bash
./dist/dokbox-linux-x86_64/bin/dokbox
./dist/dokbox-darwin-arm64/bin/dokbox
./dist/dokbox-windows-x86_64/bin/dokbox.exe
```
