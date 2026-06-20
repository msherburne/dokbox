#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

require_tool python3

MANIFEST_INPUT="${MANIFEST_INPUT:-$DIST_DIR/release-manifest.json}"
WINGET_OUTPUT_DIR="${WINGET_OUTPUT_DIR:-$DIST_DIR/winget}"

if [ ! -f "$MANIFEST_INPUT" ]; then
  printf 'Required file missing: %s\n' "$MANIFEST_INPUT" >&2
  exit 1
fi

mkdir -p "$WINGET_OUTPUT_DIR"

python3 - "$MANIFEST_INPUT" "$WINGET_OUTPUT_DIR" <<'PY'
import json
import pathlib
import sys

manifest_path, output_dir = sys.argv[1:]
manifest = json.loads(pathlib.Path(manifest_path).read_text(encoding="utf-8"))

artifacts = {
    (item["goos"], item["goarch"]): item
    for item in manifest.get("artifacts", [])
}

required = [("windows", "amd64")]
missing = [f"{goos}/{goarch}" for goos, goarch in required if (goos, goarch) not in artifacts]
if missing:
    print(f"missing required windows artifacts: {', '.join(missing)}", file=sys.stderr)
    sys.exit(1)

identifier = manifest.get("identifier", "dokbox.dokbox")
package_name = manifest.get("package", "dokbox")
version = manifest["version"]
publisher = manifest["publisher"]
homepage = manifest["homepage"]
license_name = manifest["license"]
description = manifest["description"]

architectures = {"amd64": "x64"}
installers = []
for goarch in ("amd64",):
    item = artifacts[("windows", goarch)]
    installers.extend([
        f"- Architecture: {architectures[goarch]}",
        f"  InstallerType: zip",
        f"  NestedInstallerType: portable",
        f"  InstallerUrl: {item['url']}",
        f"  InstallerSha256: {item['sha256']}",
        f"  NestedInstallerFiles:",
        f"    - RelativeFilePath: {item.get('binary', 'bin/dokbox.exe')}",
        f"      PortableCommandAlias: {package_name}",
    ])

root = pathlib.Path(output_dir)
root.mkdir(parents=True, exist_ok=True)

(root / f"{package_name}.yaml").write_text(
    "\n".join([
        "PackageIdentifier: " + identifier,
        "PackageVersion: " + version,
        "DefaultLocale: en-US",
        "ManifestType: version",
        "ManifestVersion: 1.6.0",
        "",
    ]),
    encoding="utf-8",
)

(root / f"{package_name}.locale.en-US.yaml").write_text(
    "\n".join([
        "PackageIdentifier: " + identifier,
        "PackageVersion: " + version,
        "PackageLocale: en-US",
        "Publisher: " + publisher,
        "PackageName: " + package_name,
        "License: " + license_name,
        "ShortDescription: " + manifest.get("summary", package_name),
        "Description: " + description,
        "Homepage: " + homepage,
        "ManifestType: defaultLocale",
        "ManifestVersion: 1.6.0",
        "",
    ]),
    encoding="utf-8",
)

(root / f"{package_name}.installer.yaml").write_text(
    "\n".join([
        "PackageIdentifier: " + identifier,
        "PackageVersion: " + version,
        "InstallModes:",
        "- interactive",
        "- silent",
        "Installers:",
        *installers,
        "ManifestType: installer",
        "ManifestVersion: 1.6.0",
        "",
    ]),
    encoding="utf-8",
)
PY

printf 'Built %s\n' "$WINGET_OUTPUT_DIR"
