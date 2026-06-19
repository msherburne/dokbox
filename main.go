package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/msherburne/dokbox/internal/app"
)

func main() {
	program := tea.NewProgram(app.NewModel(app.Dependencies{}), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}
