package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/halsatif/freshctl/internal/catalog"
	"github.com/halsatif/freshctl/internal/providers"
)

var bannedPackageIDs = []string{
	"teamspeak",
	"yandex-browser",
	"docker-desktop",
	"faceit",
	"nvidia-broadcast",
	"vmwareworkstation",
	"protonvpn",
	"rufus",
	"vcredist2005",
	"vcredist2008",
}

var expectedTypes = map[string]catalog.PackageType{
	"vscode":             catalog.PackageTypeApplication,
	"helix":              catalog.PackageTypeCLITool,
	"ripgrep":            catalog.PackageTypeCLITool,
	"fzf":                catalog.PackageTypeCLITool,
	"dotnet-8.0-runtime": catalog.PackageTypeRuntime,
	"vcredist140":        catalog.PackageTypeRuntime,
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	packages := collectPackages(catalog.Default())
	fmt.Printf("[OK] %d packages loaded\n", len(packages))

	if err := validateIDs(packages); err != nil {
		return err
	}
	fmt.Println("[OK] No duplicate package IDs")

	if err := validateNames(packages); err != nil {
		return err
	}
	fmt.Println("[OK] Package names look valid")

	if err := validateMetadata(packages); err != nil {
		return err
	}
	fmt.Println("[OK] Metadata validation passed")

	if err := validateBannedPackages(packages); err != nil {
		return err
	}
	fmt.Println("[OK] Banned packages absent")

	if err := validateExpectedTypes(packages); err != nil {
		return err
	}
	fmt.Println("[OK] Known package types match expectations")

	if err := validateWebsiteCatalog(); err != nil {
		return err
	}
	fmt.Println("[OK] Website catalog synchronized")

	fmt.Println()
	fmt.Println("Catalog validation passed.")
	return nil
}

func validateIDs(packages []catalog.Application) error {
	seen := make(map[string]string, len(packages))
	for _, pkg := range packages {
		id := strings.TrimSpace(pkg.ID)
		if id == "" {
			return fmt.Errorf("package %q has empty package ID", pkg.Name)
		}
		if previous, ok := seen[id]; ok {
			return fmt.Errorf("duplicate package ID %q used by %q and %q", id, previous, pkg.Name)
		}
		seen[id] = pkg.Name
	}
	return nil
}

func validateNames(packages []catalog.Application) error {
	seen := make(map[string]string, len(packages))
	for _, pkg := range packages {
		name := strings.TrimSpace(pkg.Name)
		if name == "" {
			return fmt.Errorf("package %q has empty name", pkg.ID)
		}
		lower := strings.ToLower(name)
		if strings.Contains(lower, "todo") || strings.Contains(lower, "placeholder") {
			return fmt.Errorf("package %q has placeholder-like name %q", pkg.ID, pkg.Name)
		}
		if previous, ok := seen[lower]; ok && previous != pkg.ID {
			return fmt.Errorf("duplicate package name %q used by %q and %q", pkg.Name, previous, pkg.ID)
		}
		seen[lower] = pkg.ID
	}
	return nil
}

