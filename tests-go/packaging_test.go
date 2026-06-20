package testsgo

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func projectRoot(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine caller path")
	}

	return filepath.Dir(filepath.Dir(filename))
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()

	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(dst), err)
	}

	if err := os.WriteFile(dst, data, 0o755); err != nil {
		t.Fatalf("write %s: %v", dst, err)
	}
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func runScript(t *testing.T, root, script string, pathEntries ...string) {
	runScriptWithEnv(t, root, script, map[string]string{
		"VERSION": "1.2.3",
		"RELEASE": "1",
	}, pathEntries...)
}

func runScriptWithEnv(t *testing.T, root, script string, envVars map[string]string, pathEntries ...string) {
	t.Helper()

	cmd := exec.Command("bash", script)
	cmd.Dir = root

	env := os.Environ()
	env = append(env, "PATH="+strings.Join(append(pathEntries, os.Getenv("PATH")), string(os.PathListSeparator)))
	for key, value := range envVars {
		env = append(env, key+"="+value)
	}
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v\n%s", script, err, output)
	}
}

func runScriptExpectFailure(t *testing.T, root, script string, envVars map[string]string, pathEntries ...string) string {
	t.Helper()

	cmd := exec.Command("bash", script)
	cmd.Dir = root

	env := os.Environ()
	env = append(env, "PATH="+strings.Join(append(pathEntries, os.Getenv("PATH")), string(os.PathListSeparator)))
	for key, value := range envVars {
		env = append(env, key+"="+value)
	}
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("%s unexpectedly succeeded\n%s", script, output)
	}

	return string(output)
}

func TestPackagingDocsDescribeGoNativeLinuxBuilds(t *testing.T) {
	root := projectRoot(t)

	developmentDoc, err := os.ReadFile(filepath.Join(root, "docs", "DEVELOPMENT.md"))
	if err != nil {
		t.Fatalf("read development doc: %v", err)
	}

	packagingReadme, err := os.ReadFile(filepath.Join(root, "packaging", "README.md"))
	if err != nil {
		t.Fatalf("read packaging readme: %v", err)
	}

	docText := string(developmentDoc)
	readmeText := string(packagingReadme)

	for _, snippet := range []string{
		"./packaging/build-linux.sh",
		"./packaging/build-deb.sh",
		"./packaging/build-rpm.sh",
		"./packaging/build-pacman.sh",
		"apt install ./dist/",
		"dnf install ./dist/",
		"pacman -U ./dist/",
	} {
		if !strings.Contains(docText, snippet) && !strings.Contains(readmeText, snippet) {
			t.Fatalf("expected packaging docs to mention %q", snippet)
		}
	}
}

