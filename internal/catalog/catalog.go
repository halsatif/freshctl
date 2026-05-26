package catalog

type Package struct {
	Name         string
	Description  string
	PackageID    string
	CategoryType string
	Selected     bool
	Prerelease   bool
}

type Category struct {
	Name         string
	Description  string
	CategoryType string
	Categories   []Category
	Apps         []Package
}

func Default() []Category {
	return []Category{
		{
			Name:         "Browsers",
			CategoryType: "category",
			Description:  "Web browsers for everyday browsing, privacy-focused workflows, and alternate browser engines.",
			Apps: []Package{
				{Name: "Google Chrome", PackageID: "googlechrome", CategoryType: "browser", Description: "Google's web browser."},
				{Name: "Opera", PackageID: "opera", CategoryType: "browser", Description: "Opera web browser."},
				{Name: "Opera GX", PackageID: "opera-gx", CategoryType: "browser", Description: "Opera browser tuned for gaming."},
				{Name: "Mozilla Firefox", PackageID: "firefox", CategoryType: "browser", Description: "Mozilla Firefox web browser."},
				{Name: "Waterfox", PackageID: "waterfox", CategoryType: "browser", Description: "Firefox-derived browser focused on customization."},
				{Name: "Microsoft Edge", PackageID: "microsoft-edge", CategoryType: "browser", Description: "Microsoft Edge browser."},
				{Name: "Brave Browser", PackageID: "brave", CategoryType: "browser", Description: "Privacy-focused Chromium browser."},
				{Name: "Vivaldi", PackageID: "vivaldi", CategoryType: "browser", Description: "Highly customizable browser."},
				{Name: "Yandex Browser", PackageID: "yandex-browser", CategoryType: "browser", Description: "Yandex web browser."},
				{Name: "Tor Browser", PackageID: "tor-browser", CategoryType: "browser", Description: "Browser for Tor network access."},
				{Name: "LibreWolf", PackageID: "librewolf", CategoryType: "browser", Description: "Privacy-focused Firefox fork."},
				{Name: "Zen Browser", PackageID: "zen-browser", CategoryType: "browser", Description: "Modern Firefox-based browser.", Prerelease: true},
			},
		},
		{
			Name:         "Development",
			CategoryType: "category",
			Description:  "Programming tools, runtimes, git tools and editors.",
			Categories: []Category{
				{
					Name:         "Editors",
					CategoryType: "subcategory",
					Description:  "Code editors and development workspaces.",
					Apps: []Package{
						{Name: "Visual Studio Code", PackageID: "vscode", CategoryType: "editor", Description: "Code editor with extensions and integrated tools."},
					},
				},
				{
					Name:         "Version Control",
					CategoryType: "subcategory",
					Description:  "Tools for tracking source code changes.",
					Apps: []Package{
						{Name: "Git", PackageID: "git", CategoryType: "version-control", Description: "Distributed version control system."},
					},
				},
				{
					Name:         "Runtimes",
					CategoryType: "subcategory",
					Description:  "Language runtimes and SDKs for development.",
					Categories: []Category{
						{
							Name:         ".NET",
							CategoryType: "runtime",
							Description:  ".NET runtimes and SDKs for building and running applications.",
							Apps: []Package{
								{Name: ".NET Runtime 8", PackageID: "dotnet-8.0-runtime", CategoryType: "runtime", Description: ".NET 8 runtime."},
								{Name: ".NET Runtime 9", PackageID: "dotnet-9.0-runtime", CategoryType: "runtime", Description: ".NET 9 runtime."},
								{Name: ".NET SDK 8", PackageID: "dotnet-8.0-sdk", CategoryType: "sdk", Description: ".NET 8 SDK."},
								{Name: ".NET SDK 9", PackageID: "dotnet-9.0-sdk", CategoryType: "sdk", Description: ".NET 9 SDK."},
							},
						},
						{
							Name:         "Java",
							CategoryType: "runtime",
							Description:  "Adoptium Java runtimes and development kits.",
							Apps: []Package{
								{Name: "JDK 21 (Adoptium)", PackageID: "temurin21", CategoryType: "runtime", Description: "Adoptium Java Development Kit 21."},
								{Name: "JDK 17 (Adoptium)", PackageID: "temurin17", CategoryType: "runtime", Description: "Adoptium Java Development Kit 17."},
								{Name: "JRE 21 (Adoptium)", PackageID: "temurin21jre", CategoryType: "runtime", Description: "Adoptium Java Runtime Environment 21."},
							},
						},
						{
							Name:         "Node.js",
							CategoryType: "runtime",
							Description:  "Node.js runtime and package tooling.",
							Apps: []Package{
								{Name: "Node.js LTS", PackageID: "nodejs-lts", CategoryType: "runtime", Description: "Node.js LTS runtime and npm."},
							},
						},
						{
							Name:         "Python",
							CategoryType: "runtime",
							Description:  "Python runtime and package tools.",
							Apps: []Package{
								{Name: "Python 3", PackageID: "python", CategoryType: "runtime", Description: "Python runtime and package tools."},
							},
						},
					},
				},
			},
		},
		{
			Name:         "Media",
			CategoryType: "category",
			Description:  "Media playback, recording, audio editing, codecs, and streaming tools.",
			Apps: []Package{
				{Name: "iTunes", PackageID: "itunes", CategoryType: "media", Description: "Apple media library, playback, and device sync app."},
				{Name: "VLC", PackageID: "vlc", CategoryType: "media", Description: "Media player for video and audio."},
				{Name: "AIMP", PackageID: "aimp", CategoryType: "media", Description: "Lightweight audio player with playlist and format support."},
				{Name: "foobar2000", PackageID: "foobar2000", CategoryType: "media", Description: "Advanced audio player with a compact, customizable interface."},
				{Name: "Winamp", PackageID: "winamp", CategoryType: "media", Description: "Classic music player for local audio collections."},
				{Name: "MusicBee", PackageID: "musicbee", CategoryType: "media", Description: "Music manager and player for large local libraries."},
				{Name: "Audacious", PackageID: "audacious", CategoryType: "media", Description: "Lightweight audio player focused on sound quality and codec support."},
				{Name: "Audacity", PackageID: "audacity", CategoryType: "media", Description: "Audio recording and editing tool."},
				{Name: "K-Lite Codecs", PackageID: "k-litecodecpackfull", CategoryType: "media", Description: "Codec pack for broad video and audio playback support."},
				{Name: "GOM", PackageID: "gom-player", CategoryType: "media", Description: "Video player with wide codec and subtitle support."},
				{Name: "mpv", PackageID: "mpvio", CategoryType: "media", Description: "Minimal, keyboard-driven media player for audio and video playback."},
				{Name: "Spotify", PackageID: "spotify", CategoryType: "media", Description: "Music streaming desktop app."},
				{Name: "OBS Studio", PackageID: "obs-studio", CategoryType: "media", Description: "Recording and streaming studio."},
			},
		},
		{
			Name:         "Gaming",
			CategoryType: "category",
			Description:  "Game launchers and gaming communication tools.",
			Apps: []Package{
				{Name: "Steam", PackageID: "steam", CategoryType: "gaming", Description: "Steam game launcher and store."},
				{Name: "Discord", PackageID: "discord", CategoryType: "gaming", Description: "Voice and chat app for communities."},
			},
		},
		{
			Name:         "Utilities",
			CategoryType: "category",
			Description:  "Small Windows utilities for files, search, and system productivity.",
			Apps: []Package{
				{Name: "7-Zip", PackageID: "7zip", CategoryType: "utility", Description: "File archiver with broad format support."},
				{Name: "PowerToys", PackageID: "powertoys", CategoryType: "utility", Description: "Microsoft utilities for Windows power users."},
				{Name: "Everything", PackageID: "everything", CategoryType: "utility", Description: "Fast local file search tool."},
			},
		},
	}
}
