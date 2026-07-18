package domain

import "math/rand"

// GameState is the play state machine for a single game.
type GameState int

const (
	StateReady   GameState = iota // not yet started
	StatePlaying                  // active play
	StatePaused                   // paused
	StateGameOver                 // finished
)

// Timing constants (milliseconds).
const (
	LockDelayMs        = 500
	MaxLockResets      = 15
	ClearAnimMs        = 180
	SoftDropIntervalMs = 20
)

// GameSettings configures optional mechanics for a game.
type GameSettings struct {
	GhostEnabled  bool
	HoldEnabled   bool
	Rotate180     bool
	StartingLevel int
}

// GameStats is a running tally of notable events for the end screen / DB.
type GameStats struct {
	PiecesPlaced  int
	Tetrises      int
	TSpins        int
	MiniTSpins    int
	PerfectClears int
	LongestCombo  int
}

// Game is the core, I/O-free Tetris engine. It is driven by Tick (time advance)
// and the action methods (Move/Rotate/Drop/Hold). Rendering and persistence are
// the responsibility of higher layers.
type Game struct {
	board  *Board
	rand   *Randomizer
	scorer ScoringStrategy
	cfg    GameSettings

	state   GameState
	active  ActivePiece
	hasActive bool
	hold    PieceType
	hasHold bool
	holdUsed bool
	queue   []PieceType

	score int
	level int
	lines int
	combo int // -1 when no chain active
	b2b   bool

	lockTimer  int64
	lockResets int
	grounded   bool
	lowestY    int
	gravityTimer int64
	elapsed     int64
	softDrop    bool

	lastWasRotation bool

	// line-clear animation
	clearing    bool
	clearRows   []int
	clearCount  int
	pendingB2B  bool
	clearTimer  int64

	stats GameStats
}

// NewGame constructs a game with the given RNG and settings.
func NewGame(rng *rand.Rand, cfg GameSettings) *Game {
	if cfg.StartingLevel < 1 {
		cfg.StartingLevel = 1
	}
	g := &Game{
		board:  NewBoard(),
		rand:   NewRandomizer(rng),
		scorer: GuidelineScorer{},
		cfg:    cfg,
		level:  cfg.StartingLevel,
		combo:  -1,
	}
	g.ensureQueue()
	g.spawnNext()
	g.state = StatePlaying
	return g
}

func (g *Game) ensureQueue() {
	for len(g.queue) < 7 {
		g.queue = append(g.queue, g.rand.Next())
	}
}

// spawnNext pulls the next piece from the queue (resetting hold usage).
func (g *Game) spawnNext() {
	g.ensureQueue()
	t := g.queue[0]
	g.queue = g.queue[1:]
	g.ensureQueue()
	g.spawnSpecific(t)
	g.holdUsed = false
}

// spawnSpecific places a concrete piece type, resetting per-piece timers but NOT
// hold usage (used by the hold swap).
func (g *Game) spawnSpecific(t PieceType) {
	g.active = ActivePiece{}.Spawn(t)
	g.hasActive = true
	g.lockTimer = LockDelayMs
	g.lockResets = 0
	g.grounded = false
	g.lowestY = g.active.Y
	g.gravityTimer = 0
	g.lastWasRotation = false
	if !g.board.CanPlace(g.active.Cells()) {
		g.state = StateGameOver
	}
}

// --- Accessors -------------------------------------------------------------

func (g *Game) State() GameState        { return g.state }
func (g *Game) Board() *Board           { return g.board }
func (g *Game) Active() ActivePiece     { return g.active }
func (g *Game) HasActive() bool         { return g.hasActive }
func (g *Game) HeldPiece() (PieceType, bool) { return g.hold, g.hasHold }
func (g *Game) Score() int              { return g.score }
func (g *Game) Level() int              { return g.level }
func (g *Game) Lines() int              { return g.lines }
func (g *Game) Combo() int              { return g.combo }
func (g *Game) BackToBack() bool        { return g.b2b }
func (g *Game) Stats() GameStats        { return g.stats }
func (g *Game) ElapsedMs() int64        { return g.elapsed }
func (g *Game) IsClearing() bool        { return g.clearing }
func (g *Game) ClearRows() []int        { return g.clearRows }

// NextQueue returns up to n upcoming piece types (not yet spawned).
func (g *Game) NextQueue(n int) []PieceType {
	if n > len(g.queue) {
		n = len(g.queue)
	}
	out := make([]PieceType, n)
	copy(out, g.queue[:n])
	return out
}

