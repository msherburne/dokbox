#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

require_tool go
require_tool tar

rm -rf "$STANDALONE_DIR" "$STANDALONE_ARCHIVE"
mkdir -p "$(dirname "$STANDALONE_BIN")"

GOOS=linux GOARCH="$GO_ARCH" \
go build -trimpath -ldflags="-s -w" -o "$STANDALONE_BIN" .

tar -C "$DIST_DIR" -czf "$STANDALONE_ARCHIVE" "$ARCHIVE_ROOT"

printf 'Built %s and %s\n' "$STANDALONE_DIR" "$STANDALONE_ARCHIVE"
