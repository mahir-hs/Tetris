package database

import (
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Player is a persisted user profile.
type Player struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// Game is a single finished (or abandoned) play session.
type Game struct {
	ID        uint      `gorm:"primaryKey"`
	PlayerID  uint      `gorm:"index"`
	StartedAt time.Time `gorm:"autoCreateTime"`
	EndedAt   time.Time
	Score     int
	Level     int
	Lines     int
	Duration  int // seconds
	Result    string
}

// HighScore is one leaderboard entry.
type HighScore struct {
	ID       uint      `gorm:"primaryKey"`
	PlayerID uint      `gorm:"index"`
	Score    int       `gorm:"index"`
	Level    int
	Lines    int
	Date     time.Time `gorm:"autoCreateTime"`
}

// Statistics aggregates a player's lifetime performance.
type Statistics struct {
	ID               uint    `gorm:"primaryKey"`
	PlayerID         uint    `gorm:"uniqueIndex"`
	GamesPlayed      int     `gorm:"default:0"`
	HighestScore     int     `gorm:"default:0"`
	HighestLevel     int     `gorm:"default:0"`
	TotalLines       int     `gorm:"default:0"`
	TotalPlayTime    int     `gorm:"default:0"` // seconds
	TotalTetrises    int     `gorm:"default:0"`
	TotalTSpins      int     `gorm:"default:0"`
	TotalPerfect     int     `gorm:"default:0"`
	LongestCombo     int     `gorm:"default:0"`
	AverageScore     float64 `gorm:"default:0"`
	AverageDuration  float64 `gorm:"default:0"`
}

// HighScoreRow is a leaderboard entry joined with the player name.
type HighScoreRow struct {
	Rank     int
	Username string
	Score    int
	Level    int
	Lines    int
	Date     time.Time
}

// DBFilename is the default SQLite database file.
const DBFilename = "tetris.db"

// InitDB opens (creating if needed) the SQLite database and migrates the schema.
func InitDB(path string) (*gorm.DB, error) {
	if path == "" {
		path = DBFilename
	}
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Player{}, &Game{}, &HighScore{}, &Statistics{}); err != nil {
		return nil, err
	}
	return db, nil
}
