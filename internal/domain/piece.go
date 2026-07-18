package domain

// ActivePiece is a piece currently in play: its type, rotation state, and the
// board position of its bounding-box top-left corner.
type ActivePiece struct {
	Type  PieceType
	State int
	X     int
	Y     int
}

// Tetromino returns the piece definition.
func (p ActivePiece) Tetromino() Tetromino { return GetTetromino(p.Type) }

// Cells returns the piece's occupied board cells in its current state.
func (p ActivePiece) Cells() [][2]int { return p.Tetromino().Cells(p.State, p.X, p.Y) }

// Spawn returns a fresh piece at its spawn position.
func (p ActivePiece) Spawn(t PieceType) ActivePiece {
	tm := GetTetromino(t)
	return ActivePiece{Type: t, State: 0, X: tm.SpawnX, Y: tm.SpawnY}
}
