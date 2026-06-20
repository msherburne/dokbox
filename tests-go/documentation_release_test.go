package testsgo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseDocumentationMentionsCrossPlatformBuilds(t *testing.T) {
	root := projectRoot(t)

	developmentDoc, err := os.ReadFile(filepath.Join(root, "docs", "DEVELOPMENT.md"))
	if err != nil {
		t.Fatalf("read development doc: %v", err)
	}

	packagingReadme, err := os.ReadFile(filepath.Join(root, "packaging", "README.md"))
	if err != nil {
		t.Fatalf("read packaging readme: %v", err)
	}

	text := string(developmentDoc) + "\n" + string(packagingReadme)

	for _, snippet := range []string{
		"./packaging/build-darwin.sh",
		"./packaging/build-windows.sh",
		"./packaging/release-manifest.sh",
		"./packaging/generate-homebrew-formula.sh",
		"./packaging/generate-winget-manifest.sh",
		"./install/install.sh",
		"./install/install.ps1",
		"Homebrew",
		"winget",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("expected release documentation to mention %q", snippet)
		}
	}
}
