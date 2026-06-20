package theme

const DefaultPreset = "default"

var presetList = []Theme{
	{
		Name:       DefaultPreset,
		Background: "#101418",
		Panel:      "#171d24",
		Border:     "#2c3947",
		Text:       "#e7edf4",
		Muted:      "#93a4b5",
		Emphasis:   "#ffffff",
		Accent:     "#5cc8ff",
		Focus:      "#9bdcff",
		Success:    "#56d39b",
		Warning:    "#ffbf69",
		Error:      "#ff6b6b",
		Info:       "#73c7ff",
	},
	{
		Name:       "slate",
		Background: "#0f1720",
		Panel:      "#18212c",
		Border:     "#304255",
		Text:       "#dce6f2",
		Muted:      "#8ea1b5",
		Emphasis:   "#f8fbff",
		Accent:     "#7dd3fc",
		Focus:      "#bae6fd",
		Success:    "#6ee7b7",
		Warning:    "#fbbf24",
		Error:      "#f87171",
		Info:       "#93c5fd",
	},
	{
		Name:       "ember",
		Background: "#16100f",
		Panel:      "#211715",
		Border:     "#4a312b",
		Text:       "#f3e7e4",
		Muted:      "#b89f99",
		Emphasis:   "#fff7f2",
		Accent:     "#ff8a5b",
		Focus:      "#ffc2a8",
		Success:    "#7ddf9b",
		Warning:    "#ffcf6e",
		Error:      "#ff7a7a",
		Info:       "#ffab91",
	},
}

var presets = buildPresetIndex(presetList)

func PresetNames() []string {
	names := make([]string, len(presetList))
	for i, preset := range presetList {
		names[i] = preset.Name
	}

	return names
}

func buildPresetIndex(list []Theme) map[string]Theme {
	index := make(map[string]Theme, len(list))
	for _, preset := range list {
		index[preset.Name] = preset
	}

	return index
}
