package tui

import (
	"errors"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsatif/freshctl/internal/catalog"
	"github.com/halsatif/freshctl/internal/installer"
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
	if count := strings.Count(view, "up/down move"); count != 1 {
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
	app := catalog.Package{
		Name:        "Visual Studio Code",
		Description: "Code editor with extensions and integrated tools.",
		PackageID:   "vscode",
		Type:        catalog.PackageTypeApplication,
		Source:      catalog.PackageSourceChocolatey,
		Verified:    true,
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
		"Chocolatey",
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
	app := catalog.Package{
		Name:         "Missing CLI",
		Description:  "Test command-line tool.",
		PackageID:    "missing-cli",
		Type:         catalog.PackageTypeCLITool,
		Source:       catalog.PackageSourceChocolatey,
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
	app := catalog.Package{
		Name:        "No Detection",
		Description: "Package without detection metadata.",
		PackageID:   "no-detection",
		Type:        catalog.PackageTypeApplication,
		Source:      catalog.PackageSourceChocolatey,
		Verified:    true,
	}

	view := stripANSI(fitDetailsLines(packageDetailsLines(app, "No", ""), 44, 18))
	if strings.Contains(view, "Installed:") {
		t.Fatalf("details panel should hide installed status without detection metadata, got:\n%s", view)
	}
}

func TestInstalledStatusCachePopulatesDetectedPackages(t *testing.T) {
	app := catalog.Package{
		Name:         "Cached Tool",
		PackageID:    "cached-tool",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "cached-tool.exe",
	}
	model := Model{
		categories: []catalog.Category{{Apps: []catalog.Package{app}}},
		detectInstalled: func(pkg catalog.Package) bool {
			return pkg.PackageID == "cached-tool"
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
		categories: []catalog.Category{{Apps: []catalog.Package{{
			Name:      "No Detection",
			PackageID: "no-detection",
		}}}},
		detectInstalled: func(catalog.Package) bool {
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
	app := catalog.Package{
		Name:         "Refresh Tool",
		PackageID:    "refresh-tool",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "refresh-tool.exe",
	}
	installed := false
	model := Model{
		categories: []catalog.Category{{Apps: []catalog.Package{app}}},
		detectInstalled: func(catalog.Package) bool {
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

func TestDetailsPanelUsesCachedInstalledStatus(t *testing.T) {
	app := catalog.Package{
		Name:         "Cached Missing Tool",
		Description:  "Tool that should read installed state from cache.",
		PackageID:    "cached-missing-tool",
		Type:         catalog.PackageTypeCLITool,
		Source:       catalog.PackageSourceChocolatey,
		DetectMethod: catalog.DetectPath,
		DetectValue:  "freshctl-definitely-not-installed.exe",
		Verified:     true,
	}
	model := Model{
		screen:      screenCatalog,
		width:       100,
		height:      32,
		categories:  []catalog.Category{{Apps: []catalog.Package{app}}},
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
	app := catalog.Package{
		Name:         "Cached Installed",
		PackageID:    "cached-installed",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "cached-installed.exe",
	}
	model := Model{
		categories:  []catalog.Category{{Apps: []catalog.Package{app}}},
		catalogMode: catalogModeFull,
		selected:    map[string]bool{},
		installed: map[string]InstalledStatus{
			"cached-installed": {Installed: true, Checked: true},
		},
	}

	view := stripANSI(strings.Join(model.fullCatalogListLines(), "\n"))
	if !strings.Contains(view, "Cached Installed") || !strings.Contains(view, "Installed") {
		t.Fatalf("installed package row should show Installed from cache, got:\n%s", view)
	}
}

func TestCatalogListShowsNotInstalledStatusFromCache(t *testing.T) {
	app := catalog.Package{
		Name:         "Cached Missing",
		PackageID:    "cached-missing",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "cached-missing.exe",
	}
	model := Model{
		categories:  []catalog.Category{{Apps: []catalog.Package{app}}},
		catalogMode: catalogModeFull,
		selected:    map[string]bool{},
		installed: map[string]InstalledStatus{
			"cached-missing": {Installed: false, Checked: true},
		},
	}

	view := stripANSI(strings.Join(model.fullCatalogListLines(), "\n"))
	if !strings.Contains(view, "Cached Missing") || !strings.Contains(view, "Not installed") {
		t.Fatalf("not installed package row should show Not installed from cache, got:\n%s", view)
	}
}

func TestCatalogListHidesStatusWithoutDetectionMetadata(t *testing.T) {
	app := catalog.Package{
		Name:      "No Detection",
		PackageID: "no-detection",
	}
	model := Model{
		categories:  []catalog.Category{{Apps: []catalog.Package{app}}},
		catalogMode: catalogModeFull,
		selected:    map[string]bool{},
		installed: map[string]InstalledStatus{
			"no-detection": {Installed: true, Checked: true},
		},
	}

	view := stripANSI(strings.Join(model.fullCatalogListLines(), "\n"))
	if strings.Contains(view, "Installed") || strings.Contains(view, "Not installed") {
		t.Fatalf("package without detection metadata should not show installed status, got:\n%s", view)
	}
}

func TestCatalogSearchResultsShowInstalledStatus(t *testing.T) {
	app := catalog.Package{
		Name:         "Search Installed Tool",
		Description:  "Searchable tool.",
		PackageID:    "search-installed-tool",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "search-installed-tool.exe",
	}
	model := Model{
		categories:    []catalog.Category{{Apps: []catalog.Package{app}}},
		catalogMode:   catalogModeFull,
		searchFocused: true,
		searchQuery:   "search",
		selected:      map[string]bool{},
		installed: map[string]InstalledStatus{
			"search-installed-tool": {Installed: true, Checked: true},
		},
	}

	view := stripANSI(strings.Join(model.catalogListLines(), "\n"))
	if !strings.Contains(view, "Search Installed Tool") || !strings.Contains(view, "Installed") {
		t.Fatalf("search result row should show installed status from cache, got:\n%s", view)
	}
}

func TestCatalogListRenderDoesNotCallDetection(t *testing.T) {
	app := catalog.Package{
		Name:         "Render Cached Tool",
		Description:  "Cached render test.",
		PackageID:    "render-cached-tool",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "render-cached-tool.exe",
	}
	model := Model{
		screen:      screenCatalog,
		width:       100,
		height:      32,
		categories:  []catalog.Category{{Apps: []catalog.Package{app}}},
		catalogMode: catalogModeFull,
		selected:    map[string]bool{},
		installed: map[string]InstalledStatus{
			"render-cached-tool": {Installed: true, Checked: true},
		},
		detectInstalled: func(catalog.Package) bool {
			t.Fatal("render should not call installed detection")
			return false
		},
	}

	view := stripANSI(model.View())
	if !strings.Contains(view, "Installed") {
		t.Fatalf("render should show cached installed status, got:\n%s", view)
	}
}

func TestInstallSummaryRefreshesInstalledStatusCacheOnce(t *testing.T) {
	app := catalog.Package{
		Name:         "Refresh Once Tool",
		Description:  "Refresh once test.",
		PackageID:    "refresh-once-tool",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "refresh-once-tool.exe",
	}
	calls := 0
	model := Model{
		screen:      screenInstall,
		categories:  []catalog.Category{{Apps: []catalog.Package{app}}},
		installApps: []catalog.Package{app},
		appStatus:   map[string]string{"refresh-once-tool": "installed"},
		selected:    map[string]bool{},
		installed: map[string]InstalledStatus{
			"refresh-once-tool": {Installed: false, Checked: true},
		},
		detectInstalled: func(catalog.Package) bool {
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
	app := catalog.Package{
		Name:         "Detectable Success",
		Description:  "Detectable success test.",
		PackageID:    "detectable-success",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "detectable-success.exe",
	}
	model := Model{
		screen:      screenInstall,
		categories:  []catalog.Category{{Apps: []catalog.Package{app}}},
		installApps: []catalog.Package{app},
		appStatus:   map[string]string{"detectable-success": "installed"},
		selected:    map[string]bool{},
		installed: map[string]InstalledStatus{
			"detectable-success": {Installed: false, Checked: true},
		},
		detectInstalled: func(catalog.Package) bool {
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
	app := catalog.Package{
		Name:         "Detectable Failure",
		Description:  "Detectable failure test.",
		PackageID:    "detectable-failure",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "detectable-failure.exe",
	}
	model := Model{
		screen:      screenInstall,
		categories:  []catalog.Category{{Apps: []catalog.Package{app}}},
		installApps: []catalog.Package{app},
		appStatus:   map[string]string{"detectable-failure": "failed"},
		selected:    map[string]bool{},
		installed: map[string]InstalledStatus{
			"detectable-failure": {Installed: false, Checked: true},
		},
		detectInstalled: func(catalog.Package) bool {
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
	app := catalog.Package{
		Name:         "Reflected Tool",
		Description:  "Reflected status test.",
		PackageID:    "reflected-tool",
		DetectMethod: catalog.DetectPath,
		DetectValue:  "reflected-tool.exe",
		Type:         catalog.PackageTypeCLITool,
		Source:       catalog.PackageSourceChocolatey,
		Verified:     true,
	}
	model := Model{
		screen:      screenInstall,
		width:       100,
		height:      32,
		categories:  []catalog.Category{{Apps: []catalog.Package{app}}},
		catalogMode: catalogModeFull,
		installApps: []catalog.Package{app},
		appStatus:   map[string]string{"reflected-tool": "installed"},
		selected:    map[string]bool{},
		installed: map[string]InstalledStatus{
			"reflected-tool": {Installed: false, Checked: true},
		},
		detectInstalled: func(catalog.Package) bool {
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

func TestPackageDetailsPanelFitsNarrowWidth(t *testing.T) {
	app := catalog.Package{
		Name:        "Long App",
		Description: strings.Repeat("long metadata ", 12),
		PackageID:   "very-long-package-id-that-should-be-truncated",
		Type:        catalog.PackageTypeApplication,
		Source:      catalog.PackageSourceChocolatey,
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
		selected[item.PackageID] = true
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
		selected[app.PackageID] = true
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
	if count := strings.Count(view, "l show/hide logs"); count != 1 {
		t.Fatalf("bootstrap footer should remain visible once, got %d in:\n%s", count, view)
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
		model.appStatus[app.PackageID] = "pending"
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
	if !strings.Contains(secondView, "up/down scroll") {
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
		model.appStatus[app.PackageID] = "pending"
	}
	model.appStatus[apps[0].PackageID] = "installing"

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
	app := catalog.Package{Name: "Example App", PackageID: "example"}
	model := Model{
		screen:       screenInstall,
		width:        100,
		height:       24,
		installApps:  []catalog.Package{app},
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
		searchQuery: "codex-cli",
		selected:    map[string]bool{},
	}

	items := model.filteredFullCatalogItems()
	if len(items) != 1 || items[0].Package.PackageID != "codex-cli" {
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
		if item.Package.PackageID == "vcredist140" {
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
		if item.Package.PackageID == packageID {
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

func collectTestPackages(categories []catalog.Category) []catalog.Package {
	var apps []catalog.Package
	for _, category := range categories {
		apps = append(apps, collectTestPackages(category.Categories)...)
		apps = append(apps, category.Apps...)
	}
	return apps
}

func packagesByIDForTUITest(categories []catalog.Category) map[string]catalog.Package {
	apps := make(map[string]catalog.Package)
	for _, app := range collectTestPackages(categories) {
		apps[app.PackageID] = app
	}
	return apps
}

func fakeInstallPackages(count int) []catalog.Package {
	apps := make([]catalog.Package, 0, count)
	for i := 1; i <= count; i++ {
		apps = append(apps, catalog.Package{
			Name:      "Package " + twoDigit(i),
			PackageID: "pkg-" + twoDigit(i),
		})
	}
	return apps
}

func twoDigit(value int) string {
	if value < 10 {
		return "0" + string(rune('0'+value))
	}
	return string(rune('0'+value/10)) + string(rune('0'+value%10))
}
