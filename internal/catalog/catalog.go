package catalog

type PackageType string

const (
	PackageTypeApplication PackageType = "Application"
	PackageTypeCLITool     PackageType = "CLITool"
	PackageTypeRuntime     PackageType = "Runtime"
)

type DetectMethod string

const (
	DetectNone     DetectMethod = ""
	DetectRegistry DetectMethod = "Registry"
	DetectPath     DetectMethod = "Path"
)

type Application struct {
	ID           string
	Name         string
	Description  string
	Category     string
	Type         PackageType
	Icon         string
	DetectMethod DetectMethod
	DetectValue  string
	Verified     bool
	CategoryType string
	Selected     bool
	Providers    []Provider
}

type Category struct {
	Name         string
	Description  string
	CategoryType string
	Categories   []Category
	Apps         []Application
}

func app(name, id, category, description string) Application {
	return packageWithType(name, id, category, description, PackageTypeApplication)
}

func cli(name, id, category, description string) Application {
	return packageWithType(name, id, category, description, PackageTypeCLITool)
}

func runtimePackage(name, id, category, description string) Application {
	return packageWithType(name, id, category, description, PackageTypeRuntime)
}

func prerelease(app Application) Application {
	if len(app.Providers) > 0 {
		app.Providers[0].Metadata.Prerelease = true
	}
	return app
}

func detect(app Application, method DetectMethod, value string) Application {
	app.DetectMethod = method
	app.DetectValue = value
	return app
}

func withDirectProvider(app Application, metadata DirectInstallerMetadata) Application {
	app.Providers = []Provider{
		directProvider(app.ID, metadata),
		chocolateyProvider(app.ID),
	}
	return app
}

func packageWithType(name, id, category, description string, packageType PackageType) Application {
	return Application{
		ID:           id,
		Name:         name,
		Category:     category,
		CategoryType: category,
		Description:  description,
		Type:         packageType,
		Verified:     true,
		Providers:    []Provider{chocolateyProvider(id)},
	}
}