// GhostY returns the board row (bounding-box top) where the active piece would
// come to rest if hard-dropped, or -1 if no active piece / disabled.
func (g *Game) GhostY() int {
	if !g.cfg.GhostEnabled || !g.hasActive || g.clearing {
		return -1
	}
	y := g.active.Y
	for g.board.CanPlace(g.active.Tetromino().Cells(g.active.State, g.active.X, y+1)) {
		y++
	}
	return y
}

// --- Actions ---------------------------------------------------------------

// SetSoftDrop toggles the continuous soft-drop speed.
func (g *Game) SetSoftDrop(on bool) { g.softDrop = on }

// MoveLeft / MoveRight attempt a horizontal shift.
func (g *Game) MoveLeft()  { g.tryShift(-1) }
func (g *Game) MoveRight() { g.tryShift(1) }

func (g *Game) tryShift(dx int) {
	if !g.canAct() {
		return
	}
	if g.board.CanPlace(g.active.Tetromino().Cells(g.active.State, g.active.X+dx, g.active.Y)) {
		g.active.X += dx
		g.lastWasRotation = false
		g.onSuccessfulManipulation()
	}
}

// RotateCW / RotateCCW / Rotate180 attempt rotation with SRS kicks.
func (g *Game) RotateCW() {
	if g.canAct() {
		g.applyRotation(TryRotate(g.board, g.active.Tetromino(), g.active.State, g.active.X, g.active.Y, RotateCW))
	}
}
func (g *Game) RotateCCW() {
	if g.canAct() {
		g.applyRotation(TryRotate(g.board, g.active.Tetromino(), g.active.State, g.active.X, g.active.Y, RotateCCW))
	}
}
func (g *Game) Rotate180() {
	if g.canAct() && g.cfg.Rotate180 {
		g.applyRotation(TryRotate(g.board, g.active.Tetromino(), g.active.State, g.active.X, g.active.Y, Rotate180))
	}
}

func (g *Game) applyRotation(r RotateResult) {
	if !r.Changed {
		return
	}
	g.active.State, g.active.X, g.active.Y = r.State, r.X, r.Y
	g.lastWasRotation = true
	g.onSuccessfulManipulation()
}

// onSuccessfulManipulation handles lock-delay reset rules after a move/rotate.
func (g *Game) onSuccessfulManipulation() {
	if g.board.CanPlace(g.active.Tetromino().Cells(g.active.State, g.active.X, g.active.Y+1)) {
		// airborne again
		g.grounded = false
		return
	}
	// still resting: reset lock timer if resets remain
	g.grounded = true
	if g.lockResets < MaxLockResets {
		g.lockTimer = LockDelayMs
		g.lockResets++
	}
}

// SoftDrop steps the piece down one row if possible (awards +1).
func (g *Game) SoftDrop() {
	if !g.canAct() {
		return
	}
	if g.board.CanPlace(g.active.Tetromino().Cells(g.active.State, g.active.X, g.active.Y+1)) {
		g.active.Y++
		g.score += g.scorer.SoftDropPoints(1)
		if g.active.Y > g.lowestY {
			g.lowestY = g.active.Y
			g.lockResets = 0
		}
	}
}

// HardDrop drops the piece to the bottom instantly, awards +2/cell, and locks.
func (g *Game) HardDrop() {
	if !g.canAct() {
		return
	}
	dist := 0
	for g.board.CanPlace(g.active.Tetromino().Cells(g.active.State, g.active.X, g.active.Y+1)) {
		g.active.Y++
		dist++
	}
	g.score += g.scorer.HardDropPoints(dist)
	g.lockPiece()
}

// Hold swaps the active piece with the hold slot (once per piece).
func (g *Game) Hold() {
	if !g.canAct() || !g.cfg.HoldEnabled || g.holdUsed {
		return
	}
	current := g.active.Type
	if g.hasHold {
		swapped := g.hold
		g.hold = current
		g.spawnSpecific(swapped)
	} else {
		g.hold = current
		g.hasHold = true
		g.spawnNext()
	}
	g.holdUsed = true
}

// TogglePause flips between playing and paused.
func (g *Game) TogglePause() {
	if g.state == StatePlaying {
		g.state = StatePaused
	} else if g.state == StatePaused {
		g.state = StatePlaying
	}
}

func (g *Game) canAct() bool {
	return g.state == StatePlaying && g.hasActive && !g.clearing
}

