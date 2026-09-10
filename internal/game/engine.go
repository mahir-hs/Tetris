package game

import (
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mahir/tetris/internal/config"
	"github.com/mahir/tetris/internal/domain"
	"github.com/mahir/tetris/internal/input"
	"github.com/mahir/tetris/internal/renderer"
)

// TickMsg is sent on a fixed cadence to advance the simulation.
type TickMsg struct {
	generation uint64
}

// QuitToMenuMsg tells the root model to leave the game screen.
type QuitToMenuMsg struct{}

// GameResult carries end-of-game data for persistence.
type GameResult struct {
	PlayerID      uint
	Score         int
	Level         int
	Lines         int
	DurationMs    int64
	PiecesPlaced  int
	Tetrises      int
	TSpins        int
	MiniTSpins    int
	PerfectClears int
	LongestCombo  int
}

// GameOverMsg is emitted once when a game ends.
type GameOverMsg GameResult

// Engine is the Bubble Tea model for an active game session.
type Engine struct {
	game           *domain.Game
	theme          renderer.Theme
	bindings       *input.Bindings
	gs             domain.GameSettings
	fps            int
	playerID       uint
	lastTick       time.Time
	notifiedOver   bool
	tickGeneration uint64
}

// NewEngine constructs a game engine for the given player and settings.
func NewEngine(playerID uint, cfg config.Settings, theme renderer.Theme, bindings *input.Bindings) *Engine {
	gs := domain.GameSettings{
		GhostEnabled:  cfg.GhostEnabled,
		HoldEnabled:   cfg.HoldEnabled,
		Rotate180:     cfg.Rotate180Enabled,
		StartingLevel: cfg.StartingLevel,
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &Engine{
		game:           domain.NewGame(rng, gs),
		theme:          theme,
		bindings:       bindings,
		gs:             gs,
		fps:            cfg.FPSLimit,
		playerID:       playerID,
		lastTick:       time.Now(),
		tickGeneration: 1,
	}
}

func (e *Engine) tickInterval() time.Duration {
	fps := e.fps
	if fps < 1 {
		fps = 60
	}
	ms := 1000 / fps
	if ms < 1 {
		ms = 1
	}
	return time.Duration(ms) * time.Millisecond
}

func tickCmd(d time.Duration, generation uint64) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return TickMsg{generation: generation} })
}

// Init starts the tick loop.
func (e *Engine) Init() tea.Cmd { return tickCmd(e.tickInterval(), e.tickGeneration) }

// Update handles time and input for the game.
func (e *Engine) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		if msg.generation != 0 && msg.generation != e.tickGeneration {
			return e, nil
		}
		now := time.Now()
		dt := now.Sub(e.lastTick)
		e.lastTick = now
		if e.game.State() == domain.StatePlaying {
			if dt > 100*time.Millisecond {
				dt = 100 * time.Millisecond // clamp after stalls
			}
			e.game.Tick(dt.Milliseconds())
		}
		if e.game.State() == domain.StateGameOver {
			if !e.notifiedOver {
				e.notifiedOver = true
				return e, e.gameOverCmd()
			}
			return e, nil
		}
		// Playing or paused: keep the tick loop alive.
		return e, tickCmd(e.tickInterval(), e.tickGeneration)

	case tea.KeyMsg:
		if cmd := e.handleKey(msg); cmd != nil {
			return e, cmd
		}
	}

	// A key press (e.g. hard drop) may have ended the game.
	if e.game.State() == domain.StateGameOver && !e.notifiedOver {
		e.notifiedOver = true
		return e, e.gameOverCmd()
	}
	return e, nil
}

func (e *Engine) handleKey(msg tea.KeyMsg) tea.Cmd {
	a := e.bindings.Match(msg)

	// Keys that work in any game state.
	switch a {
	case config.ActionQuit:
		return func() tea.Msg { return QuitToMenuMsg{} }
	case config.ActionRestart:
		e.resetGame()
		e.tickGeneration++
		return tickCmd(e.tickInterval(), e.tickGeneration)
	case config.ActionPause:
		e.game.TogglePause()
		e.lastTick = time.Now()
		return nil
	}

	if e.game.State() == domain.StatePaused || e.game.State() == domain.StateGameOver {
		return nil
	}

	switch a {
	case config.ActionMoveLeft:
		e.game.MoveLeft()
	case config.ActionMoveRight:
		e.game.MoveRight()
	case config.ActionSoftDrop:
		e.game.SoftDrop()
	case config.ActionHardDrop:
		e.game.HardDrop()
	case config.ActionRotateCW:
		e.game.RotateCW()
	case config.ActionRotateCCW:
		e.game.RotateCCW()
	case config.ActionRotate180:
		e.game.Rotate180()
	case config.ActionHold:
		e.game.Hold()
	}
	return nil
}

func (e *Engine) resetGame() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	e.game = domain.NewGame(rng, e.gs)
	e.notifiedOver = false
	e.lastTick = time.Now()
}

func (e *Engine) gameOverCmd() tea.Cmd {
	s := e.game.Stats()
	res := GameResult{
		PlayerID:      e.playerID,
		Score:         e.game.Score(),
		Level:         e.game.Level(),
		Lines:         e.game.Lines(),
		DurationMs:    e.game.ElapsedMs(),
		PiecesPlaced:  s.PiecesPlaced,
		Tetrises:      s.Tetrises,
		TSpins:        s.TSpins,
		MiniTSpins:    s.MiniTSpins,
		PerfectClears: s.PerfectClears,
		LongestCombo:  s.LongestCombo,
	}
	return func() tea.Msg { return GameOverMsg(res) }
}

// View renders the playfield, HUD, and any overlay at the standard cell size.
func (e *Engine) View() string {
	layout := renderer.StandardLayout
	field := renderer.Playfield(e.game, e.theme, layout)
	hud := renderer.HUD(e.game, e.theme, layout)
	row := lipgloss.JoinHorizontal(lipgloss.Top, field, "  ", hud)

	switch e.game.State() {
	case domain.StatePaused:
		row = renderer.Overlay(row, renderer.PauseView(e.theme), e.theme)
	case domain.StateGameOver:
		row = renderer.Overlay(row, renderer.GameOverView(e.game, e.theme), e.theme)
	}
	return row
}

// Game exposes the underlying engine game (used by overlays/tests).
func (e *Engine) Game() *domain.Game { return e.game }
