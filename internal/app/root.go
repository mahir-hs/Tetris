package app

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mahir/tetris/internal/config"
	"github.com/mahir/tetris/internal/database"
	"github.com/mahir/tetris/internal/game"
	"github.com/mahir/tetris/internal/input"
	"github.com/mahir/tetris/internal/log"
	"github.com/mahir/tetris/internal/renderer"
	"github.com/mahir/tetris/internal/service"
)

type screen int

const (
	screenMenu screen = iota
	screenEnterName
	screenGame
	screenHighScores
	screenStats
	screenSettings
	screenCredits
)

var menuOptions = []string{
	"Play", "High Scores", "Statistics", "Change Player",
	"Settings", "Credits", "Exit",
}

// RootModel is the top-level Bubble Tea model orchestrating all screens.
type RootModel struct {
	config    config.Settings
	theme     renderer.Theme
	bindings  *input.Bindings
	playerSvc service.PlayerService
	lbSvc     service.LeaderboardService
	statsSvc  service.StatsService

	player *database.Player
	stats  *database.Statistics
	scores []database.HighScoreRow

	scr       screen
	menuIndex int
	engine    *game.Engine
	nameInput textinput.Model
	namePrompt string // prompt shown on the name-entry screen
	pending   string // action to run after name entry: play|stats|change
	winW      int    // terminal width (from WindowSizeMsg)
	winH      int    // terminal height (from WindowSizeMsg)
}

// NewRootModel constructs the application root.
func NewRootModel(cfg config.Settings, playerSvc service.PlayerService, lbSvc service.LeaderboardService, statsSvc service.StatsService) RootModel {
	ti := textinput.New()
	ti.CharLimit = 16
	ti.Placeholder = "username"
	return RootModel{
		config:    cfg,
		theme:     renderer.GetTheme(cfg.Theme),
		bindings:  input.NewBindings(cfg),
		playerSvc: playerSvc,
		lbSvc:     lbSvc,
		statsSvc:  statsSvc,
		scr:       screenMenu,
		nameInput: ti,
	}
}

// Init satisfies tea.Model.
func (m RootModel) Init() tea.Cmd { return nil }

// Update routes messages by screen.
func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.winW = msg.Width
		m.winH = msg.Height
		return m, nil

	case game.TickMsg:
		if m.scr == screenGame && m.engine != nil {
			model, cmd := m.engine.Update(msg)
			m.engine = model.(*game.Engine)
			return m, cmd
		}
		return m, nil

	case game.QuitToMenuMsg:
		m.scr = screenMenu
		m.engine = nil
		return m, nil

	case game.GameOverMsg:
		m.persist(game.GameResult(msg))
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m RootModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.scr {
	case screenMenu:
		return m.handleMenuKey(msg)
	case screenEnterName:
		return m.handleNameKey(msg)
	case screenGame:
		if m.engine != nil {
			model, cmd := m.engine.Update(msg)
			m.engine = model.(*game.Engine)
			return m, cmd
		}
		return m, nil
	case screenHighScores, screenStats, screenCredits:
		if msg.String() == "esc" {
			m.scr = screenMenu
		}
		return m, nil
	case screenSettings:
		return m.handleSettingsKey(msg)
	}
	return m, nil
}

func (m RootModel) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.menuIndex > 0 {
			m.menuIndex--
		}
	case "down", "j":
		if m.menuIndex < len(menuOptions)-1 {
			m.menuIndex++
		}
	case "enter":
		return m.selectMenu()
	case "esc":
		// no-op on main menu
	}
	return m, nil
}

func (m RootModel) selectMenu() (tea.Model, tea.Cmd) {
	switch menuOptions[m.menuIndex] {
	case "Play":
		if m.player == nil {
			return m.beginNameEntry("play", "Enter username to play:")
		}
		return m.startGame()
	case "High Scores":
		rows, err := m.lbSvc.Top(20)
		if err == nil {
			m.scores = rows
		}
		m.scr = screenHighScores
	case "Statistics":
		if m.player == nil {
			return m.beginNameEntry("stats", "Enter username for statistics:")
		}
		return m.loadStats()
	case "Change Player":
		return m.beginNameEntry("change", "Enter username:")
	case "Settings":
		m.scr = screenSettings
	case "Credits":
		m.scr = screenCredits
	case "Exit":
		return m, tea.Quit
	}
	return m, nil
}

func (m RootModel) beginNameEntry(pending, prompt string) (tea.Model, tea.Cmd) {
	m.pending = pending
	m.nameInput.SetValue("")
	m.nameInput.Focus()
	m.scr = screenEnterName
	m.namePrompt = prompt
	return m, textinput.Blink
}

