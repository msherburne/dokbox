package browser

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/msherburne/dokbox/internal/domain"
	"github.com/msherburne/dokbox/internal/viewmodel"
)

type focusTarget string

const (
	focusTabs  focusTarget = "tabs"
	focusTable focusTarget = "table"
)

var (
	browserPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("8")).
				Padding(0, 1)
	browserTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15"))
	browserMetaStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8"))
	browserShortcutStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8"))
	browserMenuStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("8")).
				Padding(0, 1)
)

type Model struct {
	tabs                []tab
	activeTab           int
	focus               focusTarget
	tables              map[domain.ResourceKind]*tableModel
	actionsOpen         bool
	confirmationPending *pendingConfirmation
	width               int
	height              int
}

type pendingConfirmation struct {
	action       string
	resourceID   string
	resourceName string
	kind         domain.ResourceKind
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

func (m *Model) SetWidth(width int) {
	m.width = width
	for _, table := range m.tables {
		table.SetWidth(width - browserPanelStyle.GetHorizontalBorderSize())
	}
}

func (m *Model) SetHeight(height int) {
	m.height = height
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
			if selected == nil {
				return m, nil
			}

			action := actionNameForKey(keyMsg.String())
			if action == "remove" || action == "prune" {
				m.actionsOpen = false
				m.confirmationPending = &pendingConfirmation{
					action:       action,
					resourceID:   selected.ID,
					resourceName: selected.Name,
					kind:         selected.Kind,
				}
				return m, nil
			}

			m.actionsOpen = false

			return m, func() tea.Msg {
				return ActionRequest{
					Action:       action,
					ResourceID:   selected.ID,
					ResourceName: selected.Name,
					Kind:         selected.Kind,
				}
			}
		default:
			return m, nil
		}
	}

	if m.confirmationPending != nil {
		switch keyMsg.String() {
		case "y":
			pending := *m.confirmationPending
			m.confirmationPending = nil
			m.actionsOpen = false
			return m, func() tea.Msg {
				return ActionRequest{
					Action:       pending.action,
					ResourceID:   pending.resourceID,
					ResourceName: pending.resourceName,
					Kind:         pending.kind,
				}
			}
		case "n", "q", "esc":
			m.confirmationPending = nil
			m.actionsOpen = true
			return m, nil
		default:
			return m, nil
		}
	}

	switch m.focus {
	case focusTabs:
		switch keyMsg.String() {
		case "enter":
			m.focus = focusTable
			m.currentTable().Focus()
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
			table.Blur()
		}
	}

	return m, nil
}

func (m *Model) View() string {
	table := m.currentTable()
	lines := []string{
		browserTitleStyle.Render("Resources"),
		renderTabs(m.tabs, m.activeTab),
		"",
		table.View(),
	}

	if table.Selected() != nil {
		lines = append(lines, "", browserMetaStyle.Render(table.SelectionSummary()))
	}

	if m.actionsOpen {
		lines = append(lines, "", m.renderActionsMenu())
	}

	if m.confirmationPending != nil {
		lines = append(lines, "", m.renderConfirmation())
	}

	lines = append(lines, "", renderShortcuts(m.shortcutContext()))
	content := strings.Join(lines, "\n")

	if m.height > 0 {
		maxHeight := m.height - browserPanelStyle.GetVerticalBorderSize()
		content = clipLines(content, maxHeight, browserMetaStyle.Render("More rows below"))
	}

	panelStyle := browserPanelStyle
	if m.width > 0 {
		width := m.width - panelStyle.GetHorizontalBorderSize()
		if width < 0 {
			width = 0
		}
		panelStyle = panelStyle.Width(width)
	}

	return panelStyle.Render(content)
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
	return browserShortcutStyle.Render(strings.Join(parts, " | "))
}

func clipLines(content string, maxHeight int, indicator string) string {
	if maxHeight <= 0 {
		return indicator
	}

	lines := strings.Split(content, "\n")
	if len(lines) <= maxHeight {
		return content
	}
	if maxHeight == 1 {
		return indicator
	}

	clipped := append([]string{}, lines[:maxHeight-1]...)
	clipped = append(clipped, indicator)
	return strings.Join(clipped, "\n")
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

	return browserMenuStyle.Render(strings.Join(lines, "\n"))
}

func (m *Model) renderConfirmation() string {
	if m.confirmationPending == nil {
		return ""
	}

	target := m.confirmationPending.resourceName
	if m.confirmationPending.action == "prune" {
		target = string(m.confirmationPending.kind) + "s"
	}

	lines := []string{
		"Confirm " + m.confirmationPending.action + " " + target + "?",
		"y Confirm",
		"n Cancel",
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
