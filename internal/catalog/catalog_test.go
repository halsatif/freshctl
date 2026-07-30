package catalog

import (
	"strings"
	"testing"
)

func TestDefaultPackagesHaveMetadata(t *testing.T) {
	for _, app := range collectPackages(Default()) {
		if strings.TrimSpace(app.Name) == "" {
			t.Fatal("package name should not be empty")
		}
		if strings.TrimSpace(app.ID) == "" {
			t.Fatalf("%s should have a package id", app.Name)
		}
		if strings.TrimSpace(app.Description) == "" {
			t.Fatalf("%s should have a description", app.Name)
		}
		if strings.TrimSpace(app.Category) == "" {
			t.Fatalf("%s should have a category", app.Name)
		}
		if !validPackageType(app.Type) {
			t.Fatalf("%s should have a valid package type, got %q", app.Name, app.Type)
		}
		if len(app.Providers) == 0 {
			t.Fatalf("%s should have at least one provider", app.Name)
		}
		for _, provider := range app.Providers {
			if !validProviderType(provider.Type) {
				t.Fatalf("%s should have a valid provider type, got %q", app.Name, provider.Type)
			}
			if strings.TrimSpace(provider.PackageID) == "" {
				t.Fatalf("%s should have a provider package ID", app.Name)
			}
			if !validInstallStrategy(provider.Strategy) {
				t.Fatalf("%s should have a valid install strategy, got %q", app.Name, provider.Strategy)
			}
		}
		if !app.Verified {
			t.Fatalf("%s should be marked verified", app.Name)
		}
		if !validDetectMethod(app.DetectMethod) {
			t.Fatalf("%s should have a valid detection method, got %q", app.Name, app.DetectMethod)
		}
		if app.DetectMethod != DetectNone && strings.TrimSpace(app.DetectValue) == "" {
			t.Fatalf("%s should have a detection value for method %q", app.Name, app.DetectMethod)
		}
	}
}

func TestDefaultPackageTypeExamples(t *testing.T) {
	apps := packagesByID(Default())

	examples := map[string]PackageType{
		"vscode":                     PackageTypeApplication,
		"microsoft-windows-terminal": PackageTypeApplication,
		"helix":                      PackageTypeCLITool,
		"ripgrep":                    PackageTypeCLITool,
		"fzf":                        PackageTypeCLITool,
		"golang":                     PackageTypeCLITool,
		"cmake":                      PackageTypeCLITool,
		"wezterm":                    PackageTypeApplication,
		"vcredist140":                PackageTypeRuntime,
		"dotnet-8.0-runtime":         PackageTypeRuntime,
		"nodejs-lts":                 PackageTypeRuntime,
		"googlechrome":               PackageTypeApplication,
	}

	for id, want := range examples {
		app, ok := apps[id]
		if !ok {
			t.Fatalf("expected package %s in default catalog", id)
		}
		if app.Type != want {
			t.Fatalf("%s should have type %q, got %q", app.Name, want, app.Type)
		}
	}
}

func TestOnlyVSCodeUsesDirectProvider(t *testing.T) {
	for _, app := range collectPackages(Default()) {
		if len(app.Providers) != 1 {
			t.Fatalf("%s should have exactly one provider, got %d", app.Name, len(app.Providers))
		}
		provider := app.Providers[0]
		expectedType := ProviderChocolatey
		expectedStrategy := InstallStrategyPackageManager
		if app.ID == "vscode" {
			expectedType = ProviderDirect
			expectedStrategy = InstallStrategyDirectInstaller
		}
		if provider.Type != expectedType {
			t.Fatalf("%s should use %s, got %q", app.Name, expectedType, provider.Type)
		}
		if provider.Strategy != expectedStrategy {
			t.Fatalf("%s should use strategy %s, got %q", app.Name, expectedStrategy, provider.Strategy)
		}
		if app.ID != "vscode" && provider.Type != ProviderChocolatey {
			t.Fatalf("%s should still use Chocolatey, got %q", app.Name, provider.Type)
		}
		if provider.PackageID != app.ID {
			t.Fatalf("%s should preserve its provider package ID, got %q", app.Name, provider.PackageID)
		}
	}
}

