package service_test

import (
	"path/filepath"
	"testing"

	"github.com/mahir/tetris/internal/database"
	"github.com/mahir/tetris/internal/repository"
	"github.com/mahir/tetris/internal/service"
)

func setup(t *testing.T) (service.PlayerService, service.LeaderboardService, service.StatsService, func() error) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "tetris.db")
	db, err := database.InitDB(dbPath)
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	pr := repository.NewPlayerRepo(db)
	gr := repository.NewGameRepo(db)
	hr := repository.NewHighScoreRepo(db)
	sr := repository.NewStatsRepo(db)
	closeFn := func() error {
		if c, e := db.DB(); e == nil {
			return c.Close()
		}
		return nil
	}
	return service.NewPlayerService(pr, sr), service.NewLeaderboardService(hr), service.NewStatsService(gr, hr, sr), closeFn
}

func TestPersistenceRoundTrip(t *testing.T) {
	playerSvc, lbSvc, statsSvc, closeDB := setup(t)
	defer closeDB()

	p, err := playerSvc.GetOrCreate("alice")
	if err != nil {
		t.Fatalf("create player: %v", err)
	}

	outcome := service.GameOutcome{
		PlayerID: p.ID, Score: 5000, Level: 3, Lines: 12,
		DurationSec: 100, PiecesPlaced: 40, Tetrises: 1, TSpins: 1,
		MiniTSpins: 1, PerfectClears: 0, LongestCombo: 2,
	}
	if err := statsSvc.RecordGame(outcome); err != nil {
		t.Fatalf("record game: %v", err)
	}

	rows, err := lbSvc.Top(20)
	if err != nil {
		t.Fatalf("top scores: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 high score, got %d", len(rows))
	}
	if rows[0].Score != 5000 || rows[0].Username != "alice" {
		t.Fatalf("unexpected high score row: %+v", rows[0])
	}

	st, err := playerSvc.Stats(p.ID)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if st.GamesPlayed != 1 || st.HighestScore != 5000 || st.TotalLines != 12 {
		t.Fatalf("unexpected stats: %+v", st)
	}
	if st.TotalTSpins != 2 { // full + mini
		t.Fatalf("expected 2 total tspins, got %d", st.TotalTSpins)
	}

	// Second game updates aggregates.
	outcome2 := outcome
	outcome2.Score = 3000
	outcome2.Lines = 8
	if err := statsSvc.RecordGame(outcome2); err != nil {
		t.Fatalf("record second game: %v", err)
	}
	st2, _ := playerSvc.Stats(p.ID)
	if st2.GamesPlayed != 2 {
		t.Fatalf("expected 2 games played, got %d", st2.GamesPlayed)
	}
	if st2.HighestScore != 5000 {
		t.Fatalf("highest score should remain 5000, got %d", st2.HighestScore)
	}
	if st2.AverageScore < 3000 || st2.AverageScore > 5000 {
		t.Fatalf("average score out of range: %v", st2.AverageScore)
	}
	rows2, _ := lbSvc.Top(20)
	if len(rows2) != 2 {
		t.Fatalf("expected 2 leaderboard rows, got %d", len(rows2))
	}
	if rows2[0].Score != 5000 {
		t.Fatalf("expected highest first, got %d", rows2[0].Score)
	}
}
