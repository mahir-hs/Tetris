package service

import (
	"github.com/mahir/tetris/internal/database"
	"github.com/mahir/tetris/internal/repository"
)

// LeaderboardService queries the high-score table.
type LeaderboardService interface {
	Top(n int) ([]database.HighScoreRow, error)
}

type leaderboardService struct{ hs repository.HighScoreRepository }

// NewLeaderboardService builds a LeaderboardService.
func NewLeaderboardService(r repository.HighScoreRepository) LeaderboardService {
	return &leaderboardService{hs: r}
}

func (s *leaderboardService) Top(n int) ([]database.HighScoreRow, error) {
	return s.hs.TopN(n)
}
