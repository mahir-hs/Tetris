package repository

import (
	"errors"

	"github.com/mahir/tetris/internal/database"
	"gorm.io/gorm"
)

// StatsRepository persists aggregated player statistics.
type StatsRepository interface {
	FindByPlayer(id uint) (*database.Statistics, error)
	Upsert(s *database.Statistics) error
}

type statsRepo struct{ db *gorm.DB }

// NewStatsRepo builds a StatsRepository.
func NewStatsRepo(db *gorm.DB) StatsRepository { return &statsRepo{db: db} }

func (r *statsRepo) FindByPlayer(id uint) (*database.Statistics, error) {
	var s database.Statistics
	err := r.db.Where("player_id = ?", id).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &database.Statistics{PlayerID: id}, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *statsRepo) Upsert(s *database.Statistics) error {
	// Save inserts when ID is unset and updates otherwise.
	return r.db.Save(s).Error
}
