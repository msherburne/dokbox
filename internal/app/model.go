package app

import tea "github.com/charmbracelet/bubbletea"

type Dependencies struct{}

type startupCompleteMsg struct{}

type Model struct {
	ready bool
}

func NewModel(_ Dependencies) *Model {
	return &Model{}
}

func (m *Model) Init() tea.Cmd {
	return func() tea.Msg {
		return startupCompleteMsg{}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case startupCompleteMsg:
		m.ready = true
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", QuitKey:
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *Model) View() string {
	if !m.ready {
		return ShellStyle.Render("dokbox-go\n\nBootstrapping shell...")
	}

	return ShellStyle.Render("dokbox-go\n\nStarting rewrite shell...\n\nPress q to quit.")
}
