package theme

func Resolve(name string) (Theme, string) {
	preset, ok := presets[name]
	if ok {
		return preset, name
	}

	fallback := presets[DefaultPreset]
	return fallback, DefaultPreset
}
