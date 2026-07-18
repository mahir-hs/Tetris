package renderer

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mahir/tetris/internal/domain"
)

// Layout describes the on-screen size of a single board cell. CellW is the number
// of terminal columns a cell occupies; CellH is the number of terminal rows.
type Layout struct {
	CellW int
	CellH int
}

// StandardLayout is the fixed cell size for a classic, compact Tetris board: 2
// columns wide and 1 row tall per cell (renders as visually square blocks). The
// board is centered and its cells aligned by the renderer; the size does not
// auto-scale.
var StandardLayout = Layout{CellW: 2, CellH: 1}

// painter caches lipgloss styles per cell type to avoid per-cell allocation.
type painter struct {
	filled [8]lipgloss.Style
	ghost  lipgloss.Style
	clear  lipgloss.Style
	border lipgloss.Style
}

func newPainter(t Theme) painter {
	p := painter{}
	for c := 1; c <= 7; c++ {
		p.filled[c] = lipgloss.NewStyle().Foreground(t.Pieces[c]).Background(t.Pieces[c])
	}
	p.ghost = lipgloss.NewStyle().Foreground(t.Ghost).Faint(true)
	p.clear = lipgloss.NewStyle().Foreground(t.Clearing).Background(t.Clearing)
	p.border = lipgloss.NewStyle().Foreground(t.Border)
	return p
}

// Playfield renders the visible 20 rows of the board inside a bordered frame,
// including the active piece, ghost, and line-clear flash. Each board cell is
// drawn CellW columns wide and CellH rows tall so blocks stay aligned and scale
// with the terminal.
func Playfield(g *domain.Game, t Theme, layout Layout) string {
	p := newPainter(t)
	cw := layout.CellW
	ch := layout.CellH
	fill := strings.Repeat("█", cw)
	ghost := strings.Repeat("▒", cw)
	blank := strings.Repeat(" ", cw)

	grid := g.Board().Grid()

	// Base visible grid (rows HiddenRows..BoardHeight-1).
	visible := make([][domain.BoardWidth]int, domain.VisibleHeight)
	for y := 0; y < domain.VisibleHeight; y++ {
		copy(visible[y][:], grid[y+domain.HiddenRows][:])
	}

	ghostSet := make([][domain.BoardWidth]bool, domain.VisibleHeight)
	if gy := g.GhostY(); gy >= 0 && g.HasActive() {
		for _, c := range g.Active().Tetromino().Cells(g.Active().State, g.Active().X, gy) {
			ry := c[0] - domain.HiddenRows
			rx := c[1]
			if ry >= 0 && ry < domain.VisibleHeight && rx >= 0 && rx < domain.BoardWidth {
				ghostSet[ry][rx] = true
			}
		}
	}

	if g.HasActive() && !g.IsClearing() {
		for _, c := range g.Active().Cells() {
			ry := c[0] - domain.HiddenRows
			rx := c[1]
			if ry >= 0 && ry < domain.VisibleHeight && rx >= 0 && rx < domain.BoardWidth {
				visible[ry][rx] = g.Active().Tetromino().Color
			}
		}
	}

	clearing := make(map[int]bool)
	for _, y := range g.ClearRows() {
		clearing[y-domain.HiddenRows] = true
	}

	borderRune := func(r string) string { return p.border.Render(r) }
	var b strings.Builder

	top := borderRune("┌") + strings.Repeat(borderRune("─"), cw*domain.BoardWidth) + borderRune("┐")
	bottom := borderRune("└") + strings.Repeat(borderRune("─"), cw*domain.BoardWidth) + borderRune("┘")
	b.WriteString(top)
	b.WriteByte('\n')

	for y := 0; y < domain.VisibleHeight; y++ {
		for sr := 0; sr < ch; sr++ {
			b.WriteString(borderRune("│"))
			if clearing[y] {
				// Flash the whole cell (all sub-rows) during a clear.
				b.WriteString(p.clear.Render(strings.Repeat(fill, domain.BoardWidth)))
			} else if sr == 0 {
				for x := 0; x < domain.BoardWidth; x++ {
					c := visible[y][x]
					if c > 0 {
						b.WriteString(p.filled[c].Render(fill))
					} else if ghostSet[y][x] {
						b.WriteString(p.ghost.Render(ghost))
					} else {
						b.WriteString(blank)
					}
				}
			} else {
				b.WriteString(strings.Repeat(" ", cw*domain.BoardWidth))
			}
			b.WriteString(borderRune("│"))
			b.WriteByte('\n')
		}
	}
	b.WriteString(bottom)
	return b.String()
}

// miniPiece renders a single piece type as a small bordered preview (used for
// the hold slot and next queue), scaled to the current cell size.
func MiniPiece(pt domain.PieceType, t Theme, layout Layout) string {
	p := newPainter(t)
	tm := domain.GetTetromino(pt)
	cw := layout.CellW
	ch := layout.CellH
	fill := strings.Repeat("█", cw)
	blank := strings.Repeat(" ", cw)

	minR, maxR, minC, maxC := tm.BBox, 0, tm.BBox, 0
	for _, c := range tm.States[0] {
		if c[0] < minR {
			minR = c[0]
		}
		if c[0] > maxR {
			maxR = c[0]
		}
		if c[1] < minC {
			minC = c[1]
		}
		if c[1] > maxC {
			maxC = c[1]
		}
	}

	var rows []string
	for r := minR; r <= maxR; r++ {
		for sr := 0; sr < ch; sr++ {
			var line strings.Builder
			for c := minC; c <= maxC; c++ {
				filled := false
				if sr == 0 {
					for _, cell := range tm.States[0] {
						if cell[0] == r && cell[1] == c {
							filled = true
							break
						}
					}
				}
				if filled {
					line.WriteString(p.filled[tm.Color].Render(fill))
				} else {
					line.WriteString(blank)
				}
			}
			rows = append(rows, line.String())
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// EmptySlot renders a placeholder for an unused hold/next slot, sized to match a
// two-cell-tall piece preview.
func EmptySlot(t Theme, layout Layout) string {
	p := newPainter(t)
	line := p.border.Render(strings.Repeat("┄", layout.CellW))
	var rows []string
	for sr := 0; sr < layout.CellH*2; sr++ {
		rows = append(rows, line)
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
