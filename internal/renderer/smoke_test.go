package renderer

import (
	"math/rand"
	"testing"

	"github.com/mahir/tetris/internal/domain"
)

func TestPlayfieldRenders(t *testing.T) {
	g := domain.NewGame(rand.New(rand.NewSource(5)), domain.GameSettings{StartingLevel: 1, GhostEnabled: true})
	// Drop a couple of pieces to create a non-empty board.
	g.HardDrop()
	g.Tick(domain.ClearAnimMs + 1)
	g.HardDrop()
	g.Tick(domain.ClearAnimMs + 1)

	out := Playfield(g, GetTheme("classic"), Layout{CellW: 3, CellH: 1})
	if out == "" {
		t.Fatal("empty playfield render")
	}
	t.Logf("\n%s", out)
}

func TestMiniPieceRenders(t *testing.T) {
	for _, p := range domain.AllPieceTypes() {
		out := MiniPiece(p, GetTheme("classic"), Layout{CellW: 3, CellH: 1})
		if out == "" {
			t.Fatalf("empty mini render for %v", p)
		}
	}
	t.Logf("T piece preview:\n%s", MiniPiece(domain.PieceT, GetTheme("neon"), Layout{CellW: 3, CellH: 1}))
}

func TestHUDFull(t *testing.T) {
	g := domain.NewGame(rand.New(rand.NewSource(9)), domain.GameSettings{StartingLevel: 2, GhostEnabled: true, HoldEnabled: true})
	g.HardDrop()
	g.Tick(domain.ClearAnimMs + 1)
	out := HUD(g, GetTheme("classic"), Layout{CellW: 3, CellH: 1})
	if out == "" {
		t.Fatal("empty HUD render")
	}
	t.Logf("\n%s", out)
}

func TestMenuRenders(t *testing.T) {
	out := MainMenu([]string{"Play", "High Scores", "Exit"}, 0, GetTheme("retro"))
	if out == "" {
		t.Fatal("empty menu render")
	}
	t.Logf("\n%s", out)
}
