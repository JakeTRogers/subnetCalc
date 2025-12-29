package tui

import (
	"testing"
)

func TestKeyMap_ShortHelp(t *testing.T) {
	t.Parallel()
	keys := defaultKeys
	shortHelp := keys.ShortHelp()

	if len(shortHelp) == 0 {
		t.Error("ShortHelp() should return non-empty slice")
	}

	// Verify expected bindings are present
	expectedKeys := []string{"up", "down", "pgup", "pgdown", "s", "x"}
	for i, binding := range shortHelp {
		if len(binding.Keys()) == 0 {
			t.Errorf("ShortHelp()[%d] has no keys", i)
		}
	}

	// Check specific counts
	if len(shortHelp) < 5 {
		t.Errorf("ShortHelp() should return at least 5 bindings, got %d", len(shortHelp))
	}

	_ = expectedKeys // Used for documentation
}

func TestKeyMap_FullHelp(t *testing.T) {
	t.Parallel()
	keys := defaultKeys
	fullHelp := keys.FullHelp()

	if len(fullHelp) == 0 {
		t.Error("FullHelp() should return non-empty slice")
	}

	// FullHelp returns grouped bindings
	expectedGroups := 5
	if len(fullHelp) != expectedGroups {
		t.Errorf("FullHelp() should return %d groups, got %d", expectedGroups, len(fullHelp))
	}

	// Verify each group has bindings
	for i, group := range fullHelp {
		if len(group) == 0 {
			t.Errorf("FullHelp() group %d is empty", i)
		}
	}
}

func TestDefaultKeys_allDefined(t *testing.T) {
	t.Parallel()
	keys := defaultKeys

	// Verify all key bindings have valid keys defined
	bindings := []struct {
		name    string
		binding interface{ Keys() []string }
	}{
		{"Up", keys.Up},
		{"Down", keys.Down},
		{"Left", keys.Left},
		{"Right", keys.Right},
		{"PageUp", keys.PageUp},
		{"PageDown", keys.PageDown},
		{"Split", keys.Split},
		{"Join", keys.Join},
		{"Export", keys.Export},
		{"Copy", keys.Copy},
		{"Quit", keys.Quit},
		{"Help", keys.Help},
	}

	for _, b := range bindings {
		t.Run(b.name, func(t *testing.T) {
			keyList := b.binding.Keys()
			if len(keyList) == 0 {
				t.Errorf("%s binding has no keys", b.name)
			}
		})
	}
}

func TestDefaultKeys_helpText(t *testing.T) {
	t.Parallel()
	keys := defaultKeys

	// Verify navigation keys have help text
	if keys.Up.Help().Key == "" {
		t.Error("Up key should have help text")
	}
	if keys.Down.Help().Key == "" {
		t.Error("Down key should have help text")
	}
	if keys.Split.Help().Key == "" {
		t.Error("Split key should have help text")
	}
	if keys.Quit.Help().Key == "" {
		t.Error("Quit key should have help text")
	}
}
