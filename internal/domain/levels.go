package domain

import "math"

// LinesPerLevel is how many cleared lines advance the level by one.
const LinesPerLevel = 10

// GravityIntervalMs returns the time (in milliseconds) a piece takes to fall
// one row at the given level, following the official Tetris Guideline curve:
//
//	seconds_per_row = (0.8 - (level-1)*0.007) ^ (level-1)
//
// At level 1 this yields 1000 ms (1 cell/sec), accelerating with level.
func GravityIntervalMs(level int) int64 {
	if level < 1 {
		level = 1
	}
	base := 0.8 - float64(level-1)*0.007
	if base <= 0 {
		base = 0.0001
	}
	seconds := math.Pow(base, float64(level-1))
	if seconds < 0.001 {
		seconds = 0.001
	}
	return int64(seconds * 1000)
}

// LevelForLines returns the level reached after clearing the given total lines
// (level 1 is the start; every LinesPerLevel lines increments it).
func LevelForLines(totalLines int) int {
	return 1 + totalLines/LinesPerLevel
}
