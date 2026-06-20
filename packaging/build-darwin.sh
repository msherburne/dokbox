#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

require_tool go
require_tool tar

TARGET_GOOS="darwin"
TARGET_ARCHIVE_ROOT="$(archive_root_for "$TARGET_GOOS" "$ARCHIVE_ARCH")"
TARGET_STANDALONE_DIR="$(standalone_dir_for "$TARGET_GOOS" "$ARCHIVE_ARCH")"
TARGET_STANDALONE_ARCHIVE="$(standalone_archive_for "$TARGET_GOOS" "$ARCHIVE_ARCH")"
TARGET_STANDALONE_BIN="$(standalone_bin_for "$TARGET_GOOS" "$ARCHIVE_ARCH")"

rm -rf "$TARGET_STANDALONE_DIR" "$TARGET_STANDALONE_ARCHIVE"
mkdir -p "$(dirname "$TARGET_STANDALONE_BIN")"

GOOS="$TARGET_GOOS" GOARCH="$GO_ARCH" \
go build -trimpath -ldflags="-s -w" -o "$TARGET_STANDALONE_BIN" .

tar -C "$DIST_DIR" -czf "$TARGET_STANDALONE_ARCHIVE" "$TARGET_ARCHIVE_ROOT"

printf 'Built %s and %s\n' "$TARGET_STANDALONE_DIR" "$TARGET_STANDALONE_ARCHIVE"
