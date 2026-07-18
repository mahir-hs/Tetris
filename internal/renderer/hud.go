package renderer

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mahir/tetris/internal/domain"
)

func formatTime(ms int64) string {
	total := ms / 1000
	m := total / 60
	s := total % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

// HUD renders the side panel with hold, next queue, and live statistics.
func HUD(g *domain.Game, t Theme, layout Layout) string {
	label := lipgloss.NewStyle().Foreground(t.Text).Bold(true)
	value := lipgloss.NewStyle().Foreground(t.Pieces[3])

	section := func(title string, body string) string {
		head := label.Render(title)
		return lipgloss.JoinVertical(lipgloss.Left, head, body)
	}

	// Hold slot.
	var holdBody string
	if pt, ok := g.HeldPiece(); ok {
		holdBody = MiniPiece(pt, t, layout)
	} else {
		holdBody = EmptySlot(t, layout)
	}

	// Next queue (up to 5).
	var nextRows []string
	next := g.NextQueue(5)
	for i, pt := range next {
		nextRows = append(nextRows, MiniPiece(pt, t, layout))
		if i < len(next)-1 {
			nextRows = append(nextRows, "")
		}
	}
	nextBody := lipgloss.JoinVertical(lipgloss.Left, nextRows...)

	// Stats.
	combo := "—"
	if g.Combo() > 0 {
		combo = fmt.Sprintf("%d", g.Combo())
	}
	b2b := "—"
	if g.BackToBack() {
		b2b = "YES"
	}

	stats := lipgloss.JoinVertical(lipgloss.Left,
		fmt.Sprintf("%s %d", label.Render("SCORE  "), g.Score()),
		fmt.Sprintf("%s %d", label.Render("LEVEL  "), g.Level()),
		fmt.Sprintf("%s %d", label.Render("LINES  "), g.Lines()),
		fmt.Sprintf("%s %s", label.Render("COMBO  "), value.Render(combo)),
		fmt.Sprintf("%s %s", label.Render("B2B    "), value.Render(b2b)),
		fmt.Sprintf("%s %s", label.Render("TIME   "), formatTime(g.ElapsedMs())),
	)

	parts := []string{
		section("HOLD", holdBody),
		"",
		section("NEXT", nextBody),
		"",
		stats,
	}
	panel := lipgloss.JoinVertical(lipgloss.Left, parts...)

	// Constrain width with a left border accent.
	panel = lipgloss.NewStyle().Width(layout.CellW*4 + 6).Render(panel)
	return panel
}

// GameOverView renders the end-of-game statistics block.
func GameOverView(g *domain.Game, t Theme) string {
	s := g.Stats()
	lines := []string{
		lipgloss.NewStyle().Foreground(t.Pieces[3]).Bold(true).Render("GAME OVER"),
		"",
		fmt.Sprintf("Score:    %d", g.Score()),
		fmt.Sprintf("Level:    %d", g.Level()),
		fmt.Sprintf("Lines:    %d", g.Lines()),
		fmt.Sprintf("Duration: %s", formatTime(g.ElapsedMs())),
		fmt.Sprintf("Pieces:   %d", s.PiecesPlaced),
		fmt.Sprintf("Tetrises: %d", s.Tetrises),
		fmt.Sprintf("T-Spins:  %d", s.TSpins),
		fmt.Sprintf("Mini:     %d", s.MiniTSpins),
		fmt.Sprintf("Perfect:  %d", s.PerfectClears),
		fmt.Sprintf("Max Combo:%d", s.LongestCombo),
		"",
		"Press R to restart · Q to quit",
	}
	return strings.Join(lines, "\n")
}

// PauseView renders the pause banner.
func PauseView(t Theme) string {
	return lipgloss.NewStyle().Foreground(t.Pieces[2]).Bold(true).Render("PAUSED\n\nPress P to resume")
}
