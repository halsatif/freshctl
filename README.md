# freshctl

![windows](https://img.shields.io/badge/platform-windows-0A84FF?style=flat-square)
![go](https://img.shields.io/badge/built%20with-go-00ADD8?style=flat-square&logo=go)
![license](https://img.shields.io/badge/license-MIT-8b5cf6?style=flat-square)

windows bootstrap utility

Install apps from a clean terminal interface.

---

## Screenshot

![freshctl screenshot](./assets/screenshot.png)

---

## Features

- full catalog search
- category browser
- chocolatey bootstrap
- package selection
- terminal-native tui
- fast setup for fresh Windows installs
- clean minimal interface

---

## Install

PowerShell:

```powershell
irm https://freshctl.tech/install.ps1 | iex
```

---

## Build

```powershell
go mod tidy
go build -o freshctl.exe .
```

---

## Run

```powershell
.\freshctl.exe
```

---

## Requirements

- Windows 10/11
- PowerShell
- Administrator privileges may be required
- Internet connection

---

## Package Source

Currently supported:

- Chocolatey

---

## Included Packages

freshctl currently includes packages for:

- browsers
- development tools
- runtimes
- terminals
- media tools
- gaming utilities
- networking
- virtualization
- productivity
- privacy & security

Examples:

- Google Chrome
- Firefox
- VSCode
- Git
- Docker Desktop
- Python
- Node.js
- OBS Studio
- Discord
- Steam
- Tailscale
- PowerToys
- ShareX
- qBittorrent
- VirtualBox

---

## Roadmap

Planned features:

- presets
- package metadata
- custom installers
- auto-update
- package filtering
- improved install progress ui

---

## License

MIT License

See [LICENSE](./LICENSE).
