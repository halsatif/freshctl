package detection

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/halsatif/freshctl/internal/catalog"
)

func TestMatchRegistryDisplayName(t *testing.T) {
	if !MatchRegistryDisplayName("Microsoft Visual Studio Code (User)", "Visual Studio Code") {
		t.Fatal("registry display name should match detection value case-insensitively")
	}
	if MatchRegistryDisplayName("Mozilla Firefox", "Google Chrome") {
		t.Fatal("registry display name should not match unrelated detection value")
	}
	if MatchRegistryDisplayName("", "Google Chrome") || MatchRegistryDisplayName("Google Chrome", "") {
		t.Fatal("empty registry match values should not match")
	}
}

func TestDetectPath(t *testing.T) {
	dir := t.TempDir()
	name := "freshctl-detect-test"
	if os.PathSeparator == '\\' {
		name += ".exe"
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(""), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	if !DetectPath(name) {
		t.Fatalf("DetectPath should find %s on PATH", name)
	}
	if DetectPath("freshctl-not-real.exe") {
		t.Fatal("DetectPath should not find missing executable")
	}
}

func TestDetectInstalledWithoutMetadata(t *testing.T) {
	if DetectInstalled(catalog.Package{Name: "No Metadata"}) {
		t.Fatal("package without detection metadata should not be detected")
	}
	if HasDetectionMetadata(catalog.Package{Name: "No Metadata"}) {
		t.Fatal("package without detection metadata should report no metadata")
	}
}

func TestHasDetectionMetadata(t *testing.T) {
	pkg := catalog.Package{
		Name:         "ripgrep",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "rg.exe",
	}
	if !HasDetectionMetadata(pkg) {
		t.Fatal("package with detection method and value should report metadata")
	}

	pkg.DetectValue = ""
	if HasDetectionMetadata(pkg) {
		t.Fatal("package with missing detection value should report no metadata")
	}
}
