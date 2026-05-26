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
								{Name: ".NET Runtime 10", PackageID: "dotnet-10.0-runtime", CategoryType: "runtime", Description: ".NET 10 runtime."},
								{Name: ".NET Runtime 9", PackageID: "dotnet-9.0-runtime", CategoryType: "runtime", Description: ".NET 9 runtime."},
								{Name: ".NET Runtime 8", PackageID: "dotnet-8.0-runtime", CategoryType: "runtime", Description: ".NET 8 runtime."},
								{Name: ".NET Runtime 7", PackageID: "dotnet-7.0-runtime", CategoryType: "runtime", Description: ".NET 7 runtime."},
								{Name: ".NET Runtime 6", PackageID: "dotnet-6.0-runtime", CategoryType: "runtime", Description: ".NET 6 runtime."},
								{Name: ".NET Runtime 5", PackageID: "dotnet-5.0-runtime", CategoryType: "runtime", Description: ".NET 5 runtime."},
								{Name: ".NET SDK 10", PackageID: "dotnet-10.0-sdk", CategoryType: "sdk", Description: ".NET 10 SDK."},
								{Name: ".NET SDK 9", PackageID: "dotnet-9.0-sdk", CategoryType: "sdk", Description: ".NET 9 SDK."},
								{Name: ".NET SDK 8", PackageID: "dotnet-8.0-sdk", CategoryType: "sdk", Description: ".NET 8 SDK."},
								{Name: ".NET SDK 7", PackageID: "dotnet-7.0-sdk", CategoryType: "sdk", Description: ".NET 7 SDK."},
								{Name: ".NET SDK 6", PackageID: "dotnet-6.0-sdk", CategoryType: "sdk", Description: ".NET 6 SDK."},
								{Name: ".NET SDK 5", PackageID: "dotnet-5.0-sdk", CategoryType: "sdk", Description: ".NET 5 SDK."},
							},
						},
						{
							Name:         "Java",
							CategoryType: "runtime",
							Description:  "Adoptium Java runtimes and development kits.",
							Apps: []Package{
								{Name: "JDK 25 (Adoptium)", PackageID: "temurin25", CategoryType: "runtime", Description: "Adoptium Java Development Kit 25."},
								{Name: "JDK 21 (Adoptium)", PackageID: "temurin21", CategoryType: "runtime", Description: "Adoptium Java Development Kit 21."},
								{Name: "JDK 17 (Adoptium)", PackageID: "temurin17", CategoryType: "runtime", Description: "Adoptium Java Development Kit 17."},
								{Name: "JDK 11 (Adoptium)", PackageID: "temurin11", CategoryType: "runtime", Description: "Adoptium Java Development Kit 11."},
								{Name: "JDK 8 (Adoptium)", PackageID: "temurin8", CategoryType: "runtime", Description: "Adoptium Java Development Kit 8."},
								{Name: "JRE 25 (Adoptium)", PackageID: "temurin25jre", CategoryType: "runtime", Description: "Adoptium Java Runtime Environment 25."},
								{Name: "JRE 21 (Adoptium)", PackageID: "temurin21jre", CategoryType: "runtime", Description: "Adoptium Java Runtime Environment 21."},
								{Name: "JRE 17 (Adoptium)", PackageID: "temurin17jre", CategoryType: "runtime", Description: "Adoptium Java Runtime Environment 17."},
								{Name: "JRE 11 (Adoptium)", PackageID: "temurin11jre", CategoryType: "runtime", Description: "Adoptium Java Runtime Environment 11."},
								{Name: "JRE 8 (Adoptium)", PackageID: "temurin8jre", CategoryType: "runtime", Description: "Adoptium Java Runtime Environment 8."},
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
						{
							Name:         "Visual C++ Redistributables",
							CategoryType: "runtime",
							Description:  "Microsoft Visual C++ runtime packages for legacy and current Windows applications.",
							Apps: []Package{
								{Name: "VC++ Redistributable 2005 x86/x64", PackageID: "vcredist2005", CategoryType: "runtime", Description: "Microsoft Visual C++ 2005 runtime components."},
								{Name: "VC++ Redistributable 2008 x86/x64", PackageID: "vcredist2008", CategoryType: "runtime", Description: "Microsoft Visual C++ 2008 runtime components."},
								{Name: "VC++ Redistributable 2010 x86/x64", PackageID: "vcredist2010", CategoryType: "runtime", Description: "Microsoft Visual C++ 2010 runtime components."},
								{Name: "VC++ Redistributable 2012 x86/x64", PackageID: "vcredist2012", CategoryType: "runtime", Description: "Microsoft Visual C++ 2012 runtime components."},
								{Name: "VC++ Redistributable 2013 x86/x64", PackageID: "vcredist2013", CategoryType: "runtime", Description: "Microsoft Visual C++ 2013 runtime components."},
								{Name: "VC++ Redistributable 2015 x86/x64", PackageID: "vcredist2015", CategoryType: "runtime", Description: "Microsoft Visual C++ 2015 runtime components."},
								{Name: "VC++ Redistributable 2017 x86/x64", PackageID: "vcredist2017", CategoryType: "runtime", Description: "Microsoft Visual C++ 2017 runtime components."},
								{Name: "VC++ Redistributable 2015-2026 x86/x64", PackageID: "vcredist140", CategoryType: "runtime", Description: "Current Microsoft Visual C++ runtime components for Visual Studio 2015 through 2026."},
							},
						},
					},
				},
			},
		},
		{
			Name:         "Media",
			CategoryType: "category",
			Description:  "Media playback, recording, image editing, graphics, screenshots, and disc tools.",
			Categories: []Category{
				{
					Name:         "Playback & Audio",
					CategoryType: "subcategory",
					Description:  "Media players, audio tools, codecs, and streaming apps.",
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
					Name:         "Images & Graphics",
					CategoryType: "subcategory",
					Description:  "Image editors, viewers, screenshots, illustration, and 3D tools.",
					Apps: []Package{
						{Name: "Krita", PackageID: "krita", CategoryType: "graphics", Description: "Digital painting and illustration app."},
						{Name: "Blender", PackageID: "blender", CategoryType: "graphics", Description: "3D creation suite for modeling, rendering, and animation."},
						{Name: "Paint.NET", PackageID: "paint.net", CategoryType: "graphics", Description: "Lightweight image editor for Windows."},
						{Name: "GIMP", PackageID: "gimp", CategoryType: "graphics", Description: "Open source image editor."},
						{Name: "IrfanView", PackageID: "irfanview", CategoryType: "graphics", Description: "Fast image viewer and basic editor."},
						{Name: "XnView", PackageID: "xnview", CategoryType: "graphics", Description: "Image viewer and organizer."},
						{Name: "Inkscape", PackageID: "inkscape", CategoryType: "graphics", Description: "Vector graphics editor."},
						{Name: "FastStone Image Viewer", PackageID: "fsviewer", CategoryType: "graphics", Description: "Image browser, converter, and editor."},
						{Name: "Greenshot", PackageID: "greenshot", CategoryType: "graphics", Description: "Screenshot tool with annotation support."},
						{Name: "Lightshot", PackageID: "lightshot", CategoryType: "graphics", Description: "Simple screenshot capture and sharing tool."},
					},
				},
				{
					Name:         "Disc Tools",
					CategoryType: "subcategory",
					Description:  "CD, DVD, and image burning utilities.",
					Apps: []Package{
						{Name: "ImgBurn", PackageID: "imgburn", CategoryType: "disc", Description: "Disc image and optical media burning tool."},
						{Name: "CDBurnerXP", PackageID: "cdburnerxp", CategoryType: "disc", Description: "CD, DVD, Blu-ray, and ISO burning utility."},
						{Name: "InfraRecorder", PackageID: "infrarecorder", CategoryType: "disc", Description: "Open source CD and DVD burning utility."},
					},
				},
			},
		},
		{
			Name:         "Gaming",
			CategoryType: "category",
			Description:  "Game launchers and gaming communication tools.",
			Apps: []Package{
				{Name: "Steam", PackageID: "steam", CategoryType: "gaming", Description: "Steam game launcher and store."},
				{Name: "Epic Games Launcher", PackageID: "epicgameslauncher", CategoryType: "gaming", Description: "Epic Games Store launcher."},
				{Name: "Discord", PackageID: "discord", CategoryType: "gaming", Description: "Voice and chat app for communities."},
			},
		},
		{
			Name:         "Utilities",
			CategoryType: "category",
			Description:  "Windows utilities for files, remote access, security, productivity, and system maintenance.",
			Categories: []Category{
				{
					Name:         "Remote Access",
					CategoryType: "subcategory",
					Description:  "Remote desktop and VNC tools.",
					Apps: []Package{
						{Name: "AnyDesk", PackageID: "anydesk", CategoryType: "remote", Description: "Remote desktop access tool."},
						{Name: "TeamViewer", PackageID: "teamviewer", CategoryType: "remote", Description: "Remote access and support tool."},
						{Name: "RealVNC Server", PackageID: "vnc-connect", CategoryType: "remote", Description: "RealVNC server for remote desktop access."},
						{Name: "RealVNC Viewer", PackageID: "vnc-viewer", CategoryType: "remote", Description: "RealVNC viewer for connecting to VNC hosts."},
						{Name: "TightVNC", PackageID: "tightvnc", CategoryType: "remote", Description: "VNC remote control software."},
					},
				},
				{
					Name:         "File & System",
					CategoryType: "subcategory",
					Description:  "File copy, cleanup, search, launchers, and system shell utilities.",
					Apps: []Package{
						{Name: "Everything", PackageID: "everything", CategoryType: "utility", Description: "Fast local file search tool."},
						{Name: "TeraCopy", PackageID: "teracopy", CategoryType: "utility", Description: "File copy utility with verification and queueing."},
						{Name: "Revo Uninstaller", PackageID: "revo-uninstaller", CategoryType: "utility", Description: "Application uninstaller and cleanup tool."},
						{Name: "Launchy", PackageID: "launchy", CategoryType: "utility", Description: "Keyboard-driven application launcher."},
						{Name: "WinDirStat", PackageID: "windirstat", CategoryType: "utility", Description: "Disk usage analyzer and cleanup helper."},
						{Name: "WizTree", PackageID: "wiztree", CategoryType: "utility", Description: "Fast disk space analyzer."},
						{Name: "Glary Utilities", PackageID: "glaryutilities-free", CategoryType: "utility", Description: "System cleanup and optimization utility."},
						{Name: "Open-Shell", PackageID: "open-shell", CategoryType: "utility", Description: "Classic Start menu and shell enhancements."},
						{Name: "CCleaner", PackageID: "ccleaner", CategoryType: "utility", Description: "System cleanup utility."},
						{Name: "PowerToys", PackageID: "powertoys", CategoryType: "utility", Description: "Microsoft utilities for Windows power users."},
						{Name: "Google Earth", PackageID: "googleearthpro", CategoryType: "utility", Description: "Desktop globe, maps, and geographic exploration app."},
					},
				},
				{
					Name:         "Archives",
					CategoryType: "subcategory",
					Description:  "Archive managers and compression tools.",
					Apps: []Package{
						{Name: "7-Zip", PackageID: "7zip", CategoryType: "archive", Description: "File archiver with broad format support."},
						{Name: "WinRAR", PackageID: "winrar", CategoryType: "archive", Description: "Archive manager for RAR, ZIP, and other formats."},
						{Name: "PeaZip", PackageID: "peazip", CategoryType: "archive", Description: "Open source archive manager."},
					},
				},
				{
					Name:         "Security & Passwords",
					CategoryType: "subcategory",
					Description:  "Password managers and malware removal tools.",
					Apps: []Package{
						{Name: "Bitwarden", PackageID: "bitwarden", CategoryType: "security", Description: "Password manager desktop app."},
						{Name: "KeePass 2", PackageID: "keepass", CategoryType: "security", Description: "Local password manager."},
						{Name: "Malwarebytes", PackageID: "malwarebytes", CategoryType: "security", Description: "Anti-malware scanning and cleanup tool."},
					},
				},
				{
					Name:         "Network & Transfer",
					CategoryType: "subcategory",
					Description:  "FTP, SSH, torrent, and file transfer clients.",
					Apps: []Package{
						{Name: "FileZilla", PackageID: "filezilla", CategoryType: "network", Description: "FTP, FTPS, and SFTP client."},
						{Name: "WinSCP", PackageID: "winscp", CategoryType: "network", Description: "SFTP, SCP, FTP, and WebDAV file transfer client."},
						{Name: "PuTTY", PackageID: "putty", CategoryType: "network", Description: "SSH and Telnet client."},
						{Name: "qBittorrent", PackageID: "qbittorrent", CategoryType: "network", Description: "BitTorrent client."},
					},
				},
				{
					Name:         "Cloud & Documents",
					CategoryType: "subcategory",
					Description:  "Cloud sync, office suites, notes, and document readers.",
					Apps: []Package{
						{Name: "Dropbox", PackageID: "dropbox", CategoryType: "productivity", Description: "Cloud file sync desktop app."},
						{Name: "Google Drive", PackageID: "googledrive", CategoryType: "productivity", Description: "Google Drive desktop sync app."},
						{Name: "LibreOffice", PackageID: "libreoffice-fresh", CategoryType: "productivity", Description: "Open source office suite."},
						{Name: "OpenOffice", PackageID: "openoffice", CategoryType: "productivity", Description: "Apache OpenOffice productivity suite."},
						{Name: "Foxit Reader", PackageID: "foxitreader", CategoryType: "productivity", Description: "PDF reader."},
						{Name: "Evernote", PackageID: "evernote", CategoryType: "productivity", Description: "Notes and organization app."},
					},
				},
				{
					Name:         "Editors",
					CategoryType: "subcategory",
					Description:  "Text editors and code-oriented desktop apps.",
					Apps: []Package{
						{Name: "Notepad++", PackageID: "notepadplusplus", CategoryType: "editor", Description: "Fast text and source code editor."},
						{Name: "Cursor", PackageID: "cursoride", CategoryType: "editor", Description: "AI-powered code editor based on VS Code."},
					},
				},
			},
		},
	}
}
