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
	tabs      []tab
	activeTab int
	focus     focusTarget
	tables    map[domain.ResourceKind]*tableModel
}

func NewModel(resources []domain.ResourceSummary) *Model {
	tables := map[domain.ResourceKind]*tableModel{}
	for _, kind := range []domain.ResourceKind{
		domain.ResourceKindContainer,
		domain.ResourceKindImage,
		domain.ResourceKindVolume,
		domain.ResourceKindNetwork,
	} {
		table := newTable(kind, resources)
		tables[kind] = &table
	}

	return &Model{
		tabs: []tab{
			{kind: string(domain.ResourceKindContainer), title: "Containers"},
			{kind: string(domain.ResourceKindImage), title: "Images"},
			{kind: string(domain.ResourceKindVolume), title: "Volumes"},
			{kind: string(domain.ResourceKindNetwork), title: "Networks"},
		},
		focus:  focusTabs,
		tables: tables,
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
		switch keyMsg.String() {
		case "enter":
			m.focus = focusTable
		case "left":
			if m.activeTab > 0 {
				m.activeTab--
			}
		case "right":
			if m.activeTab < len(m.tabs)-1 {
				m.activeTab++
			}
		}
	case focusTable:
		table := m.currentTable()
		switch keyMsg.String() {
		case "up":
			table.MoveUp()
		case "down":
			table.MoveDown()
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
		m.currentTable().View(m.focus == focusTable),
		"",
		renderShortcuts(m.shortcutContext()),
	}

	return strings.Join(lines, "\n")
}

func (m *Model) shortcutContext() string {
	if m.focus == focusTable {
		switch m.currentTabKind() {
		case string(domain.ResourceKindImage):
			return "images-table"
		case string(domain.ResourceKindVolume):
			return "volumes-table"
		case string(domain.ResourceKindNetwork):
			return "networks-table"
		default:
			return "containers-table"
		}
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

func (m *Model) currentTable() *tableModel {
	return m.tables[domain.ResourceKind(m.currentTabKind())]
}

func (m *Model) currentTabKind() string {
	return m.tabs[m.activeTab].kind
}
