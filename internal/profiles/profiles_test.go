package profiles

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/halsatif/freshctl/internal/catalog"
)

func TestValidateAcceptsValidProfile(t *testing.T) {
	profile := Profile{
		Version:  Version,
		Name:     "my setup",
		Packages: []string{"vscode", "git", "firefox"},
	}

	if err := Validate(profile, testCatalogPackages()); err != nil {
		t.Fatalf("valid profile should pass validation: %v", err)
	}
}

func TestValidateRejectsUnsupportedVersion(t *testing.T) {
	profile := Profile{
		Version:  2,
		Name:     "future setup",
		Packages: []string{"vscode"},
	}

	if err := Validate(profile, testCatalogPackages()); err == nil {
		t.Fatal("unsupported profile version should fail validation")
	}
}

func TestValidateRejectsEmptyPackages(t *testing.T) {
	profile := Profile{
		Version: Version,
		Name:    "empty setup",
	}

	if err := Validate(profile, testCatalogPackages()); err == nil {
		t.Fatal("empty package list should fail validation")
	}
}

func TestValidateRejectsDuplicatePackageIDs(t *testing.T) {
	profile := Profile{
		Version:  Version,
		Name:     "duplicate setup",
		Packages: []string{"vscode", "vscode"},
	}

	if err := Validate(profile, testCatalogPackages()); err == nil {
		t.Fatal("duplicate package ids should fail validation")
	}
}

func TestValidateRejectsUnknownPackageID(t *testing.T) {
	profile := Profile{
		Version:  Version,
		Name:     "unknown setup",
		Packages: []string{"vscode", "missing-package"},
	}

	if err := Validate(profile, testCatalogPackages()); err == nil {
		t.Fatal("unknown package id should fail validation")
	}
}

func TestValidateAllowsOptionalName(t *testing.T) {
	profile := Profile{
		Version:  Version,
		Packages: []string{"vscode"},
	}

	if err := Validate(profile, testCatalogPackages()); err != nil {
		t.Fatalf("profile name should be optional: %v", err)
	}
}

func TestPackageIDsPreservesOrder(t *testing.T) {
	profile := Profile{
		Version:  Version,
		Name:     "ordered setup",
		Packages: []string{"vscode", "git", "firefox"},
	}

	got := PackageIDs(profile)
	want := []string{"vscode", "git", "firefox"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("package ids should preserve order, got %#v", got)
	}

	got[0] = "changed"
	if profile.Packages[0] != "vscode" {
		t.Fatal("PackageIDs should return a copy")
	}
}

func TestFromPackagesPreservesSelectionOrder(t *testing.T) {
	packages := []catalog.Package{
		{Name: "VS Code", PackageID: "vscode"},
		{Name: "Git", PackageID: "git"},
		{Name: "Firefox", PackageID: "firefox"},
	}

	profile := FromPackages(DefaultProfileName, packages)

	if profile.Version != Version || profile.Name != DefaultProfileName {
		t.Fatalf("profile should use default metadata, got %#v", profile)
	}
	if !reflect.DeepEqual(profile.Packages, []string{"vscode", "git", "firefox"}) {
		t.Fatalf("profile package ids should preserve order, got %#v", profile.Packages)
	}
}

func TestWriteJSONWritesValidProfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), DefaultExportPath)
	profile := Profile{
		Version:  Version,
		Name:     DefaultProfileName,
		Packages: []string{"vscode", "git", "firefox"},
	}

	if err := WriteJSON(path, profile, testCatalogPackages()); err != nil {
		t.Fatalf("WriteJSON should write valid profile: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("profile file should exist: %v", err)
	}
	var decoded Profile
	if err := json.Unmarshal(content, &decoded); err != nil {
		t.Fatalf("profile file should contain valid JSON: %v", err)
	}
	if !reflect.DeepEqual(decoded.Packages, profile.Packages) {
		t.Fatalf("written profile should preserve package ids, got %#v", decoded.Packages)
	}
}

func TestWriteJSONSurfacesValidationError(t *testing.T) {
	path := filepath.Join(t.TempDir(), DefaultExportPath)
	profile := Profile{
		Version:  Version,
		Name:     DefaultProfileName,
		Packages: []string{"missing-package"},
	}

	if err := WriteJSON(path, profile, testCatalogPackages()); err == nil {
		t.Fatal("WriteJSON should fail before writing invalid profile")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("invalid profile should not be written, stat err=%v", err)
	}
}

func TestWriteJSONOverwritesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), DefaultExportPath)
	if err := os.WriteFile(path, []byte("old content"), 0644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}
	profile := Profile{
		Version:  Version,
		Name:     DefaultProfileName,
		Packages: []string{"firefox"},
	}

	if err := WriteJSON(path, profile, testCatalogPackages()); err != nil {
		t.Fatalf("WriteJSON should overwrite existing file: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read overwritten profile: %v", err)
	}
	if string(content) == "old content" {
		t.Fatal("existing profile file should be overwritten")
	}
}

func testCatalogPackages() []catalog.Package {
	packages := make([]catalog.Package, 0)
	var walk func([]catalog.Category)
	walk = func(categories []catalog.Category) {
		for _, category := range categories {
			walk(category.Categories)
			packages = append(packages, category.Apps...)
		}
	}
	walk(catalog.Default())
	return packages
}