func TestBuildLinuxHelperProducesStandaloneArchive(t *testing.T) {
	root := t.TempDir()
	packagingDir := filepath.Join(root, "packaging")
	fakeBin := filepath.Join(root, "fake-bin")

	copyFile(t, filepath.Join(projectRoot(t), "packaging", "common.sh"), filepath.Join(packagingDir, "common.sh"))
	copyFile(t, filepath.Join(projectRoot(t), "packaging", "build-linux.sh"), filepath.Join(packagingDir, "build-linux.sh"))

	writeExecutable(t, filepath.Join(fakeBin, "go"), `#!/usr/bin/env bash
set -euo pipefail
if [ "$1" = "env" ] && [ "$2" = "GOARCH" ]; then
  printf 'amd64\n'
  exit 0
fi
if [ "$1" = "build" ]; then
  output=""
  while [ "$#" -gt 0 ]; do
    if [ "$1" = "-o" ]; then
      output="$2"
      shift 2
      continue
    fi
    shift
  done
  mkdir -p "$(dirname "$output")"
  printf '#!/usr/bin/env bash\necho dokbox\n' >"$output"
  chmod +x "$output"
  exit 0
fi
printf 'unexpected go invocation: %s\n' "$*" >&2
exit 1
`)

	runScript(t, root, filepath.Join("packaging", "build-linux.sh"), fakeBin)

	if _, err := os.Stat(filepath.Join(root, "dist", "dokbox-linux-x86_64", "bin", "dokbox")); err != nil {
		t.Fatalf("expected standalone binary: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "dist", "dokbox-linux-x86_64.tar.gz")); err != nil {
		t.Fatalf("expected standalone archive: %v", err)
	}
}

func TestBuildDebHelperStagesUsrBinPackageWithoutDockerDependency(t *testing.T) {
	root := t.TempDir()
	packagingDir := filepath.Join(root, "packaging")
	fakeBin := filepath.Join(root, "fake-bin")

	copyFile(t, filepath.Join(projectRoot(t), "packaging", "common.sh"), filepath.Join(packagingDir, "common.sh"))
	copyFile(t, filepath.Join(projectRoot(t), "packaging", "build-deb.sh"), filepath.Join(packagingDir, "build-deb.sh"))
	copyFile(t, filepath.Join(projectRoot(t), "packaging", "build-linux.sh"), filepath.Join(packagingDir, "build-linux.sh"))

	writeExecutable(t, filepath.Join(root, "dist", "dokbox-linux-x86_64", "bin", "dokbox"), "#!/usr/bin/env bash\necho dokbox\n")

	writeExecutable(t, filepath.Join(fakeBin, "dpkg-deb"), `#!/usr/bin/env bash
set -euo pipefail
stage_dir="$2"
output="$3"
control_file="$stage_dir/DEBIAN/control"
grep -q '^Package: dokbox$' "$control_file"
grep -q '^Description: ' "$control_file"
if grep -q '^Depends: .*docker' "$control_file"; then
  printf 'docker dependency should not be present\n' >&2
  exit 1
fi
test "$(stat -c '%a' "$stage_dir")" = "755"
test "$(stat -c '%a' "$stage_dir/DEBIAN")" = "755"
test -x "$stage_dir/usr/bin/dokbox"
mkdir -p "$(dirname "$output")"
touch "$output"
`)

	runScript(t, root, filepath.Join("packaging", "build-deb.sh"), fakeBin)

	if _, err := os.Stat(filepath.Join(root, "dist", "dokbox_1.2.3-1_amd64.deb")); err != nil {
		t.Fatalf("expected deb package: %v", err)
	}
}

func TestBuildRpmHelperStagesUsrBinPackage(t *testing.T) {
	root := t.TempDir()
	packagingDir := filepath.Join(root, "packaging")
	fakeBin := filepath.Join(root, "fake-bin")

	copyFile(t, filepath.Join(projectRoot(t), "packaging", "common.sh"), filepath.Join(packagingDir, "common.sh"))
	copyFile(t, filepath.Join(projectRoot(t), "packaging", "build-rpm.sh"), filepath.Join(packagingDir, "build-rpm.sh"))
	copyFile(t, filepath.Join(projectRoot(t), "packaging", "build-linux.sh"), filepath.Join(packagingDir, "build-linux.sh"))

	writeExecutable(t, filepath.Join(root, "dist", "dokbox-linux-x86_64", "bin", "dokbox"), "#!/usr/bin/env bash\necho dokbox\n")

	writeExecutable(t, filepath.Join(fakeBin, "rpmbuild"), `#!/usr/bin/env bash
set -euo pipefail
spec_file="${@: -1}"
grep -q '^Name: dokbox$' "$spec_file"
grep -q '/usr/bin/dokbox' "$spec_file"
topdir=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "--define" ] && [ "${2%% *}" = "_topdir" ]; then
    topdir="${2#_topdir }"
    shift 2
    continue
  fi
  shift
done
test -x "$topdir/BUILDROOT/dokbox-1.2.3-1.x86_64/usr/bin/dokbox"
mkdir -p "$topdir/RPMS/x86_64"
touch "$topdir/RPMS/x86_64/dokbox-1.2.3-1.x86_64.rpm"
`)

	runScript(t, root, filepath.Join("packaging", "build-rpm.sh"), fakeBin)

	if _, err := os.Stat(filepath.Join(root, "dist", "dokbox-1.2.3-1.x86_64.rpm")); err != nil {
		t.Fatalf("expected rpm package: %v", err)
	}
}

func TestBuildRpmHelperUsesSafeDefaultVersion(t *testing.T) {
	root := t.TempDir()
	packagingDir := filepath.Join(root, "packaging")
	fakeBin := filepath.Join(root, "fake-bin")

	copyFile(t, filepath.Join(projectRoot(t), "packaging", "common.sh"), filepath.Join(packagingDir, "common.sh"))
	copyFile(t, filepath.Join(projectRoot(t), "packaging", "build-rpm.sh"), filepath.Join(packagingDir, "build-rpm.sh"))

	writeExecutable(t, filepath.Join(root, "dist", "dokbox-linux-x86_64", "bin", "dokbox"), "#!/usr/bin/env bash\necho dokbox\n")

	writeExecutable(t, filepath.Join(fakeBin, "rpmbuild"), `#!/usr/bin/env bash
set -euo pipefail
spec_file="${@: -1}"
version="$(awk '/^Version:/ { print $2 }' "$spec_file")"
case "$version" in
  *-*)
    printf 'rpm version contains invalid hyphen: %s\n' "$version" >&2
    exit 1
    ;;
esac
topdir=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "--define" ] && [ "${2%% *}" = "_topdir" ]; then
    topdir="${2#_topdir }"
    shift 2
    continue
  fi
  shift
done
mkdir -p "$topdir/RPMS/x86_64"
touch "$topdir/RPMS/x86_64/dokbox-0.1.0-1.x86_64.rpm"
`)

	runScriptWithEnv(t, root, filepath.Join("packaging", "build-rpm.sh"), map[string]string{
		"RELEASE": "1",
	}, fakeBin)

	if _, err := os.Stat(filepath.Join(root, "dist", "dokbox-0.1.0-1.x86_64.rpm")); err != nil {
		t.Fatalf("expected rpm package with safe fallback version: %v", err)
	}
}

func TestBuildPacmanHelperCreatesPkgArchive(t *testing.T) {
	root := t.TempDir()
	packagingDir := filepath.Join(root, "packaging")
	fakeBin := filepath.Join(root, "fake-bin")

	copyFile(t, filepath.Join(projectRoot(t), "packaging", "common.sh"), filepath.Join(packagingDir, "common.sh"))
	copyFile(t, filepath.Join(projectRoot(t), "packaging", "build-pacman.sh"), filepath.Join(packagingDir, "build-pacman.sh"))
	copyFile(t, filepath.Join(projectRoot(t), "packaging", "build-linux.sh"), filepath.Join(packagingDir, "build-linux.sh"))
	copyFile(t, filepath.Join(projectRoot(t), "packaging", "LICENSE"), filepath.Join(packagingDir, "LICENSE"))

	writeExecutable(t, filepath.Join(root, "dist", "dokbox-linux-x86_64", "bin", "dokbox"), "#!/usr/bin/env bash\necho dokbox\n")
	writeExecutable(t, filepath.Join(fakeBin, "bsdtar"), "#!/usr/bin/env bash\nexit 0\n")

	writeExecutable(t, filepath.Join(fakeBin, "makepkg"), `#!/usr/bin/env bash
set -euo pipefail
test "${PACMAN:-}" = "true"
grep -q '^pkgname=dokbox$' PKGBUILD
grep -q '^pkgver=1.2.3$' PKGBUILD
grep -q '/usr/bin/dokbox' PKGBUILD
grep -q "^source=('dokbox' 'LICENSE')$" PKGBUILD
grep -q "^sha256sums=('SKIP' 'SKIP')$" PKGBUILD
grep -q '/usr/share/licenses/dokbox/LICENSE' PKGBUILD
test -f LICENSE
touch "dokbox-1.2.3-1-x86_64.pkg.tar.zst"
`)

	runScript(t, root, filepath.Join("packaging", "build-pacman.sh"), fakeBin)

	if _, err := os.Stat(filepath.Join(root, "dist", "dokbox-1.2.3-1-x86_64.pkg.tar.zst")); err != nil {
		t.Fatalf("expected pacman package: %v", err)
	}
}

func TestBuildPacmanHelperUsesSafeDefaultVersion(t *testing.T) {
	root := t.TempDir()
	packagingDir := filepath.Join(root, "packaging")
	fakeBin := filepath.Join(root, "fake-bin")

	copyFile(t, filepath.Join(projectRoot(t), "packaging", "common.sh"), filepath.Join(packagingDir, "common.sh"))
	copyFile(t, filepath.Join(projectRoot(t), "packaging", "build-pacman.sh"), filepath.Join(packagingDir, "build-pacman.sh"))
	copyFile(t, filepath.Join(projectRoot(t), "packaging", "LICENSE"), filepath.Join(packagingDir, "LICENSE"))

	writeExecutable(t, filepath.Join(root, "dist", "dokbox-linux-x86_64", "bin", "dokbox"), "#!/usr/bin/env bash\necho dokbox\n")
	writeExecutable(t, filepath.Join(fakeBin, "bsdtar"), "#!/usr/bin/env bash\nexit 0\n")

	writeExecutable(t, filepath.Join(fakeBin, "makepkg"), `#!/usr/bin/env bash
set -euo pipefail
test "${PACMAN:-}" = "true"
pkgver="$(awk -F= '/^pkgver=/ { print $2 }' PKGBUILD)"
case "$pkgver" in
  *-*)
    printf 'pkgver contains invalid hyphen: %s\n' "$pkgver" >&2
    exit 1
    ;;
esac
touch "dokbox-0.1.0-1-x86_64.pkg.tar.zst"
`)

	runScriptWithEnv(t, root, filepath.Join("packaging", "build-pacman.sh"), map[string]string{
		"RELEASE": "1",
	}, fakeBin)

	if _, err := os.Stat(filepath.Join(root, "dist", "dokbox-0.1.0-1-x86_64.pkg.tar.zst")); err != nil {
		t.Fatalf("expected pacman package with safe fallback version: %v", err)
	}
}

func TestBuildPacmanHelperFailsFastWhenBsdtarIsMissing(t *testing.T) {
	root := t.TempDir()
	packagingDir := filepath.Join(root, "packaging")

	copyFile(t, filepath.Join(projectRoot(t), "packaging", "common.sh"), filepath.Join(packagingDir, "common.sh"))
	copyFile(t, filepath.Join(projectRoot(t), "packaging", "build-pacman.sh"), filepath.Join(packagingDir, "build-pacman.sh"))
	copyFile(t, filepath.Join(projectRoot(t), "packaging", "LICENSE"), filepath.Join(packagingDir, "LICENSE"))

	writeExecutable(t, filepath.Join(root, "dist", "dokbox-linux-x86_64", "bin", "dokbox"), "#!/usr/bin/env bash\necho dokbox\n")
	writeExecutable(t, filepath.Join(root, "empty-bin", "apt"), "#!/usr/bin/env bash\nexit 1\n")

	output := runScriptExpectFailure(t, root, filepath.Join("packaging", "build-pacman.sh"), map[string]string{
		"VERSION": "1.2.3",
		"RELEASE": "1",
	}, filepath.Join(root, "empty-bin"))

	if !strings.Contains(output, "Required tool missing: bsdtar") {
		t.Fatalf("expected missing bsdtar error, got:\n%s", output)
	}
}
