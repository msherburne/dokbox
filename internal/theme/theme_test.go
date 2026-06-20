package theme

import "testing"

func TestResolveThemeFallsBackToDefault(t *testing.T) {
	resolved, name := Resolve("aurora")

	if name != DefaultPreset {
		t.Fatalf("expected resolved name %q, got %q", DefaultPreset, name)
	}

	if resolved.Name != DefaultPreset {
		t.Fatalf("expected default preset theme, got %q", resolved.Name)
	}
}

func TestResolveReturnsNamedPreset(t *testing.T) {
	resolved, name := Resolve("ember")

	if name != "ember" {
		t.Fatalf("expected resolved name %q, got %q", "ember", name)
	}

	if resolved.Name != "ember" {
		t.Fatalf("expected ember theme, got %q", resolved.Name)
	}
}

func TestResolveFallsBackToDefault(t *testing.T) {
	resolved, name := Resolve("unknown")

	if name != DefaultPreset {
		t.Fatalf("expected fallback name %q, got %q", DefaultPreset, name)
	}

	if resolved.Name != DefaultPreset {
		t.Fatalf("expected fallback theme %q, got %q", DefaultPreset, resolved.Name)
	}
}
