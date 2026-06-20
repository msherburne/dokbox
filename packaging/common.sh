#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="$ROOT_DIR/dist"

PACKAGE_NAME="dokbox"
PACKAGE_SUMMARY="Terminal UI for Docker resource management"
PACKAGE_DESCRIPTION="Dokbox is a Bubble Tea terminal application for browsing and managing Docker resources."
PACKAGE_LICENSE="Unspecified"
PACKAGE_URL="https://github.com/msherburne/dokbox"
PACKAGE_MAINTAINER="dokbox maintainers"

require_tool() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'Required tool missing: %s\n' "$1" >&2
    exit 1
  fi
}

resolve_version() {
  if [ -n "${VERSION:-}" ]; then
    printf '%s\n' "$VERSION"
    return
  fi

  if command -v git >/dev/null 2>&1; then
    version="$(git -C "$ROOT_DIR" describe --tags --abbrev=0 2>/dev/null || true)"
    version="${version#v}"
    if [ -n "$version" ]; then
      printf '%s\n' "$version"
      return
    fi
  fi

  printf '0.1.0\n'
}

sanitize_package_version() {
  printf '%s\n' "$1" | sed 's/[^A-Za-z0-9._+~^]/_/g'
}

resolve_go_arch() {
  if [ -n "${GOARCH:-}" ]; then
    printf '%s\n' "$GOARCH"
    return
  fi

  require_tool go
  go env GOARCH
}

map_archive_arch() {
  case "$1" in
    amd64) printf 'x86_64\n' ;;
    arm64) printf 'arm64\n' ;;
    *)
      printf 'Unsupported Go architecture: %s\n' "$1" >&2
      exit 1
      ;;
  esac
}

map_deb_arch() {
  case "$1" in
    x86_64) printf 'amd64\n' ;;
    arm64) printf 'arm64\n' ;;
    *)
      printf 'Unsupported Debian architecture: %s\n' "$1" >&2
      exit 1
      ;;
  esac
}

map_rpm_arch() {
  case "$1" in
    x86_64) printf 'x86_64\n' ;;
    arm64) printf 'aarch64\n' ;;
    *)
      printf 'Unsupported RPM architecture: %s\n' "$1" >&2
      exit 1
      ;;
  esac
}

RAW_VERSION="$(resolve_version)"
PACKAGE_VERSION="$(sanitize_package_version "$RAW_VERSION")"
PACKAGE_RELEASE="${RELEASE:-1}"
GO_ARCH="$(resolve_go_arch)"
ARCHIVE_ARCH="$(map_archive_arch "$GO_ARCH")"
ARCHIVE_ROOT="$PACKAGE_NAME-linux-$ARCHIVE_ARCH"
STANDALONE_DIR="$DIST_DIR/$ARCHIVE_ROOT"
STANDALONE_ARCHIVE="$DIST_DIR/$ARCHIVE_ROOT.tar.gz"
STANDALONE_BIN="$STANDALONE_DIR/bin/$PACKAGE_NAME"
BIN_PATH="/usr/bin/$PACKAGE_NAME"

ensure_standalone_payload() {
  if [ ! -x "$STANDALONE_BIN" ]; then
    "$ROOT_DIR/packaging/build-linux.sh"
  fi
}
