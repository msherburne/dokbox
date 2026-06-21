package browser

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/msherburne/dokbox/internal/domain"
	"github.com/msherburne/dokbox/internal/viewmodel"
)

type tableModel struct {
	kind    domain.ResourceKind
	columns []string
	rows    []domain.ResourceSummary
	cursor  int
	focused bool
	width   int
}

func newTable(kind domain.ResourceKind, resources []domain.ResourceSummary) tableModel {
	return tableModel{
		kind:    kind,
		columns: viewmodel.ResourceColumns(kind),
		rows:    filterResourcesByKind(resources, kind),
	}
}

func (t *tableModel) MoveDown() {
	if len(t.rows) == 0 {
		return
	}
	if t.cursor < len(t.rows)-1 {
		t.cursor++
	}
}

func (t *tableModel) MoveUp() {
	if len(t.rows) == 0 {
		return
	}
	if t.cursor > 0 {
		t.cursor--
	}
}

func (t *tableModel) Focus() {
	t.focused = true
}

func (t *tableModel) Blur() {
	t.focused = false
}

func (t *tableModel) SetWidth(width int) {
	t.width = width
}

func (t tableModel) View() string {
	if len(t.rows) == 0 {
		return "No resources found."
	}

	widths := t.columnWidths()
	widths = t.fitWidths(widths)
	lines := []string{
		t.renderBorder("top", widths),
		t.renderRow(widths, t.columns, false),
		t.renderBorder("middle", widths),
	}

	for index, resource := range t.rows {
		lines = append(lines, t.renderRow(widths, viewmodel.ResourceRow(resource), t.focused && index == t.cursor))
	}

	lines = append(lines, t.renderBorder("bottom", widths))
	return strings.Join(lines, "\n")
}

func (t tableModel) Selected() *domain.ResourceSummary {
	if len(t.rows) == 0 || t.cursor < 0 || t.cursor >= len(t.rows) {
		return nil
	}

	selected := t.rows[t.cursor]
	return &selected
}

func (t tableModel) SelectionSummary() string {
	selected := t.Selected()
	if selected == nil {
		return "Selected: none"
	}

	return fmt.Sprintf("Selected: %s", selected.Name)
}

func (t tableModel) columnWidths() []int {
	widths := make([]int, len(t.columns))
	for index, column := range t.columns {
		widths[index] = len(column)
	}

	for _, resource := range t.rows {
		values := viewmodel.ResourceRow(resource)
		for index, value := range values {
			if index >= len(widths) {
				continue
			}
			if len(value) > widths[index] {
				widths[index] = len(value)
			}
		}
	}

	return widths
}

func (t tableModel) renderRow(widths []int, values []string, selected bool) string {
	cells := make([]string, 0, len(values))
	for index, value := range values {
		cell := lipgloss.NewStyle().Width(widths[index]).Render(truncateValue(value, widths[index]))
		cells = append(cells, " "+cell+" ")
	}

	row := "│" + strings.Join(cells, "│") + "│"
	if selected {
		return selectedRowStyle.Render(row)
	}
	return row
}

func (t tableModel) renderBorder(position string, widths []int) string {
	var left string
	var middle string
	var right string

	switch position {
	case "top":
		left, middle, right = "┌", "┬", "┐"
	case "middle":
		left, middle, right = "├", "┼", "┤"
	default:
		left, middle, right = "└", "┴", "┘"
	}

	parts := make([]string, 0, len(widths))
	for _, width := range widths {
		parts = append(parts, strings.Repeat("─", width+2))
	}

	return left + strings.Join(parts, middle) + right
}

func filterResourcesByKind(resources []domain.ResourceSummary, kind domain.ResourceKind) []domain.ResourceSummary {
	filtered := make([]domain.ResourceSummary, 0, len(resources))
	for _, resource := range resources {
		if resource.Kind == kind {
			filtered = append(filtered, resource)
		}
	}
	return filtered
}

func (t tableModel) fitWidths(widths []int) []int {
	if t.width <= 0 {
		return widths
	}

	available := t.width - (len(widths)*3 + 1)
	if available <= 0 {
		out := make([]int, len(widths))
		for i := range out {
			out[i] = 1
		}
		return out
	}

	out := append([]int(nil), widths...)
	minWidth := 4
	for total(out) > available {
		shrunk := false
		for i := range out {
			if out[i] > minWidth && total(out) > available {
				out[i]--
				shrunk = true
			}
		}
		if !shrunk {
			break
		}
	}

	return out
}

func total(values []int) int {
	sum := 0
	for _, value := range values {
		sum += value
	}
	return sum
}

func truncateValue(value string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= width {
		return value
	}

	if width == 1 {
		return "…"
	}

	return string(runes[:width-1]) + "…"
}

var selectedRowStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
