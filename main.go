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

	client, err := docker.NewClient(cfg.DockerHost)
	if err == nil {
		defer client.Close()

		service := docker.NewService(client)
		deps.ConnectionChecker = service
		deps.ActionRunner = service
		deps.LogProvider = service
		deps.MetricsProvider = service
		deps.ShellProvider = service
		deps.FileProvider = service

		resources, loadErr := service.ListResourceSummaries()
		if loadErr == nil {
			deps.InitialContainers = resources
		}
	}

	program := tea.NewProgram(
		app.NewModel(deps),
		tea.WithAltScreen(),
	)
	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}
