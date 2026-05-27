param(
    [switch]$Silent,
    [switch]$NoLaunch,
    [switch]$Uninstall
)

$ErrorActionPreference = "Stop"

$AppName = "freshctl"
$Owner = "halsatif"
$Repo = "freshctl"
$ExeName = "freshctl.exe"
$InstallDir = Join-Path $env:ProgramFiles $AppName
$InstallPath = Join-Path $InstallDir $ExeName
$TempRoot = Join-Path $env:TEMP "freshctl-installer"
$ShortcutPath = Join-Path ([Environment]::GetFolderPath("CommonPrograms")) "freshctl.lnk"
$InstallerUrl = "https://freshctl.tech/install.ps1"
$GitHubApiUrl = "https://api.github.com/repos/$Owner/$Repo/releases/latest"

function Write-Info {
    param([string]$Message)
    if (-not $Silent) {
        Write-Host "[*] $Message" -ForegroundColor Cyan
    }
}

function Write-Ok {
    param([string]$Message)
    if (-not $Silent) {
        Write-Host "[+] $Message" -ForegroundColor Green
    }
}

function Write-Warn {
    param([string]$Message)
    if (-not $Silent) {
        Write-Host "[!] $Message" -ForegroundColor Yellow
    }
}

function Write-Fail {
    param([string]$Message)
    Write-Host "[x] $Message" -ForegroundColor Red
}

