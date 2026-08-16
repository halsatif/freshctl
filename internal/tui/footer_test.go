package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/halsatif/freshctl/internal/catalog"
	"github.com/muesli/termenv"
)

func TestFooterSeparatesAndStylesHintParts(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(profile)

	rendered := renderFooter(120, []keyHint{
		hint("↑↓", "move", footerPriorityNavigation),
		hint("enter", "continue", footerPriorityPrimary),
		hint("?", "help", footerPriorityTertiary),
	})
	plain := stripANSI(rendered)

	if !strings.Contains(plain, "↑↓ move  •  enter continue  •  ? help") {
		t.Fatalf("footer should render key, description, and bullet separators, got %q", plain)
	}
	if strings.Contains(plain, "|") {
		t.Fatalf("footer should not use ASCII pipe separators, got %q", plain)
	}
	if !strings.Contains(rendered, "\x1b[") {
		t.Fatalf("footer keys and descriptions should be styled independently, got %q", rendered)
	}
}

func TestFooterDegradesByRemovingWholeLowPriorityHints(t *testing.T) {
	hints := []keyHint{
		hint("↑↓", "move", footerPriorityNavigation),
		hint("space", "select", footerPrioritySecondary),
		hint("/", "search", footerPriorityTertiary),
		hint("enter", "continue", footerPriorityPrimary),
		hint("?", "help", footerPriorityTertiary),
	}

	full := stripANSI(renderFooter(120, hints))
	narrowRendered := renderFooter(24, hints)
	narrow := stripANSI(narrowRendered)
	if !strings.Contains(full, "/ search") || !strings.Contains(full, "? help") {
		t.Fatalf("full footer should include secondary hints, got %q", full)
	}
	if !strings.Contains(narrow, "enter continue") {
		t.Fatalf("narrow footer should preserve the primary action, got %q", narrow)
	}
	if strings.Contains(narrow, "? help") || strings.Contains(narrow, "/ search") {
		t.Fatalf("narrow footer should hide low-priority hints as complete units, got %q", narrow)
	}
	if ansi.StringWidth(narrowRendered) > 24 {
		t.Fatalf("narrow footer exceeds available width: width=%d output=%q", ansi.StringWidth(narrowRendered), narrow)
	}
}

func TestHelpFooterStaysVisibleAndFitsNarrowTerminal(t *testing.T) {
	model := Model{
		screen:      screenHelp,
		helpBack:    screenCatalog,
		width:       34,
		height:      8,
		catalogMode: catalogModeFull,
	}
	view := stripANSI(model.View())
	if !strings.Contains(view, "esc back") {
		t.Fatalf("help footer should remain visible at narrow sizes:\n%s", view)
	}
	for _, line := range strings.Split(view, "\n") {
		if ansi.StringWidth(line) > model.width {
			t.Fatalf("help line exceeds terminal width %d: %q", model.width, line)
		}
	}
}

func TestContextualFootersOnlyShowAvailablePrimaryActions(t *testing.T) {
	catalogModel := Model{width: 100, catalogMode: catalogModeFull}
	catalogFooter := stripANSI(catalogModel.footerFor(screenCatalog))
	for _, expected := range []string{"↑↓ move", "space select", "/ search", "enter review", "? help"} {
		if !strings.Contains(catalogFooter, expected) {
			t.Fatalf("catalog footer missing %q: %s", expected, catalogFooter)
		}
	}
	for _, secondary := range []string{"p presets", "o profile", "i install", "q quit"} {
		if strings.Contains(catalogFooter, secondary) {
			t.Fatalf("catalog footer should leave secondary action %q for help: %s", secondary, catalogFooter)
		}
	}
	catalogModel.searchQuery = "code"
	if footer := stripANSI(catalogModel.footerFor(screenCatalog)); !strings.Contains(footer, "esc clear") {
		t.Fatalf("catalog footer should describe esc as clear while a search query is active: %s", footer)
	}

	installModel := Model{width: 100, installDone: true}
	if footer := stripANSI(installModel.footerFor(screenInstall)); strings.Contains(footer, "s skip") {
		t.Fatalf("completed install must not advertise unavailable skip action: %s", footer)
	}
	installModel.installDone = false
	installModel.currentApp = catalog.Application{ID: "firefox"}
	installModel.skipInstall = make(chan struct{}, 1)
	if footer := stripANSI(installModel.footerFor(screenInstall)); !strings.Contains(footer, "s skip") {
		t.Fatalf("active install should advertise skip action: %s", footer)
	}
}