func validateMetadata(packages []catalog.Application) error {
	for _, pkg := range packages {
		if strings.TrimSpace(pkg.Description) == "" {
			return fmt.Errorf("%s (%s) has empty description", pkg.Name, pkg.ID)
		}
		if strings.TrimSpace(pkg.Category) == "" {
			return fmt.Errorf("%s (%s) has empty category", pkg.Name, pkg.ID)
		}
		if !validType(pkg.Type) {
			return fmt.Errorf("%s (%s) has invalid type %q", pkg.Name, pkg.ID, pkg.Type)
		}
		if len(pkg.Providers) == 0 {
			return fmt.Errorf("%s (%s) has no install providers", pkg.Name, pkg.ID)
		}
		seenProviders := make(map[catalog.ProviderType]bool, len(pkg.Providers))
		for _, provider := range pkg.Providers {
			if !validProviderType(provider.Type) {
				return fmt.Errorf("%s (%s) has invalid provider type %q", pkg.Name, pkg.ID, provider.Type)
			}
			if seenProviders[provider.Type] {
				return fmt.Errorf("%s (%s) has duplicate provider %q", pkg.Name, pkg.ID, provider.Type)
			}
			seenProviders[provider.Type] = true
			if strings.TrimSpace(provider.PackageID) == "" {
				return fmt.Errorf("%s (%s) has empty package ID for provider %q", pkg.Name, pkg.ID, provider.Type)
			}
			if !validInstallStrategy(provider.Strategy) {
				return fmt.Errorf("%s (%s) has invalid install strategy %q", pkg.Name, pkg.ID, provider.Strategy)
			}
			if _, ok := providers.Get(provider.Type); !ok {
				return fmt.Errorf("%s (%s) uses provider %q without an installer implementation", pkg.Name, pkg.ID, provider.Type)
			}
		}
		if !pkg.Verified {
			return fmt.Errorf("%s (%s) should be verified in default catalog", pkg.Name, pkg.ID)
		}
	}
	return nil
}

func validateBannedPackages(packages []catalog.Application) error {
	byID := packagesByID(packages)
	for _, id := range bannedPackageIDs {
		if pkg, ok := byID[id]; ok {
			return fmt.Errorf("banned package %q reappeared as %q", id, pkg.Name)
		}
	}
	return nil
}

func validateExpectedTypes(packages []catalog.Application) error {
	byID := packagesByID(packages)
	for id, expected := range expectedTypes {
		pkg, ok := byID[id]
		if !ok {
			return fmt.Errorf("expected package %q is missing", id)
		}
		if pkg.Type != expected {
			return fmt.Errorf("%s (%s) should have type %q, got %q", pkg.Name, id, expected, pkg.Type)
		}
	}
	return nil
}

func validateWebsiteCatalog() error {
	path := filepath.Join("site", "catalog.generated.js")
	current, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read website catalog: %w", err)
	}

	expected, err := catalog.GeneratedCatalogJS(catalog.Default())
	if err != nil {
		return err
	}
	if !bytes.Equal(current, expected) {
		return fmt.Errorf("%s is stale; run go run ./cmd/export-catalog", path)
	}

	exported, err := catalog.DecodeGeneratedCatalogJS(current)
	if err != nil {
		return err
	}
	packages := collectPackages(catalog.Default())
	if len(exported) != len(packages) {
		return fmt.Errorf("website catalog has %d packages, Go catalog has %d", len(exported), len(packages))
	}
	return nil
}

func validType(packageType catalog.PackageType) bool {
	switch packageType {
	case catalog.PackageTypeApplication, catalog.PackageTypeCLITool, catalog.PackageTypeRuntime:
		return true
	default:
		return false
	}
}

func validProviderType(providerType catalog.ProviderType) bool {
	switch providerType {
	case catalog.ProviderChocolatey, catalog.ProviderWinget, catalog.ProviderDirect, catalog.ProviderCommunity:
		return true
	default:
		return false
	}
}

func validInstallStrategy(strategy catalog.InstallStrategy) bool {
	switch strategy {
	case catalog.InstallStrategyPackageManager, catalog.InstallStrategyDirectInstaller:
		return true
	default:
		return false
	}
}

func collectPackages(categories []catalog.Category) []catalog.Application {
	packages := make([]catalog.Application, 0)
	for _, category := range categories {
		packages = append(packages, collectPackages(category.Categories)...)
		packages = append(packages, category.Apps...)
	}
	return packages
}

func packagesByID(packages []catalog.Application) map[string]catalog.Application {
	byID := make(map[string]catalog.Application, len(packages))
	for _, pkg := range packages {
		byID[pkg.ID] = pkg
	}
	return byID
}
