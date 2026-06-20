#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

require_tool makepkg
ensure_standalone_payload

PACMAN_ARCH="$ARCHIVE_ARCH"
PACMAN_STAGE_DIR="$DIST_DIR/pacman"
PACMAN_OUTPUT="${PACKAGE_NAME}-${PACKAGE_VERSION}-${PACKAGE_RELEASE}-${PACMAN_ARCH}.pkg.tar.zst"

rm -rf "$PACMAN_STAGE_DIR" "$DIST_DIR/$PACMAN_OUTPUT"
install -d -m 0755 "$PACMAN_STAGE_DIR"

install -m 0755 "$STANDALONE_BIN" "$PACMAN_STAGE_DIR/$PACKAGE_NAME"
install -m 0644 "$ROOT_DIR/packaging/LICENSE" "$PACMAN_STAGE_DIR/LICENSE"

cat >"$PACMAN_STAGE_DIR/PKGBUILD" <<EOF
pkgname=${PACKAGE_NAME}
pkgver=${PACKAGE_VERSION}
pkgrel=${PACKAGE_RELEASE}
pkgdesc='${PACKAGE_SUMMARY}'
arch=('${PACMAN_ARCH}')
license=('custom')
source=('${PACKAGE_NAME}')
sha256sums=('SKIP')

package() {
  install -Dm755 "\$srcdir/${PACKAGE_NAME}" "\$pkgdir${BIN_PATH}"
  install -Dm644 "\$srcdir/LICENSE" "\$pkgdir/usr/share/licenses/${PACKAGE_NAME}/LICENSE"
}
EOF

(
  cd "$PACMAN_STAGE_DIR"
  makepkg --force --nodeps
)

mv "$PACMAN_STAGE_DIR/$PACMAN_OUTPUT" "$DIST_DIR/$PACMAN_OUTPUT"

printf 'Built %s\n' "$DIST_DIR/$PACMAN_OUTPUT"
