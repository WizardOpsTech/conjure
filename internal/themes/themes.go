package themes

// TargetBackground is the canonical dark background all themes are designed and
// validated against. Every foreground color in this file must achieve a minimum
// WCAG AA contrast ratio of 4.5:1 against this value.
const TargetBackground = "#121212"

// lipgloss-compatible color values (hex or ANSI 256) for each UI element.
type Theme struct {
	Title       string
	Prompt      string
	Description string
	Input       string
	Success     string
	Required    string
}

var Default = Theme{
	Title:       "99",
	Prompt:      "86",
	Description: "#9E9E9E", // was "241" (~#626262, 3.07:1 — failed WCAG AA)
	Input:       "219",
	Success:     "42",
	Required:    "196",
}

var catalog = map[string]Theme{
	"arcane-ember": {
		Title:       "#E05C2A",
		Prompt:      "#E5603A", // was #C84B31 (4.02:1 — failed WCAG AA)
		Description: "#9E9E9E", // was #7A7A7A (4.36:1 — failed WCAG AA)
		Input:       "#D4A017",
		Success:     "#E05C2A",
		Required:    "#FF3300",
	},
	"moonlit-mana": {
		Title:       "#B39DDB",
		Prompt:      "#81D4FA",
		Description: "#90A4AE",
		Input:       "#B2EBF2",
		Success:     "#81D4FA",
		Required:    "#EF9A9A",
	},
	"runestone-grove": {
		Title:       "#7CB342",
		Prompt:      "#A5845A",
		Description: "#9E9E9E",
		Input:       "#4DB6AC",
		Success:     "#7CB342",
		Required:    "#EF5350",
	},
	"spellforge": {
		Title:       "#FF7043",
		Prompt:      "#CD5C5C",
		Description: "#78909C",
		Input:       "#CD7F32",
		Success:     "#FF7043",
		Required:    "#FF4444", // was #DC143C (3.75:1 — failed WCAG AA)
	},
	"celestial-grimoire": {
		Title:       "#CE4DE4", // was #9C27B0 (2.97:1 — failed WCAG AA)
		Prompt:      "#F1C40F",
		Description: "#C8B89A",
		Input:       "#4FC3F7",
		Success:     "#CE4DE4", // was #9C27B0 (2.97:1 — failed WCAG AA)
		Required:    "#E74C3C",
	},
	"mystic-marsh": {
		Title:       "#558B2F",
		Prompt:      "#4ECDC4",
		Description: "#9E9E9E",
		Input:       "#B2DFDB",
		Success:     "#558B2F",
		Required:    "#EF5350",
	},
	"dragon-hoard": {
		Title:       "#2ECC71",
		Prompt:      "#E74C3C",
		Description: "#808080",
		Input:       "#F39C12",
		Success:     "#2ECC71",
		Required:    "#E74C3C",
	},
	"enchanted-aurora": {
		Title:       "#00BCD4",
		Prompt:      "#FF4D8C", // was #E91E63 (4.31:1 — failed WCAG AA)
		Description: "#7986CB",
		Input:       "#00E676",
		Success:     "#00BCD4",
		Required:    "#FF4081",
	},
	"hexfire": {
		Title:       "#CE4DE4", // was #9C27B0 (2.97:1 — failed WCAG AA)
		Prompt:      "#F06292",
		Description: "#9E9E9E", // was #757575 (4.07:1 — failed WCAG AA)
		Input:       "#FF8C00",
		Success:     "#CE4DE4", // was #9C27B0 (2.97:1 — failed WCAG AA)
		Required:    "#FF1744",
	},
	"potionmaker": {
		Title:       "#00BFA5",
		Prompt:      "#8BC34A",
		Description: "#9E9E9E",
		Input:       "#C868D8", // was #AB47BC (3.89:1 — failed WCAG AA)
		Success:     "#00BFA5",
		Required:    "#FF5252",
	},
	"feywild-bloom": {
		Title:       "#F48FB1",
		Prompt:      "#CE93D8",
		Description: "#A5D6A7",
		Input:       "#FFCCBC",
		Success:     "#F48FB1",
		Required:    "#EF9A9A",
	},
	"storm-sorcerer": {
		Title:       "#2979FF",
		Prompt:      "#78909C",
		Description: "#82A5B4", // was #607D8B (4.29:1 — failed WCAG AA)
		Input:       "#E3F2FD",
		Success:     "#2979FF",
		Required:    "#EF5350",
	},
	"necromancers-ledger": {
		Title:       "#80CBC4",
		Prompt:      "#66BB6A",
		Description: "#9E9E9E",
		Input:       "#E0E0E0",
		Success:     "#66BB6A",
		Required:    "#EF9A9A",
	},
	"sunspell-sanctum": {
		Title:       "#FFB74D",
		Prompt:      "#F57C00",
		Description: "#A1887F",
		Input:       "#FFD54F",
		Success:     "#FFB74D",
		Required:    "#EF5350",
	},
	"crystal-familiar": {
		Title:       "#81D4FA",
		Prompt:      "#CE93D8",
		Description: "#B0BEC5",
		Input:       "#E1F5FE",
		Success:     "#81D4FA",
		Required:    "#F48FB1",
	},
}

func Get(name string) Theme {
	if name == "" {
		return Default
	}
	if t, ok := catalog[name]; ok {
		return t
	}
	return Default
}

func ValidNames() []string {
	names := make([]string, 0, len(catalog))
	for k := range catalog {
		names = append(names, k)
	}
	return names
}

func IsValid(name string) bool {
	if name == "" {
		return true
	}
	_, ok := catalog[name]
	return ok
}
