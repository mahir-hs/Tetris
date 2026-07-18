# CLI Tetris Game Specification

## Overview

Build a **production-quality, fully-featured CLI Tetris game** that follows the **modern Tetris Guideline** as closely as possible.

The game should run entirely inside the terminal while maintaining smooth gameplay, responsive controls, persistent local storage, and a clean, extensible architecture.

---

# Tech Stack

## Language

- Go (Preferred)
- Rust (Alternative)

## Recommended Libraries

### Go

- Bubble Tea
- Lip Gloss
- Bubbles
- GORM
- SQLite

### Rust

- Ratatui
- Crossterm
- sqlx or Diesel
- SQLite

---

# Project Goals

- Follow official Tetris mechanics
- Smooth gameplay
- Persistent player profiles
- Persistent leaderboard
- Local statistics
- Clean Architecture
- Easily extensible
- Well documented
- Fully tested

---

# Game Board

Implement the official board size.

- Width: 10 columns
- Visible Height: 20 rows
- Hidden Spawn Area: 2 rows

The board should display:

- Colored tetrominoes
- Unicode block rendering
- Border around playfield
- Ghost piece
- Line clear animations

---

# Tetrominoes

Support all seven official tetrominoes.

- I
- O
- T
- S
- Z
- J
- L

Each piece must have:

- Official spawn position
- Official colors
- Proper rotation states

---

# Random Piece Generation

Use the official **7-Bag Randomizer**.

Requirements:

- Generate all seven pieces
- Shuffle the bag
- Consume pieces one by one
- Automatically refill when empty

Do **NOT** use purely random generation.

---

# Rotation System

Implement the **Super Rotation System (SRS)**.

Include:

- Wall kicks
- Floor kicks
- I-piece kick table
- Standard kick table

Support:

- Clockwise rotation
- Counter-clockwise rotation
- Optional 180° rotation

---

# Movement

Support the following controls:

| Action      | Key   |
| ----------- | ----- |
| Move Left   | ←     |
| Move Right  | →     |
| Soft Drop   | ↓     |
| Hard Drop   | Space |
| Rotate CW   | ↑     |
| Rotate CCW  | Z     |
| Rotate 180° | A     |
| Hold Piece  | C     |
| Pause       | P     |
| Restart     | R     |
| Quit        | Q     |

Key bindings should be configurable.

---

# Gravity

Implement gravity according to level.

Example:

- Level 1 = 1 cell/sec
- Higher levels increase falling speed

---

# Lock Delay

Implement proper lock delay.

Include:

- Lock delay timer
- Movement reset
- Rotation reset

Behavior should closely match modern Tetris.

---

# Ghost Piece

Display the landing position of the active piece.

Should update in real time.

---

# Hold Piece

Support holding pieces.

Rules:

- One hold slot
- Can hold only once before locking
- Swapping follows official behavior

---

# Next Queue

Display the next **5 pieces**.

---

# Line Clears

Support:

- Single
- Double
- Triple
- Tetris

Include line clear animation.

---

# T-Spins

Implement proper T-Spin detection.

Support:

- Mini T-Spin
- T-Spin Single
- T-Spin Double
- T-Spin Triple

---

# Combos

Track combo chains.

Display current combo count.

---

# Back-to-Back

Support Back-to-Back bonuses.

Applies to:

- Consecutive Tetrises
- Consecutive T-Spins

---

# Perfect Clear

Detect Perfect Clears.

Award bonus points.

---

# Scoring

Implement official guideline scoring.

Example:

| Action    | Base Score |
| --------- | ---------- |
| Single    | 100        |
| Double    | 300        |
| Triple    | 500        |
| Tetris    | 800        |
| Soft Drop | +1/cell    |
| Hard Drop | +2/cell    |

Also implement:

- Level multiplier
- Combo bonus
- Back-to-Back bonus
- Perfect Clear bonus

---

# Levels

Increase level every **10 cleared lines**.

Each level should:

- Increase gravity
- Increase challenge

---

# Game Over

Occurs when a new piece cannot spawn.

Display:

- Final score
- Level reached
- Total lines
- Duration
- Statistics
- Restart option

---

# Terminal UI

Display:

