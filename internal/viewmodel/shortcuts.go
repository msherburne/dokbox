package viewmodel

type ShortcutHint struct {
	Key   string
	Label string
}

func BrowserShortcuts(context string) []ShortcutHint {
	switch context {
	case "containers-table":
		return []ShortcutHint{
			{Key: "Up/Down", Label: "Rows"},
			{Key: "Enter", Label: "Open Details"},
			{Key: "q", Label: "Back To Tabs"},
		}
	case "images-table", "volumes-table", "networks-table":
		return []ShortcutHint{
			{Key: "Up/Down", Label: "Rows"},
			{Key: "Enter", Label: "Open Details"},
			{Key: "q", Label: "Back To Tabs"},
		}
	default:
		return []ShortcutHint{
			{Key: "Left/Right", Label: "Switch Tabs"},
			{Key: "Enter", Label: "Focus Table"},
			{Key: "q", Label: "Quit"},
		}
	}
}
