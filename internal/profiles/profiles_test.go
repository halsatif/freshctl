package profiles

import (
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
