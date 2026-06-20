#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

require_tool dpkg-deb
ensure_standalone_payload

DEB_ARCH="$(map_deb_arch "$ARCHIVE_ARCH")"
DEB_STAGE_DIR="$DIST_DIR/deb"
DEB_CONTROL_DIR="$DEB_STAGE_DIR/DEBIAN"
DEB_OUTPUT="$DIST_DIR/${PACKAGE_NAME}_${PACKAGE_VERSION}-${PACKAGE_RELEASE}_${DEB_ARCH}.deb"

rm -rf "$DEB_STAGE_DIR" "$DEB_OUTPUT"
install -d -m 0755 "$DEB_CONTROL_DIR" "$(dirname "$DEB_STAGE_DIR$BIN_PATH")"

install -m 0755 "$STANDALONE_BIN" "$DEB_STAGE_DIR$BIN_PATH"

{
  printf 'Package: %s\n' "$PACKAGE_NAME"
  printf 'Version: %s-%s\n' "$PACKAGE_VERSION" "$PACKAGE_RELEASE"
  printf 'Section: %s\n' "utils"
  printf 'Priority: %s\n' "optional"
  printf 'Architecture: %s\n' "$DEB_ARCH"
  printf 'Maintainer: %s\n' "$PACKAGE_MAINTAINER"
  printf 'Homepage: %s\n' "$PACKAGE_URL"
  printf 'Description: %s\n' "$PACKAGE_SUMMARY"
  printf ' %s\n' "$PACKAGE_DESCRIPTION"
} >"$DEB_CONTROL_DIR/control"

dpkg-deb --build "$DEB_STAGE_DIR" "$DEB_OUTPUT"

printf 'Built %s\n' "$DEB_OUTPUT"
