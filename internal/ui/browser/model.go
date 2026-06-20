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
	actionsOpen bool
}

type ActionRequest struct {
	Action       string
	ResourceID   string
	ResourceName string
	Kind         domain.ResourceKind
}

type OpenContainerDetailRequest struct {
	ResourceID   string
	ResourceName string
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

	if m.actionsOpen {
		switch keyMsg.String() {
		case "q", "esc":
			m.actionsOpen = false
			return m, nil
		case "s", "t", "r", "x", "p":
			selected := m.currentTable().Selected()
			m.actionsOpen = false
			if selected == nil {
				return m, nil
			}

			return m, func() tea.Msg {
				return ActionRequest{
					Action:       actionNameForKey(keyMsg.String()),
					ResourceID:   selected.ID,
					ResourceName: selected.Name,
					Kind:         selected.Kind,
				}
			}
		default:
			return m, nil
		}
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
		case "enter":
			if m.currentTabKind() == string(domain.ResourceKindContainer) {
				selected := table.Selected()
				if selected != nil {
					return m, func() tea.Msg {
						return OpenContainerDetailRequest{
							ResourceID:   selected.ID,
							ResourceName: selected.Name,
						}
					}
				}
			}
		case "up":
			table.MoveUp()
		case "down":
			table.MoveDown()
		case "o":
			if m.currentTabKind() == string(domain.ResourceKindContainer) && table.Selected() != nil {
				m.actionsOpen = true
			}
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
	}

	if m.actionsOpen {
		lines = append(lines, "", m.renderActionsMenu())
	}

	lines = append(lines, "", renderShortcuts(m.shortcutContext()))

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

func (m *Model) renderActionsMenu() string {
	selected := m.currentTable().Selected()
	if selected == nil {
		return ""
	}

	lines := []string{
		"Actions: " + selected.Name,
		"s Start",
		"t Stop",
		"r Restart",
		"x Remove",
		"p Prune",
		"q Cancel",
	}

	return strings.Join(lines, "\n")
}

func actionNameForKey(key string) string {
	switch key {
	case "s":
		return "start"
	case "t":
		return "stop"
	case "r":
		return "restart"
	case "x":
		return "remove"
	case "p":
		return "prune"
	default:
		return ""
	}
}
