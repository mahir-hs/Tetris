package domain

// RotationDir enumerates the supported rotation directions.
type RotationDir int

const (
	RotateCW  RotationDir = 1
	RotateCCW RotationDir = 3 // +3 mod 4 == -1
	Rotate180 RotationDir = 2
)

// nextState computes the rotation state resulting from a direction.
func nextState(state int, dir RotationDir) int {
	return (state + int(dir)) % 4
}

// kick is an SRS offset {dx, dy} in the (x-right, y-up) convention. When applied
// to board coordinates (y-down) the row offset becomes -dy.
type kick = [2]int

// Standard JLSTZ kick data (x-right, y-up).
var (
	jlstzA = []kick{{0, 0}, {-1, 0}, {-1, 1}, {0, -2}, {-1, -2}}
	jlstzB = []kick{{0, 0}, {1, 0}, {1, -1}, {0, 2}, {1, 2}}
	jlstzC = []kick{{0, 0}, {1, 0}, {1, 1}, {0, -2}, {1, -2}}
	jlstzD = []kick{{0, 0}, {-1, 0}, {-1, -1}, {0, 2}, {-1, 2}}
)

// Standard I-piece kick data (x-right, y-up).
var (
	i0R = []kick{{0, 0}, {-2, 0}, {1, 0}, {-2, -1}, {1, 2}}
	iR0 = []kick{{0, 0}, {2, 0}, {-1, 0}, {2, 1}, {-1, -2}}
	iR2 = []kick{{0, 0}, {-1, 0}, {2, 0}, {-1, 2}, {2, -1}}
	i2R = []kick{{0, 0}, {1, 0}, {-2, 0}, {1, -2}, {-2, 1}}
	i2L = []kick{{0, 0}, {2, 0}, {-1, 0}, {2, 1}, {-1, -2}}
	iL2 = []kick{{0, 0}, {-2, 0}, {1, 0}, {-2, -1}, {1, 2}}
	iL0 = []kick{{0, 0}, {1, 0}, {-2, 0}, {1, -2}, {-2, 1}}
	i0L = []kick{{0, 0}, {-1, 0}, {2, 0}, {-1, 2}, {2, -1}}
)

// flipKicks is a simple symmetric set used for the optional 180° rotation,
// which the guideline does not formally specify.
var flipKicks = []kick{{0, 0}, {0, -1}, {1, 0}, {-1, 0}, {1, -1}, {-1, -1}, {0, 1}, {1, 1}, {-1, 1}}

// jlstzTable and iTable map (from, to) rotation states to kick lists.
var (
	jlstzTable = [4][4][]kick{}
	iTable     = [4][4][]kick{}
)

func init() {
	set := func(t *[4][4][]kick, from, to int, k []kick) { t[from][to] = k }
	set(&jlstzTable, 0, 1, jlstzA)
	set(&jlstzTable, 1, 0, jlstzB)
	set(&jlstzTable, 1, 2, jlstzB)
	set(&jlstzTable, 2, 1, jlstzA)
	set(&jlstzTable, 2, 3, jlstzC)
	set(&jlstzTable, 3, 2, jlstzD)
	set(&jlstzTable, 3, 0, jlstzD)
	set(&jlstzTable, 0, 3, jlstzC)

	set(&iTable, 0, 1, i0R)
	set(&iTable, 1, 0, iR0)
	set(&iTable, 1, 2, iR2)
	set(&iTable, 2, 1, i2R)
	set(&iTable, 2, 3, i2L)
	set(&iTable, 3, 2, iL2)
	set(&iTable, 3, 0, iL0)
	set(&iTable, 0, 3, i0L)
}

// kicksFor returns the kick list for a piece type and rotation transition.
func kicksFor(t PieceType, from, to int) []kick {
	switch t {
	case PieceO:
		return []kick{{0, 0}}
	}
	if (to-from+4)%4 == 2 {
		return flipKicks
	}
	switch t {
	case PieceI:
		if k := iTable[from][to]; k != nil {
			return k
		}
	default:
		if k := jlstzTable[from][to]; k != nil {
			return k
		}
	}
	return []kick{{0, 0}}
}

// RotateResult is the outcome of attempting a rotation.
type RotateResult struct {
	State   int
	X       int
	Y       int
	Changed bool
	Kicked  bool // true when a non-zero offset was required
}

// TryRotate attempts to rotate the piece at (px,py,state) in the given
// direction, applying SRS wall/floor kicks. It returns whether the rotation
// succeeded and the resulting placement.
func TryRotate(b *Board, t Tetromino, state, px, py int, dir RotationDir) RotateResult {
	ns := nextState(state, dir)
	kicks := kicksFor(t.Type, state, ns)
	for _, k := range kicks {
		nx, ny := px+k[0], py-k[1] // y-up -> y-down
		if b.CanPlace(t.Cells(ns, nx, ny)) {
			return RotateResult{State: ns, X: nx, Y: ny, Changed: true, Kicked: k[0] != 0 || k[1] != 0}
		}
	}
	return RotateResult{State: state, X: px, Y: py, Changed: false}
}
