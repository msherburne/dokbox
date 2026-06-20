package theme

import "testing"

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

func TestPresetNamesReturnsBuiltInPresetsInOrder(t *testing.T) {
	got := PresetNames()
	want := []string{"default", "slate", "ember"}

	if len(got) != len(want) {
		t.Fatalf("expected %d preset names, got %d (%v)", len(want), len(got), got)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected preset names %v, got %v", want, got)
		}
	}
}
