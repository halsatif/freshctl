$ErrorActionPreference = "Stop"

$Owner = "halsatif"
$Repo = "freshctl"
$ExeName = "freshctl.exe"

$InstallDir = Join-Path $env:LOCALAPPDATA "freshctl"
$ExePath = Join-Path $InstallDir $ExeName

Write-Host "freshctl installer" -ForegroundColor Cyan

if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}

$ApiUrl = "https://api.github.com/repos/$Owner/$Repo/releases/latest"

Write-Host "Fetching latest release..."
$Release = Invoke-RestMethod -Uri $ApiUrl -Headers @{
    "User-Agent" = "freshctl-installer"
}

$Asset = $Release.assets | Where-Object {
    $_.name -eq $ExeName
} | Select-Object -First 1

if (!$Asset) {
    throw "Could not find $ExeName in latest release."
}

Write-Host "Downloading $ExeName..."
Invoke-WebRequest -Uri $Asset.browser_download_url -OutFile $ExePath

Write-Host "Starting freshctl..." -ForegroundColor Green
Start-Process -FilePath $ExePath -Wait