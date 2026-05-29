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
	if len(packages) != 189 {
		t.Fatalf("expected 189 exported packages, got %d", len(packages))
	}

	byID := packagesByID(packages)
	codex := byID["codex-cli"]
	if codex.Name != "Codex CLI" {
		t.Fatalf("expected Codex CLI export, got %#v", codex)
	}
	if codex.Type != string(catalog.PackageTypeCLITool) {
		t.Fatalf("expected Codex CLI type %q, got %q", catalog.PackageTypeCLITool, codex.Type)
	}
	if codex.Source != string(catalog.PackageSourceChocolatey) {
		t.Fatalf("expected Codex CLI source %q, got %q", catalog.PackageSourceChocolatey, codex.Source)
	}
	if !codex.Verified {
		t.Fatal("expected Codex CLI to be verified")
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
	if len(packages) != 189 {
		t.Fatalf("expected generated site catalog to contain 189 packages, got %d", len(packages))
	}
}

func packagesByID(packages []catalog.ExportPackage) map[string]catalog.ExportPackage {
	byID := make(map[string]catalog.ExportPackage, len(packages))
	for _, pkg := range packages {
		byID[pkg.PackageID] = pkg
	}
	return byID
}
