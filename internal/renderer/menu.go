package renderer

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mahir/tetris/internal/config"
	"github.com/mahir/tetris/internal/database"
)

const titleArt = `
 ████████╗███████╗████████╗██████╗ ██╗███████╗
 ╚══██╔══╝██╔════╝╚══██╔══╝██╔══██╗██║██╔════╝
    ██║   ███████╗   ██║   ██████╔╝██║███████╗
    ██║   ╚════██║   ██║   ██╔══██╗██║╚════██║
    ██║   ███████║   ██║   ██║  ██║██║███████║
    ╚═╝   ╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝╚══════╝
`

// MainMenu renders the title and option list.
func MainMenu(options []string, selected int, t Theme) string {
	title := lipgloss.NewStyle().Foreground(t.Pieces[3]).Render(titleArt)
	lines := []string{title, ""}
	for i, o := range options {
		if i == selected {
			lines = append(lines, lipgloss.NewStyle().Foreground(t.Pieces[2]).Bold(true).Render("▶ "+o))
		} else {
			lines = append(lines, lipgloss.NewStyle().Foreground(t.Text).Render("  "+o))
		}
	}
	lines = append(lines, "", lipgloss.NewStyle().Faint(true).Render("↑/↓ navigate · Enter select · Esc back"))
	return strings.Join(lines, "\n")
}

// HighScores renders the leaderboard table (top 20).
func HighScores(rows []database.HighScoreRow, t Theme) string {
	head := lipgloss.NewStyle().Foreground(t.Pieces[3]).Bold(true)
	var b strings.Builder
	b.WriteString(head.Render("HIGH SCORES\n\n"))
	header := fmt.Sprintf("%-4s %-12s %8s %6s %6s %12s", "Rank", "Player", "Score", "Level", "Lines", "Date")
	b.WriteString(head.Render(header))
	b.WriteString("\n")
	for _, r := range rows {
		date := r.Date.Format("2006-01-02")
		line := fmt.Sprintf("%-4d %-12s %8d %6d %6d %12s", r.Rank, r.Username, r.Score, r.Level, r.Lines, date)
		b.WriteString(lipgloss.NewStyle().Foreground(t.Text).Render(line))
		b.WriteString("\n")
	}
	if len(rows) == 0 {
		b.WriteString(lipgloss.NewStyle().Faint(true).Render("(no scores yet)\n"))
	}
	b.WriteString("\n" + lipgloss.NewStyle().Faint(true).Render("Esc to return"))
	return b.String()
}

// Statistics renders the player statistics screen.
func Statistics(username string, s database.Statistics, t Theme) string {
	label := lipgloss.NewStyle().Foreground(t.Pieces[3]).Bold(true)
	val := lipgloss.NewStyle().Foreground(t.Text)
	row := func(k string, v interface{}) string {
		return fmt.Sprintf("%-18s %v", label.Render(k), val.Render(fmt.Sprintf("%v", v)))
	}
	lines := []string{
		lipgloss.NewStyle().Foreground(t.Pieces[3]).Bold(true).Render("STATISTICS — " + username),
		"",
		row("Games Played", s.GamesPlayed),
		row("Highest Score", s.HighestScore),
		row("Average Score", fmt.Sprintf("%.1f", s.AverageScore)),
		row("Highest Level", s.HighestLevel),
		row("Total Lines", s.TotalLines),
		row("Total Play Time", formatDuration(s.TotalPlayTime)),
		row("Average Duration", formatDuration(int(s.AverageDuration))),
		row("Total Tetrises", s.TotalTetrises),
		row("Total T-Spins", s.TotalTSpins),
		row("Perfect Clears", s.TotalPerfect),
		row("Longest Combo", s.LongestCombo),
		"",
		lipgloss.NewStyle().Faint(true).Render("Esc to return"),
	}
	return strings.Join(lines, "\n")
}

func formatDuration(sec int) string {
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// SettingsView renders the current settings and hints.
func SettingsView(s config.Settings, t Theme) string {
	label := lipgloss.NewStyle().Foreground(t.Pieces[3]).Bold(true)
	val := lipgloss.NewStyle().Foreground(t.Text)
	row := func(k string, v interface{}) string {
		return fmt.Sprintf("%-16s %v", label.Render(k), val.Render(fmt.Sprintf("%v", v)))
	}
	lines := []string{
		lipgloss.NewStyle().Foreground(t.Pieces[3]).Bold(true).Render("SETTINGS"),
		"",
		row("Starting Level", s.StartingLevel),
		row("Ghost Piece", s.GhostEnabled),
		row("Hold System", s.HoldEnabled),
		row("180° Rotation", s.Rotate180Enabled),
		row("Theme", s.Theme),
		row("FPS Limit", s.FPSLimit),
		"",
		lipgloss.NewStyle().Faint(true).Render("1-4: cycle theme · g: toggle ghost · h: toggle hold · r: toggle 180°"),
		lipgloss.NewStyle().Faint(true).Render("Esc back · key bindings are edited in settings.json"),
	}
	return strings.Join(lines, "\n")
}

// Credits renders the credits screen.
func Credits(t Theme) string {
	lines := []string{
		lipgloss.NewStyle().Foreground(t.Pieces[3]).Bold(true).Render("CREDITS"),
		"",
		lipgloss.NewStyle().Foreground(t.Text).Render("CLI Tetris — a modern Tetris Guideline implementation"),
		"",
		"Engine:    Go + Bubble Tea",
		"Storage:   SQLite (pure-Go)",
		"Mechanics: 7-Bag, SRS, T-Spin, B2B, Perfect Clear",
		"",
		"Built with the charmbracelet ecosystem.",
		"",
		lipgloss.NewStyle().Faint(true).Render("Esc to return"),
	}
	return strings.Join(lines, "\n")
}

// UsernamePrompt renders the name-entry screen.
func UsernamePrompt(input, prompt string, t Theme) string {
	title := lipgloss.NewStyle().Foreground(t.Pieces[3]).Bold(true).Render(prompt)
	field := lipgloss.NewStyle().Foreground(t.Text).Render("> " + input + "_")
	return strings.Join([]string{title, "", field, "", lipgloss.NewStyle().Faint(true).Render("Enter to confirm · Esc cancel")}, "\n")
}
