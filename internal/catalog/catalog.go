package catalog

type App struct {
	Name string
	ID   string
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
				{Name: "Mozilla Firefox", ID: "firefox"},
				{Name: "Brave Browser", ID: "brave"},
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
				{Name: "7-zip", ID: "7zip"},
				{Name: "PowerToys", ID: "powertoys"},
				{Name: "Everything", ID: "everything"},
			},
		},
	}
}
