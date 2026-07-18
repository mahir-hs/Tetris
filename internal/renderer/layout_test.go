package renderer

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/mahir/tetris/internal/domain"
)

// TestPlayfieldRectangular verifies every board cell is rendered with an
// identical width, so the frame and pieces stay aligned and the playfield is a
// proper rectangle at any cell size.
func TestPlayfieldRectangular(t *testing.T) {
	g := domain.NewGame(rand.New(rand.NewSource(5)), domain.GameSettings{StartingLevel: 1, GhostEnabled: true})
	g.HardDrop()
	g.Tick(domain.ClearAnimMs + 1)
	g.HardDrop()
	g.Tick(domain.ClearAnimMs + 1)
	g.MoveLeft()
	g.HardDrop()
	g.Tick(domain.ClearAnimMs + 1)

	for _, l := range []Layout{{1, 1}, {2, 2}, {3, 3}} {
		out := Playfield(g, GetTheme("classic"), l)
		lines := strings.Split(out, "\n")
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		wantH := domain.VisibleHeight*l.CellH + 2
		if len(lines) != wantH {
			t.Fatalf("layout %+v: %d lines, want %d", l, len(lines), wantH)
		}
		wantW := l.CellW*domain.BoardWidth + 2
		for i, ln := range lines {
			if w := lipgloss.Width(ln); w != wantW {
				t.Fatalf("layout %+v line %d: width %d, want %d\n%q", l, i, w, wantW, ln)
			}
		}
	}
}

// TestStandardLayout checks the game uses a fixed, standard-sized board cell.
func TestStandardLayout(t *testing.T) {
	if StandardLayout.CellW != 2 || StandardLayout.CellH != 1 {
		t.Fatalf("StandardLayout = %+v, want {2, 1}", StandardLayout)
	}
	out := Playfield(domain.NewGame(rand.New(rand.NewSource(5)), domain.GameSettings{}), GetTheme("classic"), StandardLayout)
	lines := strings.Split(out, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	wantH := domain.VisibleHeight*StandardLayout.CellH + 2
	wantW := StandardLayout.CellW*domain.BoardWidth + 2
	if len(lines) != wantH {
		t.Fatalf("standard playfield: %d lines, want %d", len(lines), wantH)
	}
	for i, ln := range lines {
		if w := lipgloss.Width(ln); w != wantW {
			t.Fatalf("standard playfield line %d: width %d, want %d", i, w, wantW)
		}
	}
}
