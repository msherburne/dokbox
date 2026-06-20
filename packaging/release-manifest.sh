#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

require_tool python3

MANIFEST_OUTPUT="${MANIFEST_OUTPUT:-$DIST_DIR/release-manifest.json}"
RELEASE_BASE_URL="$(resolve_release_base_url)"

mkdir -p "$(dirname "$MANIFEST_OUTPUT")"

entries=()
for target in \
  "linux amd64" \
  "linux arm64" \
  "darwin amd64" \
  "darwin arm64" \
  "windows amd64" \
  "windows arm64"
do
  set -- $target
  target_goos="$1"
  target_goarch="$2"
  target_archive_arch="$(map_archive_arch "$target_goarch")"
  target_archive_name="$(archive_name_for "$target_goos" "$target_archive_arch")"
  target_archive_path="$DIST_DIR/$target_archive_name"

  if [ ! -f "$target_archive_path" ]; then
    continue
  fi

  target_sha256="$(sha256_file "$target_archive_path")"
  target_url="$RELEASE_BASE_URL/$target_archive_name"
  target_archive_root="$(archive_root_for "$target_goos" "$target_archive_arch")"
  target_binary_path="$(binary_path_for "$target_goos")"

  entry="$(
    python3 - "$target_goos" "$target_goarch" "$target_archive_name" "$target_archive_root" "$target_sha256" "$target_url" "$target_binary_path" <<'PY'
import json
import sys

goos, goarch, archive, archive_root, sha256, url, binary = sys.argv[1:]
print(json.dumps({
    "goos": goos,
    "goarch": goarch,
    "archive": archive,
    "archive_root": archive_root,
    "sha256": sha256,
    "checksum": "sha256",
    "url": url,
    "binary": binary,
}, sort_keys=True))
PY
)"
  entries+=("$entry")
done

artifacts_json="[]"
if [ "${#entries[@]}" -gt 0 ]; then
  IFS=,
  artifacts_json="[$(printf '%s' "${entries[*]}")]"
  unset IFS
fi

python3 - "$MANIFEST_OUTPUT" "$PACKAGE_NAME" "$RAW_VERSION" "$PACKAGE_RELEASE" "$PACKAGE_URL" "$PACKAGE_SUMMARY" "$PACKAGE_DESCRIPTION" "$PACKAGE_LICENSE" "$PACKAGE_MAINTAINER" "$PACKAGE_IDENTIFIER" "$RELEASE_BASE_URL" "$artifacts_json" <<'PY'
import json
import pathlib
import sys

(
    output_path,
    package,
    version,
    release,
    homepage,
    summary,
    description,
    license_name,
    publisher,
    identifier,
    release_base_url,
    artifacts_json,
) = sys.argv[1:]

manifest = {
    "package": package,
    "version": version,
    "release": release,
    "homepage": homepage,
    "summary": summary,
    "description": description,
    "license": license_name,
    "publisher": publisher,
    "identifier": identifier,
    "release_base_url": release_base_url,
    "artifacts": json.loads(artifacts_json),
}

path = pathlib.Path(output_path)
path.write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
PY

printf 'Built %s\n' "$MANIFEST_OUTPUT"