func TestProviderLookup(t *testing.T) {
	app := packagesByID(Default())["vscode"]
	provider, ok := app.ProviderByType(ProviderDirect)
	if !ok {
		t.Fatal("VS Code should expose its Direct provider")
	}
	if provider.PackageID != "vscode" {
		t.Fatalf("unexpected VS Code provider package ID %q", provider.PackageID)
	}
	if provider.Metadata.Direct == nil {
		t.Fatal("VS Code should include direct installer metadata")
	}
	if provider.Metadata.Direct.InstallerType != InstallerTypeExecutable {
		t.Fatalf("unexpected VS Code installer type %q", provider.Metadata.Direct.InstallerType)
	}
	if len(provider.Metadata.Direct.Downloads) != 2 {
		t.Fatalf("VS Code should provide x64 and arm64 downloads, got %d", len(provider.Metadata.Direct.Downloads))
	}
	if len(provider.Metadata.Direct.SilentArgs) == 0 {
		t.Fatal("VS Code should include silent installer arguments")
	}
}

func TestPrereleaseMetadataBelongsToProvider(t *testing.T) {
	zen := packagesByID(Default())["zen-browser"]
	provider, ok := zen.ProviderByType(ProviderChocolatey)
	if !ok {
		t.Fatal("Zen Browser should expose its Chocolatey provider")
	}
	if !provider.Metadata.Prerelease {
		t.Fatal("Zen Browser prerelease flag should be stored in provider metadata")
	}
}

func TestKnownPackageDetectionMetadata(t *testing.T) {
	apps := packagesByID(Default())
	examples := map[string]struct {
		method DetectMethod
		value  string
	}{
		"googlechrome":               {method: DetectRegistry, value: "Google Chrome"},
		"firefox":                    {method: DetectRegistry, value: "Mozilla Firefox"},
		"brave":                      {method: DetectRegistry, value: "Brave"},
		"microsoft-edge":             {method: DetectRegistry, value: "Microsoft Edge"},
		"librewolf":                  {method: DetectRegistry, value: "LibreWolf"},
		"zen-browser":                {method: DetectRegistry, value: "Zen Browser"},
		"vscode":                     {method: DetectRegistry, value: "Visual Studio Code"},
		"git":                        {method: DetectPath, value: "git.exe"},
		"nodejs-lts":                 {method: DetectPath, value: "node.exe"},
		"python":                     {method: DetectPath, value: "python.exe"},
		"golang":                     {method: DetectPath, value: "go.exe"},
		"rustup.install":             {method: DetectPath, value: "rustup.exe"},
		"cmake":                      {method: DetectPath, value: "cmake.exe"},
		"powershell-core":            {method: DetectPath, value: "pwsh.exe"},
		"microsoft-windows-terminal": {method: DetectRegistry, value: "Windows Terminal"},
		"wezterm":                    {method: DetectRegistry, value: "WezTerm"},
		"discord":                    {method: DetectRegistry, value: "Discord"},
		"telegram":                   {method: DetectRegistry, value: "Telegram Desktop"},
		"signal":                     {method: DetectRegistry, value: "Signal"},
		"7zip":                       {method: DetectRegistry, value: "7-Zip"},
		"everything":                 {method: DetectRegistry, value: "Everything"},
		"powertoys":                  {method: DetectRegistry, value: "PowerToys"},
		"sharex":                     {method: DetectRegistry, value: "ShareX"},
		"notepadplusplus":            {method: DetectRegistry, value: "Notepad++"},
		"windirstat":                 {method: DetectRegistry, value: "WinDirStat"},
		"wiztree":                    {method: DetectRegistry, value: "WizTree"},
		"vlc":                        {method: DetectRegistry, value: "VLC media player"},
		"spotify":                    {method: DetectRegistry, value: "Spotify"},
		"obs-studio":                 {method: DetectRegistry, value: "OBS Studio"},
		"handbrake":                  {method: DetectRegistry, value: "HandBrake"},
		"steam":                      {method: DetectRegistry, value: "Steam"},
		"prismlauncher":              {method: DetectRegistry, value: "Prism Launcher"},
		"heroic-games-launcher":      {method: DetectRegistry, value: "Heroic"},
		"helix":                      {method: DetectPath, value: "hx.exe"},
		"ripgrep":                    {method: DetectPath, value: "rg.exe"},
		"fzf":                        {method: DetectPath, value: "fzf.exe"},
		"fastfetch":                  {method: DetectPath, value: "fastfetch.exe"},
		"yt-dlp":                     {method: DetectPath, value: "yt-dlp.exe"},
		"ffmpeg":                     {method: DetectPath, value: "ffmpeg.exe"},
		"adb":                        {method: DetectPath, value: "adb.exe"},
		"vcredist140":                {method: DetectRegistry, value: "Microsoft Visual C++"},
	}

	for id, want := range examples {
		app, ok := apps[id]
		if !ok {
			t.Fatalf("expected package %s in default catalog", id)
		}
		if app.DetectMethod != want.method || app.DetectValue != want.value {
			t.Fatalf("%s detection metadata = %q/%q, want %q/%q", id, app.DetectMethod, app.DetectValue, want.method, want.value)
		}
	}
}

