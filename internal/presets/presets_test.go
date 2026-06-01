package presets

import (
	"strings"
	"testing"

	"github.com/halsatif/freshctl/internal/catalog"
)

func TestDefaultPresetIDsAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, preset := range Default() {
		id := strings.TrimSpace(preset.ID)
		if id == "" {
			t.Fatal("preset id should not be empty")
		}
		if seen[id] {
			t.Fatalf("duplicate preset id %q", id)
		}
		seen[id] = true
	}
}

func TestDefaultPresetNamesAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, preset := range Default() {
		name := strings.TrimSpace(preset.Name)
		if name == "" {
			t.Fatalf("preset %q should have a name", preset.ID)
		}
		key := strings.ToLower(name)
		if seen[key] {
			t.Fatalf("duplicate preset name %q", name)
		}
		seen[key] = true
	}
}

func TestDefaultPresetsAreNonEmpty(t *testing.T) {
	for _, preset := range Default() {
		if strings.TrimSpace(preset.Description) == "" {
			t.Fatalf("preset %q should have a description", preset.ID)
		}
		if len(preset.Packages) == 0 {
			t.Fatalf("preset %q should include packages", preset.ID)
		}
	}
}

func TestDefaultPresetPackageReferencesExist(t *testing.T) {
	catalogIDs := packageIDs(catalog.Default())
	for _, preset := range Default() {
		for _, packageID := range preset.Packages {
			if strings.TrimSpace(packageID) == "" {
				t.Fatalf("preset %q has an empty package reference", preset.ID)
			}
			if !catalogIDs[packageID] {
				t.Fatalf("preset %q references unknown package %q", preset.ID, packageID)
			}
		}
	}
}

func TestByIDFindsPreset(t *testing.T) {
	preset, ok := ByID("developer")
	if !ok {
		t.Fatal("expected developer preset")
	}
	if preset.Name != "Developer" {
		t.Fatalf("expected Developer preset name, got %q", preset.Name)
	}
}

func TestByIDMissesUnknownPreset(t *testing.T) {
	if preset, ok := ByID("unknown"); ok {
		t.Fatalf("unknown preset should not resolve, got %#v", preset)
	}
}

func packageIDs(categories []catalog.Category) map[string]bool {
	ids := map[string]bool{}
	var walk func([]catalog.Category)
	walk = func(nodes []catalog.Category) {
		for _, node := range nodes {
			walk(node.Categories)
			for _, app := range node.Apps {
				ids[app.PackageID] = true
			}
		}
	}
	walk(categories)
	return ids
}
