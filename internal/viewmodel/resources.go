package viewmodel

import (
	"fmt"

	"github.com/msherburne/dokbox/internal/domain"
)

var resourceColumns = map[domain.ResourceKind][]string{
	domain.ResourceKindContainer: {
		"Stack",
		"Name",
		"Image",
		"State",
		"Status",
	},
}

func ResourceColumns(kind domain.ResourceKind) []string {
	columns := resourceColumns[kind]
	out := make([]string, len(columns))
	copy(out, columns)
	return out
}

func ResourceRow(summary domain.ResourceSummary) []string {
	if summary.Kind != domain.ResourceKindContainer {
		return []string{summary.Name}
	}

	stack := summary.Group
	if stack == "" {
		stack = "Ungrouped"
	}

	return []string{
		stack,
		summary.Name,
		valueOrDash(summary.Columns, "Image"),
		valueOrDash(summary.Columns, "State"),
		valueOrDash(summary.Columns, "Status"),
	}
}

func FormatRow(columns []string, row []string) string {
	formatted := make([]any, len(row))
	format := ""
	for i, value := range row {
		width := len(columns[i])
		if len(value) > width {
			width = len(value)
		}
		formatted[i] = value
		format += fmt.Sprintf("%%-%ds", width+2)
	}
	return fmt.Sprintf(format, formatted...)
}

func valueOrDash(values map[string]string, key string) string {
	value := values[key]
	if value == "" {
		return "-"
	}
	return value
}
