package testsgo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runScriptWithPath(t *testing.T, root, script string, pathEntries ...string) {
	t.Helper()

	cmd := exec.Command("bash", script)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "PATH="+strings.Join(append(pathEntries, os.Getenv("PATH")), string(os.PathListSeparator)))

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v\n%s", script, err, output)
	}
}

func TestInstallScriptUsesAptForLinuxDeb(t *testing.T) {
	root := t.TempDir()
	installDir := filepath.Join(root, "install")
	fakeBin := filepath.Join(root, "fake-bin")
	logFile := filepath.Join(root, "calls.log")

	copyFile(t, filepath.Join(projectRoot(t), "install", "install.sh"), filepath.Join(installDir, "install.sh"))

	writeExecutable(t, filepath.Join(fakeBin, "uname"), `#!/usr/bin/env bash
set -euo pipefail
case "${1:-}" in
  -s) printf 'Linux\n' ;;
  -m) printf 'x86_64\n' ;;
  *) printf 'Linux\n' ;;
esac
`)

	writeExecutable(t, filepath.Join(fakeBin, "id"), `#!/usr/bin/env bash
set -euo pipefail
if [ "${1:-}" = "-u" ]; then
  printf '1000\n'
  exit 0
fi
printf 'unexpected id invocation\n' >&2
exit 1
`)

	writeExecutable(t, filepath.Join(fakeBin, "curl"), fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
output=""
url=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    -o)
      output="$2"
      shift 2
      ;;
    -*)
      shift
      ;;
    *)
      url="$1"
      shift
      ;;
  esac
done
printf 'curl %%s\n' "$url" >>%q
printf 'payload\n' >"$output"
`, logFile))

	writeExecutable(t, filepath.Join(fakeBin, "sudo"), fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
printf 'sudo %%s\n' "$*" >>%q
"$@"
`, logFile))

	writeExecutable(t, filepath.Join(fakeBin, "apt"), fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
printf 'apt %%s\n' "$*" >>%q
`, logFile))

	runScriptWithPath(t, root, filepath.Join("install", "install.sh"), fakeBin)

	logData, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}

	logText := string(logData)
	if !strings.Contains(logText, "dokbox-linux-amd64.deb") {
		t.Fatalf("expected apt installer to download a deb asset, got %q", logText)
	}
	if !strings.Contains(logText, "sudo apt install -y") {
		t.Fatalf("expected apt installer to use sudo apt install, got %q", logText)
	}
}

func TestInstallScriptUsesBrewFormulaOnMacOS(t *testing.T) {
	root := t.TempDir()
	installDir := filepath.Join(root, "install")
	fakeBin := filepath.Join(root, "fake-bin")
	logFile := filepath.Join(root, "calls.log")

	copyFile(t, filepath.Join(projectRoot(t), "install", "install.sh"), filepath.Join(installDir, "install.sh"))

	writeExecutable(t, filepath.Join(fakeBin, "uname"), `#!/usr/bin/env bash
set -euo pipefail
case "${1:-}" in
  -s) printf 'Darwin\n' ;;
  -m) printf 'arm64\n' ;;
  *) printf 'Darwin\n' ;;
esac
`)

	writeExecutable(t, filepath.Join(fakeBin, "curl"), fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
output=""
url=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    -o)
      output="$2"
      shift 2
      ;;
    -*)
      shift
      ;;
    *)
      url="$1"
      shift
      ;;
  esac
done
printf 'curl %%s\n' "$url" >>%q
printf 'class Dokbox < Formula>\nend\n' >"$output"
`, logFile))

	writeExecutable(t, filepath.Join(fakeBin, "brew"), fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
printf 'brew %%s\n' "$*" >>%q
`, logFile))

	runScriptWithPath(t, root, filepath.Join("install", "install.sh"), fakeBin)

	logData, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}

	logText := string(logData)
	if !strings.Contains(logText, "dokbox.rb") {
		t.Fatalf("expected macOS installer to download dokbox.rb, got %q", logText)
	}
	if !strings.Contains(logText, "brew install --formula") {
		t.Fatalf("expected macOS installer to use brew install, got %q", logText)
	}
}

func TestWindowsInstallerAndWorkflowReferenceReleaseAssets(t *testing.T) {
	root := projectRoot(t)

	powershellInstaller, err := os.ReadFile(filepath.Join(root, "install", "install.ps1"))
	if err != nil {
		t.Fatalf("read PowerShell installer: %v", err)
	}

	releaseWorkflow, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatalf("read release workflow: %v", err)
	}

	for _, snippet := range []string{
		"dokbox-winget-manifests.zip",
		"winget install --manifest",
		"./packaging/build-darwin.sh",
		"./packaging/build-windows.sh",
		"./packaging/release-manifest.sh",
		"softprops/action-gh-release",
		"dist/dokbox.rb",
		"dist/install.sh",
		"dist/install.ps1",
	} {
		if !strings.Contains(string(powershellInstaller), snippet) && !strings.Contains(string(releaseWorkflow), snippet) {
			t.Fatalf("expected release automation to mention %q", snippet)
		}
	}
}

func TestReleaseWorkflowUsesReducedArchitectureMatrix(t *testing.T) {
	root := projectRoot(t)

	releaseWorkflow, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatalf("read release workflow: %v", err)
	}

	text := string(releaseWorkflow)

	required := []string{
		"goarch: [amd64, arm64]",
		"GOARCH=amd64 ./packaging/build-linux.sh",
		"GOARCH=amd64 ./packaging/build-windows.sh",
		`cp "dist/dokbox_${VERSION}-1_amd64.deb" dist/dokbox-linux-amd64.deb`,
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("expected workflow to contain %q", snippet)
		}
	}

	for _, snippet := range []string{
		"GOARCH=arm64 ./packaging/build-linux.sh",
		"GOARCH=arm64 ./packaging/build-deb.sh",
		`cp "dist/dokbox_${VERSION}-1_arm64.deb" dist/dokbox-linux-arm64.deb`,
		"name: Build Windows standalone (${{ matrix.goarch }})",
	} {
		if strings.Contains(text, snippet) {
			t.Fatalf("expected workflow to omit %q", snippet)
		}
	}
}
