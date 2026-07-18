package renderer

import (
	"github.com/charmbracelet/lipgloss"
)

// Overlay centers the given banner text over the playfield, preserving the
// underlying board rendering.
func Overlay(base, banner string, t Theme) string {
	w := lipgloss.Width(base)
	h := lipgloss.Height(base)
	styled := lipgloss.NewStyle().
		Foreground(t.Text).
		Background(t.BG).
		Render(banner)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, styled)
}
