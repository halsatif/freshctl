package presets

type Preset struct {
	ID          string
	Name        string
	Description string
	Packages    []string
}

func Default() []Preset {
	return []Preset{
		{
			ID:          "developer",
			Name:        "Developer",
			Description: "Common tools for coding on a fresh Windows install.",
			Packages: []string{
				"vscode",
				"git",
				"powershell-core",
				"microsoft-windows-terminal",
				"nodejs-lts",
				"python",
				"golang",
			},
		},
		{
			ID:          "gaming",
			Name:        "Gaming",
			Description: "Launchers and communication tools for gaming setups.",
			Packages: []string{
				"steam",
				"discord",
				"obs-studio",
			},
		},
		{
			ID:          "streaming",
			Name:        "Streaming",
			Description: "Recording, streaming, chat, and media playback tools.",
			Packages: []string{
				"obs-studio",
				"discord",
				"vlc",
			},
		},
		{
			ID:          "minimal",
			Name:        "Minimal",
			Description: "Small starter set for a clean everyday Windows setup.",
			Packages: []string{
				"firefox",
				"7zip",
				"everything",
			},
		},
		{
			ID:          "privacy",
			Name:        "Privacy",
			Description: "Browser, password manager, and private messaging basics.",
			Packages: []string{
				"firefox",
				"bitwarden",
				"signal",
			},
		},
	}
}

func ByID(id string) (Preset, bool) {
	for _, preset := range Default() {
		if preset.ID == id {
			return preset, true
		}
	}
	return Preset{}, false
}
