package browser

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/msherburne/dokbox/internal/domain"
	"github.com/msherburne/dokbox/internal/viewmodel"
)

type focusTarget string

const (
	focusTabs  focusTarget = "tabs"
	focusTable focusTarget = "table"
)

type Model struct {
	tabs       []tab
	activeTab  int
	focus      focusTarget
	containers tableModel
}

func NewModel(containers []domain.ResourceSummary) *Model {
	return &Model{
		tabs: []tab{
			{id: "containers", title: "Containers"},
		},
		focus:      focusTabs,
		containers: newContainerTable(containers),
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}


func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch m.focus {
	case focusTabs:
		if keyMsg.String() == "enter" {
			m.focus = focusTable
		}
	case focusTable:
		switch keyMsg.String() {
		case "up":
			m.containers.MoveUp()
		case "down":
			m.containers.MoveDown()
		case "q":
			m.focus = focusTabs
		}
	}

	return m, nil
}

func (m *Model) View() string {
	lines := []string{
		renderTabs(m.tabs, m.activeTab),
		"",
		m.containers.View(m.focus == focusTable),
		"",
		renderShortcuts(shortcutContext(m.focus)),
	}

	return strings.Join(lines, "\n")
}

func shortcutContext(focus focusTarget) string {
	if focus == focusTable {
		return "containers-table"
	}
	return "resource-tabs"
}

func renderShortcuts(context string) string {
	hints := viewmodel.BrowserShortcuts(context)
	parts := make([]string, 0, len(hints))
	for _, hint := range hints {
		parts = append(parts, hint.Key+" "+hint.Label)
	}
	return strings.Join(parts, " | ")
}

func (m *Model) TableFocused() bool {
	return m.focus == focusTable
}
