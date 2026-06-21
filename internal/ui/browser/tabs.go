package browser

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type tab struct {
	kind  string
	title string
}

func renderTabs(tabs []tab, active int) string {
	parts := make([]string, 0, len(tabs))
	for index, tab := range tabs {
		if index == active {
			parts = append(parts, activeTabStyle.Render("["+tab.title+"]"))
			continue
		}
		parts = append(parts, inactiveTabStyle.Render(tab.title))
	}
	return strings.Join(parts, " ")
}

var (
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("12"))
	inactiveTabStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("8"))
)