func TestHelpOpensAndReturnsWithoutChangingCatalogState(t *testing.T) {
	model := Model{
		screen:        screenCatalog,
		width:         100,
		height:        28,
		categories:    catalog.Default(),
		catalogMode:   catalogModeFull,
		catalogCursor: 4,
		catalogScroll: 2,
		searchQuery:   "code",
		selected:      map[string]bool{"git": true},
	}

	updated, _ := model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	help := updated.(Model)
	if help.screen != screenHelp || help.helpBack != screenCatalog {
		t.Fatalf("? should open help for catalog, screen=%v back=%v", help.screen, help.helpBack)
	}
	view := stripANSI(help.View())
	for _, expected := range []string{"shortcuts for catalog", "j/k", "p", "open presets", "ctrl+c"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("catalog help missing %q:\n%s", expected, view)
		}
	}

	updated, _ = help.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	closed := updated.(Model)
	if closed.screen != screenCatalog || closed.catalogCursor != 4 || closed.catalogScroll != 2 ||
		closed.searchQuery != "code" || !closed.selected["git"] {
		t.Fatalf("closing help should restore exact catalog state: %#v", closed)
	}
}

func TestQuestionMarkAlsoClosesHelp(t *testing.T) {
	model := Model{screen: screenHelp, helpBack: screenReview}
	updated, _ := model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if got := updated.(Model).screen; got != screenReview {
		t.Fatalf("? should close help and restore review, got %v", got)
	}
}

func TestHelpPreservesActiveInstallAndTicking(t *testing.T) {
	model := Model{
		screen:        screenInstall,
		installDone:   false,
		currentStep:   3,
		currentStatus: "Downloading... 70%",
		logScroll:     4,
		installLogs:   fakeInstallLogEntries(10),
	}
	updated, _ := model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	help := updated.(Model)
	ticked, cmd := help.handleInstallTick()
	got := ticked.(Model)
	if cmd == nil || got.spinnerFrame != help.spinnerFrame+1 {
		t.Fatal("installation ticks should continue while help is open")
	}
	if got.currentStep != 3 || got.currentStatus != "Downloading... 70%" || got.logScroll != 4 || len(got.installLogs) != 10 {
		t.Fatal("help must not reset installation or log state")
	}
}

func TestEnterIsPrimaryCatalogAction(t *testing.T) {
	full := Model{
		screen:      screenCatalog,
		categories:  catalog.Default(),
		catalogMode: catalogModeFull,
		selected:    map[string]bool{},
	}
	updated, _ := full.handleCatalogKey(tea.KeyMsg{Type: tea.KeyEnter})
	if got := updated.(Model).screen; got != screenReview {
		t.Fatalf("enter in full catalog should open review, got %v", got)
	}

	app := catalog.Application{Name: "Example", ID: "example"}
	categories := Model{
		screen:      screenCatalog,
		categories:  []catalog.Category{{Name: "Tools", Apps: []catalog.Application{app}}},
		catalogMode: catalogModeCategories,
		selected:    map[string]bool{},
	}
	updated, _ = categories.handleCatalogKey(tea.KeyMsg{Type: tea.KeyEnter})
	opened := updated.(Model)
	if opened.screen != screenCatalog || len(opened.catalogPath) != 1 {
		t.Fatalf("enter on category should open it, got screen=%v path=%v", opened.screen, opened.catalogPath)
	}
	updated, _ = opened.handleCatalogKey(tea.KeyMsg{Type: tea.KeyEnter})
	if got := updated.(Model).screen; got != screenReview {
		t.Fatalf("enter on package should open review, got %v", got)
	}
}

