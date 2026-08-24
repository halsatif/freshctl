package tui

import (
	"errors"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsatif/freshctl/internal/catalog"
	"github.com/halsatif/freshctl/internal/installer"
	presetpkg "github.com/halsatif/freshctl/internal/presets"
	"github.com/halsatif/freshctl/internal/profiles"
)

func TestCatalogViewRendersSingleCleanScreen(t *testing.T) {
	model := Model{
		screen:      screenCatalog,
		width:       100,
		height:      32,
		categories:  catalog.Default(),
		catalogMode: catalogModeCategories,
		selected:    map[string]bool{},
	}

	view := stripANSI(model.View())

	if !strings.Contains(view, "Browsers >") {
		t.Fatalf("catalog view should render category names, got:\n%s", view)
	}
	if !strings.Contains(view, "Web browsers for everyday") || !strings.Contains(view, "Contains:") {
		t.Fatalf("catalog view should render details panel for highlighted category, got:\n%s", view)
	}
	if strings.Contains(view, "[BR]") || strings.Contains(view, "[PY]") {
		t.Fatalf("catalog view should not render icon tokens, got:\n%s", view)
	}
	if strings.Contains(view, "fresh windows setup, but not painful") {
		t.Fatalf("catalog view should not contain welcome screen content, got:\n%s", view)
	}
	if count := strings.Count(view, "freshctl"); count != 1 {
		t.Fatalf("catalog view should render one title, got %d in:\n%s", count, view)
	}
	if count := strings.Count(view, "↑↓ move"); count != 1 {
		t.Fatalf("catalog view should render one footer, got %d in:\n%s", count, view)
	}
}

func TestCatalogViewHeightStaysStableAcrossNavigation(t *testing.T) {
	root := Model{
		screen:      screenCatalog,
		width:       100,
		height:      32,
		categories:  catalog.Default(),
		catalogMode: catalogModeCategories,
		selected:    map[string]bool{},
	}
	browsers := root
	browsers.catalogPath = []int{0}

	rootLines := strings.Split(root.View(), "\n")
	browserLines := strings.Split(browsers.View(), "\n")
	if len(rootLines) != len(browserLines) {
		t.Fatalf("catalog view line count should stay stable, root=%d browsers=%d", len(rootLines), len(browserLines))
	}
}

func TestCatalogPKeyOpensPresetPicker(t *testing.T) {
	model := Model{
		screen:      screenCatalog,
		width:       100,
		height:      32,
		categories:  catalog.Default(),
		catalogMode: catalogModeFull,
		selected:    map[string]bool{},
	}

	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	got := updated.(Model)

	if got.screen != screenPresetPicker {
		t.Fatalf("p should open preset picker, got screen %v", got.screen)
	}
}

func TestPresetPickerRendersPresets(t *testing.T) {
	model := Model{
		screen:     screenPresetPicker,
		width:      100,
		height:     32,
		categories: catalog.Default(),
		presets: []presetpkg.Preset{{
			ID:          "developer",
			Name:        "Developer",
			Description: "Common tools for coding.",
			Packages:    []string{"vscode", "git"},
		}},
	}

	view := stripANSI(model.View())
	if !strings.Contains(view, "Presets") || !strings.Contains(view, "Developer") || !strings.Contains(view, "2 packages") {
		t.Fatalf("preset picker should render preset name, description, and package count, got:\n%s", view)
	}
	if !strings.Contains(view, "enter apply") || !strings.Contains(view, "esc back") {
		t.Fatalf("preset picker should render footer, got:\n%s", view)
	}
}

func TestPresetPreviewRendersPackageNames(t *testing.T) {
	model := Model{
		screen:     screenPresetPicker,
		width:      100,
		height:     32,
		categories: catalog.Default(),
		presets: []presetpkg.Preset{{
			ID:          "developer",
			Name:        "Developer",
			Description: "Common tools for coding.",
			Packages:    []string{"vscode", "git"},
		}},
	}

	view := stripANSI(model.View())
	if !strings.Contains(view, "Visual Studio Code") || !strings.Contains(view, "Git") {
		t.Fatalf("preset preview should render package names, got:\n%s", view)
	}
	if strings.Contains(view, "vscode") {
		t.Fatalf("preset preview should not render package ids, got:\n%s", view)
	}
}

func TestPresetPreviewChangesWithSelection(t *testing.T) {
	model := Model{
		screen:     screenPresetPicker,
		width:      100,
		height:     32,
		categories: catalog.Default(),
		presets: []presetpkg.Preset{
			{ID: "minimal", Name: "Minimal", Description: "Small setup.", Packages: []string{"firefox"}},
			{ID: "streaming", Name: "Streaming", Description: "Streaming setup.", Packages: []string{"obs-studio"}},
		},
	}

	first := stripANSI(model.View())
	updated, _ := model.handlePresetPickerKey(tea.KeyMsg{Type: tea.KeyDown})
	second := stripANSI(updated.(Model).View())

	if !strings.Contains(first, "Mozilla Firefox") {
		t.Fatalf("first preset preview should show Firefox, got:\n%s", first)
	}
	if !strings.Contains(second, "OBS Studio") {
		t.Fatalf("second preset preview should show OBS Studio, got:\n%s", second)
	}
}

func TestPresetPreviewClipsLongPackageList(t *testing.T) {
	model := Model{
		screen:     screenPresetPicker,
		width:      100,
		height:     18,
		categories: catalog.Default(),
		presets: []presetpkg.Preset{{
			ID:          "large",
			Name:        "Large",
			Description: "Large setup.",
			Packages:    []string{"firefox", "7zip", "everything", "vscode", "git", "vlc", "discord", "steam", "signal", "bitwarden"},
		}},
	}

	view := stripANSI(model.View())
	if !strings.Contains(view, "...and") {
		t.Fatalf("long preset preview should show clipped package count, got:\n%s", view)
	}
	if !strings.Contains(view, "? help") {
		t.Fatalf("preset picker footer should remain visible, got:\n%s", view)
	}
}

func TestPresetPreviewIgnoresMissingPackageReferences(t *testing.T) {
	model := Model{
		screen:     screenPresetPicker,
		width:      100,
		height:     32,
		categories: catalog.Default(),
		presets: []presetpkg.Preset{{
			ID:          "partial",
			Name:        "Partial",
			Description: "Partial setup.",
			Packages:    []string{"firefox", "missing-package-id"},
		}},
	}

	view := stripANSI(model.View())
	if !strings.Contains(view, "Mozilla Firefox") {
		t.Fatalf("preview should render valid package refs, got:\n%s", view)
	}
	if strings.Contains(view, "missing-package-id") {
		t.Fatalf("preview should skip missing package refs, got:\n%s", view)
	}
}

