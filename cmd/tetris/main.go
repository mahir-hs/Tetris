package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mahir/tetris/internal/app"
	"github.com/mahir/tetris/internal/config"
	"github.com/mahir/tetris/internal/database"
	"github.com/mahir/tetris/internal/log"
	"github.com/mahir/tetris/internal/repository"
	"github.com/mahir/tetris/internal/service"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		log.Errorf("load config: %v", err)
		fmt.Fprintf(os.Stderr, "failed to load settings: %v\n", err)
		os.Exit(1)
	}

	db, err := database.InitDB("")
	if err != nil {
		log.Errorf("init db: %v", err)
		fmt.Fprintf(os.Stderr, "failed to open database: %v\n", err)
		os.Exit(1)
	}

	playerRepo := repository.NewPlayerRepo(db)
	gameRepo := repository.NewGameRepo(db)
	hsRepo := repository.NewHighScoreRepo(db)
	statsRepo := repository.NewStatsRepo(db)

	playerSvc := service.NewPlayerService(playerRepo, statsRepo)
	lbSvc := service.NewLeaderboardService(hsRepo)
	statsSvc := service.NewStatsService(gameRepo, hsRepo, statsRepo)

	root := app.NewRootModel(cfg, playerSvc, lbSvc, statsSvc)
	log.Infof("application started")

	p := tea.NewProgram(root, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Errorf("run: %v", err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