func Default() []Category {
	categories := []Category{
		{
			Name:         "Browsers",
			CategoryType: "category",
			Description:  "Web browsers for everyday browsing, privacy-focused workflows, and alternate browser engines.",
			Apps: []Application{
				withDirectProvider(
					detect(app("Google Chrome", "googlechrome", "browser", "Google's web browser."), DetectRegistry, "Google Chrome"),
					DirectInstallerMetadata{
						Downloads: []DirectDownload{
							{URL: "https://dl.google.com/dl/chrome/install/googlechromestandaloneenterprise64.msi", Architecture: InstallerArchitectureX64},
						},
						Filename:      "GoogleChrome.msi",
						SilentArgs:    []string{"/qn", "/norestart"},
						InstallerType: InstallerTypeMSI,
					},
				),
				app("Opera", "opera", "browser", "Opera web browser."),
				app("Opera GX", "opera-gx", "browser", "Opera browser tuned for gaming."),
				withDirectProvider(
					detect(app("Mozilla Firefox", "firefox", "browser", "Mozilla Firefox web browser."), DetectRegistry, "Mozilla Firefox"),
					DirectInstallerMetadata{
						Downloads: []DirectDownload{
							{URL: "https://download.mozilla.org/?product=firefox-latest-ssl&os=win64&lang=en-US", Architecture: InstallerArchitectureX64},
							{URL: "https://download.mozilla.org/?product=firefox-latest-ssl&os=win64-aarch64&lang=en-US", Architecture: InstallerArchitectureARM64},
						},
						Filename:      "FirefoxSetup.exe",
						SilentArgs:    []string{"/S"},
						InstallerType: InstallerTypeExecutable,
					},
				),
				app("Waterfox", "waterfox", "browser", "Firefox-derived browser focused on customization."),
				detect(app("Microsoft Edge", "microsoft-edge", "browser", "Microsoft Edge browser."), DetectRegistry, "Microsoft Edge"),
				detect(app("Brave Browser", "brave", "browser", "Privacy-focused Chromium browser."), DetectRegistry, "Brave"),
				app("Vivaldi", "vivaldi", "browser", "Highly customizable browser."),
				app("Tor Browser", "tor-browser", "browser", "Browser for Tor network access."),
				prerelease(detect(app("Zen Browser", "zen-browser", "browser", "Modern Firefox-based browser."), DetectRegistry, "Zen Browser")),
			},
		},
		{
			Name:         "Communication",
			CategoryType: "category",
			Description:  "Messaging, voice chat, video meetings, and team communication apps.",
			Apps: []Application{
				withDirectProvider(
					detect(app("Telegram Desktop", "telegram", "communication", "Telegram desktop messenger."), DetectRegistry, "Telegram Desktop"),
					DirectInstallerMetadata{
						Downloads: []DirectDownload{
							{URL: "https://telegram.org/dl/desktop/win64", Architecture: InstallerArchitectureX64},
						},
						Filename:      "TelegramSetup.exe",
						SilentArgs:    []string{"/VERYSILENT", "/NORESTART"},
						InstallerType: InstallerTypeExecutable,
					},
				),
				detect(app("Signal", "signal", "communication", "Private messaging desktop app."), DetectRegistry, "Signal"),
				app("Element", "element-desktop", "communication", "Matrix-based secure chat client."),
				app("Zoom", "zoom", "communication", "Video meetings and conferencing app."),
				app("Dorion", "dorion", "communication", "Lightweight alternative Discord client."),
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
					Apps: []Application{
						withDirectProvider(
							detect(app("Visual Studio Code", "vscode", "editor", "Code editor with extensions and integrated tools."), DetectRegistry, "Visual Studio Code"),
							DirectInstallerMetadata{
								Downloads: []DirectDownload{
									{URL: "https://update.code.visualstudio.com/latest/win32-x64/stable", Architecture: InstallerArchitectureX64},
									{URL: "https://update.code.visualstudio.com/latest/win32-arm64/stable", Architecture: InstallerArchitectureARM64},
								},
								Filename:      "VSCodeSetup.exe",
								SilentArgs:    []string{"/VERYSILENT", "/NORESTART", "/MERGETASKS=!runcode"},
								InstallerType: InstallerTypeExecutable,
							},
						),
						app("Zed", "zed-editor", "editor", "Fast collaborative code editor."),
						app("Sublime Text", "sublimetext4", "editor", "Fast text and code editor."),
						detect(cli("Neovim", "neovim", "editor", "Terminal-based code editor. Run with nvim."), DetectPath, "nvim.exe"),
						detect(cli("Helix", "helix", "editor", "Terminal-based code editor. Run with hx."), DetectPath, "hx.exe"),
						app("JetBrains Toolbox", "jetbrainstoolbox", "editor", "JetBrains IDE manager."),
						app("IntelliJ IDEA Community", "intellijidea-community", "editor", "JetBrains Java and JVM IDE."),
						app("PyCharm Community", "pycharm-community", "editor", "JetBrains Python IDE."),
						app("Android Studio", "androidstudio", "editor", "Android app development IDE."),
					},
				},
				{
					Name:         "Version Control",
					CategoryType: "subcategory",
					Description:  "Tools for tracking source code changes.",
					Apps: []Application{
						detect(cli("Git", "git", "version-control", "Command-line version control tool. Run with git."), DetectPath, "git.exe"),
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
							Apps: []Application{
								runtimePackage(".NET Runtime 10", "dotnet-10.0-runtime", "runtime", ".NET 10 runtime required by some Windows apps."),
								runtimePackage(".NET Runtime 9", "dotnet-9.0-runtime", "runtime", ".NET 9 runtime required by some Windows apps."),
								runtimePackage(".NET Runtime 8", "dotnet-8.0-runtime", "runtime", ".NET 8 runtime required by many Windows apps."),
								runtimePackage(".NET Runtime 7", "dotnet-7.0-runtime", "runtime", ".NET 7 runtime required by some Windows apps."),
								runtimePackage(".NET Runtime 6", "dotnet-6.0-runtime", "runtime", ".NET 6 runtime required by some Windows apps."),
								runtimePackage(".NET Runtime 5", "dotnet-5.0-runtime", "runtime", ".NET 5 runtime required by older Windows apps."),
								runtimePackage(".NET SDK 10", "dotnet-10.0-sdk", "sdk", ".NET 10 tools for building .NET apps."),
								runtimePackage(".NET SDK 9", "dotnet-9.0-sdk", "sdk", ".NET 9 tools for building .NET apps."),
								runtimePackage(".NET SDK 8", "dotnet-8.0-sdk", "sdk", ".NET 8 tools for building .NET apps."),
								runtimePackage(".NET SDK 7", "dotnet-7.0-sdk", "sdk", ".NET 7 tools for building .NET apps."),
								runtimePackage(".NET SDK 6", "dotnet-6.0-sdk", "sdk", ".NET 6 tools for building .NET apps."),
								runtimePackage(".NET SDK 5", "dotnet-5.0-sdk", "sdk", ".NET 5 tools for building older .NET apps."),
							},
						},
						{
							Name:         "Java",
							CategoryType: "runtime",
							Description:  "Adoptium Java runtimes and development kits.",
							Apps: []Application{
								runtimePackage("JDK 25 (Adoptium)", "temurin25", "runtime", "Adoptium Java Development Kit 25."),
								runtimePackage("JDK 21 (Adoptium)", "temurin21", "runtime", "Adoptium Java Development Kit 21."),
								runtimePackage("JDK 17 (Adoptium)", "temurin17", "runtime", "Adoptium Java Development Kit 17."),
								runtimePackage("JDK 11 (Adoptium)", "temurin11", "runtime", "Adoptium Java Development Kit 11."),
								runtimePackage("JDK 8 (Adoptium)", "temurin8", "runtime", "Adoptium Java Development Kit 8."),
								runtimePackage("JRE 25 (Adoptium)", "temurin25jre", "runtime", "Adoptium Java Runtime Environment 25."),
								runtimePackage("JRE 21 (Adoptium)", "temurin21jre", "runtime", "Adoptium Java Runtime Environment 21."),
								runtimePackage("JRE 17 (Adoptium)", "temurin17jre", "runtime", "Adoptium Java Runtime Environment 17."),
								runtimePackage("JRE 11 (Adoptium)", "temurin11jre", "runtime", "Adoptium Java Runtime Environment 11."),
								runtimePackage("JRE 8 (Adoptium)", "temurin8jre", "runtime", "Adoptium Java Runtime Environment 8."),
							},
						},
						{
							Name:         "Node.js",
							CategoryType: "runtime",
							Description:  "Node.js runtime and package tooling.",
							Apps: []Application{
								detect(runtimePackage("Node.js LTS", "nodejs-lts", "runtime", "Node.js LTS runtime and npm."), DetectPath, "node.exe"),
							},
						},
						{
							Name:         "Python",
							CategoryType: "runtime",
							Description:  "Python runtime and package tools.",
							Apps: []Application{
								detect(runtimePackage("Python 3", "python", "runtime", "Python runtime and package tools."), DetectPath, "python.exe"),
							},
						},
						{
							Name:         "Toolchains",
							CategoryType: "runtime",
							Description:  "Compiler toolchains and build systems.",
							Apps: []Application{
								detect(cli("Go", "golang", "toolchain", "Go command-line toolchain. Run with go."), DetectPath, "go.exe"),
								detect(cli("Rustup", "rustup.install", "toolchain", "Rust command-line toolchain installer. Run with rustup."), DetectPath, "rustup.exe"),
								detect(cli("LLVM", "llvm", "toolchain", "Compiler toolchain for C, C++, and LLVM-based tools."), DetectPath, "clang.exe"),
								cli("MinGW", "mingw", "toolchain", "GNU compiler toolchain for Windows."),
								detect(cli("CMake", "cmake", "toolchain", "Command-line build configuration tool. Run with cmake."), DetectPath, "cmake.exe"),
							},
						},
						{
							Name:         "Visual C++ Redistributables",
							CategoryType: "runtime",
							Description:  "Microsoft Visual C++ runtime packages for legacy and current Windows applications.",
							Apps: []Application{
								runtimePackage("VC++ Redist 2010 x86/x64", "vcredist2010", "runtime", "Microsoft runtime required by older Windows apps and games."),
								runtimePackage("VC++ Redist 2012 x86/x64", "vcredist2012", "runtime", "Microsoft runtime required by older Windows apps and games."),
								runtimePackage("VC++ Redist 2013 x86/x64", "vcredist2013", "runtime", "Microsoft runtime required by older Windows apps and games."),
								detect(runtimePackage("VC++ Redist 2015-2022 x86/x64", "vcredist140", "runtime", "Microsoft runtime required by many Windows apps and games."), DetectRegistry, "Microsoft Visual C++"),
							},
						},
					},
				},
				{
					Name:         "Terminals & CLI",
					CategoryType: "subcategory",
					Description:  "Terminal emulators, shells, and command-line utilities.",
					Apps: []Application{
						detect(app("Windows Terminal", "microsoft-windows-terminal", "terminal", "Microsoft terminal app for shells and command-line tools."), DetectRegistry, "Windows Terminal"),
						detect(cli("PowerShell 7", "powershell-core", "terminal", "Command-line shell and scripting environment. Run with pwsh."), DetectPath, "pwsh.exe"),
						detect(app("WezTerm", "wezterm", "terminal", "GPU-accelerated terminal emulator."), DetectRegistry, "WezTerm"),
						detect(cli("Fastfetch", "fastfetch", "terminal", "Command-line system information tool. Run with fastfetch."), DetectPath, "fastfetch.exe"),
						detect(cli("FZF", "fzf", "terminal", "Command-line fuzzy finder. Run with fzf."), DetectPath, "fzf.exe"),
						detect(cli("ripgrep", "ripgrep", "terminal", "Command-line search tool. Run with rg."), DetectPath, "rg.exe"),
					},
				},
				{
					Name:         "Containers",
					CategoryType: "subcategory",
					Description:  "Container runtimes and desktop container managers.",
					Apps: []Application{
						app("Podman Desktop", "podman-desktop", "container", "Desktop manager for Podman and containers."),
					},
				},
				{
					Name:         "API & Databases",
					CategoryType: "subcategory",
					Description:  "API clients, database clients, and local database servers.",
					Apps: []Application{
						app("Postman", "postman", "database", "API development and testing client."),
						app("Bruno", "bruno", "database", "Git-friendly API client."),
						app("Insomnia", "insomnia-rest-api-client", "database", "REST, GraphQL, and API client."),
						app("DBeaver", "dbeaver", "database", "Universal database client."),
						app("PostgreSQL", "postgresql", "database", "PostgreSQL database server."),
						app("MySQL", "mysql", "database", "MySQL database server."),
						app("MongoDB Compass", "mongodb-compass", "database", "MongoDB desktop GUI client."),
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
					Apps: []Application{
						app("iTunes", "itunes", "media", "Apple media library, playback, and device sync app."),
						detect(app("VLC", "vlc", "media", "Media player for video and audio."), DetectRegistry, "VLC media player"),
						app("foobar2000", "foobar2000", "media", "Advanced audio player with a compact, customizable interface."),
						app("Winamp", "winamp", "media", "Classic music player for local audio collections."),
						app("MusicBee", "musicbee", "media", "Music manager and player for large local libraries."),
						app("Audacity", "audacity", "media", "Audio recording and editing tool."),
						app("K-Lite Codecs", "k-litecodecpackfull", "media", "Codec pack for broad video and audio playback support."),
						app("mpv", "mpvio", "media", "Minimal, keyboard-driven media player for audio and video playback."),
						detect(app("Spotify", "spotify", "media", "Music streaming desktop app."), DetectRegistry, "Spotify"),
						detect(app("OBS Studio", "obs-studio", "media", "Recording and streaming studio."), DetectRegistry, "OBS Studio"),
						app("Kdenlive", "kdenlive", "media", "Open source video editor."),
						detect(cli("yt-dlp", "yt-dlp", "media", "Command-line video downloader. Run with yt-dlp."), DetectPath, "yt-dlp.exe"),
						detect(cli("FFmpeg", "ffmpeg", "media", "Command-line audio and video toolkit. Run with ffmpeg."), DetectPath, "ffmpeg.exe"),
					},
				},
				{
					Name:         "Images & Graphics",
					CategoryType: "subcategory",
					Description:  "Image editors, viewers, screenshots, illustration, and 3D tools.",
					Apps: []Application{
						app("Krita", "krita", "graphics", "Digital painting and illustration app."),
						app("Blender", "blender", "graphics", "3D creation suite for modeling, rendering, and animation."),
						app("Paint.NET", "paint.net", "graphics", "Lightweight image editor for Windows."),
						app("GIMP", "gimp", "graphics", "Open source image editor."),
						app("IrfanView", "irfanview", "graphics", "Fast image viewer and basic editor."),
						app("Inkscape", "inkscape", "graphics", "Vector graphics editor."),
						app("FastStone Image Viewer", "fsviewer", "graphics", "Image browser, converter, and editor."),
						app("Greenshot", "greenshot", "graphics", "Screenshot tool with annotation support."),
						app("Lightshot", "lightshot", "graphics", "Simple screenshot capture and sharing tool."),
						app("ImageGlass", "imageglass", "graphics", "Modern image viewer."),
						detect(app("ShareX", "sharex", "graphics", "Screenshot, screen capture, and sharing tool."), DetectRegistry, "ShareX"),
						app("ScreenToGif", "screentogif", "graphics", "Screen, webcam, and sketchboard recorder."),
						app("Flameshot", "flameshot", "graphics", "Screenshot tool with annotation features."),
					},
				},
				{
					Name:         "Disc Tools",
					CategoryType: "subcategory",
					Description:  "CD, DVD, and image burning utilities.",
					Apps: []Application{
						app("CDBurnerXP", "cdburnerxp", "disc", "CD, DVD, Blu-ray, and ISO burning utility."),
						app("InfraRecorder", "infrarecorder", "disc", "Open source CD and DVD burning utility."),
					},
				},
			},
		},
		{
			Name:         "Gaming",
			CategoryType: "category",
			Description:  "Game launchers and gaming communication tools.",
			Apps: []Application{
				detect(app("Steam", "steam", "gaming", "Steam game launcher and store."), DetectRegistry, "Steam"),
				detect(app("Heroic Games Launcher", "heroic-games-launcher", "gaming", "Open source launcher for Epic, GOG, and Amazon games."), DetectRegistry, "Heroic"),
				detect(app("Prism Launcher", "prismlauncher", "gaming", "Minecraft launcher for multiple instances and modded setups."), DetectRegistry, "Prism Launcher"),
				detect(app("Discord", "discord", "gaming", "Voice and chat app for communities."), DetectRegistry, "Discord"),
				app("Moonlight", "moonlight", "gaming", "GameStream client for remote gaming."),
				app("Sunshine", "sunshine", "gaming", "Self-hosted game streaming host."),
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
					Apps: []Application{
						app("AnyDesk", "anydesk", "remote", "Remote desktop access tool."),
						app("TeamViewer", "teamviewer", "remote", "Remote access and support tool."),
						app("RealVNC Server", "vnc-connect", "remote", "RealVNC server for remote desktop access."),
						app("RealVNC Viewer", "vnc-viewer", "remote", "RealVNC viewer for connecting to VNC hosts."),
						app("TightVNC", "tightvnc", "remote", "VNC remote control software."),
						app("RustDesk", "rustdesk", "remote", "Open source remote desktop tool."),
						app("Barrier", "barrier", "remote", "Keyboard and mouse sharing across computers."),
						cli("scrcpy", "scrcpy", "remote", "Android screen mirroring and control tool."),
						detect(cli("ADB Platform Tools", "adb", "remote", "Android command-line debugging tools. Run with adb."), DetectPath, "adb.exe"),
					},
				},
				{
					Name:         "File & System",
					CategoryType: "subcategory",
					Description:  "File copy, cleanup, search, launchers, and system shell utilities.",
					Apps: []Application{
						detect(app("Everything", "everything", "utility", "Fast local file search tool."), DetectRegistry, "Everything"),
						detect(app("WinDirStat", "windirstat", "utility", "Disk usage analyzer and cleanup helper."), DetectRegistry, "WinDirStat"),
						detect(app("WizTree", "wiztree", "utility", "Fast disk space analyzer."), DetectRegistry, "WizTree"),
						app("Glary Utilities Free", "glaryutilities-free", "utility", "System cleanup and optimization utility."),
						app("Open-Shell", "open-shell", "utility", "Classic Start menu and shell enhancements."),
						detect(app("PowerToys", "powertoys", "utility", "Microsoft utilities for Windows power users."), DetectRegistry, "PowerToys"),
						app("Google Earth", "googleearthpro", "utility", "Desktop globe, maps, and geographic exploration app."),
						app("AutoHotkey", "autohotkey", "utility", "Automation and hotkey scripting tool."),
						app("Ventoy", "ventoy", "utility", "Multiboot USB drive creator."),
						app("Bulk Crap Uninstaller", "bulk-crap-uninstaller", "utility", "Bulk application uninstaller and cleanup tool."),
						app("HWiNFO64", "hwinfo", "utility", "Hardware information and monitoring tool."),
						app("HWMonitor", "hwmonitor", "utility", "Hardware sensor monitoring tool."),
						app("GPU-Z", "gpu-z", "utility", "Graphics card information utility."),
						app("System Informer", "systeminformer", "utility", "Modern process, service, and system monitor."),
						app("Process Explorer", "procexp", "utility", "Sysinternals process inspection tool."),
						app("EarTrumpet", "eartrumpet", "utility", "Per-app volume control for Windows."),
						app("StartAllBack", "startallback", "utility", "Windows taskbar, Start menu, and Explorer customization."),
						app("Twinkle Tray", "twinkle-tray", "utility", "External monitor brightness control."),
						app("UniGetUI", "unigetui", "utility", "GUI for package managers including Chocolatey and winget."),
					},
				},
				{
					Name:         "Archives",
					CategoryType: "subcategory",
					Description:  "Archive managers and compression tools.",
					Apps: []Application{
						detect(app("7-Zip", "7zip", "archive", "File archiver with broad format support."), DetectRegistry, "7-Zip"),
						app("WinRAR", "winrar", "archive", "Archive manager for RAR, ZIP, and other formats."),
						app("PeaZip", "peazip", "archive", "Open source archive manager."),
					},
				},
				{
					Name:         "Security & Passwords",
					CategoryType: "subcategory",
					Description:  "Password managers and malware removal tools.",
					Apps: []Application{
						withDirectProvider(
							detect(app("Bitwarden", "bitwarden", "security", "Password manager desktop app."), DetectRegistry, "Bitwarden"),
							DirectInstallerMetadata{
								Downloads: []DirectDownload{
									{URL: "https://vault.bitwarden.com/download/?app=desktop&platform=windows", Architecture: InstallerArchitectureAny},
								},
								Filename:      "BitwardenInstaller.exe",
								SilentArgs:    []string{"/allusers", "/S"},
								InstallerType: InstallerTypeExecutable,
							},
						),
						app("KeePass 2", "keepass", "security", "Local password manager."),
						app("BleachBit", "bleachbit", "security", "Privacy-focused cleanup utility."),
						app("SimpleWall", "simplewall", "security", "Windows Filtering Platform firewall control tool."),
					},
				},
				{
					Name:         "Network & Transfer",
					CategoryType: "subcategory",
					Description:  "FTP, SSH, torrent, and file transfer clients.",
					Apps: []Application{
						app("FileZilla", "filezilla", "network", "FTP, FTPS, and SFTP client."),
						app("WinSCP", "winscp", "network", "SFTP, SCP, FTP, and WebDAV file transfer client."),
						app("PuTTY", "putty", "network", "SSH and Telnet client."),
						app("qBittorrent", "qbittorrent", "network", "BitTorrent client."),
						withDirectProvider(
							detect(app("Tailscale", "tailscale", "network", "Mesh VPN client."), DetectRegistry, "Tailscale"),
							DirectInstallerMetadata{
								Downloads: []DirectDownload{
									{URL: "https://pkgs.tailscale.com/stable/tailscale-setup-latest-amd64.msi", Architecture: InstallerArchitectureX64},
									// Tailscale explicitly directs Windows ARM64 users to its x86 MSI.
									{URL: "https://pkgs.tailscale.com/stable/tailscale-setup-latest-x86.msi", Architecture: InstallerArchitectureARM64},
								},
								Filename:      "Tailscale.msi",
								SilentArgs:    []string{"/qn", "/norestart"},
								InstallerType: InstallerTypeMSI,
							},
						),
						app("WireGuard", "wireguard", "network", "WireGuard VPN client."),
						app("ZeroTier", "zerotier-one", "network", "Virtual networking and mesh VPN client."),
						app("Wireshark", "wireshark", "network", "Network protocol analyzer."),
						cli("Nmap", "nmap", "network", "Network discovery and security scanner."),
						app("LocalSend", "localsend", "network", "Local network file sharing app."),
					},
				},
				{
					Name:         "Cloud & Documents",
					CategoryType: "subcategory",
					Description:  "Cloud sync, office suites, notes, and document readers.",
					Apps: []Application{
						app("Dropbox", "dropbox", "productivity", "Cloud file sync desktop app."),
						app("Google Drive", "googledrive", "productivity", "Google Drive desktop sync app."),
						app("LibreOffice", "libreoffice-fresh", "productivity", "Open source office suite."),
						app("OpenOffice", "openoffice", "productivity", "Apache OpenOffice productivity suite."),
						app("Evernote", "evernote", "productivity", "Notes and organization app."),
						app("OnlyOffice Community", "onlyoffice", "productivity", "Office suite for documents, spreadsheets, and presentations."),
						app("SumatraPDF", "sumatrapdf", "productivity", "Lightweight PDF and ebook reader."),
						app("Claude Desktop", "claude", "productivity", "Anthropic Claude desktop app."),
					},
				},
				{
					Name:         "Editors",
					CategoryType: "subcategory",
					Description:  "Text editors and code-oriented desktop apps.",
					Apps: []Application{
						detect(app("Notepad++", "notepadplusplus", "editor", "Fast text and source code editor."), DetectRegistry, "Notepad++"),
						app("Cursor", "cursoride", "editor", "AI-powered code editor based on VS Code."),
						app("WinMerge", "winmerge", "editor", "File and folder comparison tool."),
					},
				},
				{
					Name:         "Imaging & Virtualization",
					CategoryType: "subcategory",
					Description:  "USB imaging, virtual machines, and installer utilities.",
					Apps: []Application{
						app("balenaEtcher", "etcher", "virtualization", "Bootable USB and SD card image writer."),
						app("VirtualBox", "virtualbox", "virtualization", "Virtual machine platform."),
					},
				},
			},
		},
	}

	normalizeCategories(categories)
	return categories
}

func normalizeCategories(categories []Category) {
	for categoryIndex := range categories {
		category := &categories[categoryIndex]
		normalizeCategories(category.Categories)
		for appIndex := range category.Apps {
			app := &category.Apps[appIndex]
			if app.Category == "" {
				app.Category = app.CategoryType
			}
			app.Verified = true
		}
	}
}
