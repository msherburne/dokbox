package testsgo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/msherburne/dokbox/internal/config"
	"github.com/msherburne/dokbox/internal/theme"
)

func TestConfigPathUsesHomeDokboxJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := config.Path()
	if err != nil {
		t.Fatalf("expected config path, got error: %v", err)
	}

	want := filepath.Join(home, ".dokbox.json")
	if got != want {
		t.Fatalf("expected config path %q, got %q", want, got)
	}
}

func TestResolveCurrentConfigMigratesAndNormalizesSavedFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(home, ".dokbox.json")
	if err := os.WriteFile(configPath, []byte(`{"docker_host":"unix:///var/run/docker.sock"}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := config.ResolveCurrentConfig()
	if err != nil {
		t.Fatalf("resolve config: %v", err)
	}

	if cfg == nil {
		t.Fatal("expected resolved config")
	}

	if cfg.ConfigVersion != config.CurrentConfigVersion {
		t.Fatalf("expected config version %d, got %d", config.CurrentConfigVersion, cfg.ConfigVersion)
	}

	if cfg.DockerHost != "unix:///var/run/docker.sock" {
		t.Fatalf("expected docker host to be preserved, got %q", cfg.DockerHost)
	}

	if cfg.Theme != "default" {
		t.Fatalf("expected missing theme to resolve to default, got %q", cfg.Theme)
	}

	savedBytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read saved config: %v", err)
	}

	var saved map[string]any
	if err := json.Unmarshal(savedBytes, &saved); err != nil {
		t.Fatalf("unmarshal saved config: %v", err)
	}

	if saved["config_version"] != float64(config.CurrentConfigVersion) {
		t.Fatalf("expected saved config_version %d, got %#v", config.CurrentConfigVersion, saved["config_version"])
	}

	if saved["docker_host"] != "unix:///var/run/docker.sock" {
		t.Fatalf("expected saved docker_host to be preserved, got %#v", saved["docker_host"])
	}

	if saved["theme"] != "default" {
		t.Fatalf("expected saved theme to be default, got %#v", saved["theme"])
	}
}

func TestResolveCurrentConfigDefaultsMissingThemeToDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(home, ".dokbox.json")
	if err := os.WriteFile(configPath, []byte(`{"config_version":1,"docker_host":"unix:///var/run/docker.sock"}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := config.ResolveCurrentConfig()
	if err != nil {
		t.Fatalf("resolve config: %v", err)
	}

	if cfg == nil {
		t.Fatal("expected resolved config")
	}

	if cfg.Theme != "default" {
		t.Fatalf("expected missing theme to resolve to default, got %q", cfg.Theme)
	}
}

func TestResolveCurrentConfigPreservesUnknownTheme(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(home, ".dokbox.json")
	if err := os.WriteFile(configPath, []byte(`{"config_version":1,"docker_host":"unix:///var/run/docker.sock","theme":"aurora"}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := config.ResolveCurrentConfig()
	if err != nil {
		t.Fatalf("resolve config: %v", err)
	}

	if cfg == nil {
		t.Fatal("expected resolved config")
	}

	if cfg.Theme != "aurora" {
		t.Fatalf("expected unknown theme to remain configured as %q, got %q", "aurora", cfg.Theme)
	}

	resolvedTheme, resolvedThemeName := theme.Resolve(cfg.Theme)
	if resolvedThemeName != "default" {
		t.Fatalf("expected unknown theme to resolve at runtime as %q, got %q", "default", resolvedThemeName)
	}

	if resolvedTheme.Name != "default" {
		t.Fatalf("expected runtime-resolved fallback theme %q, got %q", "default", resolvedTheme.Name)
	}

	savedBytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read saved config: %v", err)
	}

	var saved map[string]any
	if err := json.Unmarshal(savedBytes, &saved); err != nil {
		t.Fatalf("unmarshal saved config: %v", err)
	}

	if saved["theme"] != "aurora" {
		t.Fatalf("expected saved unknown theme to remain %q, got %#v", "aurora", saved["theme"])
	}
}

func TestResolveCurrentConfigPreservesSupportedTheme(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := filepath.Join(home, ".dokbox.json")
	if err := os.WriteFile(configPath, []byte(`{"config_version":1,"docker_host":"unix:///var/run/docker.sock","theme":"slate"}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := config.ResolveCurrentConfig()
	if err != nil {
		t.Fatalf("resolve config: %v", err)
	}

	if cfg == nil {
		t.Fatal("expected resolved config")
	}

	if cfg.Theme != "slate" {
		t.Fatalf("expected supported theme to remain %q, got %q", "slate", cfg.Theme)
	}

	resolvedTheme, resolvedThemeName := theme.Resolve(cfg.Theme)
	if resolvedThemeName != "slate" {
		t.Fatalf("expected supported theme to resolve as %q, got %q", "slate", resolvedThemeName)
	}

	if resolvedTheme.Name != "slate" {
		t.Fatalf("expected resolved theme preset %q, got %q", "slate", resolvedTheme.Name)
	}

	savedBytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read saved config: %v", err)
	}

	var saved map[string]any
	if err := json.Unmarshal(savedBytes, &saved); err != nil {
		t.Fatalf("unmarshal saved config: %v", err)
	}

	if saved["theme"] != "slate" {
		t.Fatalf("expected saved supported theme to remain %q, got %#v", "slate", saved["theme"])
	}
}

func TestLoadOrSetupConfigPromptsBeforeSavingGeneratedDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("DOCKER_HOST", "tcp://docker.example:2375")

	var output strings.Builder
	cfg, err := config.LoadOrSetupConfig(strings.NewReader("n\n"), &output)
	if err != nil {
		t.Fatalf("load or setup config: %v", err)
	}

	if cfg.ConfigVersion != config.CurrentConfigVersion {
		t.Fatalf("expected generated config version %d, got %d", config.CurrentConfigVersion, cfg.ConfigVersion)
	}

	if cfg.DockerHost != "tcp://docker.example:2375" {
		t.Fatalf("expected generated docker host from env, got %q", cfg.DockerHost)
	}

	rendered := output.String()
	if !strings.Contains(rendered, "Using default configuration:") {
		t.Fatalf("expected generated default config output, got %q", rendered)
	}

	if !strings.Contains(rendered, "Do you want to save this configuration? [Y/n]") {
		t.Fatalf("expected save prompt in output, got %q", rendered)
	}

	configPath := filepath.Join(home, ".dokbox.json")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("expected config not to be saved on declined prompt, stat err=%v", err)
	}
}

