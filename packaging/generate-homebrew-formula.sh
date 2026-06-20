#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

require_tool python3

MANIFEST_INPUT="${MANIFEST_INPUT:-$DIST_DIR/release-manifest.json}"
FORMULA_OUTPUT="${FORMULA_OUTPUT:-$DIST_DIR/dokbox.rb}"

if [ ! -f "$MANIFEST_INPUT" ]; then
  printf 'Required file missing: %s\n' "$MANIFEST_INPUT" >&2
  exit 1
fi

mkdir -p "$(dirname "$FORMULA_OUTPUT")"

python3 - "$MANIFEST_INPUT" "$FORMULA_OUTPUT" <<'PY'
import json
import pathlib
import sys

manifest_path, output_path = sys.argv[1:]
manifest = json.loads(pathlib.Path(manifest_path).read_text(encoding="utf-8"))

artifacts = {
    (item["goos"], item["goarch"]): item
    for item in manifest.get("artifacts", [])
}

required = [("darwin", "amd64"), ("darwin", "arm64")]
missing = [f"{goos}/{goarch}" for goos, goarch in required if (goos, goarch) not in artifacts]
if missing:
    print(f"missing required darwin artifacts: {', '.join(missing)}", file=sys.stderr)
    sys.exit(1)

intel = artifacts[("darwin", "amd64")]
arm = artifacts[("darwin", "arm64")]
summary = manifest.get("summary") or manifest.get("description") or manifest.get("package", "dokbox")

formula = f"""class Dokbox < Formula
  desc {summary!r}
  homepage {manifest["homepage"]!r}
  version {manifest["version"]!r}
  license {manifest["license"]!r}

  on_macos do
    if Hardware::CPU.intel?
      url {intel["url"]!r}
      sha256 {intel["sha256"]!r}
    else
      url {arm["url"]!r}
      sha256 {arm["sha256"]!r}
    end
  end

  def install
    bin.install "bin/dokbox"
  end
end
"""

pathlib.Path(output_path).write_text(formula.replace("'", '"'), encoding="utf-8")
PY

printf 'Built %s\n' "$FORMULA_OUTPUT"