func TestPresetPickerEnterAppliesPreset(t *testing.T) {
	model := Model{
		screen:       screenPresetPicker,
		width:        100,
		height:       32,
		categories:   catalog.Default(),
		selected:     map[string]bool{},
		presetCursor: 0,
		presets: []presetpkg.Preset{{
			ID:          "minimal",
			Name:        "Minimal",
			Description: "Small setup.",
			Packages:    []string{"firefox", "7zip", "everything"},
		}},
	}

	updated, _ := model.handlePresetPickerKey(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(Model)

	if got.screen != screenCatalog {
		t.Fatalf("applying preset should return to catalog, got screen %v", got.screen)
	}
	for _, id := range []string{"firefox", "7zip", "everything"} {
		if !got.selected[id] {
			t.Fatalf("preset package %q should be selected", id)
		}
	}
	if got.appliedPreset != "Minimal" {
		t.Fatalf("applied preset header should remember preset name, got %q", got.appliedPreset)
	}
}

func TestPresetApplyPreservesManualSelections(t *testing.T) {
	model := Model{
		screen:     screenPresetPicker,
		categories: catalog.Default(),
		selected:   map[string]bool{"vlc": true},
		presets: []presetpkg.Preset{{
			ID:       "privacy",
			Name:     "Privacy",
			Packages: []string{"firefox", "signal"},
		}},
	}

	updated, _ := model.handlePresetPickerKey(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(Model)

	if !got.selected["vlc"] || !got.selected["firefox"] || !got.selected["signal"] {
		t.Fatalf("preset should preserve existing selections and add preset packages, got %#v", got.selected)
	}
}

func TestPresetPickerEscReturnsWithoutApplying(t *testing.T) {
	model := Model{
		screen:     screenPresetPicker,
		categories: catalog.Default(),
		selected:   map[string]bool{},
		presets: []presetpkg.Preset{{
			ID:       "minimal",
			Name:     "Minimal",
			Packages: []string{"firefox"},
		}},
	}

	updated, _ := model.handlePresetPickerKey(tea.KeyMsg{Type: tea.KeyEsc})
	got := updated.(Model)

	if got.screen != screenCatalog {
		t.Fatalf("esc should return to catalog, got screen %v", got.screen)
	}
	if got.selected["firefox"] {
		t.Fatal("esc should not apply preset")
	}
}

func TestPresetPickerQQuits(t *testing.T) {
	model := Model{screen: screenPresetPicker}

	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("q should return quit command from preset picker")
	}
}

func TestPresetApplyIgnoresMissingPackageReferences(t *testing.T) {
	model := Model{
		screen:     screenPresetPicker,
		categories: catalog.Default(),
		selected:   map[string]bool{},
		presets: []presetpkg.Preset{{
			ID:       "broken",
			Name:     "Broken",
			Packages: []string{"firefox", "missing-package-id"},
		}},
	}

	updated, _ := model.handlePresetPickerKey(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(Model)

	if !got.selected["firefox"] {
		t.Fatal("valid preset package should be selected")
	}
	if got.selected["missing-package-id"] {
		t.Fatal("missing preset package reference should be ignored")
	}
}

func TestPresetApplyUpdatesSelectedCountAndClearsSearch(t *testing.T) {
	model := Model{
		screen:        screenPresetPicker,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		searchFocused: false,
		searchQuery:   "fire",
		selected:      map[string]bool{},
		presets: []presetpkg.Preset{{
			ID:       "minimal",
			Name:     "Minimal",
			Packages: []string{"firefox", "7zip", "everything"},
		}},
	}

	updated, _ := model.handlePresetPickerKey(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(Model)
	view := stripANSI(got.View())

	if got.searchFocused || got.searchQuery != "" {
		t.Fatalf("applying preset should clear search, focused=%v query=%q", got.searchFocused, got.searchQuery)
	}
	if !strings.Contains(view, "3 selected") || !strings.Contains(view, "Preset: Minimal") {
		t.Fatalf("catalog should show updated selected count and preset header, got:\n%s", view)
	}
}

func TestReviewShowsAppliedPreset(t *testing.T) {
	model := Model{
		screen:        screenReview,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		selected:      map[string]bool{"firefox": true},
		appliedPreset: "Minimal",
	}

	view := stripANSI(model.View())
	if !strings.Contains(view, "Preset: Minimal") {
		t.Fatalf("review should show applied preset, got:\n%s", view)
	}
}

func TestReviewHidesPresetLineWhenNoneApplied(t *testing.T) {
	model := Model{
		screen:     screenReview,
		width:      100,
		height:     32,
		categories: catalog.Default(),
		selected:   map[string]bool{"firefox": true},
	}

	view := stripANSI(model.View())
	if strings.Contains(view, "Preset:") {
		t.Fatalf("review should not show preset line without applied preset, got:\n%s", view)
	}
}

func TestReviewExportsSelectedPackagesToProfile(t *testing.T) {
	var gotPath string
	var gotProfile profiles.Profile
	model := Model{
		screen:     screenReview,
		width:      100,
		height:     32,
		categories: catalog.Default(),
		selected:   map[string]bool{"firefox": true, "git": true, "vscode": true},
		exportProfile: func(path string, profile profiles.Profile, packages []catalog.Application) error {
			gotPath = path
			gotProfile = profile
			return profiles.Validate(profile, packages)
		},
	}

	_, cmd := model.handleReviewKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if cmd == nil {
		t.Fatal("profile export should return a command")
	}
	msg := cmd().(profileExportMsg)
	updated, _ := model.handleProfileExportMsg(msg)
	got := updated.(Model)

	if gotPath != profiles.DefaultExportPath {
		t.Fatalf("profile should export to default path, got %q", gotPath)
	}
	if gotProfile.Version != profiles.Version || gotProfile.Name != profiles.DefaultProfileName {
		t.Fatalf("exported profile should use default metadata, got %#v", gotProfile)
	}
	want := []string{"firefox", "vscode", "git"}
	if !reflect.DeepEqual(gotProfile.Packages, want) {
		t.Fatalf("exported package ids should match review order, got %#v", gotProfile.Packages)
	}
	if !strings.Contains(got.notice, "Profile exported: freshctl-profile.json") {
		t.Fatalf("successful export should show compact notice, got %q", got.notice)
	}
}

func TestReviewExportWithNoSelectedPackagesIsBlocked(t *testing.T) {
	called := false
	model := Model{
		screen:     screenReview,
		categories: catalog.Default(),
		selected:   map[string]bool{},
		exportProfile: func(string, profiles.Profile, []catalog.Application) error {
			called = true
			return nil
		},
	}

	updated, cmd := model.handleReviewKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	got := updated.(Model)

	if cmd != nil {
		t.Fatal("empty selection should not start profile export command")
	}
	if called {
		t.Fatal("empty selection should not call profile exporter")
	}
	if !strings.Contains(got.notice, "No apps selected") {
		t.Fatalf("empty selection should show friendly notice, got %q", got.notice)
	}
}

func TestReviewProfileExportValidationErrorIsSurfaced(t *testing.T) {
	model := Model{
		screen:     screenReview,
		categories: catalog.Default(),
		selected:   map[string]bool{"firefox": true},
		exportProfile: func(string, profiles.Profile, []catalog.Application) error {
			return errors.New("validation failed")
		},
	}

	_, cmd := model.handleReviewKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if cmd == nil {
		t.Fatal("profile export should return command")
	}
	msg := cmd().(profileExportMsg)
	updated, _ := model.handleProfileExportMsg(msg)
	got := updated.(Model)

	if !strings.Contains(got.notice, "Profile export failed: validation failed") {
		t.Fatalf("validation error should be surfaced, got %q", got.notice)
	}
}

func TestCatalogOKeyImportsProfile(t *testing.T) {
	var gotPath string
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		searchFocused: false,
		searchQuery:   "fire",
		selected:      map[string]bool{},
		importProfile: func(path string, packages []catalog.Application) (profiles.Profile, error) {
			gotPath = path
			return profiles.Profile{
				Version:  profiles.Version,
				Name:     "freshctl profile",
				Packages: []string{"firefox", "git"},
			}, nil
		},
	}

	updated, cmd := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	if cmd == nil {
		t.Fatal("profile import should return a command")
	}
	importing := updated.(Model)
	if importing.notice != "Importing profile..." {
		t.Fatalf("catalog should show importing notice, got %q", importing.notice)
	}
	msg := cmd().(profileImportMsg)
	updated, _ = importing.handleProfileImportMsg(msg)
	got := updated.(Model)

	if gotPath != profiles.DefaultExportPath {
		t.Fatalf("profile should import from default path, got %q", gotPath)
	}
	if !got.selected["firefox"] || !got.selected["git"] {
		t.Fatalf("profile packages should become selected, got %#v", got.selected)
	}
	if got.searchFocused || got.searchQuery != "" {
		t.Fatalf("profile import should clear search, focused=%v query=%q", got.searchFocused, got.searchQuery)
	}
	if got.appliedProfile != "freshctl profile" || got.appliedPreset != "" {
		t.Fatalf("profile should become current source context, profile=%q preset=%q", got.appliedProfile, got.appliedPreset)
	}
}

func TestProfileImportPreservesExistingSelections(t *testing.T) {
	model := Model{
		screen:     screenCatalog,
		categories: catalog.Default(),
		selected:   map[string]bool{"vlc": true},
		importProfile: func(string, []catalog.Application) (profiles.Profile, error) {
			return profiles.Profile{
				Version:  profiles.Version,
				Name:     "freshctl profile",
				Packages: []string{"firefox"},
			}, nil
		},
	}

	_, cmd := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	msg := cmd().(profileImportMsg)
	updated, _ := model.handleProfileImportMsg(msg)
	got := updated.(Model)

	if !got.selected["vlc"] || !got.selected["firefox"] {
		t.Fatalf("profile import should preserve existing selections and add packages, got %#v", got.selected)
	}
}

func TestProfileImportUsesDefaultLabelForEmptyName(t *testing.T) {
	model := Model{
		screen:     screenCatalog,
		categories: catalog.Default(),
		selected:   map[string]bool{},
		importProfile: func(string, []catalog.Application) (profiles.Profile, error) {
			return profiles.Profile{
				Version:  profiles.Version,
				Packages: []string{"firefox"},
			}, nil
		},
	}

	_, cmd := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	msg := cmd().(profileImportMsg)
	updated, _ := model.handleProfileImportMsg(msg)
	got := updated.(Model)

	if got.appliedProfile != profiles.DefaultExportPath {
		t.Fatalf("empty profile name should use default path label, got %q", got.appliedProfile)
	}
}

func TestProfileLabelAppearsInCatalogReviewAndInstall(t *testing.T) {
	app := catalog.Application{Name: "Firefox", ID: "firefox"}
	model := Model{
		screen:         screenCatalog,
		width:          100,
		height:         32,
		categories:     catalog.Default(),
		catalogMode:    catalogModeFull,
		selected:       map[string]bool{"firefox": true},
		appliedProfile: "freshctl profile",
	}

	catalogView := stripANSI(model.View())
	model.screen = screenReview
	reviewView := stripANSI(model.View())
	model.screen = screenInstall
	model.installApps = []catalog.Application{app}
	model.appStatus = map[string]string{"firefox": "pending"}
	model.appElapsed = map[string]time.Duration{}
	installView := stripANSI(model.View())

	for name, view := range map[string]string{
		"catalog": catalogView,
		"review":  reviewView,
		"install": installView,
	} {
		if !strings.Contains(view, "Profile: freshctl profile") {
			t.Fatalf("%s view should show profile label, got:\n%s", name, view)
		}
	}
}

func TestProfileImportErrorsAreSurfaced(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{name: "missing file", err: errors.New("freshctl-profile.json not found. Export a profile from review with e first"), want: "Profile import failed: freshctl-profile.json not found. Export a profile from review with e first"},
		{name: "invalid json", err: errors.New("invalid JSON"), want: "Profile import failed: invalid JSON"},
		{name: "validation", err: errors.New(`unknown package id "missing-package"`), want: `Profile import failed: unknown package id "missing-package"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			model := Model{
				screen:     screenCatalog,
				categories: catalog.Default(),
				selected:   map[string]bool{},
				importProfile: func(string, []catalog.Application) (profiles.Profile, error) {
					return profiles.Profile{}, tc.err
				},
			}

			_, cmd := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
			msg := cmd().(profileImportMsg)
			updated, _ := model.handleProfileImportMsg(msg)
			got := updated.(Model)

			if got.notice != tc.want {
				t.Fatalf("wrong import error notice, got %q want %q", got.notice, tc.want)
			}
		})
	}
}

func TestPresetAndProfileContextSwitching(t *testing.T) {
	model := Model{
		screen:     screenCatalog,
		categories: catalog.Default(),
		selected:   map[string]bool{},
		presets: []presetpkg.Preset{{
			ID:       "minimal",
			Name:     "Minimal",
			Packages: []string{"firefox"},
		}},
		importProfile: func(string, []catalog.Application) (profiles.Profile, error) {
			return profiles.Profile{
				Version:  profiles.Version,
				Name:     "freshctl profile",
				Packages: []string{"git"},
			}, nil
		},
	}

	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	picker := updated.(Model)
	updated, _ = picker.handlePresetPickerKey(tea.KeyMsg{Type: tea.KeyEnter})
	afterPreset := updated.(Model)
	if afterPreset.appliedPreset != "Minimal" || afterPreset.appliedProfile != "" {
		t.Fatalf("preset should become current context, profile=%q preset=%q", afterPreset.appliedProfile, afterPreset.appliedPreset)
	}

	_, cmd := afterPreset.handleCatalogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	msg := cmd().(profileImportMsg)
	updated, _ = afterPreset.handleProfileImportMsg(msg)
	afterProfile := updated.(Model)
	if afterProfile.appliedProfile != "freshctl profile" || afterProfile.appliedPreset != "" {
		t.Fatalf("profile import should replace preset context, profile=%q preset=%q", afterProfile.appliedProfile, afterProfile.appliedPreset)
	}

	updated, _ = afterProfile.handleCatalogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	picker = updated.(Model)
	updated, _ = picker.handlePresetPickerKey(tea.KeyMsg{Type: tea.KeyEnter})
	afterSecondPreset := updated.(Model)
	if afterSecondPreset.appliedPreset != "Minimal" || afterSecondPreset.appliedProfile != "" {
		t.Fatalf("preset apply should replace profile context, profile=%q preset=%q", afterSecondPreset.appliedProfile, afterSecondPreset.appliedPreset)
	}
	if !afterSecondPreset.selected["firefox"] || !afterSecondPreset.selected["git"] {
		t.Fatalf("context switching should preserve additive selections, got %#v", afterSecondPreset.selected)
	}
}

func TestInstallPlanShowsAppliedPreset(t *testing.T) {
	app := catalog.Application{Name: "Firefox", ID: "firefox"}
	model := Model{
		screen:        screenInstall,
		width:         100,
		height:        24,
		installApps:   []catalog.Application{app},
		appStatus:     map[string]string{"firefox": "pending"},
		appElapsed:    map[string]time.Duration{},
		appliedPreset: "Minimal",
	}

	view := stripANSI(model.View())
	if !strings.Contains(view, "Preset: Minimal") || !strings.Contains(view, "Selected: 1") {
		t.Fatalf("install plan should show applied preset and counts, got:\n%s", view)
	}
}

func TestInstallPlanHidesPresetLineWhenNoneApplied(t *testing.T) {
	app := catalog.Application{Name: "Firefox", ID: "firefox"}
	model := Model{
		screen:      screenInstall,
		width:       100,
		height:      24,
		installApps: []catalog.Application{app},
		appStatus:   map[string]string{"firefox": "pending"},
		appElapsed:  map[string]time.Duration{},
	}

	view := stripANSI(model.View())
	if strings.Contains(view, "Preset:") {
		t.Fatalf("install plan should not show preset line without applied preset, got:\n%s", view)
	}
}

func TestApplyingSecondPresetUpdatesDisplayedPreset(t *testing.T) {
	model := Model{
		screen:       screenPresetPicker,
		width:        100,
		height:       32,
		categories:   catalog.Default(),
		selected:     map[string]bool{},
		presetCursor: 0,
		presets: []presetpkg.Preset{
			{ID: "minimal", Name: "Minimal", Packages: []string{"firefox"}},
			{ID: "privacy", Name: "Privacy", Packages: []string{"signal"}},
		},
	}

	updated, _ := model.handlePresetPickerKey(tea.KeyMsg{Type: tea.KeyEnter})
	first := updated.(Model)
	first.screen = screenPresetPicker
	first.presetCursor = 1
	updated, _ = first.handlePresetPickerKey(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(Model)

	if got.appliedPreset != "Privacy" {
		t.Fatalf("second preset should update applied preset name, got %q", got.appliedPreset)
	}
	if !got.selected["firefox"] || !got.selected["signal"] {
		t.Fatalf("second preset should preserve additive selections, got %#v", got.selected)
	}
}

func TestManualSelectionAfterPresetKeepsPresetLabel(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		selected:      map[string]bool{"firefox": true},
		appliedPreset: "Minimal",
	}
	items := model.filteredFullCatalogItems()
	for i, item := range items {
		if item.Package.ID == "vlc" {
			model.catalogCursor = i
			break
		}
	}

	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeySpace})
	got := updated.(Model)

	if !got.selected["vlc"] {
		t.Fatal("manual selection should still toggle package")
	}
	if got.appliedPreset != "Minimal" {
		t.Fatalf("manual selection should not clear applied preset label, got %q", got.appliedPreset)
	}
}

func TestEscAfterPresetApplyStaysInCatalogAndClearsPresetLayer(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		selected:      map[string]bool{"firefox": true},
		appliedPreset: "Minimal",
	}

	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyEsc})
	got := updated.(Model)

	if got.screen != screenCatalog {
		t.Fatalf("esc after preset apply should stay in catalog, got screen %v", got.screen)
	}
	if got.appliedPreset != "" {
		t.Fatalf("esc should clear preset layer before leaving catalog, got %q", got.appliedPreset)
	}
	if !got.selected["firefox"] {
		t.Fatal("clearing preset layer should not clear selected packages")
	}
}

func TestEscAfterProfileImportStaysInCatalogAndClearsProfileLayer(t *testing.T) {
	model := Model{
		screen:         screenCatalog,
		width:          100,
		height:         32,
		categories:     catalog.Default(),
		catalogMode:    catalogModeFull,
		selected:       map[string]bool{"git": true},
		appliedProfile: "freshctl profile",
	}

	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyEsc})
	got := updated.(Model)

	if got.screen != screenCatalog {
		t.Fatalf("esc after profile import should stay in catalog, got screen %v", got.screen)
	}
	if got.appliedProfile != "" {
		t.Fatalf("esc should clear profile layer before leaving catalog, got %q", got.appliedProfile)
	}
	if !got.selected["git"] {
		t.Fatal("clearing profile layer should not clear selected packages")
	}
}

func TestCatalogSearchPanelHeightStaysStableWithShortResults(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		searchFocused: true,
		searchQuery:   "discord",
		selected:      map[string]bool{},
	}

	view := stripANSI(model.View())
	top, bottom := catalogPanelBorderRows(t, view)
	if got := bottom - top; got != model.catalogPanelHeight()+1 {
		t.Fatalf("catalog panel height should stay fixed for short search results, got border distance %d, want %d\n%s", got, model.catalogPanelHeight()+1, view)
	}
}

func TestCatalogSearchPanelHeightStaysStableWithEmptyResults(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		searchFocused: true,
		searchQuery:   "definitely-not-a-package",
		selected:      map[string]bool{},
	}

	view := stripANSI(model.View())
	top, bottom := catalogPanelBorderRows(t, view)
	if got := bottom - top; got != model.catalogPanelHeight()+1 {
		t.Fatalf("catalog panel height should stay fixed for empty search results, got border distance %d, want %d\n%s", got, model.catalogPanelHeight()+1, view)
	}
}

func TestCatalogSearchPanelsStayAligned(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		searchFocused: true,
		searchQuery:   "zzzzzz",
		selected:      map[string]bool{},
	}

	view := stripANSI(model.View())
	top, bottom := catalogPanelBorderRows(t, view)
	lines := strings.Split(view, "\n")
	for _, row := range []int{top, bottom} {
		if count := strings.Count(lines[row], "+"); count < 4 {
			t.Fatalf("left and right panel borders should share row %d, got %d plus signs in %q\n%s", row, count, lines[row], view)
		}
	}
}

func TestCatalogBreadcrumbIncludesRoot(t *testing.T) {
	model := Model{
		categories:  catalog.Default(),
		catalogPath: []int{3, 1},
	}

	if got := model.currentBreadcrumb(); got != "Catalog > Media > Images & Graphics" {
		t.Fatalf("breadcrumb should include catalog root, got %q", got)
	}
}

func TestRussianKeyboardAliasesWorkForGlobalQuit(t *testing.T) {
	model := Model{
		screen:   screenCatalog,
		selected: map[string]bool{},
	}

	_, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'й'}})
	if cmd == nil {
		t.Fatal("russian q key alias should quit")
	}
}

func TestRussianKeyboardAliasesWorkForCatalogNavigation(t *testing.T) {
	model := Model{
		screen:      screenCatalog,
		width:       100,
		height:      32,
		categories:  catalog.Default(),
		catalogMode: catalogModeFull,
		selected:    map[string]bool{},
	}

	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'о'}})
	got := updated.(Model)
	if got.catalogCursor != 1 {
		t.Fatalf("russian j key alias should move down, got cursor %d", got.catalogCursor)
	}

	updated, _ = got.handleCatalogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'л'}})
	got = updated.(Model)
	if got.catalogCursor != 0 {
		t.Fatalf("russian k key alias should move up, got cursor %d", got.catalogCursor)
	}
}

func TestRussianKeyboardAliasesWorkForCatalogActions(t *testing.T) {
	model := Model{
		screen:      screenCatalog,
		width:       100,
		height:      32,
		categories:  catalog.Default(),
		catalogMode: catalogModeFull,
		selected:    map[string]bool{},
	}

	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
	got := updated.(Model)
	if !got.searchFocused {
		t.Fatal("russian slash key alias should focus search")
	}

	got.searchFocused = false
	got.searchQuery = ""
	updated, _ = got.handleCatalogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'ш'}})
	got = updated.(Model)
	if got.screen != screenReview {
		t.Fatalf("russian i key alias should open review screen, got screen %v", got.screen)
	}
}

func TestPackageDetailsPanelShowsMetadata(t *testing.T) {
	app, ok := packagesByIDForTUITest(catalog.Default())["vscode"]
	if !ok {
		t.Fatal("expected Visual Studio Code in default catalog")
	}

	view := stripANSI(fitDetailsLines(packageDetailsLines(app, "No", ""), 40, 18))
	for _, want := range []string{
		"Package:",
		"Visual Studio Code",
		"ID:",
		"vscode",
		"Type:",
		"Application",
		"Manager:",
		"Direct",
		"Verified:",
		"Yes",
		"Description:",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("details panel should contain %q, got:\n%s", want, view)
		}
	}
	if strings.Contains(view, "Package:\nvscode") {
		t.Fatalf("details panel should show human-readable name under Package, got:\n%s", view)
	}
	if !strings.Contains(view, "Package:\nVisual Studio Code") || !strings.Contains(view, "ID:\nvscode") {
		t.Fatalf("details panel should show name and id separately, got:\n%s", view)
	}
}

func TestPackageDetailsPanelShowsCLIToolMetadata(t *testing.T) {
	apps := packagesByIDForTUITest(catalog.Default())
	app, ok := apps["helix"]
	if !ok {
		t.Fatal("expected Helix in default catalog")
	}

	view := stripANSI(fitDetailsLines(packageDetailsLines(app, "No", ""), 44, 18))
	if !strings.Contains(view, "CLI Tool") {
		t.Fatalf("CLI package should render CLI Tool type, got:\n%s", view)
	}
	if !strings.Contains(view, "hx") {
		t.Fatalf("Helix description should mention hx command, got:\n%s", view)
	}
}

func TestPackageDetailsPanelShowsInstalledStatusWhenDetectionExists(t *testing.T) {
	app := catalog.Application{
		Name:         "Missing CLI",
		Description:  "Test command-line tool.",
		ID:           "missing-cli",
		Type:         catalog.PackageTypeCLITool,
		Providers:    testChocolateyProviders("missing-cli"),
		DetectMethod: catalog.DetectPath,
		DetectValue:  "freshctl-definitely-not-installed.exe",
		Verified:     true,
	}

	view := stripANSI(fitDetailsLines(packageDetailsLines(app, "No", "No"), 44, 18))
	if !strings.Contains(view, "Installed: No") {
		t.Fatalf("details panel should show installed status when detection metadata exists, got:\n%s", view)
	}
}

func TestPackageDetailsPanelHidesInstalledStatusWithoutDetection(t *testing.T) {
	app := catalog.Application{
		Name:        "No Detection",
		Description: "Package without detection metadata.",
		ID:          "no-detection",
		Type:        catalog.PackageTypeApplication,
		Providers:   testChocolateyProviders("no-detection"),
		Verified:    true,
	}

	view := stripANSI(fitDetailsLines(packageDetailsLines(app, "No", ""), 44, 18))
	if strings.Contains(view, "Installed:") {
		t.Fatalf("details panel should hide installed status without detection metadata, got:\n%s", view)
	}
}

func TestInstalledStatusCachePopulatesDetectedPackages(t *testing.T) {
	app := catalog.Application{
		Name:         "Cached Tool",
		ID:           "cached-tool",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "cached-tool.exe",
	}
	model := Model{
		categories: []catalog.Category{{Apps: []catalog.Application{app}}},
		detectInstalled: func(pkg catalog.Application) bool {
			return pkg.ID == "cached-tool"
		},
	}

	model.RefreshInstalledStatus()
	status, ok := model.installed["cached-tool"]
	if !ok || !status.Checked || !status.Installed {
		t.Fatalf("refresh should populate checked installed status, got %#v ok=%v", status, ok)
	}
}

func TestInstalledStatusCacheSkipsPackagesWithoutDetectionMetadata(t *testing.T) {
	model := Model{
		categories: []catalog.Category{{Apps: []catalog.Application{{
			Name: "No Detection",
			ID:   "no-detection",
		}}}},
		detectInstalled: func(catalog.Application) bool {
			t.Fatal("detector should not be called for package without detection metadata")
			return true
		},
	}

	model.RefreshInstalledStatus()
	if _, ok := model.installed["no-detection"]; ok {
		t.Fatal("package without detection metadata should not be cached")
	}
}

func TestInstalledStatusRefreshUpdatesCache(t *testing.T) {
	app := catalog.Application{
		Name:         "Refresh Tool",
		ID:           "refresh-tool",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "refresh-tool.exe",
	}
	installed := false
	model := Model{
		categories: []catalog.Category{{Apps: []catalog.Application{app}}},
		detectInstalled: func(catalog.Application) bool {
			return installed
		},
	}

	model.RefreshInstalledStatus()
	if model.installed["refresh-tool"].Installed {
		t.Fatal("first refresh should cache not installed")
	}

	installed = true
	model.RefreshInstalledStatus()
	if !model.installed["refresh-tool"].Installed {
		t.Fatal("second refresh should update cached installed status")
	}
}

func TestNewModelScansInstalledStatusAtStartup(t *testing.T) {
	model := NewModel(nil)
	status, ok := model.installed["googlechrome"]
	if !ok || !status.Checked {
		t.Fatalf("NewModel should populate installed status cache for packages with detection metadata, got %#v ok=%v", status, ok)
	}
}

func TestApplicationsUseProvider(t *testing.T) {
	directApp := catalog.Application{
		ID: "direct-app",
		Providers: []catalog.Provider{
			{
				Type:      catalog.ProviderDirect,
				PackageID: "direct-app",
				Strategy:  catalog.InstallStrategyDirectInstaller,
			},
			{
				Type:      catalog.ProviderChocolatey,
				PackageID: "direct-app",
				Strategy:  catalog.InstallStrategyPackageManager,
			},
		},
	}
	chocolateyApp := catalog.Application{
		ID:        "chocolatey-app",
		Providers: testChocolateyProviders("chocolatey-app"),
	}

	if applicationsUseProvider([]catalog.Application{directApp}, catalog.ProviderChocolatey) {
		t.Fatal("Direct-primary selection should not require its secondary Chocolatey provider")
	}
	if !applicationsUseProvider([]catalog.Application{directApp, chocolateyApp}, catalog.ProviderChocolatey) {
		t.Fatal("mixed selection should require Chocolatey")
	}
}

func TestNewModelDoesNotRequireChocolateyBeforeBrowsing(t *testing.T) {
	model := NewModel(nil)
	if model.screen != screenWelcome {
		t.Fatalf("fresh startup should open welcome before provider checks, got screen %v", model.screen)
	}

	selectedModel := NewModel([]string{"--selected=vscode"})
	if selectedModel.screen != screenReview {
		t.Fatalf("elevated relaunch selection should return to review, got screen %v", selectedModel.screen)
	}
}

func TestDetailsPanelUsesCachedInstalledStatus(t *testing.T) {
	app := catalog.Application{
		Name:         "Cached Missing Tool",
		Description:  "Tool that should read installed state from cache.",
		ID:           "cached-missing-tool",
		Type:         catalog.PackageTypeCLITool,
		Providers:    testChocolateyProviders("cached-missing-tool"),
		DetectMethod: catalog.DetectPath,
		DetectValue:  "freshctl-definitely-not-installed.exe",
		Verified:     true,
	}
	model := Model{
		screen:      screenCatalog,
		width:       100,
		height:      32,
		categories:  []catalog.Category{{Apps: []catalog.Application{app}}},
		catalogMode: catalogModeFull,
		selected:    map[string]bool{},
		installed: map[string]InstalledStatus{
			"cached-missing-tool": {Installed: true, Checked: true},
		},
	}

	view := stripANSI(model.View())
	if !strings.Contains(view, "Installed: Yes") {
		t.Fatalf("details panel should use cached installed status, got:\n%s", view)
	}
}

func TestCatalogListShowsInstalledStatusFromCache(t *testing.T) {
	app := catalog.Application{
		Name:         "Cached Installed",
		ID:           "cached-installed",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "cached-installed.exe",
	}
	model := Model{
		categories:  []catalog.Category{{Apps: []catalog.Application{app}}},
		catalogMode: catalogModeFull,
		selected:    map[string]bool{},
		installed: map[string]InstalledStatus{
			"cached-installed": {Installed: true, Checked: true},
		},
	}

	view := stripANSI(strings.Join(model.fullCatalogListLines(64), "\n"))
	if !strings.Contains(view, "Cached Installed") || !strings.Contains(view, "OK") {
		t.Fatalf("installed package row should show Installed from cache, got:\n%s", view)
	}
}

func TestCatalogListShowsNotInstalledStatusFromCache(t *testing.T) {
	app := catalog.Application{
		Name:         "Cached Missing",
		ID:           "cached-missing",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "cached-missing.exe",
	}
	model := Model{
		categories:  []catalog.Category{{Apps: []catalog.Application{app}}},
		catalogMode: catalogModeFull,
		selected:    map[string]bool{},
		installed: map[string]InstalledStatus{
			"cached-missing": {Installed: false, Checked: true},
		},
	}

	view := stripANSI(strings.Join(model.fullCatalogListLines(64), "\n"))
	if !strings.Contains(view, "Cached Missing") || !strings.Contains(view, "--") {
		t.Fatalf("not installed package row should show Not installed from cache, got:\n%s", view)
	}
}

func TestCatalogListHidesStatusWithoutDetectionMetadata(t *testing.T) {
	app := catalog.Application{
		Name: "No Detection",
		ID:   "no-detection",
	}
	model := Model{
		categories:  []catalog.Category{{Apps: []catalog.Application{app}}},
		catalogMode: catalogModeFull,
		selected:    map[string]bool{},
		installed: map[string]InstalledStatus{
			"no-detection": {Installed: true, Checked: true},
		},
	}

	view := stripANSI(strings.Join(model.fullCatalogListLines(64), "\n"))
	if strings.Contains(view, "OK") || strings.Contains(view, "--") {
		t.Fatalf("package without detection metadata should not show installed status, got:\n%s", view)
	}
}

func TestCatalogSearchResultsShowInstalledStatus(t *testing.T) {
	app := catalog.Application{
		Name:         "Search Installed Tool",
		Description:  "Searchable tool.",
		ID:           "search-installed-tool",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "search-installed-tool.exe",
	}
	model := Model{
		categories:    []catalog.Category{{Apps: []catalog.Application{app}}},
		catalogMode:   catalogModeFull,
		searchFocused: true,
		searchQuery:   "search",
		selected:      map[string]bool{},
		installed: map[string]InstalledStatus{
			"search-installed-tool": {Installed: true, Checked: true},
		},
	}

	view := stripANSI(strings.Join(model.catalogListLines(64), "\n"))
	if !strings.Contains(view, "Search Installed Tool") || !strings.Contains(view, "OK") {
		t.Fatalf("search result row should show installed status from cache, got:\n%s", view)
	}
}

func TestCatalogListRenderDoesNotCallDetection(t *testing.T) {
	app := catalog.Application{
		Name:         "Render Cached Tool",
		Description:  "Cached render test.",
		ID:           "render-cached-tool",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "render-cached-tool.exe",
	}
	model := Model{
		screen:      screenCatalog,
		width:       100,
		height:      32,
		categories:  []catalog.Category{{Apps: []catalog.Application{app}}},
		catalogMode: catalogModeFull,
		selected:    map[string]bool{},
		installed: map[string]InstalledStatus{
			"render-cached-tool": {Installed: true, Checked: true},
		},
		detectInstalled: func(catalog.Application) bool {
			t.Fatal("render should not call installed detection")
			return false
		},
	}

	view := stripANSI(model.View())
	if !strings.Contains(view, "OK") {
		t.Fatalf("render should show cached installed status, got:\n%s", view)
	}
}

func TestInstallSummaryRefreshesInstalledStatusCacheOnce(t *testing.T) {
	app := catalog.Application{
		Name:         "Refresh Once Tool",
		Description:  "Refresh once test.",
		ID:           "refresh-once-tool",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "refresh-once-tool.exe",
	}
	calls := 0
	model := Model{
		screen:      screenInstall,
		categories:  []catalog.Category{{Apps: []catalog.Application{app}}},
		installApps: []catalog.Application{app},
		appStatus:   map[string]string{"refresh-once-tool": "installed"},
		selected:    map[string]bool{},
		installed: map[string]InstalledStatus{
			"refresh-once-tool": {Installed: false, Checked: true},
		},
		detectInstalled: func(catalog.Application) bool {
			calls++
			return true
		},
	}

	updated, _ := model.handleInstallEvent(installEventMsg{
		ok: true,
		event: installer.Event{
			Kind:    installer.EventSummary,
			Results: []installer.Result{{App: app, Success: true}},
		},
	})
	got := updated.(Model)
	if calls != 1 {
		t.Fatalf("install summary should refresh installed status once, got %d calls", calls)
	}
	if !got.installed["refresh-once-tool"].Installed {
		t.Fatalf("install summary should refresh cache to installed, got %#v", got.installed["refresh-once-tool"])
	}

	updated, _ = got.handleInstallEvent(installEventMsg{
		ok: true,
		event: installer.Event{
			Kind:    installer.EventSummary,
			Results: []installer.Result{{App: app, Success: true}},
		},
	})
	if calls != 1 {
		t.Fatalf("second install summary should not refresh cache again, got %d calls", calls)
	}
}

func TestSuccessfulInstallUpdatesStatusWhenDetectable(t *testing.T) {
	app := catalog.Application{
		Name:         "Detectable Success",
		Description:  "Detectable success test.",
		ID:           "detectable-success",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "detectable-success.exe",
	}
	model := Model{
		screen:      screenInstall,
		categories:  []catalog.Category{{Apps: []catalog.Application{app}}},
		installApps: []catalog.Application{app},
		appStatus:   map[string]string{"detectable-success": "installed"},
		selected:    map[string]bool{},
		installed: map[string]InstalledStatus{
			"detectable-success": {Installed: false, Checked: true},
		},
		detectInstalled: func(catalog.Application) bool {
			return true
		},
	}

	updated, _ := model.handleInstallEvent(installEventMsg{
		ok: true,
		event: installer.Event{
			Kind:    installer.EventSummary,
			Results: []installer.Result{{App: app, Success: true}},
		},
	})
	got := updated.(Model)
	if !got.installed["detectable-success"].Installed {
		t.Fatalf("successful install should use detection refresh result, got %#v", got.installed["detectable-success"])
	}
}

func TestFailedInstallDoesNotFakeInstalledStatus(t *testing.T) {
	app := catalog.Application{
		Name:         "Detectable Failure",
		Description:  "Detectable failure test.",
		ID:           "detectable-failure",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "detectable-failure.exe",
	}
	model := Model{
		screen:      screenInstall,
		categories:  []catalog.Category{{Apps: []catalog.Application{app}}},
		installApps: []catalog.Application{app},
		appStatus:   map[string]string{"detectable-failure": "failed"},
		selected:    map[string]bool{},
		installed: map[string]InstalledStatus{
			"detectable-failure": {Installed: false, Checked: true},
		},
		detectInstalled: func(catalog.Application) bool {
			return false
		},
	}

	updated, _ := model.handleInstallEvent(installEventMsg{
		ok: true,
		event: installer.Event{
			Kind:    installer.EventSummary,
			Results: []installer.Result{{App: app, Success: false, Err: errors.New("install failed")}},
		},
	})
	got := updated.(Model)
	if got.installed["detectable-failure"].Installed {
		t.Fatalf("failed install should not force installed status, got %#v", got.installed["detectable-failure"])
	}
}

func TestCatalogReflectsRefreshedInstalledStatusAfterInstall(t *testing.T) {
	app := catalog.Application{
		Name:         "Reflected Tool",
		Description:  "Reflected status test.",
		ID:           "reflected-tool",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "reflected-tool.exe",
		Type:         catalog.PackageTypeCLITool,
		Providers:    testChocolateyProviders("reflected-tool"),
		Verified:     true,
	}
	model := Model{
		screen:      screenInstall,
		width:       100,
		height:      32,
		categories:  []catalog.Category{{Apps: []catalog.Application{app}}},
		catalogMode: catalogModeFull,
		installApps: []catalog.Application{app},
		appStatus:   map[string]string{"reflected-tool": "installed"},
		selected:    map[string]bool{},
		installed: map[string]InstalledStatus{
			"reflected-tool": {Installed: false, Checked: true},
		},
		detectInstalled: func(catalog.Application) bool {
			return true
		},
	}

	updated, _ := model.handleInstallEvent(installEventMsg{
		ok: true,
		event: installer.Event{
			Kind:    installer.EventSummary,
			Results: []installer.Result{{App: app, Success: true}},
		},
	})
	got := updated.(Model)
	got.screen = screenCatalog

	view := stripANSI(got.View())
	if !strings.Contains(view, "Installed") || !strings.Contains(view, "Installed: Yes") {
		t.Fatalf("catalog list and details should reflect refreshed installed cache, got:\n%s", view)
	}
}

func TestStartInstallSkipsInstalledPackages(t *testing.T) {
	installedApp := catalog.Application{
		Name:         "Installed App",
		ID:           "installed-app",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "installed-app.exe",
	}
	m := Model{
		selected: map[string]bool{},
		installed: map[string]InstalledStatus{
			"installed-app": {Installed: true, Checked: true},
		},
	}

	updated, _ := m.startInstall([]catalog.Application{installedApp})
	got := updated.(Model)

	if got.appStatus["installed-app"] != "skipped" {
		t.Fatalf("installed package should be skipped, got status %q", got.appStatus["installed-app"])
	}
	if !got.installDone {
		t.Fatal("all-installed selection should complete without launching Chocolatey")
	}
	if len(got.results) != 1 || !got.results[0].Skipped {
		t.Fatalf("skipped package should appear in results, got %#v", got.results)
	}
}

func TestStartInstallLeavesNotInstalledPackagesPending(t *testing.T) {
	app := catalog.Application{
		Name:         "Missing App",
		ID:           "missing-app",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "missing-app.exe",
	}
	m := Model{
		selected: map[string]bool{},
		installed: map[string]InstalledStatus{
			"missing-app": {Installed: false, Checked: true},
		},
	}

	updated, _ := m.startInstall([]catalog.Application{app})
	got := updated.(Model)

	if got.appStatus["missing-app"] != "pending" {
		t.Fatalf("not installed package should remain pending for install, got %q", got.appStatus["missing-app"])
	}
	if got.installDone {
		t.Fatal("not installed package should start normal install flow")
	}
	if len(got.initialSkippedResults) != 0 {
		t.Fatalf("not installed package should not be pre-skipped, got %#v", got.initialSkippedResults)
	}
}

func TestStartInstallInstallsPackagesWithoutDetectionMetadata(t *testing.T) {
	app := catalog.Application{Name: "Unknown Detection App", ID: "unknown-detection-app"}
	m := Model{
		selected: map[string]bool{},
		installed: map[string]InstalledStatus{
			"unknown-detection-app": {Installed: true, Checked: true},
		},
	}

	updated, _ := m.startInstall([]catalog.Application{app})
	got := updated.(Model)

	if got.appStatus["unknown-detection-app"] != "pending" {
		t.Fatalf("package without detection metadata should install normally, got %q", got.appStatus["unknown-detection-app"])
	}
	if len(got.initialSkippedResults) != 0 {
		t.Fatalf("package without detection metadata should not be skipped, got %#v", got.initialSkippedResults)
	}
}

func TestSkippedPackagesAppearInInstallSummaryWithoutFailure(t *testing.T) {
	skippedApp := catalog.Application{
		Name:         "Already There",
		ID:           "already-there",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "already-there.exe",
	}
	m := Model{
		width:    100,
		height:   24,
		selected: map[string]bool{},
		installed: map[string]InstalledStatus{
			"already-there": {Installed: true, Checked: true},
		},
	}

	updated, _ := m.startInstall([]catalog.Application{skippedApp})
	got := updated.(Model)
	view := stripANSI(got.View())

	if !strings.Contains(view, "skipped") || !strings.Contains(view, "Already There") {
		t.Fatalf("skipped package should render in install summary, got:\n%s", view)
	}
	if strings.Contains(view, "FAIL") || strings.Contains(view, "Failures:") {
		t.Fatalf("skipped package should not count as a failure, got:\n%s", view)
	}
	if !strings.Contains(view, "0 failed") {
		t.Fatalf("skipped package should produce zero failures, got:\n%s", view)
	}
	if !strings.Contains(view, "Selected: 1") || !strings.Contains(view, "Already installed: 1") || !strings.Contains(view, "Remaining: 0") {
		t.Fatalf("install counts should include skipped packages, got:\n%s", view)
	}
}

func TestInstallSummaryMergesPreSkippedResults(t *testing.T) {
	skippedApp := catalog.Application{Name: "Skipped App", ID: "skipped-app"}
	installedApp := catalog.Application{Name: "Installed App", ID: "installed-app"}
	model := Model{
		screen:      screenInstall,
		categories:  []catalog.Category{{Apps: []catalog.Application{skippedApp, installedApp}}},
		installApps: []catalog.Application{skippedApp, installedApp},
		appStatus: map[string]string{
			"skipped-app":   "skipped",
			"installed-app": "installed",
		},
		selected: map[string]bool{},
		initialSkippedResults: []installer.Result{{
			App:     skippedApp,
			Skipped: true,
			Err:     installer.ErrInstallSkipped,
		}},
		detectInstalled: func(catalog.Application) bool {
			return false
		},
	}

	updated, _ := model.handleInstallEvent(installEventMsg{
		ok: true,
		event: installer.Event{
			Kind:    installer.EventSummary,
			Results: []installer.Result{{App: installedApp, Success: true}},
		},
	})
	got := updated.(Model)

	if len(got.results) != 2 {
		t.Fatalf("summary should include pre-skipped and installed results, got %#v", got.results)
	}
	if !got.results[0].Skipped || !got.results[1].Success {
		t.Fatalf("summary result states are wrong, got %#v", got.results)
	}
}

func TestPackageDetailsPanelFitsNarrowWidth(t *testing.T) {
	app := catalog.Application{
		Name:        "Long App",
		Description: strings.Repeat("long metadata ", 12),
		ID:          "very-long-package-id-that-should-be-truncated",
		Type:        catalog.PackageTypeApplication,
		Providers:   testChocolateyProviders("very-long-package-id-that-should-be-truncated"),
		Verified:    true,
	}

	view := stripANSI(fitDetailsLines(packageDetailsLines(app, "Yes", ""), 26, 18))
	for _, line := range strings.Split(view, "\n") {
		if len(line) > 27 {
			t.Fatalf("details line should be constrained, got %d chars in %q\n%s", len(line), line, view)
		}
	}
}

func TestEscFromLeafCategoryReturnsToCategoryRoot(t *testing.T) {
	model := Model{
		screen:      screenCatalog,
		width:       100,
		height:      32,
		categories:  catalog.Default(),
		catalogMode: catalogModeCategories,
		catalogPath: []int{0},
		selected:    map[string]bool{},
	}

	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyEsc})
	got := updated.(Model)
	if got.screen != screenCatalog {
		t.Fatalf("esc from leaf category should stay in catalog, got screen %v", got.screen)
	}
	if len(got.catalogPath) != 0 {
		t.Fatalf("esc from leaf category should return to category root, got path %v", got.catalogPath)
	}
}

func TestEscFromCategoryRootReturnsToModeSelect(t *testing.T) {
	model := Model{
		screen:      screenCatalog,
		width:       100,
		height:      32,
		categories:  catalog.Default(),
		catalogMode: catalogModeCategories,
		selected:    map[string]bool{},
	}

	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyEsc})
	got := updated.(Model)
	if got.screen != screenModeSelect {
		t.Fatalf("esc from category root should return to mode select, got screen %v", got.screen)
	}
}

func TestReviewScreenSummarizesLargeSelection(t *testing.T) {
	selected := map[string]bool{}
	for _, item := range collectTestPackages(catalog.Default()) {
		selected[item.ID] = true
	}
	model := Model{
		screen:     screenReview,
		width:      100,
		height:     24,
		categories: catalog.Default(),
		selected:   selected,
	}

	view := stripANSI(model.View())
	if !strings.Contains(view, "Packages selected:") {
		t.Fatalf("review should show package count, got:\n%s", view)
	}
	if strings.Contains(view, "Commands:") || strings.Contains(view, "choco install") {
		t.Fatalf("review should not render every install command, got:\n%s", view)
	}
	if count := strings.Count(view, "enter install"); count != 1 {
		t.Fatalf("review footer should remain visible once, got %d in:\n%s", count, view)
	}
}

func TestReviewScreenScrollsSelection(t *testing.T) {
	selected := map[string]bool{}
	packages := collectTestPackages(catalog.Default())
	for _, app := range packages {
		selected[app.ID] = true
	}
	model := Model{
		screen:       screenReview,
		width:        100,
		height:       24,
		categories:   catalog.Default(),
		selected:     selected,
		reviewScroll: 0,
	}

	firstView := stripANSI(model.View())
	updated, _ := model.handleReviewKey(tea.KeyMsg{Type: tea.KeyDown})
	scrolled := updated.(Model)
	secondView := stripANSI(scrolled.View())

	if scrolled.reviewScroll != 1 {
		t.Fatalf("down should scroll review list by one row, got %d", scrolled.reviewScroll)
	}
	if firstView == secondView {
		t.Fatalf("scrolling review list should change visible content")
	}
}

func TestBootstrapLogsHiddenByDefault(t *testing.T) {
	model := Model{
		screen:          screenBootstrap,
		width:           90,
		height:          24,
		selected:        map[string]bool{},
		bootstrapStatus: "Bootstrapping Chocolatey...",
		bootstrapLog: []string{
			"RAW CHOCOLATEY POWERSHELL OUTPUT",
			"Downloading chocolatey package from source https://community.chocolatey.org/api/v2/",
		},
	}

	view := stripANSI(model.View())
	if strings.Contains(view, "RAW CHOCOLATEY POWERSHELL OUTPUT") ||
		strings.Contains(view, "community.chocolatey.org/api/v2") {
		t.Fatalf("bootstrap raw logs should be hidden by default, got:\n%s", view)
	}
	if !strings.Contains(view, "Logs hidden. Press l to show full logs.") {
		t.Fatalf("bootstrap view should show hidden logs hint, got:\n%s", view)
	}
	if count := strings.Count(view, "enter bootstrap"); count != 1 {
		t.Fatalf("bootstrap footer should remain visible once, got %d in:\n%s", count, view)
	}
}

func TestBootstrapLogToggleShowsClippedLogs(t *testing.T) {
	longLine := "Downloading " + strings.Repeat("very-long-output-", 20)
	model := Model{
		screen:           screenBootstrap,
		width:            90,
		height:           22,
		selected:         map[string]bool{},
		bootstrapStatus:  "Bootstrapping Chocolatey...",
		showBootstrapLog: true,
		bootstrapLog: []string{
			"first line should scroll away",
			"second line should scroll away",
			"third line should scroll away",
			"fourth line should scroll away",
			"fifth line should scroll away",
			"Installing Chocolatey...",
			longLine,
		},
	}

	view := stripANSI(model.View())
	if !strings.Contains(view, "full logs") || !strings.Contains(view, "Installing Chocolatey") {
		t.Fatalf("bootstrap logs should render when enabled, got:\n%s", view)
	}
	if strings.Contains(view, longLine) {
		t.Fatalf("long bootstrap log lines should be truncated, got:\n%s", view)
	}
	if strings.Contains(view, "first line should scroll away") {
		t.Fatalf("bootstrap logs should be clipped to visible height, got:\n%s", view)
	}
	if count := strings.Count(view, "•  l logs"); count != 1 {
		t.Fatalf("bootstrap footer should remain visible once, got %d in:\n%s", count, view)
	}
}

func TestInstallScreenNeverRendersRawLogs(t *testing.T) {
	app := catalog.Application{
		Name:      "Visual Studio Code",
		ID:        "vscode",
		Providers: []catalog.Provider{{Type: catalog.ProviderDirect}},
	}
	model := Model{
		screen:        screenInstall,
		width:         100,
		height:        24,
		installApps:   []catalog.Application{app},
		appStatus:     map[string]string{app.ID: "installing"},
		appElapsed:    map[string]time.Duration{},
		currentApp:    app,
		currentStep:   1,
		currentStatus: "Installing...",
		installLogs: []installLogEntry{{
			Application: app.Name,
			Line:        "RAW PROVIDER OUTPUT THAT MUST STAY OFF THE INSTALL SCREEN",
		}},
	}

	view := stripANSI(model.View())
	if strings.Contains(view, "RAW PROVIDER OUTPUT") {
		t.Fatalf("install screen must not render raw logs, got:\n%s", view)
	}
	if !strings.Contains(view, "l logs") {
		t.Fatalf("install footer should open the dedicated logs screen, got:\n%s", view)
	}
	if strings.Contains(view, "show/hide logs") {
		t.Fatalf("install footer should no longer advertise inline logs, got:\n%s", view)
	}
}

func TestInstallLogScreenNavigationPreservesInstallState(t *testing.T) {
	app := catalog.Application{Name: "VLC", ID: "vlc"}
	model := Model{
		screen:        screenInstall,
		width:         100,
		height:        20,
		installApps:   []catalog.Application{app},
		appStatus:     map[string]string{app.ID: "installing"},
		currentApp:    app,
		currentStep:   1,
		currentStatus: "Downloading... 50%",
		installLogs:   []installLogEntry{{Application: app.Name, Line: "downloading 50%"}},
		logFollow:     true,
	}

	updated, _ := model.handleInstallKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	logs := updated.(Model)
	if logs.screen != screenInstallLogs {
		t.Fatalf("l should open logs screen, got screen %v", logs.screen)
	}
	if logs.currentStep != model.currentStep || logs.currentStatus != model.currentStatus ||
		logs.appStatus[app.ID] != model.appStatus[app.ID] || len(logs.installLogs) != len(model.installLogs) {
		t.Fatal("opening logs must preserve installation state")
	}

	updated, _ = logs.handleInstallLogsKey(tea.KeyMsg{Type: tea.KeyEsc})
	back := updated.(Model)
	if back.screen != screenInstall {
		t.Fatalf("esc should return to install screen, got screen %v", back.screen)
	}

	back.screen = screenInstallLogs
	updated, _ = back.handleInstallLogsKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if got := updated.(Model).screen; got != screenInstall {
		t.Fatalf("l should also return to install screen, got screen %v", got)
	}
}

func TestInstallLogsScrollAndKeepFooterVisible(t *testing.T) {
	model := Model{
		screen:      screenInstallLogs,
		width:       80,
		height:      14,
		installLogs: fakeInstallLogEntries(12),
		logFollow:   true,
	}
	model.clampLogScroll()
	start := model.logScroll

	updated, _ := model.handleInstallLogsKey(tea.KeyMsg{Type: tea.KeyUp})
	scrolled := updated.(Model)
	if scrolled.logScroll >= start {
		t.Fatalf("up should move log viewport, start=%d got=%d", start, scrolled.logScroll)
	}

	view := stripANSI(scrolled.View())
	if !strings.Contains(view, "↑↓ scroll") || !strings.Contains(view, "esc back") {
		t.Fatalf("logs footer should remain visible after scrolling, got:\n%s", view)
	}
}

func TestInstallLogsTruncateLongLines(t *testing.T) {
	longLine := strings.Repeat("provider-output-", 20)
	model := Model{
		screen: screenInstallLogs,
		width:  50,
		height: 12,
		installLogs: []installLogEntry{{
			Application: "Very Long Application Name",
			Line:        "\x1b[31m" + longLine + "\x1b[0m\t\x00",
		}},
		logFollow: true,
	}

	view := stripANSI(model.View())
	if strings.Contains(view, longLine) {
		t.Fatalf("long log line should be truncated, got:\n%s", view)
	}
	if strings.Contains(view, "[31m") {
		t.Fatalf("log control sequences should be sanitized, got:\n%s", view)
	}
	for _, line := range strings.Split(view, "\n") {
		if len([]rune(line)) > model.width {
			t.Fatalf("log line exceeds terminal width %d: %q", model.width, line)
		}
	}
}

func TestInstallLogEventsAppendWhileViewing(t *testing.T) {
	app := catalog.Application{Name: "Firefox", ID: "firefox"}
	model := Model{
		screen:        screenInstallLogs,
		width:         80,
		height:        14,
		currentApp:    app,
		currentStatus: "Installing...",
		installLogs:   []installLogEntry{{Application: app.Name, Line: "first line"}},
		logFollow:     true,
	}

	updated, _ := model.handleInstallEvent(installEventMsg{
		ok: true,
		event: installer.Event{
			Kind: installer.EventLog,
			App:  app,
			Line: "second line",
		},
	})
	got := updated.(Model)
	if got.screen != screenInstallLogs {
		t.Fatal("receiving an install event should not close the logs screen")
	}
	if len(got.installLogs) != 2 || got.installLogs[1].Line != "second line" {
		t.Fatalf("new install event should append to shared log buffer, got %#v", got.installLogs)
	}
	if !strings.Contains(stripANSI(got.View()), "[Firefox] second line") {
		t.Fatalf("logs should identify the producing application, got:\n%s", stripANSI(got.View()))
	}
}

func TestInstallLogAutoFollowStopsAndResumes(t *testing.T) {
	model := Model{
		screen:    screenInstallLogs,
		width:     80,
		height:    12,
		logFollow: true,
	}
	for _, entry := range fakeInstallLogEntries(10) {
		model.appendInstallLog(catalog.Application{Name: entry.Application}, entry.Line)
	}
	if model.logScroll != model.maxLogScroll() {
		t.Fatalf("following logs should stay at bottom, scroll=%d max=%d", model.logScroll, model.maxLogScroll())
	}

	model.moveLogScroll(-2)
	pausedAt := model.logScroll
	if model.logFollow {
		t.Fatal("manual upward scrolling should disable auto-follow")
	}
	model.appendInstallLog(catalog.Application{Name: "Package"}, "new while paused")
	if model.logScroll != pausedAt {
		t.Fatalf("new logs should not force a manually scrolled viewport down, got %d want %d", model.logScroll, pausedAt)
	}

	updated, _ := model.handleInstallLogsKey(tea.KeyMsg{Type: tea.KeyEnd})
	model = updated.(Model)
	if !model.logFollow || model.logScroll != model.maxLogScroll() {
		t.Fatalf("end should resume auto-follow, scroll=%d max=%d follow=%v", model.logScroll, model.maxLogScroll(), model.logFollow)
	}
	model.appendInstallLog(catalog.Application{Name: "Package"}, "new at bottom")
	if model.logScroll != model.maxLogScroll() {
		t.Fatalf("auto-follow should track appended logs, scroll=%d max=%d", model.logScroll, model.maxLogScroll())
	}
}

func TestInstallScreenUsesProviderNeutralStatus(t *testing.T) {
	app := catalog.Application{Name: "Example App", ID: "example"}
	model := Model{
		screen:      screenInstall,
		width:       90,
		height:      20,
		installApps: []catalog.Application{app},
		appStatus:   map[string]string{app.ID: "installing"},
		appElapsed:  map[string]time.Duration{},
		currentApp:  app,
		currentStep: 1,
		logFollow:   true,
	}

	updated, _ := model.handleInstallEvent(installEventMsg{
		ok: true,
		event: installer.Event{
			Kind: installer.EventLog,
			App:  app,
			Line: "Downloading package from provider-internal-endpoint 70%",
		},
	})
	got := updated.(Model)
	view := stripANSI(got.View())
	if !strings.Contains(view, "Status: Downloading... 70%") {
		t.Fatalf("install screen should show compact provider-neutral progress, got:\n%s", view)
	}
	if strings.Contains(view, "provider-internal-endpoint") {
		t.Fatalf("install screen should not expose raw provider output, got:\n%s", view)
	}
}

func TestInstallSummaryScrollsLongPackageList(t *testing.T) {
	apps := fakeInstallPackages(24)
	model := Model{
		screen:      screenInstall,
		width:       100,
		height:      24,
		installApps: apps,
		appStatus:   map[string]string{},
		appElapsed:  map[string]time.Duration{},
	}
	for _, app := range apps {
		model.appStatus[app.ID] = "pending"
	}

	firstView := stripANSI(model.View())
	updated, _ := model.handleInstallKey(tea.KeyMsg{Type: tea.KeyPgDown})
	scrolled := updated.(Model)
	secondView := stripANSI(scrolled.View())

	if scrolled.installScroll == 0 {
		t.Fatal("pgdown should scroll the install summary")
	}
	if firstView == secondView {
		t.Fatalf("scrolling install summary should change visible rows")
	}
	if !strings.Contains(secondView, "l logs") {
		t.Fatalf("install footer should remain visible after scrolling, got:\n%s", secondView)
	}
}

func TestInstallSummaryDoesNotDuplicateVisibleRows(t *testing.T) {
	apps := fakeInstallPackages(20)
	model := Model{
		screen:      screenInstall,
		width:       100,
		height:      24,
		installApps: apps,
		appStatus:   map[string]string{},
		appElapsed:  map[string]time.Duration{},
	}
	for _, app := range apps {
		model.appStatus[app.ID] = "pending"
	}
	model.appStatus[apps[0].ID] = "installing"

	updated, _ := model.handleInstallTick()
	ticked := updated.(Model)
	view := stripANSI(ticked.View())

	visible := 0
	for _, app := range apps {
		count := strings.Count(view, app.Name)
		if count > 1 {
			t.Fatalf("expected %s to render at most once, got %d in:\n%s", app.Name, count, view)
		}
		if count == 1 {
			visible++
		}
	}
	if visible == 0 {
		t.Fatalf("expected install summary to render visible package rows, got:\n%s", view)
	}
}

func TestInstallElapsedFreezesAfterCompletion(t *testing.T) {
	app := catalog.Application{Name: "Example App", ID: "example"}
	model := Model{
		screen:       screenInstall,
		width:        100,
		height:       24,
		installApps:  []catalog.Application{app},
		appStatus:    map[string]string{"example": "installed"},
		appElapsed:   map[string]time.Duration{"example": 3 * time.Second},
		currentApp:   app,
		currentStart: time.Now().Add(-1 * time.Hour),
		currentStep:  1,
		installDone:  true,
	}

	view := stripANSI(model.View())
	if !strings.Contains(view, "00:03") {
		t.Fatalf("completed install should show frozen elapsed time, got:\n%s", view)
	}
	if strings.Contains(view, "60:") || strings.Contains(view, "59:") {
		t.Fatalf("completed install elapsed should not keep increasing, got:\n%s", view)
	}
}

func TestFullCatalogSearchFiltersByPackageMetadata(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        40,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		searchFocused: true,
		searchQuery:   "python runtime",
		selected:      map[string]bool{},
	}

	view := stripANSI(model.View())
	if !strings.Contains(view, "Python 3") {
		t.Fatalf("full catalog search should match package description, got:\n%s", view)
	}
	if strings.Contains(view, "Google Chrome") {
		t.Fatalf("full catalog search should filter unrelated packages, got:\n%s", view)
	}
	if strings.Contains(view, "Development > Runtimes > Python") {
		t.Fatalf("full catalog list should not show package path, got:\n%s", view)
	}
}

func TestCatalogSearchMatchesCaseInsensitiveName(t *testing.T) {
	model := Model{
		categories:  catalog.Default(),
		catalogMode: catalogModeFull,
		searchQuery: "fireFOX",
		selected:    map[string]bool{},
	}

	items := model.filteredFullCatalogItems()
	if !containsPackage(items, "firefox") {
		t.Fatalf("search should match package name case-insensitively, got %#v", itemNames(items))
	}
}

func TestCatalogSearchMatchesDescription(t *testing.T) {
	model := Model{
		categories:  catalog.Default(),
		catalogMode: catalogModeFull,
		searchQuery: "runtime",
		selected:    map[string]bool{},
	}

	items := model.filteredFullCatalogItems()
	if !containsPackage(items, "vcredist140") || !containsPackage(items, "dotnet-8.0-runtime") {
		t.Fatalf("search should match runtime descriptions, got %#v", itemNames(items))
	}
}

func TestCatalogSearchMatchesPackageID(t *testing.T) {
	model := Model{
		categories:  catalog.Default(),
		catalogMode: catalogModeFull,
		searchQuery: "googlechrome",
		selected:    map[string]bool{},
	}

	items := model.filteredFullCatalogItems()
	if len(items) != 1 || items[0].Package.ID != "googlechrome" {
		t.Fatalf("search should match package id, got %#v", itemNames(items))
	}
}

func TestCatalogSearchMatchesFuzzyInitialsAndSubsequence(t *testing.T) {
	model := Model{
		categories:  catalog.Default(),
		catalogMode: catalogModeFull,
		searchQuery: "vsc",
		selected:    map[string]bool{},
	}
	if !containsPackage(model.filteredFullCatalogItems(), "vscode") {
		t.Fatalf("search query vsc should find VS Code")
	}

	model.searchQuery = "rg"
	if !containsPackage(model.filteredFullCatalogItems(), "ripgrep") {
		t.Fatalf("search query rg should find ripgrep")
	}
}

func TestCatalogSearchClearsWithEscape(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeCategories,
		catalogPath:   []int{3, 1},
		searchFocused: true,
		searchQuery:   "code",
		selected:      map[string]bool{},
	}

	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyEsc})
	got := updated.(Model)
	if got.searchFocused || got.searchQuery != "" {
		t.Fatalf("esc should clear active search, focused=%v query=%q", got.searchFocused, got.searchQuery)
	}
	if got.screen != screenCatalog || got.catalogMode != catalogModeCategories || got.currentBreadcrumb() != "Catalog > Media > Images & Graphics" {
		t.Fatalf("esc should return to normal category browsing, got screen=%v mode=%v path=%q", got.screen, got.catalogMode, got.currentBreadcrumb())
	}
}

func TestCatalogSearchEscapeClearsInactiveSearchBeforeLeavingCatalog(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		searchFocused: false,
		searchQuery:   "discord",
		catalogCursor: 2,
		catalogScroll: 1,
		selected:      map[string]bool{},
	}

	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyEsc})
	got := updated.(Model)
	if got.screen != screenCatalog {
		t.Fatalf("esc should clear inactive search before leaving catalog, got screen %v", got.screen)
	}
	if got.searchQuery != "" || got.catalogCursor != 0 || got.catalogScroll != 0 {
		t.Fatalf("esc should clear inactive search state, query=%q cursor=%d scroll=%d", got.searchQuery, got.catalogCursor, got.catalogScroll)
	}
}

func TestCatalogSearchEnterStopsEditingWithoutSelectingPackage(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeCategories,
		searchFocused: true,
		searchQuery:   "discord",
		selected:      map[string]bool{},
	}

	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(Model)
	if got.selected["discord"] {
		t.Fatalf("enter should not select highlighted search result")
	}
	if got.searchFocused {
		t.Fatalf("enter should stop editing search")
	}
	if got.searchQuery != "discord" {
		t.Fatalf("enter should keep current search query, got %q", got.searchQuery)
	}
}

func TestCatalogSearchTypingAppendsLetters(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		searchFocused: true,
		searchQuery:   "d",
		selected:      map[string]bool{},
	}

	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	got := updated.(Model)
	if got.searchQuery != "di" {
		t.Fatalf("typing while search is focused should append letters, got %q", got.searchQuery)
	}
}

func TestCatalogSearchPrintableHotkeysAreTypedBeforeGlobalRouting(t *testing.T) {
	tests := []struct {
		name  string
		input rune
	}{
		{name: "vim down", input: 'j'},
		{name: "vim up", input: 'k'},
		{name: "logs", input: 'l'},
		{name: "quit", input: 'q'},
		{name: "skip", input: 's'},
		{name: "help", input: '?'},
		{name: "russian physical j", input: 'о'},
		{name: "russian physical k", input: 'л'},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := Model{
				screen:        screenCatalog,
				width:         100,
				height:        32,
				categories:    catalog.Default(),
				catalogMode:   catalogModeFull,
				catalogCursor: 0,
				searchFocused: true,
				selected:      map[string]bool{},
			}

			updated, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{test.input}})
			got := updated.(Model)
			if got.searchQuery != string(test.input) {
				t.Fatalf("printable key should be inserted verbatim, got query %q", got.searchQuery)
			}
			if got.screen != screenCatalog || !got.searchFocused {
				t.Fatalf("printable key should not trigger a global action, screen=%v focused=%v", got.screen, got.searchFocused)
			}
			if got.catalogCursor != 0 {
				t.Fatalf("printable key should not navigate while search is focused, cursor=%d", got.catalogCursor)
			}
			if cmd != nil {
				t.Fatalf("printable key should not return a global command, got %T", cmd)
			}
		})
	}
}

func TestCatalogSearchArrowKeysStillNavigateResults(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		catalogCursor: 1,
		searchFocused: true,
		selected:      map[string]bool{},
	}

	updated, _ := model.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	got := updated.(Model)
	if got.catalogCursor != 2 || got.searchQuery != "" {
		t.Fatalf("down should navigate search results without editing query, cursor=%d query=%q", got.catalogCursor, got.searchQuery)
	}

	updated, _ = got.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	got = updated.(Model)
	if got.catalogCursor != 1 || got.searchQuery != "" {
		t.Fatalf("up should navigate search results without editing query, cursor=%d query=%q", got.catalogCursor, got.searchQuery)
	}
}

func TestCatalogSearchExitPreservesSelectionAndRestoresHotkeys(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		searchFocused: true,
		searchQuery:   "git",
		selected:      map[string]bool{"firefox": true},
	}

	updated, _ := model.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	got := updated.(Model)
	if got.searchFocused || got.searchQuery != "" || !got.selected["firefox"] {
		t.Fatalf("escape should clear only search state, focused=%v query=%q selected=%v", got.searchFocused, got.searchQuery, got.selected)
	}

	updated, _ = got.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	got = updated.(Model)
	if got.catalogCursor != 1 || got.searchQuery != "" {
		t.Fatalf("j should navigate again after search closes, cursor=%d query=%q", got.catalogCursor, got.searchQuery)
	}
}

func TestCatalogSearchShowsBlinkingInputCursor(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		searchFocused: true,
		searchCursor:  true,
		searchQuery:   "code",
		selected:      map[string]bool{},
	}

	view := stripANSI(model.View())
	if !strings.Contains(view, "Search: code|") {
		t.Fatalf("focused search should show visible input cursor, got:\n%s", view)
	}

	updated, _ := model.handleSearchCursorTick()
	got := updated.(Model)
	if got.searchCursor {
		t.Fatal("search cursor tick should toggle cursor visibility")
	}
	view = stripANSI(got.View())
	if !strings.Contains(view, "Search: code ") {
		t.Fatalf("hidden cursor should preserve input spacing, got:\n%s", view)
	}
}

func TestCatalogSearchSpaceTypesSpaceWhileFocused(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		searchFocused: true,
		searchQuery:   "visual",
		selected:      map[string]bool{},
	}

	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeySpace})
	got := updated.(Model)
	if got.searchQuery != "visual " {
		t.Fatalf("space should be typed into focused search input, got %q", got.searchQuery)
	}
	if len(got.selected) != 0 {
		t.Fatalf("space should not select packages while search input is focused, got %#v", got.selected)
	}
}

func TestCatalogSearchEmptyResultMessage(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		searchFocused: true,
		searchQuery:   "definitely-not-a-package",
		selected:      map[string]bool{},
	}

	view := stripANSI(model.View())
	if !strings.Contains(view, "No packages found.") || !strings.Contains(view, "Try a different search term.") {
		t.Fatalf("empty search should show friendly message, got:\n%s", view)
	}
}

func TestFullCatalogItemsAreSortedByName(t *testing.T) {
	model := Model{
		categories: catalog.Default(),
	}

	items := model.allCatalogItems()
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, strings.ToLower(item.Package.Name))
	}

	sorted := append([]string{}, names...)
	sort.Strings(sorted)

	if strings.Join(names, "\n") != strings.Join(sorted, "\n") {
		t.Fatalf("full catalog items should be sorted alphabetically")
	}
}

func TestFullCatalogScrollsWithCursor(t *testing.T) {
	model := Model{
		screen:      screenCatalog,
		width:       100,
		height:      32,
		categories:  catalog.Default(),
		catalogMode: catalogModeFull,
		selected:    map[string]bool{},
	}

	items := model.filteredFullCatalogItems()
	model.catalogCursor = len(items) - 1
	model.ensureCatalogCursorVisible()

	view := stripANSI(model.View())
	target := items[len(items)-1].Package.Name
	if !strings.Contains(view, "> [ ] "+target) {
		t.Fatalf("full catalog should render the highlighted item after scrolling, got:\n%s", view)
	}
	if strings.Contains(view, "Google Chrome") {
		t.Fatalf("full catalog should scroll past first-page items, got:\n%s", view)
	}
}

func TestFullCatalogTruncatesLongNamesInsidePane(t *testing.T) {
	model := Model{
		screen:      screenCatalog,
		width:       100,
		height:      32,
		categories:  catalog.Default(),
		catalogMode: catalogModeFull,
		selected:    map[string]bool{},
	}

	for i, item := range model.filteredFullCatalogItems() {
		if item.Package.ID == "vcredist140" {
			model.catalogCursor = i
			break
		}
	}
	model.ensureCatalogCursorVisible()

	view := stripANSI(model.View())
	if strings.Contains(view, "\nx86/x64") {
		t.Fatalf("long package names should not wrap into a stray line, got:\n%s", view)
	}
	if !strings.Contains(view, "VC++ Redist 2015-2022 x86/x64") {
		t.Fatalf("expected long highlighted package to remain visible, got:\n%s", view)
	}
}

func TestBackspaceEditsCatalogSearch(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		searchFocused: true,
		searchQuery:   "python",
		selected:      map[string]bool{},
	}

	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyBackspace})
	got := updated.(Model)
	if !got.searchFocused {
		t.Fatal("backspace should keep search active")
	}
	if got.searchQuery != "pytho" {
		t.Fatalf("backspace should edit search query, got %q", got.searchQuery)
	}
}

func TestCtrlBackspaceDeletesPreviousSearchWord(t *testing.T) {
	previousControlKeyPressed := controlKeyPressed
	controlKeyPressed = func() bool { return true }
	t.Cleanup(func() { controlKeyPressed = previousControlKeyPressed })

	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		searchFocused: true,
		searchQuery:   "visual studio",
		selected:      map[string]bool{},
	}

	updated, _ := model.handleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	got := updated.(Model)
	if got.searchQuery != "visual " {
		t.Fatalf("ctrl+backspace should delete the previous word, got %q", got.searchQuery)
	}
	if !got.searchFocused {
		t.Fatal("ctrl+backspace should keep search focused")
	}
}

func TestCtrlWDoesNothingInCatalogSearch(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		searchFocused: true,
		searchQuery:   "visual studio",
		selected:      map[string]bool{},
	}

	updated, _ := model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlW})
	got := updated.(Model)
	if got.searchQuery != "visual studio" {
		t.Fatalf("ctrl+w should not edit search, got %q", got.searchQuery)
	}
	if !got.searchFocused {
		t.Fatal("ctrl+w should keep search focused")
	}
}

func TestDropLastWordHandlesUnicode(t *testing.T) {
	if got := dropLastWord("поиск браузер"); got != "поиск " {
		t.Fatalf("word deletion should be unicode-safe, got %q", got)
	}
}

func TestBackspaceRemovesOneUnicodeRuneFromCatalogSearch(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		searchFocused: true,
		searchQuery:   "браузер",
		selected:      map[string]bool{},
	}

	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyBackspace})
	got := updated.(Model)
	if got.searchQuery != "браузе" {
		t.Fatalf("backspace should remove one unicode rune, got %q", got.searchQuery)
	}
	if strings.Contains(stripANSI(got.View()), "�") {
		t.Fatalf("backspace should not leave replacement glyphs, got:\n%s", stripANSI(got.View()))
	}
}

func TestCatalogSearchIgnoresAltAndControlRunes(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        32,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		searchFocused: true,
		searchQuery:   "code",
		selected:      map[string]bool{},
	}

	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}, Alt: true})
	got := updated.(Model)
	if got.searchQuery != "code" {
		t.Fatalf("alt-modified runes should not be inserted into search, got %q", got.searchQuery)
	}

	updated, _ = got.handleCatalogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{0, '\x1b', unicode.ReplacementChar}})
	got = updated.(Model)
	if got.searchQuery != "code" {
		t.Fatalf("control/replacement runes should not be inserted into search, got %q", got.searchQuery)
	}
}

func containsPackage(items []fullCatalogItem, packageID string) bool {
	for _, item := range items {
		if item.Package.ID == packageID {
			return true
		}
	}
	return false
}

func itemNames(items []fullCatalogItem) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Package.Name)
	}
	return names
}

func catalogPanelBorderRows(t *testing.T, view string) (int, int) {
	t.Helper()

	rows := []int{}
	for i, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "+---") && strings.Count(line, "+") >= 4 {
			rows = append(rows, i)
		}
	}
	if len(rows) < 2 {
		t.Fatalf("expected top and bottom rows for aligned catalog panels, got rows=%v\n%s", rows, view)
	}
	return rows[0], rows[len(rows)-1]
}

func stripANSI(value string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(value, "")
}

func collectTestPackages(categories []catalog.Category) []catalog.Application {
	var apps []catalog.Application
	for _, category := range categories {
		apps = append(apps, collectTestPackages(category.Categories)...)
		apps = append(apps, category.Apps...)
	}
	return apps
}

func packagesByIDForTUITest(categories []catalog.Category) map[string]catalog.Application {
	apps := make(map[string]catalog.Application)
	for _, app := range collectTestPackages(categories) {
		apps[app.ID] = app
	}
	return apps
}

func fakeInstallPackages(count int) []catalog.Application {
	apps := make([]catalog.Application, 0, count)
	for i := 1; i <= count; i++ {
		apps = append(apps, catalog.Application{
			Name: "Package " + twoDigit(i),
			ID:   "pkg-" + twoDigit(i),
		})
	}
	return apps
}

func fakeInstallLogEntries(count int) []installLogEntry {
	entries := make([]installLogEntry, 0, count)
	for i := 1; i <= count; i++ {
		entries = append(entries, installLogEntry{
			Application: "Package " + twoDigit(i),
			Line:        "log line " + twoDigit(i),
		})
	}
	return entries
}

func testChocolateyProviders(packageID string) []catalog.Provider {
	return []catalog.Provider{{
		Type:      catalog.ProviderChocolatey,
		PackageID: packageID,
		Strategy:  catalog.InstallStrategyPackageManager,
	}}
}

func twoDigit(value int) string {
	if value < 10 {
		return "0" + string(rune('0'+value))
	}
	return string(rune('0'+value/10)) + string(rune('0'+value%10))
}
