<h1 align="center">
  <img src="./site/favicon.png" width="35">
  freshctl
</h1>

<p align="center">
	<span style="font-size: 30px; font-weight: 300;">
  windows bootstrap utility
  </span>
</p>

<p align="center">
	<span style="font-size: 22px; font-weight: 300;">
  install apps from a clean terminal interface.
  </span>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/platform-windows-0078D6">
  <img src="https://img.shields.io/badge/built%20with-go-00ADD8">
  <img src="https://img.shields.io/badge/license-MIT-purple">
</p>

## Install

PowerShell:

```powershell
irm https://freshctl.tech/install.ps1 | iex
```

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
