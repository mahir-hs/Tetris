package repository

import (
	"errors"

	"github.com/mahir/tetris/internal/database"
	"gorm.io/gorm"
)

// PlayerRepository persists player profiles.
type PlayerRepository interface {
	FindByUsername(name string) (*database.Player, error)
	FindOrCreate(name string) (*database.Player, error)
}

type playerRepo struct{ db *gorm.DB }

// NewPlayerRepository builds a PlayerRepository.
func NewPlayerRepo(db *gorm.DB) PlayerRepository { return &playerRepo{db: db} }

func (r *playerRepo) FindByUsername(name string) (*database.Player, error) {
	var p database.Player
	err := r.db.Where("username = ?", name).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *playerRepo) FindOrCreate(name string) (*database.Player, error) {
	p, err := r.FindByUsername(name)
	if err == nil {
		return p, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	p = &database.Player{Username: name}
	if err := r.db.Create(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}