func TestLoadOrSetupConfigShowsThemeInGeneratedDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("DOCKER_HOST", "tcp://docker.example:2375")

	var output strings.Builder
	cfg, err := config.LoadOrSetupConfig(strings.NewReader("n\n"), &output)
	if err != nil {
		t.Fatalf("load or setup config: %v", err)
	}

	if cfg.Theme != "default" {
		t.Fatalf("expected generated default theme %q, got %q", "default", cfg.Theme)
	}

	if !strings.Contains(output.String(), "\"theme\": \"default\"") {
		t.Fatalf("expected generated default output to include default theme, got %q", output.String())
	}
}

func TestGenerateDefaultConfigUsesPlatformDockerHostWhenEnvUnset(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("DOCKER_HOST", "")

	cfg, err := config.GenerateDefaultConfig(nil)
	if err != nil {
		t.Fatalf("generate default config: %v", err)
	}

	want := "unix:///var/run/docker.sock"
	if runtime.GOOS == "windows" {
		want = "npipe:////./pipe/docker_engine"
	}

	if cfg.DockerHost != want {
		t.Fatalf("expected docker host %q from platform default, got %q", want, cfg.DockerHost)
	}
}

func TestGenerateDefaultConfigIncludesTheme(t *testing.T) {
	cfg, err := config.GenerateDefaultConfig(nil)
	if err != nil {
		t.Fatalf("generate default config: %v", err)
	}

	if cfg.Theme != "default" {
		t.Fatalf("expected generated default theme %q, got %q", "default", cfg.Theme)
	}
}
