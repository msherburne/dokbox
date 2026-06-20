#!/usr/bin/env bash
set -euo pipefail

PACKAGE_NAME="dokbox"
DEFAULT_REPO="msherburne/dokbox"
PACKAGE_RELEASE="${DOKBOX_PACKAGE_RELEASE:-1}"

usage() {
  cat <<'EOF'
Usage: install.sh [options]

Install Dokbox using the native package manager for the current platform.

Options:
  --version <version>    Release version without the leading "v"
  --repo <owner/repo>    GitHub repository to install from
  --base-url <url>       Override the release asset base URL
  --no-sudo              Do not use sudo for package manager commands
  --help                 Show this help text
EOF
}

require_tool() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'Required tool missing: %s\n' "$1" >&2
    exit 1
  fi
}

normalize_tag() {
  local version="$1"
  if [ "$version" = "latest" ]; then
    printf 'latest\n'
    return
  fi

  version="${version#v}"
  printf 'v%s\n' "$version"
}

normalize_version() {
  local version="$1"
  if [ "$version" = "latest" ]; then
    printf 'latest\n'
    return
  fi

  printf '%s\n' "${version#v}"
}

resolve_base_url() {
  local repo="$1"
  local version="$2"

  if [ -n "${BASE_URL:-}" ]; then
    printf '%s\n' "$BASE_URL"
    return
  fi

  if [ "$version" = "latest" ]; then
    printf 'https://github.com/%s/releases/latest/download\n' "$repo"
    return
  fi

  printf 'https://github.com/%s/releases/download/%s\n' "$repo" "$(normalize_tag "$version")"
}

resolve_linux_arch() {
  case "$(uname -m)" in
    x86_64|amd64) printf 'amd64 x86_64\n' ;;
    aarch64|arm64) printf 'arm64 arm64\n' ;;
    *)
      printf 'Unsupported Linux architecture: %s\n' "$(uname -m)" >&2
      exit 1
      ;;
  esac
}

run_with_privilege() {
  if [ "$USE_SUDO" -eq 0 ]; then
    "$@"
    return
  fi

  if [ "$(id -u)" -eq 0 ]; then
    "$@"
    return
  fi

  require_tool sudo
  sudo "$@"
}

download_file() {
  local url="$1"
  local output="$2"

  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$output"
    return
  fi

  if command -v wget >/dev/null 2>&1; then
    wget -qO "$output" "$url"
    return
  fi

  printf 'Required tool missing: curl or wget\n' >&2
  exit 1
}

install_linux() {
  local version="$1"
  local base_url="$2"
  local deb_arch rpm_arch asset url package_file
  read -r deb_arch rpm_arch <<<"$(resolve_linux_arch)"

  if [ "$deb_arch" = "arm64" ]; then
    printf 'Linux ARM package-manager installs are not currently published for Dokbox.\n' >&2
    exit 1
  fi

  if command -v apt >/dev/null 2>&1; then
    require_tool apt
    if [ "$version" = "latest" ]; then
      asset="${PACKAGE_NAME}-linux-${deb_arch}.deb"
    else
      asset="${PACKAGE_NAME}_$(normalize_version "$version")-${PACKAGE_RELEASE}_${deb_arch}.deb"
    fi
    package_file="$TEMP_DIR/$asset"
    url="$base_url/$asset"
    download_file "$url" "$package_file"
    run_with_privilege apt install -y "$package_file"
    return
  fi

  if command -v dnf >/dev/null 2>&1; then
    if [ "$version" = "latest" ]; then
      asset="${PACKAGE_NAME}-linux-${rpm_arch}.rpm"
    else
      asset="${PACKAGE_NAME}-$(normalize_version "$version")-${PACKAGE_RELEASE}.${rpm_arch}.rpm"
    fi
    package_file="$TEMP_DIR/$asset"
    url="$base_url/$asset"
    download_file "$url" "$package_file"
    run_with_privilege dnf install -y "$package_file"
    return
  fi

  if command -v yum >/dev/null 2>&1; then
    if [ "$version" = "latest" ]; then
      asset="${PACKAGE_NAME}-linux-${rpm_arch}.rpm"
    else
      asset="${PACKAGE_NAME}-$(normalize_version "$version")-${PACKAGE_RELEASE}.${rpm_arch}.rpm"
    fi
    package_file="$TEMP_DIR/$asset"
    url="$base_url/$asset"
    download_file "$url" "$package_file"
    run_with_privilege yum install -y "$package_file"
    return
  fi

  if command -v pacman >/dev/null 2>&1; then
    if [ "$version" = "latest" ]; then
      asset="${PACKAGE_NAME}-linux-${rpm_arch}.pkg.tar.zst"
    else
      asset="${PACKAGE_NAME}-$(normalize_version "$version")-${PACKAGE_RELEASE}-${rpm_arch}.pkg.tar.zst"
    fi
    package_file="$TEMP_DIR/$asset"
    url="$base_url/$asset"
    download_file "$url" "$package_file"
    run_with_privilege pacman -U --noconfirm "$package_file"
    return
  fi

  printf 'No supported Linux package manager found.\n' >&2
  exit 1
}

install_macos() {
  local base_url="$1"
  local formula_file="$TEMP_DIR/${PACKAGE_NAME}.rb"

  require_tool brew
  download_file "$base_url/${PACKAGE_NAME}.rb" "$formula_file"
  brew install --formula "$formula_file"
}

VERSION="latest"
REPO="$DEFAULT_REPO"
BASE_URL=""
USE_SUDO=1

while [ "$#" -gt 0 ]; do
  case "$1" in
    --version)
      VERSION="$2"
      shift 2
      ;;
    --repo)
      REPO="$2"
      shift 2
      ;;
    --base-url)
      BASE_URL="$2"
      shift 2
      ;;
    --no-sudo)
      USE_SUDO=0
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      printf 'Unknown option: %s\n' "$1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

TEMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TEMP_DIR"' EXIT

platform="$(uname -s)"
base_url="$(resolve_base_url "$REPO" "$VERSION")"

case "$platform" in
  Linux)
    install_linux "$VERSION" "$base_url"
    ;;
  Darwin)
    install_macos "$base_url"
    ;;
  *)
    printf 'Unsupported platform for install.sh: %s\n' "$platform" >&2
    exit 1
    ;;
esac

printf 'Dokbox installation completed.\n'
