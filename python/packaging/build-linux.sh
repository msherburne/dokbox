#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="dist"
STANDALONE_DIR="$DIST_DIR/dokbox-linux-x86_64.dist"
STANDALONE_ARCHIVE="$DIST_DIR/dokbox-linux-x86_64.tar.gz"

cd "$ROOT_DIR"

if ! command -v patchelf >/dev/null 2>&1; then
  printf 'patchelf is required for Linux standalone builds.\n' >&2
  printf 'Install it first, for example: sudo apt-get install patchelf\n' >&2
  exit 1
fi

mkdir -p "$DIST_DIR"
rm -rf "$STANDALONE_DIR" "$STANDALONE_ARCHIVE"

PYTHONPATH="$ROOT_DIR/src${PYTHONPATH:+:$PYTHONPATH}" \
uv run python -m nuitka \
  --standalone \
  --assume-yes-for-downloads \
  --remove-output \
  --output-dir=dist \
  --output-folder-name=dokbox-linux-x86_64 \
  --output-filename=dokbox \
  --include-package=dokbox \
  src/main.py

tar -C "$DIST_DIR" -czf "$STANDALONE_ARCHIVE" dokbox-linux-x86_64.dist

printf 'Built %s and %s\n' "$STANDALONE_DIR" "$STANDALONE_ARCHIVE"
