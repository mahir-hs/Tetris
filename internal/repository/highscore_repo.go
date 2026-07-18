package repository

import (
	"github.com/mahir/tetris/internal/database"
	"gorm.io/gorm"
)

// HighScoreRepository persists and queries leaderboard entries.
type HighScoreRepository interface {
	Create(hs *database.HighScore) error
	TopN(n int) ([]database.HighScoreRow, error)
}

type highScoreRepo struct{ db *gorm.DB }

// NewHighScoreRepo builds a HighScoreRepository.
func NewHighScoreRepo(db *gorm.DB) HighScoreRepository { return &highScoreRepo{db: db} }

func (r *highScoreRepo) Create(hs *database.HighScore) error {
	return r.db.Create(hs).Error
}

func (r *highScoreRepo) TopN(n int) ([]database.HighScoreRow, error) {
	var rows []database.HighScoreRow
	err := r.db.Table("high_scores AS h").
		Select("h.score AS score, h.level AS level, h.lines AS lines, h.date AS date, p.username AS username").
		Joins("JOIN players p ON p.id = h.player_id").
		Order("h.score DESC").
		Limit(n).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].Rank = i + 1
	}
	return rows, nil
}