func (m RootModel) handleNameKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.nameInput.Blur()
		m.scr = screenMenu
		return m, nil
	case "enter":
		name := m.nameInput.Value()
		if name == "" {
			return m, nil
		}
		p, err := m.playerSvc.GetOrCreate(name)
		if err != nil {
			return m, nil
		}
		m.player = p
		m.nameInput.Blur()
		switch m.pending {
		case "play":
			return m.startGame()
		case "stats":
			return m.loadStats()
		default:
			m.scr = screenMenu
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

func (m RootModel) loadStats() (tea.Model, tea.Cmd) {
	if m.player == nil {
		m.scr = screenMenu
		return m, nil
	}
	st, err := m.playerSvc.Stats(m.player.ID)
	if err == nil {
		m.stats = st
	}
	m.scr = screenStats
	return m, nil
}

func (m RootModel) startGame() (tea.Model, tea.Cmd) {
	if m.player == nil {
		m.scr = screenMenu
		return m, nil
	}
	m.engine = game.NewEngine(m.player.ID, m.config, m.theme, m.bindings)
	m.scr = screenGame
	log.Infof("game start player=%d", m.player.ID)
	return m, m.engine.Init()
}

func (m RootModel) handleSettingsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.scr = screenMenu
		return m, nil
	case "1", "2", "3", "4":
		idx := int(msg.Runes[0]-'1')
		if idx < len(renderer.ThemeNames) {
			m.config.Theme = renderer.ThemeNames[idx]
			m.theme = renderer.GetTheme(m.config.Theme)
			saveConfig(m.config)
		}
	case "g":
		m.config.GhostEnabled = !m.config.GhostEnabled
		saveConfig(m.config)
	case "h":
		m.config.HoldEnabled = !m.config.HoldEnabled
		saveConfig(m.config)
	case "r":
		m.config.Rotate180Enabled = !m.config.Rotate180Enabled
		saveConfig(m.config)
	}
	return m, nil
}

func (m RootModel) persist(res game.GameResult) {
	outcome := service.GameOutcome{
		PlayerID:      res.PlayerID,
		Score:         res.Score,
		Level:         res.Level,
		Lines:         res.Lines,
		DurationSec:   int(res.DurationMs / 1000),
		PiecesPlaced:  res.PiecesPlaced,
		Tetrises:      res.Tetrises,
		TSpins:        res.TSpins,
		MiniTSpins:    res.MiniTSpins,
		PerfectClears: res.PerfectClears,
		LongestCombo:  res.LongestCombo,
	}
	if err := m.statsSvc.RecordGame(outcome); err == nil {
		if m.player != nil {
			if st, e := m.playerSvc.Stats(m.player.ID); e == nil {
				m.stats = st
			}
		}
		log.Infof("game over player=%d score=%d lines=%d level=%d", res.PlayerID, res.Score, res.Lines, res.Level)
	} else {
		log.Errorf("persist game: %v", err)
	}
}

// View renders the current screen, centered within the terminal.
func (m RootModel) View() string {
	var v string
	switch m.scr {
	case screenMenu:
		v = renderer.MainMenu(menuOptions, m.menuIndex, m.theme)
	case screenEnterName:
		v = renderer.UsernamePrompt(m.nameInput.View(), m.namePrompt, m.theme)
	case screenGame:
		if m.engine != nil {
			v = m.engine.View()
		} else {
			v = renderer.MainMenu(menuOptions, m.menuIndex, m.theme)
		}
	case screenHighScores:
		v = renderer.HighScores(m.scores, m.theme)
	case screenStats:
		name := "unknown"
		if m.player != nil {
			name = m.player.Username
		}
		st := database.Statistics{}
		if m.stats != nil {
			st = *m.stats
		}
		v = renderer.Statistics(name, st, m.theme)
	case screenSettings:
		v = renderer.SettingsView(m.config, m.theme)
	case screenCredits:
		v = renderer.Credits(m.theme)
	}
	return m.place(v)
}

// place centers the content within the terminal when its size is known.
func (m RootModel) place(v string) string {
	if m.winW <= 0 || m.winH <= 0 {
		return v
	}
	return lipgloss.Place(m.winW, m.winH, lipgloss.Center, lipgloss.Center, v)
}

// saveConfig persists settings to disk (errors are non-fatal).
func saveConfig(cfg config.Settings) {
	_ = config.Save(cfg, "")
}
