#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

map_deb_arch() {
  case "$1" in
    x86_64) printf 'amd64\n' ;;
    aarch64 | arm64) printf 'arm64\n' ;;
    *)
      printf 'Unsupported Debian architecture: %s\n' "$1" >&2
      exit 1
      ;;
  esac
}

require_tool dpkg-deb
ensure_standalone_payload

DEB_ARCH="$(map_deb_arch "$PACKAGE_ARCH")"
DEB_STAGE_DIR="$DIST_DIR/deb"
DEB_CONTROL_DIR="$DEB_STAGE_DIR/DEBIAN"
DEB_OUTPUT="$DIST_DIR/${PACKAGE_NAME}_${PACKAGE_VERSION}-${PACKAGE_RELEASE}_${DEB_ARCH}.deb"

rm -rf "$DEB_STAGE_DIR" "$DEB_OUTPUT"
mkdir -p "$DEB_CONTROL_DIR"

stage_standalone_payload "$DEB_STAGE_DIR"

{
  printf 'Package: %s\n' "$PACKAGE_NAME"
  printf 'Version: %s-%s\n' "$PACKAGE_VERSION" "$PACKAGE_RELEASE"
  printf 'Section: %s\n' "utils"
  printf 'Priority: %s\n' "optional"
  printf 'Architecture: %s\n' "$DEB_ARCH"
  printf 'Maintainer: %s\n' "$PACKAGE_MAINTAINER"
  if [ -n "$PACKAGE_RUNTIME_DEPENDS" ]; then
    printf 'Depends: %s\n' "$PACKAGE_RUNTIME_DEPENDS"
  fi
  printf 'Homepage: %s\n' "$PACKAGE_HOMEPAGE"
  printf 'Description: %s\n' "$PACKAGE_SUMMARY"
  printf ' %s\n' "$PACKAGE_DESCRIPTION"
} >"$DEB_CONTROL_DIR/control"

dpkg-deb --build "$DEB_STAGE_DIR" "$DEB_OUTPUT"

printf 'Built %s\n' "$DEB_OUTPUT"
