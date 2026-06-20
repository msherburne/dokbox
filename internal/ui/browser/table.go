package browser

import (
	"strings"

	"github.com/msherburne/dokbox/internal/domain"
	"github.com/msherburne/dokbox/internal/viewmodel"
)

type tableModel struct {
	columns []string
	rows    []domain.ResourceSummary
	cursor  int
}

func newContainerTable(resources []domain.ResourceSummary) tableModel {
	return tableModel{
		columns: viewmodel.ResourceColumns(domain.ResourceKindContainer),
		rows:    resources,
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
		return "No containers found."
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
