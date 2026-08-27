package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/halsatif/freshctl/internal/catalog"
)

func TestExportCatalogIncludesMetadata(t *testing.T) {
	packages := catalog.Export(catalog.Default())
	if len(packages) != 164 {
		t.Fatalf("expected 164 exported packages, got %d", len(packages))
	}

	byID := packagesByID(packages)
	vscode := byID["vscode"]
	if vscode.Name != "Visual Studio Code" {
		t.Fatalf("expected Visual Studio Code export, got %#v", vscode)
	}
	if vscode.Type != string(catalog.PackageTypeApplication) {
		t.Fatalf("expected Visual Studio Code type %q, got %q", catalog.PackageTypeApplication, vscode.Type)
	}
	if vscode.Source != string(catalog.ProviderDirect) {
		t.Fatalf("expected Visual Studio Code source %q, got %q", catalog.ProviderDirect, vscode.Source)
	}
	if !vscode.Verified {
		t.Fatal("expected Visual Studio Code to be verified")
	}
	powershell := byID["powershell-core"]
	if powershell.Name != "PowerShell 7" || powershell.Source != string(catalog.ProviderDirect) {
		t.Fatalf("expected PowerShell 7 to use Direct, got %#v", powershell)
	}
	if byID["docker-desktop"].PackageID != "" {
		t.Fatal("docker-desktop should not be exported")
	}
}

func TestGeneratedSiteCatalogIsCurrent(t *testing.T) {
	want, err := catalog.GeneratedCatalogJS(catalog.Default())
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join("..", "..", "site", "catalog.generated.js")
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated site catalog: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("%s is stale; run go run ./cmd/export-catalog", path)
	}

	packages, err := catalog.DecodeGeneratedCatalogJS(got)
	if err != nil {
		t.Fatal(err)
	}
	if len(packages) != 164 {
		t.Fatalf("expected generated site catalog to contain 164 packages, got %d", len(packages))
	}
}

func TestExportUsesApplicationIDAndPrimaryProvider(t *testing.T) {
	categories := []catalog.Category{{
		Name: "Example",
		Apps: []catalog.Application{{
			ID:       "stable-application-id",
			Name:     "Example App",
			Category: "example",
			Providers: []catalog.Provider{{
				Type:      catalog.ProviderChocolatey,
				PackageID: "provider-specific-id",
				Strategy:  catalog.InstallStrategyPackageManager,
			}},
		}},
	}}

	exported := catalog.Export(categories)
	if len(exported) != 1 {
		t.Fatalf("expected one exported application, got %d", len(exported))
	}
	if exported[0].PackageID != "stable-application-id" {
		t.Fatalf("website profile ID should use application ID, got %q", exported[0].PackageID)
	}
	if exported[0].Source != string(catalog.ProviderChocolatey) {
		t.Fatalf("website source should use primary provider, got %q", exported[0].Source)
	}
}

func packagesByID(packages []catalog.ExportPackage) map[string]catalog.ExportPackage {
	byID := make(map[string]catalog.ExportPackage, len(packages))
	for _, pkg := range packages {
		byID[pkg.PackageID] = pkg
	}
	return byID
}
