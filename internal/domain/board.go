package domain

// Board dimensions per the Tetris Guideline.
const (
	BoardWidth      = 10
	BoardHeight     = 22 // includes 2 hidden spawn rows at the top
	HiddenRows      = 2
	VisibleHeight   = BoardHeight - HiddenRows // 20
)

// Board is the playfield grid. cells[y][x]: 0 = empty, 1..7 = locked piece color.
// Rows 0..HiddenRows-1 are hidden spawn area; rows HiddenRows..BoardHeight-1 are visible.
type Board struct {
	cells [BoardHeight][BoardWidth]int
}

// NewBoard returns an empty board.
func NewBoard() *Board { return &Board{} }

// Cell returns the value at (x, y). Out-of-bounds horizontally or below the floor
// is reported as a filled wall/floor (used by collision and T-spin corner checks).
func (b *Board) Cell(x, y int) int {
	if x < 0 || x >= BoardWidth || y >= BoardHeight {
		return -1 // treated as occupied (wall / floor)
	}
	if y < 0 {
		return 0 // above the ceiling is open space
	}
	return b.cells[y][x]
}

// IsFilled reports whether a board position is blocked for placement.
func (b *Board) IsFilled(x, y int) bool {
	return b.Cell(x, y) != 0
}

// CanPlace reports whether the given absolute cells fit without collision.
func (b *Board) CanPlace(cells [][2]int) bool {
	for _, c := range cells {
		if b.IsFilled(c[1], c[0]) { // c = {row, col}
			return false
		}
	}
	return true
}

// Lock writes the piece color into the grid at the given cells.
func (b *Board) Lock(cells [][2]int, color int) {
	for _, c := range cells {
		y, x := c[0], c[1]
		if y >= 0 && y < BoardHeight && x >= 0 && x < BoardWidth {
			b.cells[y][x] = color
		}
	}
}

// ClearLines removes every full row and collapses the stack downward.
// It returns the indices (board rows) that were cleared and the count.
func (b *Board) ClearLines() (cleared []int, count int) {
	for y := 0; y < BoardHeight; y++ {
		full := true
		for x := 0; x < BoardWidth; x++ {
			if b.cells[y][x] == 0 {
				full = false
				break
			}
		}
		if full {
			cleared = append(cleared, y)
		}
	}
	if len(cleared) == 0 {
		return nil, 0
	}
	// Build the new grid from the bottom up, skipping cleared rows.
	var next [BoardHeight][BoardWidth]int
	write := BoardHeight - 1
	for y := BoardHeight - 1; y >= 0; y-- {
		isCleared := false
		for _, cy := range cleared {
			if cy == y {
				isCleared = true
				break
			}
		}
		if isCleared {
			continue
		}
		copy(next[write][:], b.cells[y][:])
		write--
	}
	b.cells = next
	return cleared, len(cleared)
}

// IsEmpty reports whether the board has no locked cells.
func (b *Board) IsEmpty() bool {
	for y := 0; y < BoardHeight; y++ {
		for x := 0; x < BoardWidth; x++ {
			if b.cells[y][x] != 0 {
				return false
			}
		}
	}
	return true
}

// Grid returns a copy of the full grid (including hidden rows).
func (b *Board) Grid() [BoardHeight][BoardWidth]int {
	return b.cells
}

// IsPerfectClear reports whether the board is completely empty (used for the
// perfect-clear bonus).
func (b *Board) IsPerfectClear() bool { return b.IsEmpty() }
