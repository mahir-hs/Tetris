// Package log provides a simple file logger for the game.
package log

import (
	"log"
	"os"
	"path/filepath"
)

// Logger writes game events to logs/game.log (falling back to stderr).
var Logger *log.Logger

func init() {
	_ = os.MkdirAll("logs", 0o755)
	f, err := os.OpenFile(filepath.Join("logs", "game.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		f = os.Stderr
	}
	Logger = log.New(f, "tetris ", log.LstdFlags)
}

// Infof logs an informational event.
func Infof(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Printf(format, args...)
	}
}

// Errorf logs an error event.
func Errorf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Printf("ERROR: "+format, args...)
	}
}