function Enable-ModernTls {
    try {
        $tls12 = [Net.SecurityProtocolType]3072
        $tls13 = [Net.SecurityProtocolType]12288
        [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor $tls12 -bor $tls13
    } catch {
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    }
}

function Test-IsWindows {
    if ($PSVersionTable.PSEdition -eq "Core") {
        return $IsWindows
    }
    return $true
}

function Test-IsAdmin {
    $identity = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($identity)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Get-PowerShellHostPath {
    $pwsh = Get-Command pwsh.exe -ErrorAction SilentlyContinue
    if ($pwsh) {
        return $pwsh.Source
    }
    $powershell = Get-Command powershell.exe -ErrorAction SilentlyContinue
    if ($powershell) {
        return $powershell.Source
    }
    throw "Could not locate powershell.exe."
}

function Get-InstallerArguments {
    $arguments = @()
    if ($Silent) { $arguments += "-Silent" }
    if ($NoLaunch) { $arguments += "-NoLaunch" }
    if ($Uninstall) { $arguments += "-Uninstall" }
    return $arguments
}

function Get-CurrentScriptPath {
    if ($PSCommandPath -and (Test-Path $PSCommandPath)) {
        return $PSCommandPath
    }
    if ($MyInvocation.MyCommand.Path -and (Test-Path $MyInvocation.MyCommand.Path)) {
        return $MyInvocation.MyCommand.Path
    }
    return $null
}

function Request-Administrator {
    if (Test-IsAdmin) {
        return
    }

    Write-Warn "Administrator privileges are required."
    New-Item -ItemType Directory -Path $TempRoot -Force | Out-Null

    $scriptPath = Get-CurrentScriptPath
    if (-not $scriptPath) {
        $scriptPath = Join-Path $TempRoot "install.ps1"
        Write-Info "Downloading installer for elevated restart..."
        Invoke-WebRequestCompat -Uri $InstallerUrl -OutFile $scriptPath
    }

    $argList = @(
        "-NoProfile",
        "-ExecutionPolicy", "Bypass",
        "-File", "`"$scriptPath`""
    ) + (Get-InstallerArguments)

    $hostPath = Get-PowerShellHostPath
    Write-Info "Opening Windows UAC prompt..."
    Start-Process -FilePath $hostPath -ArgumentList $argList -Verb RunAs | Out-Null
    exit 0
}

function Get-SystemArchitecture {
    $arch = $env:PROCESSOR_ARCHITEW6432
    if ([string]::IsNullOrWhiteSpace($arch)) {
        $arch = $env:PROCESSOR_ARCHITECTURE
    }

    switch -Regex ($arch) {
        "ARM64|AARCH64" { return "arm64" }
        "AMD64|X64" { return "x64" }
        default {
            throw "Unsupported CPU architecture: $arch. freshctl currently provides Windows x64 and arm64 builds."
        }
    }
}

function Invoke-WebRequestCompat {
    param(
        [Parameter(Mandatory = $true)][string]$Uri,
        [string]$OutFile,
        [string]$Method = "Get",
        [hashtable]$Headers
    )

    $params = @{
        Uri = $Uri
        Method = $Method
        UseBasicParsing = $true
        ErrorAction = "Stop"
    }
    if ($OutFile) { $params.OutFile = $OutFile }
    if ($Headers) { $params.Headers = $Headers }

    return Invoke-WebRequest @params
}

function Invoke-RestMethodCompat {
    param(
        [Parameter(Mandatory = $true)][string]$Uri,
        [hashtable]$Headers
    )

    $params = @{
        Uri = $Uri
        UseBasicParsing = $true
        ErrorAction = "Stop"
    }
    if ($Headers) { $params.Headers = $Headers }

    return Invoke-RestMethod @params
}

function Test-InternetConnectivity {
    Write-Info "Checking internet connectivity..."
    try {
        Invoke-WebRequestCompat -Uri "https://api.github.com" -Method "Head" -Headers @{ "User-Agent" = "freshctl-installer" } | Out-Null
        Write-Ok "Internet connection looks good."
    } catch {
        throw "Could not reach GitHub. Check your internet connection and try again."
    }
}

function Get-LatestRelease {
    Write-Info "Fetching latest GitHub release..."
    try {
        $release = Invoke-RestMethodCompat -Uri $GitHubApiUrl -Headers @{
            "Accept" = "application/vnd.github+json"
            "User-Agent" = "freshctl-installer"
            "X-GitHub-Api-Version" = "2022-11-28"
        }
    } catch {
        throw "Could not query GitHub releases API: $($_.Exception.Message)"
    }

    if (-not $release -or -not $release.assets -or $release.assets.Count -eq 0) {
        throw "Latest GitHub release does not contain downloadable assets."
    }

    Write-Ok "Latest release: $($release.tag_name)"
    return $release
}

function Select-ReleaseAsset {
    param(
        [Parameter(Mandatory = $true)]$Release,
        [Parameter(Mandatory = $true)][string]$Architecture
    )

    $exeAssets = @($Release.assets | Where-Object {
        $_.name -and
        $_.browser_download_url -and
        $_.name.ToLowerInvariant().EndsWith(".exe") -and
        $_.name.ToLowerInvariant() -notmatch "checksum|sha256|sig|signature"
    })

    if ($exeAssets.Count -eq 0) {
        throw "Latest release does not contain a Windows executable asset."
    }

    $archPatterns = if ($Architecture -eq "arm64") {
        @("arm64", "aarch64", "windows-arm64", "win-arm64")
    } else {
        @("x64", "amd64", "win64", "windows-x64", "windows-amd64")
    }

    foreach ($pattern in $archPatterns) {
        $match = $exeAssets | Where-Object { $_.name.ToLowerInvariant().Contains($pattern) } | Select-Object -First 1
        if ($match) {
            return $match
        }
    }

    $generic = $exeAssets | Where-Object { $_.name.ToLowerInvariant() -eq $ExeName } | Select-Object -First 1
    if ($generic) {
        Write-Warn "No exact $Architecture asset found. Falling back to generic $ExeName."
        return $generic
    }

    $windowsAsset = $exeAssets | Where-Object {
        $name = $_.name.ToLowerInvariant()
        $name -match "windows|win|freshctl"
    } | Select-Object -First 1

    if ($windowsAsset) {
        Write-Warn "No exact $Architecture or generic asset found. Falling back to first Windows executable asset: $($windowsAsset.name)"
        return $windowsAsset
    }

    Write-Warn "No exact $Architecture, generic, or clearly named Windows asset found. Falling back to first executable asset: $($exeAssets[0].name)"
    return $exeAssets[0]
}

function Test-DownloadAvailability {
    param([Parameter(Mandatory = $true)]$Asset)

    Write-Info "Checking download availability..."
    try {
        Invoke-WebRequestCompat -Uri $Asset.browser_download_url -Method "Head" -Headers @{ "User-Agent" = "freshctl-installer" } | Out-Null
        Write-Ok "Release asset is available."
    } catch {
        throw "GitHub release asset is not reachable: $($_.Exception.Message)"
    }
}

function Download-Asset {
    param([Parameter(Mandatory = $true)]$Asset)

    New-Item -ItemType Directory -Path $TempRoot -Force | Out-Null
    $downloadPath = Join-Path $TempRoot $Asset.name

    if ((Test-Path $downloadPath) -and ((Get-Item $downloadPath).Length -gt 0)) {
        Write-Ok "Using existing temporary download: $($Asset.name)"
        return $downloadPath
    }

    Write-Info "Downloading $($Asset.name)..."
    $previousProgressPreference = $ProgressPreference
    try {
        if ($Silent) {
            $ProgressPreference = "SilentlyContinue"
        }
        Invoke-WebRequestCompat -Uri $Asset.browser_download_url -OutFile $downloadPath -Headers @{ "User-Agent" = "freshctl-installer" } | Out-Null
    } finally {
        $ProgressPreference = $previousProgressPreference
    }

    if (-not (Test-Path $downloadPath) -or ((Get-Item $downloadPath).Length -eq 0)) {
        throw "Download failed or produced an empty file."
    }

    Write-Ok "Downloaded freshctl."
    return $downloadPath
}

function Install-Executable {
    param([Parameter(Mandatory = $true)][string]$DownloadedPath)

    Write-Info "Installing to $InstallDir..."
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Copy-Item -Path $DownloadedPath -Destination $InstallPath -Force

    if (-not (Test-Path $InstallPath)) {
        throw "Installed executable was not found at $InstallPath."
    }

    Write-Ok "Installed $AppName."
}

function Get-MachinePathEntries {
    $path = [Environment]::GetEnvironmentVariable("Path", "Machine")
    if ([string]::IsNullOrWhiteSpace($path)) {
        return @()
    }
    return @($path.Split(";") | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
}

function Add-ToPath {
    $entries = Get-MachinePathEntries
    $alreadyInPath = $false

    foreach ($entry in $entries) {
        if ($entry.TrimEnd("\") -ieq $InstallDir.TrimEnd("\")) {
            $alreadyInPath = $true
            break
        }
    }

    if ($alreadyInPath) {
        Write-Ok "PATH already contains $InstallDir."
        return
    }

    Write-Info "Adding freshctl to PATH..."
    $entries += $InstallDir
    $newPath = ($entries | Select-Object -Unique) -join ";"
    [Environment]::SetEnvironmentVariable("Path", $newPath, "Machine")
    $env:Path = "$env:Path;$InstallDir"
    Send-EnvironmentChanged
    Write-Ok "PATH updated."
}

function Remove-FromPath {
    Write-Info "Removing freshctl from PATH..."
    $entries = Get-MachinePathEntries | Where-Object {
        $_.TrimEnd("\") -ine $InstallDir.TrimEnd("\")
    }
    [Environment]::SetEnvironmentVariable("Path", ($entries -join ";"), "Machine")
    Send-EnvironmentChanged
    Write-Ok "PATH cleaned."
}

function Send-EnvironmentChanged {
    try {
        Add-Type -Namespace Win32 -Name NativeMethods -MemberDefinition @"
[DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
public static extern System.IntPtr SendMessageTimeout(
    System.IntPtr hWnd,
    uint Msg,
    System.IntPtr wParam,
    string lParam,
    uint fuFlags,
    uint uTimeout,
    out System.IntPtr lpdwResult);
"@ -ErrorAction SilentlyContinue

        $result = [IntPtr]::Zero
        [Win32.NativeMethods]::SendMessageTimeout([IntPtr]0xffff, 0x1A, [IntPtr]::Zero, "Environment", 0x2, 5000, [ref]$result) | Out-Null
    } catch {
        Write-Warn "Could not broadcast PATH update. New terminals will still pick it up."
    }
}

function New-StartMenuShortcut {
    Write-Info "Creating Start Menu shortcut..."
    $shell = New-Object -ComObject WScript.Shell
    $shortcut = $shell.CreateShortcut($ShortcutPath)
    $shortcut.TargetPath = $InstallPath
    $shortcut.WorkingDirectory = $InstallDir
    $shortcut.Description = "freshctl"
    $shortcut.Save()
    Write-Ok "Start Menu shortcut created."
}

function Remove-StartMenuShortcut {
    if (Test-Path $ShortcutPath) {
        Write-Info "Removing Start Menu shortcut..."
        Remove-Item -Path $ShortcutPath -Force
        Write-Ok "Start Menu shortcut removed."
    }
}

function Remove-TempFiles {
    if (Test-Path $TempRoot) {
        Remove-Item -Path $TempRoot -Recurse -Force -ErrorAction SilentlyContinue
    }
}

function Invoke-Uninstall {
    Write-Info "Uninstalling freshctl..."

    if (Test-Path $InstallDir) {
        Remove-Item -Path $InstallDir -Recurse -Force
        Write-Ok "Removed $InstallDir."
    } else {
        Write-Warn "Install directory was not found."
    }

    Remove-FromPath
    Remove-StartMenuShortcut
    Remove-TempFiles
    Write-Ok "freshctl has been uninstalled."
}

function Prompt-Launch {
    if ($Silent -or $NoLaunch) {
        return
    }

    if (-not (Test-Path $InstallPath)) {
        return
    }

    $answer = Read-Host "Launch freshctl now? [Y/N]"
    if ($answer -match "^(y|yes)$") {
        Write-Info "Launching freshctl..."
        Start-Process -FilePath $InstallPath
    }
}

function Invoke-Install {
    $architecture = Get-SystemArchitecture
    Write-Info "Detected architecture: $architecture"

    Test-InternetConnectivity
    $release = Get-LatestRelease
    $asset = Select-ReleaseAsset -Release $release -Architecture $architecture
    Write-Ok "Selected asset: $($asset.name)"

    Test-DownloadAvailability -Asset $asset
    $downloadedPath = Download-Asset -Asset $asset
    Install-Executable -DownloadedPath $downloadedPath
    Add-ToPath
    New-StartMenuShortcut
    Write-Ok "freshctl is ready."
    Prompt-Launch
}

try {
    if (-not (Test-IsWindows)) {
        throw "freshctl installer currently supports Windows only."
    }

    Enable-ModernTls
    Request-Administrator

    if (-not $Silent) {
        Write-Host ""
        Write-Host "freshctl installer" -ForegroundColor Cyan
        Write-Host "------------------" -ForegroundColor DarkGray
    }

    if ($Uninstall) {
        Invoke-Uninstall
    } else {
        Invoke-Install
    }
} catch {
    Write-Fail $_.Exception.Message
    exit 1
} finally {
    if (-not $Uninstall) {
        Remove-TempFiles
    }
}
