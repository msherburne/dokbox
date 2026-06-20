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
PACKAGE_IDENTIFIER="dokbox.dokbox"

require_tool() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'Required tool missing: %s\n' "$1" >&2
    exit 1
  fi
}

bootstrap_bsdtar() {
  if command -v bsdtar >/dev/null 2>&1; then
    return 0
  fi

  local tool_root="$ROOT_DIR/.cache/packaging-tools/libarchive-tools"
  local tool_bin="$tool_root/usr/bin"

  if [ -x "$tool_bin/bsdtar" ]; then
    PATH="$tool_bin:$PATH"
    export PATH
    return 0
  fi

  if ! command -v apt >/dev/null 2>&1 || ! command -v dpkg-deb >/dev/null 2>&1; then
    printf 'Required tool missing: bsdtar\n' >&2
    exit 1
  fi

  local download_dir="$ROOT_DIR/.cache/packaging-tools/downloads"
  mkdir -p "$download_dir" "$tool_root"

  if ! (
    cd "$download_dir"
    apt download libarchive-tools >/dev/null 2>&1
  ); then
    printf 'Required tool missing: bsdtar\n' >&2
    exit 1
  fi

  local deb_file
  deb_file="$(find "$download_dir" -maxdepth 1 -name 'libarchive-tools_*.deb' | sort | tail -n 1)"
  if [ -z "$deb_file" ]; then
    printf 'Required tool missing: bsdtar\n' >&2
    exit 1
  fi

  rm -rf "$tool_root"
  mkdir -p "$tool_root"
  dpkg-deb -x "$deb_file" "$tool_root"

  if [ ! -x "$tool_bin/bsdtar" ]; then
    printf 'Required tool missing: bsdtar\n' >&2
    exit 1
  fi

  PATH="$tool_bin:$PATH"
  export PATH
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

archive_extension_for() {
  case "$1" in
    linux|darwin) printf 'tar.gz\n' ;;
    windows) printf 'zip\n' ;;
    *)
      printf 'Unsupported target OS: %s\n' "$1" >&2
      exit 1
      ;;
  esac
}

binary_name_for() {
  case "$1" in
    windows) printf '%s.exe\n' "$PACKAGE_NAME" ;;
    linux|darwin) printf '%s\n' "$PACKAGE_NAME" ;;
    *)
      printf 'Unsupported target OS: %s\n' "$1" >&2
      exit 1
      ;;
  esac
}

archive_root_for() {
  printf '%s-%s-%s\n' "$PACKAGE_NAME" "$1" "$2"
}

archive_name_for() {
  local goos="$1"
  local archive_arch="$2"
  printf '%s.%s\n' "$(archive_root_for "$goos" "$archive_arch")" "$(archive_extension_for "$goos")"
}

binary_path_for() {
  printf 'bin/%s\n' "$(binary_name_for "$1")"
}

standalone_dir_for() {
  printf '%s/%s\n' "$DIST_DIR" "$(archive_root_for "$1" "$2")"
}

standalone_archive_for() {
  printf '%s/%s\n' "$DIST_DIR" "$(archive_name_for "$1" "$2")"
}

standalone_bin_for() {
  printf '%s/%s\n' "$(standalone_dir_for "$1" "$2")" "$(binary_path_for "$1")"
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

map_homebrew_arch() {
  case "$1" in
    amd64) printf 'intel\n' ;;
    arm64) printf 'arm\n' ;;
    *)
      printf 'Unsupported Homebrew architecture: %s\n' "$1" >&2
      exit 1
      ;;
  esac
}

map_winget_arch() {
  case "$1" in
    amd64) printf 'x64\n' ;;
    arm64) printf 'arm64\n' ;;
    *)
      printf 'Unsupported winget architecture: %s\n' "$1" >&2
      exit 1
      ;;
  esac
}

resolve_release_base_url() {
  if [ -n "${RELEASE_BASE_URL:-}" ]; then
    printf '%s\n' "${RELEASE_BASE_URL%/}"
    return
  fi

  printf '%s/releases/download/v%s\n' "$PACKAGE_URL" "$RAW_VERSION"
}

sha256_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
    return
  fi

  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
    return
  fi

  if command -v openssl >/dev/null 2>&1; then
    openssl dgst -sha256 "$1" | awk '{print $NF}'
    return
  fi

  printf 'Required tool missing: sha256 checksum utility\n' >&2
  exit 1
}

RAW_VERSION="$(resolve_version)"
PACKAGE_VERSION="$(sanitize_package_version "$RAW_VERSION")"
PACKAGE_RELEASE="${RELEASE:-1}"
GO_ARCH="$(resolve_go_arch)"
ARCHIVE_ARCH="$(map_archive_arch "$GO_ARCH")"
ARCHIVE_ROOT="$(archive_root_for linux "$ARCHIVE_ARCH")"
STANDALONE_DIR="$(standalone_dir_for linux "$ARCHIVE_ARCH")"
STANDALONE_ARCHIVE="$(standalone_archive_for linux "$ARCHIVE_ARCH")"
STANDALONE_BIN="$(standalone_bin_for linux "$ARCHIVE_ARCH")"
BIN_PATH="/usr/bin/$PACKAGE_NAME"

ensure_standalone_payload() {
  if [ ! -x "$STANDALONE_BIN" ]; then
    "$ROOT_DIR/packaging/build-linux.sh"
  fi
}
