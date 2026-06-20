#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

require_tool rpmbuild
ensure_standalone_payload

RPM_ARCH="$(map_rpm_arch "$ARCHIVE_ARCH")"
RPM_TOPDIR="$DIST_DIR/rpmbuild"
RPM_BUILDROOT="$RPM_TOPDIR/BUILDROOT/${PACKAGE_NAME}-${PACKAGE_VERSION}-${PACKAGE_RELEASE}.${RPM_ARCH}"
RPM_SPEC_DIR="$RPM_TOPDIR/SPECS"
RPM_SPEC_FILE="$RPM_SPEC_DIR/${PACKAGE_NAME}.spec"
RPM_OUTPUT="$DIST_DIR/${PACKAGE_NAME}-${PACKAGE_VERSION}-${PACKAGE_RELEASE}.${RPM_ARCH}.rpm"

rm -rf "$RPM_TOPDIR" "$RPM_OUTPUT"
mkdir -p \
  "$RPM_BUILDROOT$(dirname "$BIN_PATH")" \
  "$RPM_SPEC_DIR" \
  "$RPM_TOPDIR/BUILD" \
  "$RPM_TOPDIR/RPMS/$RPM_ARCH" \
  "$RPM_TOPDIR/SOURCES" \
  "$RPM_TOPDIR/SRPMS"

install -m 0755 "$STANDALONE_BIN" "$RPM_BUILDROOT$BIN_PATH"

cat >"$RPM_SPEC_FILE" <<EOF
Name: ${PACKAGE_NAME}
Version: ${PACKAGE_VERSION}
Release: ${PACKAGE_RELEASE}
Summary: ${PACKAGE_SUMMARY}
License: ${PACKAGE_LICENSE}
URL: ${PACKAGE_URL}
BuildArch: ${RPM_ARCH}

%description
${PACKAGE_DESCRIPTION}

%files
${BIN_PATH}
EOF

rpmbuild -bb --define "_topdir $RPM_TOPDIR" "$RPM_SPEC_FILE"

cp "$RPM_TOPDIR/RPMS/$RPM_ARCH/${PACKAGE_NAME}-${PACKAGE_VERSION}-${PACKAGE_RELEASE}.${RPM_ARCH}.rpm" "$RPM_OUTPUT"

printf 'Built %s\n' "$RPM_OUTPUT"
