package domain

import "testing"

func TestNextState(t *testing.T) {
	cases := []struct {
		state int
		dir   RotationDir
		want  int
	}{
		{0, RotateCW, 1},
		{3, RotateCW, 0},
		{1, RotateCCW, 0},
		{0, RotateCCW, 3},
		{0, Rotate180, 2},
		{1, Rotate180, 3},
	}
	for _, c := range cases {
		if got := nextState(c.state, c.dir); got != c.want {
			t.Errorf("nextState(%d,%d)=%d, want %d", c.state, c.dir, got, c.want)
		}
	}
}

func TestKickTablesStartWithZero(t *testing.T) {
	for from := 0; from < 4; from++ {
		for to := 0; to < 4; to++ {
			if k := jlstzTable[from][to]; k != nil && (k[0][0] != 0 || k[0][1] != 0) {
				t.Errorf("JLSTZ kick[%d][%d] must start with (0,0)", from, to)
			}
			if k := iTable[from][to]; k != nil && (k[0][0] != 0 || k[0][1] != 0) {
				t.Errorf("I kick[%d][%d] must start with (0,0)", from, to)
			}
		}
	}
}

func TestRotateOnEmptyBoardSucceeds(t *testing.T) {
	b := NewBoard()
	tm := GetTetromino(PieceT)
	p := ActivePiece{}.Spawn(PieceT)
	// CW and CCW both succeed on an empty board (first kick (0,0) fits).
	r := TryRotate(b, tm, p.State, p.X, p.Y, RotateCW)
	if !r.Changed || r.State != 1 {
		t.Fatalf("CW rotation failed on empty board: %+v", r)
	}
	r2 := TryRotate(b, tm, r.State, r.X, r.Y, RotateCCW)
	if !r2.Changed || r2.State != 0 {
		t.Fatalf("CCW rotation failed on empty board: %+v", r2)
	}
}

func TestORotationDoesNotMove(t *testing.T) {
	b := NewBoard()
	tm := GetTetromino(PieceO)
	p := ActivePiece{}.Spawn(PieceO)
	r := TryRotate(b, tm, p.State, p.X, p.Y, RotateCW)
	if !r.Changed {
		t.Fatal("O rotation should always succeed")
	}
	if r.X != p.X || r.Y != p.Y {
		t.Fatalf("O piece must not translate on rotation: got (%d,%d) want (%d,%d)", r.X, r.Y, p.X, p.Y)
	}
}

func TestRotate180DisabledWhenUnsupported(t *testing.T) {
	b := NewBoard()
	tm := GetTetromino(PieceT)
	p := ActivePiece{}.Spawn(PieceT)
	r := TryRotate(b, tm, p.State, p.X, p.Y, Rotate180)
	// 180 rotation tables exist, so it changes state to 2.
	if !r.Changed || r.State != 2 {
		t.Fatalf("180 rotation unexpected: %+v", r)
	}
}

func TestRotationBlockedWhenSurrounded(t *testing.T) {
	b := NewBoard()
	tm := GetTetromino(PieceT)
	p := ActivePiece{}.Spawn(PieceT)
	// Surround the T's rotation footprint with locked cells everywhere feasible.
	cells := p.Cells()
	for _, c := range cells {
		// Fill the rows just below and around to block kicks.
		for dx := -2; dx <= 2; dx++ {
			for dy := -2; dy <= 2; dy++ {
				y, x := c[0]+dy, c[1]+dx
				if y >= 0 && y < BoardHeight && x >= 0 && x < BoardWidth {
					b.cells[y][x] = 9
				}
			}
		}
	}
	// Clear the exact piece cells so the piece itself is valid but kicks are blocked.
	for _, c := range cells {
		b.cells[c[0]][c[1]] = 0
	}
	r := TryRotate(b, tm, p.State, p.X, p.Y, RotateCW)
	if r.Changed {
		t.Fatalf("rotation should fail when fully blocked, got %+v", r)
	}
}
