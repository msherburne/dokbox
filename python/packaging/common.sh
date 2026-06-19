#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="$ROOT_DIR/dist"
PACKAGE_NAME="dokbox"
resolve_python() {
  if command -v python3 >/dev/null 2>&1; then
    printf '%s\n' python3
    return
  fi

  if command -v python >/dev/null 2>&1; then
    printf '%s\n' python
    return
  fi

  printf 'Required tool missing: python3 (or python)\n' >&2
  exit 1
}

PACKAGE_VERSION="$(ROOT_DIR="$ROOT_DIR" "$(resolve_python)" - <<'PY'
import os
import re
from pathlib import Path

pyproject = Path(os.environ["ROOT_DIR"]) / "pyproject.toml"
content = pyproject.read_text(encoding="utf-8")

try:
    import tomllib
except ModuleNotFoundError:
    tomllib = None

if tomllib is not None:
    print(tomllib.loads(content)["project"]["version"])
    raise SystemExit

match = re.search(r'(?ms)^\[project\].*?^version\s*=\s*"([^"]+)"', content)
if match is None:
    raise SystemExit(f"Could not determine project version from {pyproject}")

print(match.group(1))
PY
)"
PACKAGE_RELEASE="1"
HOST_ARCH="$(uname -m)"
STANDALONE_ARCH="x86_64"
PACKAGE_ARCH="$STANDALONE_ARCH"
STANDALONE_SUFFIX="linux-$STANDALONE_ARCH"
STANDALONE_BASENAME="$PACKAGE_NAME-$STANDALONE_SUFFIX"
INSTALL_ROOT="/opt/dokbox"
BIN_PATH="/usr/bin/dokbox"
STANDALONE_DIR="$DIST_DIR/$STANDALONE_BASENAME.dist"
STANDALONE_ARCHIVE="$DIST_DIR/$STANDALONE_BASENAME.tar.gz"
PACKAGE_SUMMARY="Terminal UI for Docker resource management"
PACKAGE_DESCRIPTION="dokbox is a Textual-based terminal application for browsing and managing Docker resources."
PACKAGE_LICENSE="Unspecified"
PACKAGE_URL="https://github.com/msherburne/dokbox"
PACKAGE_HOMEPAGE="$PACKAGE_URL"
PACKAGE_MAINTAINER="dokbox maintainers"
PACKAGE_VENDOR="dokbox"
PACKAGE_RUNTIME_DEPENDS=""

require_tool() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'Required tool missing: %s\n' "$1" >&2
    exit 1
  fi
}

ensure_standalone_payload() {
  if [ ! -d "$STANDALONE_DIR" ]; then
    "$ROOT_DIR/packaging/build-linux.sh"
  fi
}

stage_standalone_payload() {
  stage_root="$1"

  mkdir -p "$stage_root$INSTALL_ROOT" "$(dirname "$stage_root$BIN_PATH")"
  cp -a "$STANDALONE_DIR"/. "$stage_root$INSTALL_ROOT"/
  ln -sf "$INSTALL_ROOT/dokbox" "$stage_root$BIN_PATH"
}
