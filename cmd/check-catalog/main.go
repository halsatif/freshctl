package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/halsatif/freshctl/internal/catalog"
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

func validateIDs(packages []catalog.Package) error {
	seen := make(map[string]string, len(packages))
	for _, pkg := range packages {
		id := strings.TrimSpace(pkg.PackageID)
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

func validateNames(packages []catalog.Package) error {
	seen := make(map[string]string, len(packages))
	for _, pkg := range packages {
		name := strings.TrimSpace(pkg.Name)
		if name == "" {
			return fmt.Errorf("package %q has empty name", pkg.PackageID)
		}
		lower := strings.ToLower(name)
		if strings.Contains(lower, "todo") || strings.Contains(lower, "placeholder") {
			return fmt.Errorf("package %q has placeholder-like name %q", pkg.PackageID, pkg.Name)
		}
		if previous, ok := seen[lower]; ok && previous != pkg.PackageID {
			return fmt.Errorf("duplicate package name %q used by %q and %q", pkg.Name, previous, pkg.PackageID)
		}
		seen[lower] = pkg.PackageID
	}
	return nil
}

func validateMetadata(packages []catalog.Package) error {
	for _, pkg := range packages {
		if strings.TrimSpace(pkg.Description) == "" {
			return fmt.Errorf("%s (%s) has empty description", pkg.Name, pkg.PackageID)
		}
		if strings.TrimSpace(pkg.Category) == "" {
			return fmt.Errorf("%s (%s) has empty category", pkg.Name, pkg.PackageID)
		}
		if !validType(pkg.Type) {
			return fmt.Errorf("%s (%s) has invalid type %q", pkg.Name, pkg.PackageID, pkg.Type)
		}
		if !validSource(pkg.Source) {
			return fmt.Errorf("%s (%s) has invalid source %q", pkg.Name, pkg.PackageID, pkg.Source)
		}
		if !pkg.Verified {
			return fmt.Errorf("%s (%s) should be verified in default catalog", pkg.Name, pkg.PackageID)
		}
	}
	return nil
}

func validateBannedPackages(packages []catalog.Package) error {
	byID := packagesByID(packages)
	for _, id := range bannedPackageIDs {
		if pkg, ok := byID[id]; ok {
			return fmt.Errorf("banned package %q reappeared as %q", id, pkg.Name)
		}
	}
	return nil
}

func validateExpectedTypes(packages []catalog.Package) error {
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

func validSource(source catalog.PackageSource) bool {
	switch source {
	case catalog.PackageSourceChocolatey:
		return true
	default:
		return false
	}
}

func collectPackages(categories []catalog.Category) []catalog.Package {
	packages := make([]catalog.Package, 0)
	for _, category := range categories {
		packages = append(packages, collectPackages(category.Categories)...)
		packages = append(packages, category.Apps...)
	}
	return packages
}

func packagesByID(packages []catalog.Package) map[string]catalog.Package {
	byID := make(map[string]catalog.Package, len(packages))
	for _, pkg := range packages {
		byID[pkg.PackageID] = pkg
	}
	return byID
}
