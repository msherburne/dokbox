package app

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/msherburne/dokbox/internal/theme"
)

type Styles struct {
	theme   theme.Theme
	Shell   lipgloss.Style
	Success lipgloss.Style
	Warning lipgloss.Style
	Error   lipgloss.Style
	Info    lipgloss.Style
	Muted   lipgloss.Style
}

func NewStyles(active theme.Theme) Styles {
	panel := active.Panel
	if panel == "" {
		panel = active.Background
	}

	border := active.Border
	if border == "" {
		border = active.Text
	}

	return Styles{
		theme: active,
		Shell: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(border)).
			Padding(1, 1).
			Foreground(lipgloss.Color(active.Text)).
			Background(lipgloss.Color(panel)),
		Success: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(active.Success)),
		Warning: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(active.Warning)),
		Error: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(active.Error)),
		Info: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(active.Info)),
		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color(active.Muted)),
	}
}

func (s Styles) Theme() theme.Theme {
	if s.theme.Name != "" {
		return s.theme
	}

	activeTheme, _ := theme.Resolve(theme.DefaultPreset)
	return activeTheme
}

func defaultStyles() Styles {
	activeTheme, _ := theme.Resolve(theme.DefaultPreset)
	return NewStyles(activeTheme)
}