```
+------------------------+
|        NEXT            |
|                        |
|                        |
|                        |
+------------------------+

+------------------------+
|                        |
|                        |
|      PLAYFIELD         |
|                        |
|                        |
+------------------------+

Score

Level

Lines

Combo

Back-to-Back

Time

Held Piece
```

Requirements:

- No flickering
- Smooth rendering
- Unicode graphics
- Colored blocks
- Pause overlay
- Game Over overlay

---

# Database

Use SQLite.

Database file:

```
tetris.db
```

Automatically create tables if they do not exist.

---

# Database Schema

## players

```sql
id
username
created_at
```

---

## games

```sql
id
player_id
started_at
ended_at
score
level
lines
duration
result
```

---

## highscores

```sql
id
player_id
score
level
lines
date
```

---

## statistics

```sql
id
player_id
games_played
highest_score
highest_level
total_lines
total_play_time
total_tetrises
total_tspins
total_perfect_clears
longest_combo
average_score
average_duration
```

---

# Persistence

At the end of every game automatically save:

- Score
- Level
- Lines
- Duration
- Statistics
- High score if applicable

On startup load:

- Settings
- Player profile
- Statistics
- Leaderboard

---

# Main Menu

```
========================

      CLI TETRIS

========================

1. Play

2. High Scores

3. Statistics

4. Change Player

5. Settings

6. Credits

7. Exit
```

---

# Player Profiles

Support multiple players.

At startup:

```
Enter username:
```

If player does not exist:

- Automatically create profile

---

# High Scores

Display:

| Rank | Player | Score | Level | Lines | Date |
| ---- | ------ | ----- | ----- | ----- | ---- |

Show top 20.

---

# Statistics Screen

Display:

- Highest Score
- Games Played
- Average Score
- Average Duration
- Longest Game
- Highest Level
- Total Lines Cleared
- Total Tetrises
- Total T-Spins
- Perfect Clears
- Longest Combo
- Total Play Time

---

# Settings

Allow users to configure:

- Starting level
- Ghost piece
- Hold system
- 180° rotation
- Theme
- Key bindings
- FPS limit
- Future sound placeholder
- Future music placeholder

---

# Themes

Include:

- Classic
- Neon
- Retro Green
- Monochrome

---

# Configuration

Store configuration in:

```
settings.json
```

---

# Project Structure

```
cmd/

internal/

domain/

application/

game/

renderer/

input/

database/

repository/

service/

config/

assets/

logs/
```

---

# Design Patterns

Use:

- Clean Architecture
- Repository Pattern
- Dependency Injection
- Factory Pattern
- Strategy Pattern (Scoring)
- State Pattern
- Observer Pattern
- Command Pattern

Follow SOLID principles.

---

# Logging

Create:

```
logs/game.log
```

Log:

- Game start
- Game end
- Errors
- Warnings
- Database failures

---

# Testing

Write unit tests for:

- Randomizer
- Collision detection
- Rotations
- Wall kicks
- Scoring
- T-Spins
- Line clearing
- Database persistence
- Statistics

---

# Performance

Requirements:

- 60 FPS rendering
- No flickering
- Efficient redraws
- Minimal allocations
- Low CPU usage

---

# Documentation

Generate a comprehensive `README.md` including:

- Installation
- Build instructions
- Controls
- Architecture
- Database schema
- Project structure
- Screenshots/GIFs
- Development roadmap
- Future improvements

---

# Bonus Features

Implement if time permits:

- Replay recording
- Replay playback
- AI autoplay
- Daily Challenge
- Sprint Mode (40 Lines)
- Marathon Mode
- Ultra Mode (2-Minute Score Attack)
- Zen Mode
- Custom board sizes
- Custom gravity
- Achievements
- Export leaderboard to CSV
- Import/export player profiles
- Save & Resume
- ASCII particle effects
- Additional themes

---

# Code Quality Requirements

The implementation should:

- Be production-ready
- Follow SOLID principles
- Follow Clean Architecture
- Avoid duplicated code
- Keep functions small and focused
- Include proper error handling
- Be fully documented
- Be easy to extend with networking, multiplayer, sound, and graphical rendering in the future

---

# Expected Deliverables

- Complete source code
- SQLite database integration
- Player profile system
- High score system
- Statistics tracking
- Configuration system
- Unit tests
- README.md
- Clean project structure
- Production-quality CLI gameplay
- Fully playable modern Tetris experience
