package catalog

type App struct {
	Name       string
	ID         string
	Prerelease bool
}

type Category struct {
	Name string
	Apps []App
}

func Default() []Category {
	return []Category{
		{
			Name: "browsers",
			Apps: []App{
				{Name: "Google Chrome", ID: "googlechrome"},
				{Name: "Opera", ID: "opera"},
				{Name: "Opera GX", ID: "opera-gx"},
				{Name: "Mozilla Firefox", ID: "firefox"},
				{Name: "Waterfox", ID: "waterfox"},
				{Name: "Microsoft Edge", ID: "microsoft-edge"},
				{Name: "Brave Browser", ID: "brave"},
				{Name: "Vivaldi", ID: "vivaldi"},
				{Name: "Yandex Browser", ID: "yandex-browser"},
				{Name: "Tor Browser", ID: "tor-browser"},
				{Name: "LibreWolf", ID: "librewolf"},
				{Name: "Zen Browser", ID: "zen-browser", Prerelease: true},
			},
		},
		{
			Name: "dev",
			Apps: []App{
				{Name: "Visual Studio Code", ID: "vscode"},
				{Name: "Git", ID: "git"},
				{Name: "Node.js LTS", ID: "nodejs-lts"},
				{Name: "Python 3", ID: "python"},
			},
		},
		{
			Name: "media",
			Apps: []App{
				{Name: "VLC", ID: "vlc"},
				{Name: "OBS Studio", ID: "obs-studio"},
			},
		},
		{
			Name: "gaming",
			Apps: []App{
				{Name: "Steam", ID: "steam"},
				{Name: "Discord", ID: "discord"},
			},
		},
		{
			Name: "utilities",
			Apps: []App{
				{Name: "7-Zip", ID: "7zip"},
				{Name: "PowerToys", ID: "powertoys"},
				{Name: "Everything", ID: "everything"},
			},
		},
	}
}
