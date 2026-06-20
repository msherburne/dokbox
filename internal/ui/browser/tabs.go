package browser

import "strings"

type tab struct {
	kind  string
	title string
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
