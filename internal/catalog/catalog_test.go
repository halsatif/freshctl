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
		if strings.TrimSpace(app.PackageID) == "" {
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
		if !validPackageSource(app.Source) {
			t.Fatalf("%s should have a valid source, got %q", app.Name, app.Source)
		}
		if !app.Verified {
			t.Fatalf("%s should be marked verified", app.Name)
		}
	}
}

func TestDefaultPackageTypeExamples(t *testing.T) {
	apps := packagesByID(Default())

	examples := map[string]PackageType{
		"vscode":       PackageTypeApplication,
		"helix":        PackageTypeCLITool,
		"ripgrep":      PackageTypeCLITool,
		"golang":       PackageTypeCLITool,
		"cmake":        PackageTypeCLITool,
		"wezterm":      PackageTypeApplication,
		"vcredist140":  PackageTypeRuntime,
		"nodejs-lts":   PackageTypeRuntime,
		"googlechrome": PackageTypeApplication,
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

func validPackageSource(source PackageSource) bool {
	switch source {
	case PackageSourceChocolatey:
		return true
	default:
		return false
	}
}

func collectPackages(categories []Category) []Package {
	apps := make([]Package, 0)
	for _, category := range categories {
		apps = append(apps, collectPackages(category.Categories)...)
		apps = append(apps, category.Apps...)
	}
	return apps
}

func packagesByID(categories []Category) map[string]Package {
	apps := make(map[string]Package)
	for _, app := range collectPackages(categories) {
		apps[app.PackageID] = app
	}
	return apps
}
