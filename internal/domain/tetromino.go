package domain

import "math/rand"

// PieceType identifies one of the seven tetrominoes.
type PieceType int

const (
	PieceI PieceType = iota
	PieceO
	PieceT
	PieceS
	PieceZ
	PieceJ
	PieceL
)

// NumPieceTypes is the number of distinct tetrominoes.
const NumPieceTypes = 7

// AllPieceTypes returns the seven piece types in a stable order.
func AllPieceTypes() []PieceType {
	return []PieceType{PieceI, PieceO, PieceT, PieceS, PieceZ, PieceJ, PieceL}
}

// Tetromino describes a piece kind: its four SRS rotation states, bounding box,
// spawn offset, and color index used inside the board grid.
type Tetromino struct {
	Type   PieceType
	Color  int // 1..7, used as the cell value when locked
	BBox   int // bounding box edge length (3 or 4)
	SpawnX int // board column of the bounding-box top-left at spawn
	SpawnY int // board row of the bounding-box top-left at spawn
	States [4][][2]int
}

// cell is a {row, col} offset within the piece bounding box.
type cell = [2]int

// tetrominoes holds the canonical SRS shapes. Coordinates are {row, col} within
// a BBox×BBox grid; rotation state 0 is the spawn orientation.
var tetrominoes = map[PieceType]Tetromino{
	PieceI: {
		Type: PieceI, Color: 1, BBox: 4, SpawnX: 3, SpawnY: 1,
		States: [4][][2]int{
			{{1, 0}, {1, 1}, {1, 2}, {1, 3}},
			{{0, 2}, {1, 2}, {2, 2}, {3, 2}},
			{{2, 0}, {2, 1}, {2, 2}, {2, 3}},
			{{0, 1}, {1, 1}, {2, 1}, {3, 1}},
		},
	},
	PieceO: {
		Type: PieceO, Color: 2, BBox: 4, SpawnX: 3, SpawnY: 1,
		States: [4][][2]int{
			{{1, 1}, {1, 2}, {2, 1}, {2, 2}},
			{{1, 1}, {1, 2}, {2, 1}, {2, 2}},
			{{1, 1}, {1, 2}, {2, 1}, {2, 2}},
			{{1, 1}, {1, 2}, {2, 1}, {2, 2}},
		},
	},
	PieceT: {
		Type: PieceT, Color: 3, BBox: 3, SpawnX: 3, SpawnY: 2,
		States: [4][][2]int{
			{{0, 1}, {1, 0}, {1, 1}, {1, 2}},
			{{0, 1}, {1, 1}, {1, 2}, {2, 1}},
			{{1, 0}, {1, 1}, {1, 2}, {2, 1}},
			{{0, 1}, {1, 0}, {1, 1}, {2, 1}},
		},
	},
	PieceS: {
		Type: PieceS, Color: 4, BBox: 3, SpawnX: 3, SpawnY: 2,
		States: [4][][2]int{
			{{0, 1}, {0, 2}, {1, 0}, {1, 1}},
			{{0, 1}, {1, 1}, {1, 2}, {2, 2}},
			{{1, 1}, {1, 2}, {2, 0}, {2, 1}},
			{{0, 0}, {1, 0}, {1, 1}, {2, 1}},
		},
	},
	PieceZ: {
		Type: PieceZ, Color: 5, BBox: 3, SpawnX: 3, SpawnY: 2,
		States: [4][][2]int{
			{{0, 0}, {0, 1}, {1, 1}, {1, 2}},
			{{0, 2}, {1, 1}, {1, 2}, {2, 1}},
			{{1, 0}, {1, 1}, {2, 1}, {2, 2}},
			{{0, 1}, {1, 0}, {1, 1}, {2, 0}},
		},
	},
	PieceJ: {
		Type: PieceJ, Color: 6, BBox: 3, SpawnX: 3, SpawnY: 2,
		States: [4][][2]int{
			{{0, 0}, {1, 0}, {1, 1}, {1, 2}},
			{{0, 1}, {1, 1}, {2, 0}, {2, 1}},
			{{1, 0}, {1, 1}, {1, 2}, {2, 2}},
			{{0, 1}, {0, 2}, {1, 1}, {2, 1}},
		},
	},
	PieceL: {
		Type: PieceL, Color: 7, BBox: 3, SpawnX: 3, SpawnY: 2,
		States: [4][][2]int{
			{{0, 2}, {1, 0}, {1, 1}, {1, 2}},
			{{0, 0}, {0, 1}, {1, 1}, {2, 1}},
			{{1, 0}, {1, 1}, {1, 2}, {2, 0}},
			{{0, 1}, {1, 1}, {2, 0}, {2, 1}},
		},
	},
}

// GetTetromino returns the definition for a piece type.
func GetTetromino(t PieceType) Tetromino { return tetrominoes[t] }

// Cells returns the absolute board coordinates of the piece's occupied cells
// given its bounding-box top-left position (px, py) and rotation state.
func (t Tetromino) Cells(state, px, py int) [][2]int {
	out := make([][2]int, 4)
	for i, c := range t.States[state] {
		out[i] = cell{py + c[0], px + c[1]}
	}
	return out
}

// NewRandomPiece draws a piece type from the provided random source's bag.
// (Convenience used by tests; production code uses the Randomizer type.)
func NewRandomPiece(rng *rand.Rand) PieceType {
	return AllPieceTypes()[rng.Intn(NumPieceTypes)]
}
