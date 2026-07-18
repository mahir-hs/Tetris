package domain

import "testing"

func TestScoringBaseLineValues(t *testing.T) {
	s := GuidelineScorer{}
	cases := []struct {
		name string
		ctx  ClearContext
		want int
	}{
		{"single", ClearContext{LinesCleared: 1, Level: 1}, 100},
		{"double", ClearContext{LinesCleared: 2, Level: 1}, 300},
		{"triple", ClearContext{LinesCleared: 3, Level: 1}, 500},
		{"tetris", ClearContext{LinesCleared: 4, Level: 1}, 800},
		{"tspin single", ClearContext{LinesCleared: 1, TSpin: TSpinFull, Level: 1}, 800},
		{"tspin double", ClearContext{LinesCleared: 2, TSpin: TSpinFull, Level: 1}, 1200},
		{"tspin triple", ClearContext{LinesCleared: 3, TSpin: TSpinFull, Level: 1}, 1600},
		{"tspin mini single", ClearContext{LinesCleared: 1, TSpin: TSpinMini, Level: 1}, 200},
	}
	for _, c := range cases {
		if got := s.LockBonus(c.ctx); got != c.want {
			t.Errorf("%s: got %d, want %d", c.name, got, c.want)
		}
	}
}

func TestScoringLevelMultiplier(t *testing.T) {
	s := GuidelineScorer{}
	// Tetris at level 3 -> 800 * 3 = 2400.
	if got := s.LockBonus(ClearContext{LinesCleared: 4, Level: 3}); got != 2400 {
		t.Fatalf("level multiplier wrong: got %d, want 2400", got)
	}
}

func TestScoringBackToBack(t *testing.T) {
	s := GuidelineScorer{}
	// Tetris with B2B at level 1 -> 800 * 1.5 = 1200.
	if got := s.LockBonus(ClearContext{LinesCleared: 4, Level: 1, BackToBack: true}); got != 1200 {
		t.Fatalf("B2B tetris wrong: got %d, want 1200", got)
	}
}

func TestScoringCombo(t *testing.T) {
	s := GuidelineScorer{}
	// Single (100) + combo bonus 50*1*1 = 50 -> 150 at combo 1, level 1.
	if got := s.LockBonus(ClearContext{LinesCleared: 1, Level: 1, Combo: 1}); got != 150 {
		t.Fatalf("combo bonus wrong: got %d, want 150", got)
	}
}

func TestScoringDropPoints(t *testing.T) {
	s := GuidelineScorer{}
	if s.SoftDropPoints(3) != 3 {
		t.Fatal("soft drop should be +1/cell")
	}
	if s.HardDropPoints(3) != 6 {
		t.Fatal("hard drop should be +2/cell")
	}
}

func TestScoringPerfectClear(t *testing.T) {
	s := GuidelineScorer{}
	// Tetris perfect clear at level 1 -> base 800 + PC 2000 = 2800.
	got := s.LockBonus(ClearContext{LinesCleared: 4, Level: 1, PerfectClear: true})
	if got != 2800 {
		t.Fatalf("perfect clear tetris wrong: got %d, want 2800", got)
	}
}
