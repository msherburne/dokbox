# Packaging

Dokbox now uses Go-native Linux packaging helpers from the repository root.

## Supported outputs

- Standalone Linux directory plus `.tar.gz`
- Debian `.deb`
- RPM `.rpm`
- Pacman `.pkg.tar.zst`

All artifacts are written to `dist/`.

## Prerequisites

Install Go first, then install the packaging tool you need:

- `build-linux.sh`: `go`, `tar`
- `build-deb.sh`: `dpkg-deb`
- `build-rpm.sh`: `rpmbuild`
- `build-pacman.sh`: `makepkg`

## Build commands

Build the standalone Linux payload and archive:

```bash
./packaging/build-linux.sh
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