func TestNavigationAliasesAndSpaceSelectionRemainConsistent(t *testing.T) {
	model := Model{
		screen:      screenCatalog,
		categories:  catalog.Default(),
		catalogMode: catalogModeFull,
		selected:    map[string]bool{},
	}
	updated, _ := model.handleCatalogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	moved := updated.(Model)
	if moved.catalogCursor != 1 {
		t.Fatalf("j should move down, got cursor %d", moved.catalogCursor)
	}
	updated, _ = moved.handleCatalogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	moved = updated.(Model)
	if moved.catalogCursor != 0 {
		t.Fatalf("k should move up, got cursor %d", moved.catalogCursor)
	}
	app, ok := moved.currentPackageSelection()
	if !ok {
		t.Fatal("expected highlighted package")
	}
	updated, _ = moved.handleCatalogKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !updated.(Model).selected[app.ID] {
		t.Fatal("space should toggle highlighted package")
	}
}

func TestEscIsBackAndObsoleteBackAliasesDoNothing(t *testing.T) {
	review := Model{screen: screenReview}
	for _, obsolete := range []rune{'b', 'h'} {
		updated, _ := review.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{obsolete}})
		if got := updated.(Model).screen; got != screenReview {
			t.Fatalf("obsolete %q alias should not leave review, got %v", obsolete, got)
		}
	}
	updated, _ := review.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if got := updated.(Model).screen; got != screenCatalog {
		t.Fatalf("esc should return from review to catalog, got %v", got)
	}

	categories := Model{
		screen:      screenCatalog,
		categories:  catalog.Default(),
		catalogMode: catalogModeCategories,
		catalogPath: []int{0},
		selected:    map[string]bool{},
	}
	updated, _ = categories.handleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	if got := updated.(Model); len(got.catalogPath) != 1 {
		t.Fatalf("backspace outside search should no longer navigate back, path=%v", got.catalogPath)
	}
	updated, _ = categories.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if got := updated.(Model); len(got.catalogPath) != 1 {
		t.Fatalf("h should no longer navigate back, path=%v", got.catalogPath)
	}
}

func TestEscReturnsOneScreenBackFromIntermediateScreens(t *testing.T) {
	mode := Model{screen: screenModeSelect}
	updated, _ := mode.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if got := updated.(Model).screen; got != screenWelcome {
		t.Fatalf("esc from mode selection should return to welcome, got %v", got)
	}

	elevation := Model{screen: screenElevation, elevationBack: screenBootstrap}
	updated, _ = elevation.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if got := updated.(Model).screen; got != screenBootstrap {
		t.Fatalf("esc from elevation should restore its source screen, got %v", got)
	}

	install := Model{screen: screenInstall, installDone: true}
	updated, _ = install.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if got := updated.(Model).screen; got != screenCatalog {
		t.Fatalf("esc from completed install should return to catalog, got %v", got)
	}
}

func TestQIsSafeQuitAndCtrlCRemainsEmergencyExit(t *testing.T) {
	cancelled := false
	model := Model{
		screen:        screenInstall,
		installDone:   false,
		cancelInstall: func() { cancelled = true },
	}
	updated, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil || cancelled || updated.(Model).screen != screenInstall {
		t.Fatal("q should not cancel or quit during an active install")
	}

	_, cmd = model.handleKey(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil || !cancelled {
		t.Fatal("ctrl+c should remain the emergency exit during installation")
	}

	model.installDone = true
	_, cmd = model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("q should quit after installation is complete")
	}
}

func TestSlashStartsCatalogSearch(t *testing.T) {
	model := Model{
		screen:      screenCatalog,
		categories:  catalog.Default(),
		catalogMode: catalogModeFull,
		selected:    map[string]bool{},
	}
	updated, cmd := model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	got := updated.(Model)
	if !got.searchFocused || cmd == nil {
		t.Fatal("/ should focus catalog search and start cursor ticking")
	}
}
