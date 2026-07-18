package renderer

import "github.com/charmbracelet/lipgloss"

// Theme describes colors and styles for rendering a piece color index (1..7).
type Theme struct {
	Name    string
	Pieces  [8]lipgloss.Color // index 0 unused; 1..7 hold piece colors
	Border  lipgloss.Color
	Text    lipgloss.Color
	BG      lipgloss.Color
	Ghost   lipgloss.Color
	Clearing lipgloss.Color
}

// PieceColor returns the color for a board cell value (1..7).
func (t Theme) PieceColor(c int) lipgloss.Color {
	if c < 1 || c > 7 {
		return t.BG
	}
	return t.Pieces[c]
}

// Themes is the registry of available themes by name.
var Themes = map[string]Theme{
	"classic": {
		Name:     "classic",
		BG:       lipgloss.Color("#101010"),
		Border:   lipgloss.Color("#5a5a5a"),
		Text:     lipgloss.Color("#e0e0e0"),
		Ghost:    lipgloss.Color("#555555"),
		Clearing: lipgloss.Color("#ffffff"),
		Pieces: [8]lipgloss.Color{
			0: "", 1: "#3ad0e8", 2: "#f7d038", 3: "#a85cd8",
			4: "#46c64a", 5: "#e0556a", 6: "#4a78e0", 7: "#e08a3c",
		},
	},
	"neon": {
		Name:     "neon",
		BG:       lipgloss.Color("#050507"),
		Border:   lipgloss.Color("#00ffff"),
		Text:     lipgloss.Color("#ffffff"),
		Ghost:    lipgloss.Color("#335577"),
		Clearing: lipgloss.Color("#ffffff"),
		Pieces: [8]lipgloss.Color{
			0: "", 1: "#00ffff", 2: "#ffff00", 3: "#ff00ff",
			4: "#00ff66", 5: "#ff0040", 6: "#3060ff", 7: "#ff8000",
		},
	},
	"retro": {
		Name:     "retro",
		BG:       lipgloss.Color("#001100"),
		Border:   lipgloss.Color("#00ff66"),
		Text:     lipgloss.Color("#33ff88"),
		Ghost:    lipgloss.Color("#0a5522"),
		Clearing: lipgloss.Color("#aaffcc"),
		Pieces: [8]lipgloss.Color{
			0: "", 1: "#00ff66", 2: "#00cc55", 3: "#00aa44",
			4: "#33ff77", 5: "#00dd55", 6: "#22cc66", 7: "#55ff88",
		},
	},
	"mono": {
		Name:     "mono",
		BG:       lipgloss.Color("#0a0a0a"),
		Border:   lipgloss.Color("#888888"),
		Text:     lipgloss.Color("#cccccc"),
		Ghost:    lipgloss.Color("#444444"),
		Clearing: lipgloss.Color("#ffffff"),
		Pieces: [8]lipgloss.Color{
			0: "", 1: "#ffffff", 2: "#dddddd", 3: "#bbbbbb",
			4: "#999999", 5: "#888888", 6: "#cccccc", 7: "#aaaaaa",
		},
	},
}

// ThemeNames lists the available themes in a stable order.
var ThemeNames = []string{"classic", "neon", "retro", "mono"}

// GetTheme returns a theme by name, defaulting to classic.
func GetTheme(name string) Theme {
	if t, ok := Themes[name]; ok {
		return t
	}
	return Themes["classic"]
}
