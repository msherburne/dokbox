package containerdetail

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/msherburne/dokbox/internal/domain"
)

type tab struct {
	title string
}

type Model struct {
	containerName string
	metrics       []domain.MetricSample
	logs          []domain.LogLine
	tabs          []tab
	activeTab     int
}

func NewModel(containerName string, metrics []domain.MetricSample, logs []domain.LogLine) *Model {
	return &Model{
		containerName: containerName,
		metrics:       append([]domain.MetricSample(nil), metrics...),
		logs:          append([]domain.LogLine(nil), logs...),
		tabs: []tab{
			{title: "Overview"},
			{title: "Logs"},
			{title: "Shell"},
			{title: "Files"},
		},
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

	switch keyMsg.String() {
	case "left":
		if m.activeTab > 0 {
			m.activeTab--
		}
	case "right":
		if m.activeTab < len(m.tabs)-1 {
			m.activeTab++
		}
	}

	return m, nil
}

func (m *Model) View() string {
	lines := []string{
		renderTabs(m.tabs, m.activeTab),
		"",
		"Container: " + m.containerName,
		"",
		m.activeContent(),
	}

	return strings.Join(lines, "\n")
}

func (m *Model) activeContent() string {
	switch m.tabs[m.activeTab].title {
	case "Logs":
		if len(m.logs) == 0 {
			return "No logs."
		}

		lines := make([]string, 0, len(m.logs))
		for _, line := range m.logs {
			lines = append(lines, line.Text)
		}
		return strings.Join(lines, "\n")
	case "Shell":
		return "Shell view coming soon."
	case "Files":
		return "Files view coming soon."
	default:
		if len(m.metrics) == 0 {
			return "No metrics."
		}

		lines := make([]string, 0, len(m.metrics))
		for _, metric := range m.metrics {
			lines = append(lines, metric.Name+": "+metric.Label)
		}
		return strings.Join(lines, "\n")
	}
}

func renderTabs(tabs []tab, active int) string {
	parts := make([]string, 0, len(tabs))
	for index, tab := range tabs {
		if index == active {
			parts = append(parts, "["+tab.title+"]")
			continue
		}
		parts = append(parts, tab.title)
	}

	return strings.Join(parts, " ")
}
