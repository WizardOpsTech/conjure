package themes

import (
	"fmt"
	"math"
	"strconv"
	"strings"
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

// verifies every hex foreground color in the catalog achieves
// a minimum 4.5:1 contrast ratio against TargetBackground (WCAG AA for normal text).
// ANSI 256 colors in the Default theme are excluded because their exact rendering
// is terminal-defined and cannot be validated at build time.
func TestWCAGContrast(t *testing.T) {
	type field struct {
		name  string
		value string
	}
	check := func(themeName string, fields []field) {
		for _, f := range fields {
			if f.value == "" || f.value[0] != '#' {
				continue // skip ANSI codes
			}
			ratio, err := wcagContrast(f.value, TargetBackground)
			if err != nil {
				t.Errorf("theme %q field %s: invalid color %q: %v", themeName, f.name, f.value, err)
				continue
			}
			if ratio < 4.5 {
				t.Errorf("theme %q field %s: color %q has contrast %.2f:1 against %s (need 4.5:1)",
					themeName, f.name, f.value, ratio, TargetBackground)
			}
		}
	}

	for name, th := range catalog {
		check(name, []field{
			{"Title", th.Title},
			{"Prompt", th.Prompt},
			{"Description", th.Description},
			{"Input", th.Input},
			{"Success", th.Success},
			{"Required", th.Required},
		})
	}
}

func wcagContrast(fg, bg string) (float64, error) {
	lr, err := hexLuminance(fg)
	if err != nil {
		return 0, fmt.Errorf("fg: %w", err)
	}
	lb, err := hexLuminance(bg)
	if err != nil {
		return 0, fmt.Errorf("bg: %w", err)
	}
	if lr < lb {
		lr, lb = lb, lr
	}
	return (lr + 0.05) / (lb + 0.05), nil
}

func hexLuminance(hex string) (float64, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, fmt.Errorf("expected 6-char hex, got %q", hex)
	}
	r, err1 := strconv.ParseInt(hex[0:2], 16, 32)
	g, err2 := strconv.ParseInt(hex[2:4], 16, 32)
	b, err3 := strconv.ParseInt(hex[4:6], 16, 32)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, fmt.Errorf("invalid hex color %q", hex)
	}
	lin := func(c int64) float64 {
		v := float64(c) / 255.0
		if v <= 0.04045 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b), nil
}
