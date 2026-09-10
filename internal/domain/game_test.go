package domain

import (
	"math/rand"
	"testing"
)

func newTestGame(t *testing.T) *Game {
	t.Helper()
	return NewGame(rand.New(rand.NewSource(1)), GameSettings{
		GhostEnabled: true, HoldEnabled: true, Rotate180: true, StartingLevel: 1,
	})
}

func TestGameStartsPlaying(t *testing.T) {
	g := newTestGame(t)
	if g.State() != StatePlaying {
		t.Fatalf("expected Playing, got %v", g.State())
	}
	if !g.HasActive() {
		t.Fatal("expected an active piece")
	}
	if _, ok := g.HeldPiece(); ok {
		t.Fatal("hold should start empty")
	}
	if g.Level() != 1 {
		t.Fatalf("expected level 1, got %d", g.Level())
	}
}

func TestHardDropLocksAndScores(t *testing.T) {
	g := newTestGame(t)
	g.board = NewBoard()
	g.spawnSpecific(PieceO)
	// O rests at Y=19 (lowest cell row 21); spawn Y=1 -> 18 cells dropped.
	g.HardDrop()
	if g.Stats().PiecesPlaced != 1 {
		t.Fatalf("expected 1 piece placed, got %d", g.Stats().PiecesPlaced)
	}
	if g.Score() != 36 { // 2 points/cell * 18 cells
		t.Fatalf("expected score 36, got %d", g.Score())
	}
	if !g.HasActive() {
		t.Fatal("a new piece should have spawned after lock")
	}
}

func TestLockDelayLocksGroundedPiece(t *testing.T) {
	g := newTestGame(t)
	g.board = NewBoard()
	g.spawnSpecific(PieceO)
	g.active.Y = 19 // resting at the floor
	g.Tick(10)      // begins lock delay (lockTimer -> 490)
	if g.Stats().PiecesPlaced != 0 {
		t.Fatal("piece should not lock yet")
	}
	g.Tick(500) // exceeds remaining lock delay
	if g.Stats().PiecesPlaced != 1 {
		t.Fatalf("grounded piece should lock after delay, placed=%d", g.Stats().PiecesPlaced)
	}
}

func TestAirbornePieceDoesNotLock(t *testing.T) {
	g := newTestGame(t)
	g.board = NewBoard()
	g.spawnSpecific(PieceO)
	g.active.Y = 5
	g.Tick(1000) // one gravity step, still airborne
	if g.Stats().PiecesPlaced != 0 {
		t.Fatal("airborne piece must not lock")
	}
}

func TestHoldUsedOnlyOnce(t *testing.T) {
	g := newTestGame(t)
	first := g.Active().Type
	g.Hold()
	held, ok := g.HeldPiece()
	if !ok || held != first {
		t.Fatalf("hold should store %v, got %v (ok=%v)", first, held, ok)
	}
	after := g.Active().Type
	g.Hold() // should be ignored (holdUsed)
	if g.Active().Type != after {
		t.Fatalf("second hold must be a no-op, changed to %v", g.Active().Type)
	}
}

func TestHoldResetsAfterLock(t *testing.T) {
	g := newTestGame(t)
	g.Hold()
	if !g.holdUsed {
		t.Fatal("holdUsed should be set")
	}
	g.HardDrop()
	g.Tick(ClearAnimMs + 1)
	if g.holdUsed {
		t.Fatal("holdUsed should reset after a lock")
	}
}

func TestGhostYAtOrBelowPiece(t *testing.T) {
	g := newTestGame(t)
	g.board = NewBoard()
	gy := g.GhostY()
	if gy < g.Active().Y {
		t.Fatalf("ghost (%d) must be at or below active Y (%d)", gy, g.Active().Y)
	}
}

func TestLineClearAndLevelUp(t *testing.T) {
	g := newTestGame(t)
	for i := 0; i < 10; i++ {
		g.board = NewBoard()
		// Fill the bottom row except column 4.
		for x := 0; x < BoardWidth; x++ {
			if x != 4 {
				g.board.Lock([][2]int{{BoardHeight - 1, x}}, 1)
			}
		}
		g.spawnSpecific(PieceI)
		g.active.State = 1 // vertical
		g.active.X = 2     // occupies column 4
		g.active.Y = 18
		g.HardDrop()
		g.Tick(ClearAnimMs + 1) // finish the line-clear animation and spawn next
	}
	if g.Lines() != 10 {
		t.Fatalf("expected 10 lines cleared, got %d", g.Lines())
	}
	if g.Level() != 2 {
		t.Fatalf("expected level 2 after 10 lines, got %d", g.Level())
	}
}

