package domain

// TSpinType classifies a T-Spin result.
type TSpinType int

const (
	NoTSpin   TSpinType = iota
	TSpinMini            // mini T-spin
	TSpinFull            // full T-spin
)

// DetectTSpin applies the 3-corner rule to determine whether the T piece
// last placed via a rotation is a (mini or full) T-Spin. Only the T piece can
// T-spin, and only when the immediately preceding action was a rotation.
func DetectTSpin(b *Board, p ActivePiece, lastWasRotation bool) TSpinType {
	if p.Type != PieceT || !lastWasRotation {
		return NoTSpin
	}
	// Center is the middle of the 3x3 box; corners are its diagonal neighbours.
	cx, cy := p.X+1, p.Y+1
	tl := b.Cell(cx-1, cy-1) != 0
	tr := b.Cell(cx+1, cy-1) != 0
	bl := b.Cell(cx-1, cy+1) != 0
	br := b.Cell(cx+1, cy+1) != 0

	occupied := 0
	if tl {
		occupied++
	}
	if tr {
		occupied++
	}
	if bl {
		occupied++
	}
	if br {
		occupied++
	}
	if occupied < 3 {
		return NoTSpin
	}

	// Front corners depend on the pointing direction (rotation state).
	var frontA, frontB bool
	switch p.State {
	case 0: // pointing up
		frontA, frontB = tl, tr
	case 1: // pointing right
		frontA, frontB = tr, br
	case 2: // pointing down
		frontA, frontB = bl, br
	case 3: // pointing left
		frontA, frontB = tl, bl
	}
	if frontA && frontB {
		return TSpinFull
	}
	return TSpinMini
}
