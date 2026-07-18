package repository

import (
	"github.com/mahir/tetris/internal/database"
	"gorm.io/gorm"
)

// GameRepository persists finished game sessions.
type GameRepository interface {
	Create(g *database.Game) error
}

type gameRepo struct{ db *gorm.DB }

// NewGameRepo builds a GameRepository.
func NewGameRepo(db *gorm.DB) GameRepository { return &gameRepo{db: db} }

func (r *gameRepo) Create(g *database.Game) error {
	return r.db.Create(g).Error
}