func TestStartingLevelAdvancesFromConfiguredLevel(t *testing.T) {
	g := NewGame(rand.New(rand.NewSource(1)), GameSettings{StartingLevel: 10})
	for i := 0; i < 10; i++ {
		g.board = NewBoard()
		for x := 0; x < BoardWidth; x++ {
			if x != 4 {
				g.board.Lock([][2]int{{BoardHeight - 1, x}}, 1)
			}
		}
		g.spawnSpecific(PieceI)
		g.active.State = 1
		g.active.X = 2
		g.active.Y = 18
		g.HardDrop()
		g.Tick(ClearAnimMs + 1)
	}
	if g.Level() != 11 {
		t.Fatalf("expected starting level 10 to advance to 11, got %d", g.Level())
	}
}

func TestZeroLineTSpinScoresAndCounts(t *testing.T) {
	g := newTestGame(t)
	g.board = NewBoard()
	g.active = ActivePiece{Type: PieceT, State: 0, X: 3, Y: 10}
	g.hasActive = true
	g.lastWasRotation = true
	fill(g.board, 10, 3, 1)
	fill(g.board, 10, 5, 1)
	fill(g.board, 12, 3, 1)

	g.lockPiece()
	if g.Score() != 400 {
		t.Fatalf("expected 400 points for a zero-line T-spin, got %d", g.Score())
	}
	if g.Stats().TSpins != 1 {
		t.Fatalf("expected one recorded T-spin, got %d", g.Stats().TSpins)
	}
}

func TestDoubleTapGestureUsesDirect180Rotation(t *testing.T) {
	g := newTestGame(t)
	g.board = NewBoard()
	g.spawnSpecific(PieceT)
	original := g.Active()
	g.RotateGesture(RotateCW)
	if g.Active().State != 1 {
		t.Fatalf("expected immediate quarter turn, got state %d", g.Active().State)
	}
	g.CompleteGesture180(RotateCW)
	if g.Active().State != 2 {
		t.Fatalf("expected double tap to end 180 degrees from state %d, got %d", original.State, g.Active().State)
	}
	if g.gestureRotation {
		t.Fatal("completed double tap should clear gesture rotation state")
	}
}

func TestGesture180FallsBackAfterMovement(t *testing.T) {
	g := newTestGame(t)
	g.board = NewBoard()
	g.spawnSpecific(PieceT)
	g.RotateGesture(RotateCW)
	g.MoveRight()
	g.CompleteGesture180(RotateCW)
	if g.Active().State != 2 {
		t.Fatalf("expected a second quarter turn after movement, got state %d", g.Active().State)
	}
}

func TestDoubleTapUses180WhenFirstQuarterTurnIsBlocked(t *testing.T) {
	g := newTestGame(t)
	g.board = NewBoard()
	g.active = ActivePiece{Type: PieceT, State: 0, X: 3, Y: 5}
	g.hasActive = true
	fill(g.board, 7, 4, 1)
	fill(g.board, 7, 3, 1)
	fill(g.board, 4, 3, 1)

	g.RotateGesture(RotateCW)
	if g.Active().State != 0 {
		t.Fatalf("expected every quarter-turn kick to be blocked, got state %d", g.Active().State)
	}
	g.CompleteGesture180(RotateCW)
	if g.Active().State != 2 || g.Active().X != 4 {
		t.Fatalf("expected direct 180 kick to state 2 at x=4, got %+v", g.Active())
	}
}

func TestZenModeRecoversFromTopOut(t *testing.T) {
	g := NewGame(rand.New(rand.NewSource(1)), GameSettings{StartingLevel: 1, ZenMode: true})
	for y := 0; y < BoardHeight; y++ {
		for x := 0; x < BoardWidth; x++ {
			g.board.cells[y][x] = 1
		}
	}
	g.spawnNext()
	if g.State() != StatePlaying {
		t.Fatalf("zen game should continue after top out, got %v", g.State())
	}
	if g.Stats().ZenResets != 1 {
		t.Fatalf("expected one zen reset, got %d", g.Stats().ZenResets)
	}
}

func TestGameOverWhenSpawnBlocked(t *testing.T) {
	g := newTestGame(t)
	g.board = NewBoard()
	// Fill every visible row so a new spawn necessarily collides.
	for y := 2; y < BoardHeight; y++ {
		for x := 0; x < BoardWidth; x++ {
			g.board.Lock([][2]int{{y, x}}, 1)
		}
	}
	g.spawnNext()
	if g.State() != StateGameOver {
		t.Fatalf("expected GameOver, got %v", g.State())
	}
}

func TestPauseToggles(t *testing.T) {
	g := newTestGame(t)
	g.TogglePause()
	if g.State() != StatePaused {
		t.Fatalf("expected Paused, got %v", g.State())
	}
	g.TogglePause()
	if g.State() != StatePlaying {
		t.Fatalf("expected Playing, got %v", g.State())
	}
}
