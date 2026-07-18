package service

import (
	"time"

	"github.com/mahir/tetris/internal/database"
	"github.com/mahir/tetris/internal/repository"
)

// GameOutcome is the data needed to persist a finished game.
type GameOutcome struct {
	PlayerID      uint
	Score         int
	Level         int
	Lines         int
	DurationSec   int
	PiecesPlaced  int
	Tetrises      int
	TSpins        int
	MiniTSpins    int
	PerfectClears int
	LongestCombo  int
}

// StatsService records finished games and maintains aggregated statistics.
type StatsService interface {
	RecordGame(o GameOutcome) error
}

type statsService struct {
	games repository.GameRepository
	hs    repository.HighScoreRepository
	stats repository.StatsRepository
}

// NewStatsService builds a StatsService.
func NewStatsService(g repository.GameRepository, h repository.HighScoreRepository, s repository.StatsRepository) StatsService {
	return &statsService{games: g, hs: h, stats: s}
}

func (s *statsService) RecordGame(o GameOutcome) error {
	ended := time.Now()
	started := ended.Add(-time.Duration(o.DurationSec) * time.Second)

	// 1. Persist the game session.
	game := &database.Game{
		PlayerID:  o.PlayerID,
		StartedAt: started,
		EndedAt:   ended,
		Score:     o.Score,
		Level:     o.Level,
		Lines:     o.Lines,
		Duration:  o.DurationSec,
		Result:    "gameover",
	}
	if err := s.games.Create(game); err != nil {
		return err
	}

	// 2. Leaderboard entry.
	if err := s.hs.Create(&database.HighScore{
		PlayerID: o.PlayerID,
		Score:    o.Score,
		Level:    o.Level,
		Lines:    o.Lines,
	}); err != nil {
		return err
	}

	// 3. Aggregate statistics.
	st, err := s.stats.FindByPlayer(o.PlayerID)
	if err != nil {
		return err
	}
	prevGames := st.GamesPlayed
	st.GamesPlayed++
	st.HighestScore = maxInt(st.HighestScore, o.Score)
	st.HighestLevel = maxInt(st.HighestLevel, o.Level)
	st.TotalLines += o.Lines
	st.TotalPlayTime += o.DurationSec
	st.TotalTetrises += o.Tetrises
	st.TotalTSpins += o.TSpins + o.MiniTSpins
	st.TotalPerfect += o.PerfectClears
	st.LongestCombo = maxInt(st.LongestCombo, o.LongestCombo)
	st.AverageScore = rollingAvg(st.AverageScore, prevGames, float64(o.Score))
	st.AverageDuration = rollingAvg(st.AverageDuration, prevGames, float64(o.DurationSec))
	return s.stats.Upsert(st)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func rollingAvg(prev float64, prevCount int, val float64) float64 {
	if prevCount == 0 {
		return val
	}
	return (prev*float64(prevCount) + val) / float64(prevCount+1)
}
