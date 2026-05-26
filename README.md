# freshctl

freshctl is a Windows-first terminal installer for setting up a fresh machine without clicking through a pile of installers.

It shows a small app catalog, lets you choose what to install, then runs real Chocolatey installs after you confirm the review screen. It is an MVP CLI/TUI version of a Ninite-like workflow.

## Requirements

- Windows
- Go
- Chocolatey

Administrator privileges are required for Chocolatey bootstrap and package installation. If freshctl needs elevation, it relaunches through Windows UAC.

## Build

```powershell
go mod tidy
go build -o freshctl.exe .
```

## Run

```powershell
.\freshctl.exe
```

## Chocolatey Bootstrap

freshctl currently uses Chocolatey as its primary package manager. On startup, freshctl checks whether `choco` is available.

If Chocolatey is missing, freshctl shows a bootstrap screen. Pressing `enter` runs Chocolatey's official PowerShell bootstrap command from:

```text
https://community.chocolatey.org/install.ps1
```

Bootstrap output is shown in the TUI. When bootstrap finishes, freshctl automatically checks for `choco` again. If bootstrap fails, run freshctl as Administrator and try again, or install Chocolatey manually from the official Chocolatey documentation.

If freshctl is not already running with administrator privileges when Chocolatey bootstrap is needed, it shows an elevation screen first. Press `enter` there to relaunch freshctl as administrator through Windows UAC.

If `C:\ProgramData\chocolatey` exists but `C:\ProgramData\chocolatey\bin\choco.exe` is missing, freshctl treats it as a broken partial install. It will not rerun bootstrap until that folder is removed from the repair screen.

## Install Logs

The install screen shows compact progress by default: current app, command, and a per-app summary. Logs are hidden by default; press `l` to toggle full logs. Press `s` to skip the current app and continue with the next selected app. Each package install has a 30 minute timeout as a last-resort guard.

## Warning

freshctl uses Chocolatey and installs real apps on your machine. Review the commands shown in the app before starting installation.
