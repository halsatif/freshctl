you are working on freshctl.

freshctl is a windows-first go tui app for installing selected apps via winget and other utils.

rules:
- build the mvp only
- do not redesign the product
- do not add backend/server/auth/cloud
- do not add linux/mac support
- do not add package managers besides winget
- do not add telemetry
- do not add update system
- do not create fake placeholder features
- do not use huge rounded card ui / ai-looking design
- keep the tui minimal, fast, and terminal-native
- prefer boring reliable code over clever abstractions

implementation goals:
- use go
- use bubble tea
- use lipgloss
- hardcode the initial app catalog
- create 4 screens:
  1. welcome
  2. catalog
  3. review
  4. install log
- selected apps must persist across category navigation
- never install anything before explicit confirmation on review screen
- install apps one by one through winget
- continue if one install fails
- show final summary
- handle missing winget gracefully

keyboard:
- welcome:
  - enter: continue
  - q / ctrl+c: quit
- catalog:
  - up/down or k/j: move
  - tab: switch focus between categories and app list
  - space: toggle selected app
  - enter: review
  - q / ctrl+c: quit
- review:
  - enter: install
  - b / esc: back
  - q / ctrl+c: quit
- install:
  - q / ctrl+c: quit after/while logs are visible

catalog:
browsers:
- google chrome: Google.Chrome
- mozilla firefox: Mozilla.Firefox
- brave browser: Brave.Brave

dev:
- visual studio code: Microsoft.VisualStudioCode
- git: Git.Git
- node.js lts: OpenJS.NodeJS.LTS
- python 3: Python.Python.3.12

media:
- vlc: VideoLAN.VLC
- obs studio: OBSProject.OBSStudio

gaming:
- steam: Valve.Steam
- discord: Discord.Discord

utilities:
- 7-zip: 7zip.7zip
- powertoys: Microsoft.PowerToys
- everything: voidtools.Everything

expected project structure:
- main.go
- internal/catalog/catalog.go
- internal/installer/winget.go
- internal/tui/model.go
- internal/tui/screens.go
- internal/tui/styles.go
- README.md

quality checks:
- run gofmt
- run go test ./...
- run go build -o freshctl.exe .
- fix all build errors before finishing

README:
- explain what freshctl is
- say it is windows-first
- say it uses winget and installs real apps
- include build/run commands