func TestDetectionMetadataCoverage(t *testing.T) {
	registryCount := 0
	pathCount := 0
	for _, app := range collectPackages(Default()) {
		switch app.DetectMethod {
		case DetectRegistry:
			registryCount++
		case DetectPath:
			pathCount++
		}
	}

	total := registryCount + pathCount
	if total < 40 || total > 60 {
		t.Fatalf("detection coverage should stay around 40-60 packages, got total=%d registry=%d path=%d", total, registryCount, pathCount)
	}
	if registryCount == 0 || pathCount == 0 {
		t.Fatalf("detection coverage should include registry and path methods, registry=%d path=%d", registryCount, pathCount)
	}
}

func TestKnownCLIToolsUsePathDetection(t *testing.T) {
	apps := packagesByID(Default())
	for _, id := range []string{
		"git",
		"golang",
		"rustup.install",
		"cmake",
		"powershell-core",
		"fastfetch",
		"helix",
		"ripgrep",
		"fzf",
		"yt-dlp",
		"ffmpeg",
		"adb",
	} {
		app, ok := apps[id]
		if !ok {
			t.Fatalf("expected CLI package %s in default catalog", id)
		}
		if app.DetectMethod != DetectPath {
			t.Fatalf("%s should use PATH detection, got %q", app.Name, app.DetectMethod)
		}
	}
}

func TestKnownCLIToolsMentionCommand(t *testing.T) {
	apps := packagesByID(Default())
	commands := map[string]string{
		"helix":           "hx",
		"neovim":          "nvim",
		"git":             "git",
		"golang":          "go",
		"rustup.install":  "rustup",
		"cmake":           "cmake",
		"powershell-core": "pwsh",
		"fastfetch":       "fastfetch",
		"fzf":             "fzf",
		"ripgrep":         "rg",
		"codex-cli":       "codex",
		"yt-dlp":          "yt-dlp",
		"ffmpeg":          "ffmpeg",
		"adb":             "adb",
	}

	for id, command := range commands {
		app, ok := apps[id]
		if !ok {
			t.Fatalf("expected CLI package %s in default catalog", id)
		}
		if app.Type != PackageTypeCLITool {
			t.Fatalf("%s should be a CLI tool, got %q", app.Name, app.Type)
		}
		if !strings.Contains(strings.ToLower(app.Description), strings.ToLower(command)) {
			t.Fatalf("%s description should mention command %q, got %q", app.Name, command, app.Description)
		}
	}
}

func validPackageType(packageType PackageType) bool {
	switch packageType {
	case PackageTypeApplication, PackageTypeCLITool, PackageTypeRuntime:
		return true
	default:
		return false
	}
}

func validProviderType(providerType ProviderType) bool {
	switch providerType {
	case ProviderChocolatey, ProviderWinget, ProviderDirect, ProviderCommunity:
		return true
	default:
		return false
	}
}

func validInstallStrategy(strategy InstallStrategy) bool {
	switch strategy {
	case InstallStrategyPackageManager, InstallStrategyDirectInstaller:
		return true
	default:
		return false
	}
}

func validDetectMethod(method DetectMethod) bool {
	switch method {
	case DetectNone, DetectRegistry, DetectPath:
		return true
	default:
		return false
	}
}

func collectPackages(categories []Category) []Application {
	apps := make([]Application, 0)
	for _, category := range categories {
		apps = append(apps, collectPackages(category.Categories)...)
		apps = append(apps, category.Apps...)
	}
	return apps
}

func packagesByID(categories []Category) map[string]Application {
	apps := make(map[string]Application)
	for _, app := range collectPackages(categories) {
		apps[app.ID] = app
	}
	return apps
}
