package app

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/msherburne/dokbox/internal/theme"
)

type Styles struct {
	Shell lipgloss.Style
}

func NewStyles(active theme.Theme) Styles {
	return Styles{
		Shell: lipgloss.NewStyle().
			Padding(1, 2).
			Foreground(lipgloss.Color(active.Text)).
			Background(lipgloss.Color(active.Background)),
	}
}

func defaultStyles() Styles {
	activeTheme, _ := theme.Resolve(theme.DefaultPreset)
	return NewStyles(activeTheme)
}
