package themes

import (
	"testing"
)

func TestGet_EmptyNameReturnsDefault(t *testing.T) {
	got := Get("")
	if got != Default {
		t.Errorf("Get(\"\") = %v, want Default %v", got, Default)
	}
}

func TestGet_UnknownNameReturnsDefault(t *testing.T) {
	got := Get("not-a-real-theme")
	if got != Default {
		t.Errorf("Get(\"not-a-real-theme\") = %v, want Default %v", got, Default)
	}
}

func TestGet_KnownThemes(t *testing.T) {
	for name, want := range catalog {
		got := Get(name)
		if got != want {
			t.Errorf("Get(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestGet_ThemeFieldsNonEmpty(t *testing.T) {
	for name := range catalog {
		th := Get(name)
		if th.Title == "" {
			t.Errorf("theme %q: Title is empty", name)
		}
		if th.Prompt == "" {
			t.Errorf("theme %q: Prompt is empty", name)
		}
		if th.Description == "" {
			t.Errorf("theme %q: Description is empty", name)
		}
		if th.Input == "" {
			t.Errorf("theme %q: Input is empty", name)
		}
		if th.Success == "" {
			t.Errorf("theme %q: Success is empty", name)
		}
		if th.Required == "" {
			t.Errorf("theme %q: Required is empty", name)
		}
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty string is valid", "", true},
		{"known theme is valid", "arcane-ember", true},
		{"all known themes are valid", "crystal-familiar", true},
		{"unknown theme is invalid", "rainbow-wizard", false},
		{"partial name is invalid", "arcane", false},
		{"uppercase is invalid", "Arcane-Ember", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValid(tt.input); got != tt.want {
				t.Errorf("IsValid(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidNames_ContainsAll15Themes(t *testing.T) {
	names := ValidNames()
	if len(names) != 15 {
		t.Errorf("ValidNames() returned %d names, want 15", len(names))
	}
}

func TestValidNames_AllAreValid(t *testing.T) {
	for _, name := range ValidNames() {
		if !IsValid(name) {
			t.Errorf("ValidNames() returned %q which IsValid() rejects", name)
		}
	}
}
