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
	domain.ResourceKindImage: {
		"Repository",
		"Tag",
		"Image ID",
		"Size",
		"Created",
	},
	domain.ResourceKindVolume: {
		"Name",
		"Driver",
		"Scope",
		"Mountpoint",
		"Created",
	},
	domain.ResourceKindNetwork: {
		"Name",
		"Driver",
		"Scope",
		"Flags",
		"Containers",
	},
}

func ResourceColumns(kind domain.ResourceKind) []string {
	columns := resourceColumns[kind]
	out := make([]string, len(columns))
	copy(out, columns)
	return out
}

func ResourceRow(summary domain.ResourceSummary) []string {
	if summary.Kind == domain.ResourceKindContainer {
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

	row := make([]string, 0, len(ResourceColumns(summary.Kind)))
	for _, column := range ResourceColumns(summary.Kind) {
		if column == "Name" {
			row = append(row, summary.Name)
			continue
		}
		if column == "Repository" {
			if value := valueOrDash(summary.Columns, "Repository"); value != "-" {
				row = append(row, value)
				continue
			}
		}
		row = append(row, valueOrDash(summary.Columns, column))
	}
	return row
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