// --- Time advance ----------------------------------------------------------

// Tick advances the simulation by dtMs milliseconds.
func (g *Game) Tick(dtMs int64) {
	if g.state != StatePlaying {
		return
	}
	g.elapsed += dtMs
	if g.clearing {
		g.clearTimer -= dtMs
		if g.clearTimer <= 0 {
			g.finishClear()
		}
		return
	}
	if !g.hasActive {
		return
	}

	canDescend := g.board.CanPlace(g.active.Tetromino().Cells(g.active.State, g.active.X, g.active.Y+1))
	if !canDescend {
		if !g.grounded {
			g.grounded = true
			g.lockTimer = LockDelayMs
		}
	} else {
		g.grounded = false
	}

	if canDescend {
		g.gravityTimer += dtMs
		interval := g.gravityInterval()
		for g.gravityTimer >= interval {
			g.gravityTimer -= interval
			if g.board.CanPlace(g.active.Tetromino().Cells(g.active.State, g.active.X, g.active.Y+1)) {
				g.active.Y++
				if g.softDrop {
					g.score += g.scorer.SoftDropPoints(1)
				}
				if g.active.Y > g.lowestY {
					g.lowestY = g.active.Y
					g.lockResets = 0
				}
			} else {
				break
			}
		}
	}

	if g.grounded {
		g.lockTimer -= dtMs
		if g.lockTimer <= 0 {
			g.lockPiece()
		}
	}
}

func (g *Game) gravityInterval() int64 {
	base := GravityIntervalMs(g.level)
	if g.softDrop && base > SoftDropIntervalMs {
		return SoftDropIntervalMs
	}
	return base
}

// lockPiece finalizes the active piece, resolves line clears / scoring / combos.
func (g *Game) lockPiece() {
	locked := g.active
	tm := locked.Tetromino()
	g.board.Lock(locked.Cells(), tm.Color)
	g.hasActive = false
	g.stats.PiecesPlaced++

	rows, count := g.board.fullRows()
	if count == 0 {
		g.combo = -1
		g.spawnNext()
		return
	}

	tspin := DetectTSpin(g.board, locked, g.lastWasRotation)
	difficult := count == 4 || tspin != NoTSpin

	// combo
	g.combo++
	if g.combo > g.stats.LongestCombo {
		g.stats.LongestCombo = g.combo
	}

	// back-to-back
	backToBack := false
	if difficult {
		if g.b2b {
			backToBack = true
		}
		g.b2b = true
	} else {
		g.b2b = false
	}

	ctx := ClearContext{
		LinesCleared: count,
		TSpin:        tspin,
		Level:        g.level,
		Combo:        g.combo,
		BackToBack:   backToBack,
	}
	g.score += g.scorer.LockBonus(ctx)

	g.lines += count
	g.level = LevelForLines(g.lines)
	g.clearRows = rows
	g.clearCount = count
	g.pendingB2B = backToBack
	g.clearing = true
	g.clearTimer = ClearAnimMs

	// stats
	if tspin == TSpinFull {
		g.stats.TSpins++
	} else if tspin == TSpinMini {
		g.stats.MiniTSpins++
	}
	if count == 4 {
		g.stats.Tetrises++
	}
}

// finishClear collapses the board after the animation and awards perfect clear.
func (g *Game) finishClear() {
	count := g.clearCount
	g.board.ClearLines()
	g.clearing = false
	g.clearRows = nil
	g.clearCount = 0
	if g.board.IsPerfectClear() {
		g.stats.PerfectClears++
		g.score += perfectClearScore(count, g.pendingB2B, g.level)
	}
	g.spawnNext()
}

// perfectClearScore mirrors the GuidelineScorer perfect-clear values.
func perfectClearScore(count int, b2b bool, level int) int {
	base := 0
	switch {
	case b2b && count == 4:
		base = 3200
	case count == 1:
		base = 800
	case count == 2:
		base = 1200
	case count == 3:
		base = 1800
	case count == 4:
		base = 2000
	}
	return base * level
}

// fullRows returns the indices of completely filled rows without clearing them.
func (b *Board) fullRows() (rows []int, count int) {
	for y := 0; y < BoardHeight; y++ {
		full := true
		for x := 0; x < BoardWidth; x++ {
			if b.cells[y][x] == 0 {
				full = false
				break
			}
		}
		if full {
			rows = append(rows, y)
		}
	}
	return rows, len(rows)
}
