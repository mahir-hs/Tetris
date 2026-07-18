package service

import (
	"github.com/mahir/tetris/internal/database"
	"github.com/mahir/tetris/internal/repository"
)

// PlayerService manages player profiles and their statistics lookup.
type PlayerService interface {
	GetOrCreate(username string) (*database.Player, error)
	Stats(playerID uint) (*database.Statistics, error)
}

type playerService struct {
	players repository.PlayerRepository
	stats   repository.StatsRepository
}

// NewPlayerService builds a PlayerService.
func NewPlayerService(p repository.PlayerRepository, s repository.StatsRepository) PlayerService {
	return &playerService{players: p, stats: s}
}

func (s *playerService) GetOrCreate(username string) (*database.Player, error) {
	return s.players.FindOrCreate(username)
}

func (s *playerService) Stats(playerID uint) (*database.Statistics, error) {
	return s.stats.FindByPlayer(playerID)
}
