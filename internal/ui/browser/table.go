package browser

import (
	"strings"

	"github.com/msherburne/dokbox/internal/domain"
	"github.com/msherburne/dokbox/internal/viewmodel"
)

type tableModel struct {
	kind    domain.ResourceKind
	columns []string
	rows    []domain.ResourceSummary
	cursor  int
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

func (t tableModel) View(focused bool) string {
	if len(t.rows) == 0 {
		return "No resources found."
	}

	lines := []string{
		viewmodel.FormatRow(t.columns, t.columns),
	}

	for index, resource := range t.rows {
		prefix := "  "
		if focused && index == t.cursor {
			prefix = "> "
		}

		lines = append(lines, prefix+strings.TrimRight(
			viewmodel.FormatRow(t.columns, viewmodel.ResourceRow(resource)),
			" ",
		))
	}

	return strings.Join(lines, "\n")
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
