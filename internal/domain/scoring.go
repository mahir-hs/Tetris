package domain

// ScoringStrategy abstracts the scoring rules so alternative schemes can be
// swapped in (Strategy pattern, per the specification).
type ScoringStrategy interface {
	// LockBonus returns points awarded for locking a piece given the clear
	// context (line clears, T-spin, combo, back-to-back, perfect clear).
	LockBonus(ctx ClearContext) int
	// SoftDropPoints returns points for soft-dropping n cells.
	SoftDropPoints(cells int) int
	// HardDropPoints returns points for hard-dropping n cells.
	HardDropPoints(cells int) int
}

// ClearContext captures everything the scorer needs to value a lock.
type ClearContext struct {
	LinesCleared int
	TSpin        TSpinType
	Level        int
	Combo        int // chain count; 0 for the first clear in a chain
	BackToBack   bool
	PerfectClear bool
}

// GuidelineScorer implements the official Tetris Guideline scoring table.
type GuidelineScorer struct{}

func (GuidelineScorer) SoftDropPoints(cells int) int { return cells }      // +1 per cell
func (GuidelineScorer) HardDropPoints(cells int) int { return cells * 2 }  // +2 per cell

func (GuidelineScorer) LockBonus(ctx ClearContext) int {
	base := lineBaseValue(ctx)
	score := 0
	if base > 0 {
		if ctx.BackToBack && isDifficult(ctx) {
			base = int(float64(base) * 1.5)
		}
		score += base * ctx.Level
	}
	if ctx.Combo > 0 {
		score += 50 * ctx.Combo * ctx.Level
	}
	if ctx.PerfectClear {
		score += perfectClearBonus(ctx) * ctx.Level
	}
	return score
}

func isDifficult(ctx ClearContext) bool {
	return ctx.LinesCleared == 4 || ctx.TSpin != NoTSpin
}

func lineBaseValue(ctx ClearContext) int {
	switch ctx.TSpin {
	case TSpinFull:
		switch ctx.LinesCleared {
		case 0:
			return 400
		case 1:
			return 800
		case 2:
			return 1200
		case 3:
			return 1600
		}
	case TSpinMini:
		switch ctx.LinesCleared {
		case 0:
			return 100
		case 1:
			return 200
		case 2:
			return 400
		}
	default:
		switch ctx.LinesCleared {
		case 1:
			return 100
		case 2:
			return 300
		case 3:
			return 500
		case 4:
			return 800
		}
	}
	return 0
}

func perfectClearBonus(ctx ClearContext) int {
	switch {
	case ctx.BackToBack && ctx.LinesCleared == 4:
		return 3200
	case ctx.LinesCleared == 1:
		return 800
	case ctx.LinesCleared == 2:
		return 1200
	case ctx.LinesCleared == 3:
		return 1800
	case ctx.LinesCleared == 4:
		return 2000
	}
	return 0
}
