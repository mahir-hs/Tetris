package game

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mahir/tetris/internal/config"
	"github.com/mahir/tetris/internal/input"
	"github.com/mahir/tetris/internal/renderer"
)

func TestEngineViewRenders(t *testing.T) {
	cfg := config.DefaultSettings()
	bindings := input.NewBindings(cfg)
	e := NewEngine(1, cfg, renderer.GetTheme("classic"), bindings)
	out := e.View()
	if out == "" {
		t.Fatal("empty engine view")
	}
	t.Logf("\n%s", out)
}

func TestEngineHardDropScores(t *testing.T) {
	cfg := config.DefaultSettings()
	bindings := input.NewBindings(cfg)
	e := NewEngine(1, cfg, renderer.GetTheme("classic"), bindings)
	_, _ = e.Update(tea.KeyMsg{Type: tea.KeySpace}) // hard drop
	if e.Game().Score() <= 0 {
		t.Fatalf("hard drop should award score, got %d", e.Game().Score())
	}
	if e.Game().Stats().PiecesPlaced != 1 {
		t.Fatalf("expected 1 piece placed, got %d", e.Game().Stats().PiecesPlaced)
	}
}

func TestEngineQuitKey(t *testing.T) {
	cfg := config.DefaultSettings()
	bindings := input.NewBindings(cfg)
	e := NewEngine(1, cfg, renderer.GetTheme("classic"), bindings)
	_, cmd := e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected a command from quit key")
	}
	msg := cmd()
	if _, ok := msg.(QuitToMenuMsg); !ok {
		t.Fatalf("expected QuitToMenuMsg, got %T", msg)
	}
}

func TestEngineTickAdvances(t *testing.T) {
	cfg := config.DefaultSettings()
	bindings := input.NewBindings(cfg)
	e := NewEngine(1, cfg, renderer.GetTheme("classic"), bindings)
	for i := 0; i < 50; i++ {
		_, _ = e.Update(TickMsg{})
		time.Sleep(20 * time.Millisecond) // let real time advance so dt > 0
	}
	// After ~1s the elapsed timer and gravity should have progressed.
	if e.Game().ElapsedMs() <= 0 {
		t.Fatal("expected elapsed time to advance")
	}
}
