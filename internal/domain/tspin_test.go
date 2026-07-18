package domain

import "testing"

// fill sets a board cell (used to craft T-spin slots).
func fill(b *Board, y, x, c int) {
	if y >= 0 && y < BoardHeight && x >= 0 && x < BoardWidth {
		b.cells[y][x] = c
	}
}

func TestTSpinFull(t *testing.T) {
	b := NewBoard()
	// T pointing down (state 2) centered at board (4,6); front corners = BL,BR.
	p := ActivePiece{Type: PieceT, State: 2, X: 3, Y: 5}
	// Fill both front corners and one side -> 3 corners, both front filled => Full.
	fill(b, p.Y+2, p.X+0, 1) // BL (3,7)
	fill(b, p.Y+2, p.X+2, 1) // BR (5,7)
	fill(b, p.Y+0, p.X+0, 1) // TL (3,5)
	if got := DetectTSpin(b, p, true); got != TSpinFull {
		t.Fatalf("expected TSpinFull, got %v", got)
	}
}

func TestTSpinMini(t *testing.T) {
	b := NewBoard()
	// Same position, but only ONE front corner filled plus back + side => 3 corners, full=false => Mini.
	p := ActivePiece{Type: PieceT, State: 2, X: 3, Y: 5}
	fill(b, p.Y+2, p.X+0, 1) // BL (front)
	fill(b, p.Y+0, p.X+0, 1) // TL (side)
	fill(b, p.Y+0, p.X+2, 1) // TR (side)
	if got := DetectTSpin(b, p, true); got != TSpinMini {
		t.Fatalf("expected TSpinMini, got %v", got)
	}
}

func TestNoTSpinWhenFewerThanThreeCorners(t *testing.T) {
	b := NewBoard()
	p := ActivePiece{Type: PieceT, State: 2, X: 3, Y: 5}
	fill(b, p.Y+2, p.X+0, 1) // only one corner
	if got := DetectTSpin(b, p, true); got != NoTSpin {
		t.Fatalf("expected NoTSpin, got %v", got)
	}
}

func TestNoTSpinWithoutRotation(t *testing.T) {
	b := NewBoard()
	p := ActivePiece{Type: PieceT, State: 2, X: 3, Y: 5}
	fill(b, p.Y+2, p.X+0, 1)
	fill(b, p.Y+2, p.X+2, 1)
	fill(b, p.Y+0, p.X+0, 1)
	if got := DetectTSpin(b, p, false); got != NoTSpin {
		t.Fatalf("expected NoTSpin when last action was not a rotation, got %v", got)
	}
}

func TestNoTSpinForNonT(t *testing.T) {
	b := NewBoard()
	p := ActivePiece{Type: PieceL, State: 2, X: 3, Y: 5}
	fill(b, p.Y+2, p.X+0, 1)
	fill(b, p.Y+2, p.X+2, 1)
	fill(b, p.Y+0, p.X+0, 1)
	if got := DetectTSpin(b, p, true); got != NoTSpin {
		t.Fatalf("expected NoTSpin for non-T piece, got %v", got)
	}
}
