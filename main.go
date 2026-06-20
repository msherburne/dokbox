package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/msherburne/dokbox/internal/app"
	"github.com/msherburne/dokbox/internal/config"
	"github.com/msherburne/dokbox/internal/docker"
)

func main() {
	cfg, err := config.LoadOrSetupConfig(os.Stdin, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	deps := app.Dependencies{
		ConnectionChecker: docker.NewStatusChecker(cfg.DockerHost),
	}

	program := tea.NewProgram(
		app.NewModel(deps),
		tea.WithAltScreen(),
	)
	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}
