package app

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mahir/tetris/internal/config"
	"github.com/mahir/tetris/internal/database"
	"github.com/mahir/tetris/internal/domain"
	"github.com/mahir/tetris/internal/game"
	"github.com/mahir/tetris/internal/repository"
	"github.com/mahir/tetris/internal/service"
)

func newTestApp(t *testing.T) (RootModel, service.PlayerService, service.StatsService) {
	t.Helper()
	db, err := database.InitDB(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	pr := repository.NewPlayerRepo(db)
	gr := repository.NewGameRepo(db)
	hr := repository.NewHighScoreRepo(db)
	sr := repository.NewStatsRepo(db)
	ps := service.NewPlayerService(pr, sr)
	ls := service.NewLeaderboardService(hr)
	ss := service.NewStatsService(gr, hr, sr)
	t.Cleanup(func() {
		if c, e := db.DB(); e == nil {
			_ = c.Close()
		}
	})
	return NewRootModel(config.DefaultSettings(), ps, ls, ss), ps, ss
}

func key(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func TestRootMenuToGameFlow(t *testing.T) {
	m, ps, ss := newTestApp(t)

	var step func(tea.Msg)
	step = func(msg tea.Msg) {
		model, cmd := m.Update(msg)
		m = model.(RootModel)
		if cmd != nil {
			produced := cmd()
			if _, isTick := produced.(game.TickMsg); !isTick {
				step(produced) // deliver GameOverMsg / QuitToMenuMsg back
			}
		}
	}

	// Select "Play" (first menu option) and confirm.
	step(tea.KeyMsg{Type: tea.KeyEnter})
	if m.scr != screenEnterName {
		t.Fatalf("expected name entry screen, got %v", m.scr)
	}

	// Type a username and confirm.
	for _, r := range "bob" {
		step(key(r))
	}
	step(tea.KeyMsg{Type: tea.KeyEnter})

	if m.scr != screenGame {
		t.Fatalf("expected game screen, got %v", m.scr)
	}
	if m.player == nil || m.player.Username != "bob" {
		t.Fatalf("expected player bob, got %+v", m.player)
	}
	if m.engine == nil {
		t.Fatal("expected an engine after starting a game")
	}

	// Drive hard drops until the game ends.
	for i := 0; i < 2000; i++ {
		step(tea.KeyMsg{Type: tea.KeySpace})
		if m.engine != nil && m.engine.Game().State() == domain.StateGameOver {
			step(game.TickMsg{}) // flush the GameOverMsg
			break
		}
		step(game.TickMsg{})
	}

	if m.engine == nil || m.engine.Game().State() != domain.StateGameOver {
		t.Fatal("expected game over after many hard drops")
	}

	// A game should have been persisted for bob.
	st, err := ps.Stats(m.player.ID)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if st.GamesPlayed < 1 {
		t.Fatalf("expected at least 1 recorded game, got %d", st.GamesPlayed)
	}
	if st.HighestScore <= 0 {
		t.Fatalf("expected a positive high score, got %d", st.HighestScore)
	}
	_ = ss
}

func TestRootPersistDirectly(t *testing.T) {
	m, ps, _ := newTestApp(t)
	p, err := ps.GetOrCreate("carol")
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	m.player = p
	m.persist(game.GameResult{PlayerID: p.ID, Score: 4321, Level: 4, Lines: 30, DurationMs: 60000})

	st, _ := ps.Stats(p.ID)
	if st.GamesPlayed != 1 || st.HighestScore != 4321 || st.TotalLines != 30 {
		t.Fatalf("unexpected stats after persist: %+v", st)
	}
}

func TestRootMenuNavigation(t *testing.T) {
	m, _, _ := newTestApp(t)
	if m.menuIndex != 0 {
		t.Fatal("menu should start at index 0")
	}
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(RootModel)
	if m.menuIndex != 1 {
		t.Fatalf("down should move to index 1, got %d", m.menuIndex)
	}
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = model.(RootModel)
	if m.menuIndex != 0 {
		t.Fatalf("up should return to index 0, got %d", m.menuIndex)
	}
}
