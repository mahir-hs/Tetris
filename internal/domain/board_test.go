package domain

import "testing"

func TestBoardEmptyAndCollision(t *testing.T) {
	b := NewBoard()
	if !b.IsEmpty() {
		t.Fatal("new board should be empty")
	}
	if !b.CanPlace([][2]int{{5, 3}, {5, 4}}) {
		t.Fatal("empty board should accept placement")
	}
	// Floor and walls are occupied.
	if b.CanPlace([][2]int{{BoardHeight, 3}}) {
		t.Fatal("below floor must be blocked")
	}
	if !b.CanPlace([][2]int{{-1, 3}}) {
		t.Fatal("above ceiling (y<0) must be open")
	}
	if b.CanPlace([][2]int{{5, -1}}) {
		t.Fatal("left wall must be blocked")
	}
	if b.CanPlace([][2]int{{5, BoardWidth}}) {
		t.Fatal("right wall must be blocked")
	}
}

func TestBoardLockAndClearSingle(t *testing.T) {
	b := NewBoard()
	// Fill the bottom row completely.
	for x := 0; x < BoardWidth; x++ {
		b.Lock([][2]int{{BoardHeight - 1, x}}, 1)
	}
	cleared, n := b.ClearLines()
	if n != 1 {
		t.Fatalf("expected 1 cleared line, got %d", n)
	}
	if cleared[0] != BoardHeight-1 {
		t.Fatalf("expected bottom row cleared, got %d", cleared[0])
	}
	if !b.IsEmpty() {
		t.Fatal("board should be empty after clearing the only filled row")
	}
}

func TestBoardClearDoesNotTriggerWithGap(t *testing.T) {
	b := NewBoard()
	for x := 0; x < BoardWidth-1; x++ {
		b.Lock([][2]int{{BoardHeight - 1, x}}, 1)
	}
	if _, n := b.ClearLines(); n != 0 {
		t.Fatalf("row with a gap must not clear, got %d", n)
	}
}

func TestBoardClearMultipleAndCollapse(t *testing.T) {
	b := NewBoard()
	// Fill bottom two rows.
	for y := BoardHeight - 2; y < BoardHeight; y++ {
		for x := 0; x < BoardWidth; x++ {
			b.Lock([][2]int{{y, x}}, 1)
		}
	}
	// Put a marker on the row above so we can verify it collapses.
	b.Lock([][2]int{{BoardHeight - 3, 0}}, 2)

	cleared, n := b.ClearLines()
	if n != 2 {
		t.Fatalf("expected 2 cleared lines, got %d", n)
	}
	if len(cleared) != 2 {
		t.Fatalf("expected 2 cleared row indices, got %d", len(cleared))
	}
	// The marker should now sit at the bottom row after collapse.
	if b.Cell(0, BoardHeight-1) != 2 {
		t.Fatalf("marker did not collapse to bottom; got %d", b.Cell(0, BoardHeight-1))
	}
}
