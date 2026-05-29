package tui

import (
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsatif/freshctl/internal/catalog"
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

func TestCatalogBreadcrumbIncludesRoot(t *testing.T) {
	model := Model{
		categories:  catalog.Default(),
		catalogPath: []int{3, 1},
	}

	if got := model.currentBreadcrumb(); got != "Catalog > Media > Images & Graphics" {
		t.Fatalf("breadcrumb should include catalog root, got %q", got)
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

	view := stripANSI(fitDetailsLines(packageDetailsLines(app, "No"), 40))
	for _, want := range []string{
		"Package:",
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
}

func TestPackageDetailsPanelShowsCLIToolMetadata(t *testing.T) {
	apps := packagesByIDForTUITest(catalog.Default())
	app, ok := apps["helix"]
	if !ok {
		t.Fatal("expected Helix in default catalog")
	}

	view := stripANSI(fitDetailsLines(packageDetailsLines(app, "No"), 44))
	if !strings.Contains(view, "CLI Tool") {
		t.Fatalf("CLI package should render CLI Tool type, got:\n%s", view)
	}
	if !strings.Contains(view, "hx") {
		t.Fatalf("Helix description should mention hx command, got:\n%s", view)
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

	view := stripANSI(fitDetailsLines(packageDetailsLines(app, "Yes"), 26))
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

func TestEnterDeactivatesCatalogSearch(t *testing.T) {
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

	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(Model)
	if got.searchFocused {
		t.Fatal("enter should deactivate active catalog search")
	}
	if got.searchQuery != "python" {
		t.Fatalf("enter should keep the current search query, got %q", got.searchQuery)
	}
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
