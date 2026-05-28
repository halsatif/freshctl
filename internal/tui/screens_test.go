package tui

import (
	"regexp"
	"sort"
	"strings"
	"testing"

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
