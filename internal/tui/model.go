package tui

import (
	"context"
	"sort"
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsatif/freshctl/internal/catalog"
	"github.com/halsatif/freshctl/internal/installer"
)

type screen int

const (
	screenWelcome screen = iota
	screenModeSelect
	screenCatalog
	screenReview
	screenInstall
	screenBootstrap
	screenElevation
	screenBrokenChocolatey
)

type catalogMode int

const (
	catalogModeFull catalogMode = iota
	catalogModeCategories
)

type Model struct {
	screen screen

	width  int
	height int

	categories    []catalog.Category
	catalogPath   []int
	catalogCursor int
	catalogScroll int
	modeCursor    int
	catalogMode   catalogMode
	searchFocused bool
	searchQuery   string
	searchCursor  bool
	selected      map[string]bool
	notice        string
	reviewScroll  int
	installScroll int

	installEvents chan installer.Event
	skipInstall   chan struct{}
	cancelInstall context.CancelFunc
	installApps   []catalog.Package
	installLog    []string
	fullLog       []string
	showFullLog   bool
	results       []installer.Result
	appStatus     map[string]string
	appElapsed    map[string]time.Duration
	currentApp    catalog.Package
	currentStep   int
	currentCmd    string
	currentStart  time.Time
	spinnerFrame  int
	installDone   bool

	bootstrapBack    screen
	bootstrapEvents  chan installer.BootstrapEvent
	bootstrapLog     []string
	bootstrapStatus  string
	showBootstrapLog bool
	bootstrapRunning bool
	cancelBootstrap  context.CancelFunc

	elevationRunning bool
	elevationError   string
	elevationArgs    []string

	brokenBack    screen
	brokenRunning bool
	brokenError   string
}

type installEventMsg struct {
	event installer.Event
	ok    bool
}

type bootstrapEventMsg struct {
	event installer.BootstrapEvent
	ok    bool
}

type retryPackageManagerMsg struct {
	ready  bool
	broken bool
}

type elevationMsg struct {
	err error
}

type brokenRecoveryMsg struct {
	err error
}

type installTickMsg struct{}

type searchCursorTickMsg struct{}

func NewModel(args []string) Model {
	initialScreen := screenWelcome
	bootstrapStatus := ""
	selected := selectedFromArgs(args)
	if selected == nil {
		selected = make(map[string]bool)
	}
	if installer.HasBrokenPackageManagerInstall() {
		initialScreen = screenBrokenChocolatey
	} else if !installer.HasPackageManager() {
		if installer.IsElevated() {
			initialScreen = screenBootstrap
			bootstrapStatus = "Chocolatey was not found on this system."
		} else {
			initialScreen = screenElevation
		}
	} else if len(selected) > 0 {
		initialScreen = screenReview
	}

	return Model{
		screen:          initialScreen,
		categories:      catalog.Default(),
		selected:        selected,
		bootstrapBack:   screenWelcome,
		bootstrapStatus: bootstrapStatus,
		brokenBack:      screenWelcome,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.clampReviewScroll()
		m.clampInstallScroll()
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	case installEventMsg:
		return m.handleInstallEvent(msg)
	case bootstrapEventMsg:
		return m.handleBootstrapEvent(msg)
	case retryPackageManagerMsg:
		return m.handleRetryPackageManagerMsg(msg)
	case elevationMsg:
		return m.handleElevationMsg(msg)
	case brokenRecoveryMsg:
		return m.handleBrokenRecoveryMsg(msg)
	case installTickMsg:
		return m.handleInstallTick()
	case searchCursorTickMsg:
		return m.handleSearchCursorTick()
	}

	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case screenWelcome:
		return m.viewWelcome()
	case screenModeSelect:
		return m.viewModeSelect()
	case screenCatalog:
		return m.viewCatalog()
	case screenReview:
		return m.viewReview()
	case screenInstall:
		return m.viewInstall()
	case screenBootstrap:
		return m.viewBootstrap()
	case screenElevation:
		return m.viewElevation()
	case screenBrokenChocolatey:
		return m.viewBrokenChocolatey()
	default:
		return ""
	}
}

func keyName(msg tea.KeyMsg) string {
	key := msg.String()
	if len(msg.Runes) != 1 {
		return key
	}
	if normalized, ok := russianKeyboardAliases[unicode.ToLower(msg.Runes[0])]; ok {
		return normalized
	}
	return key
}

func dropLastRune(value string) string {
	if value == "" {
		return ""
	}
	runes := []rune(value)
	return string(runes[:len(runes)-1])
}

func searchTextInput(msg tea.KeyMsg) (string, bool) {
	if msg.Alt || len(msg.Runes) == 0 {
		return "", false
	}

	runes := make([]rune, 0, len(msg.Runes))
	for _, r := range msg.Runes {
		if r == unicode.ReplacementChar || !unicode.IsPrint(r) || unicode.IsControl(r) {
			continue
		}
		runes = append(runes, r)
	}
	if len(runes) == 0 {
		return "", false
	}
	return string(runes), true
}

var russianKeyboardAliases = map[rune]string{
	'й': "q",
	'ц': "w",
	'у': "e",
	'к': "r",
	'е': "t",
	'н': "y",
	'г': "u",
	'ш': "i",
	'щ': "o",
	'з': "p",
	'х': "[",
	'ъ': "]",
	'ф': "a",
	'ы': "s",
	'в': "d",
	'а': "f",
	'п': "g",
	'р': "h",
	'о': "j",
	'л': "k",
	'д': "l",
	'ж': ";",
	'э': "'",
	'я': "z",
	'ч': "x",
	'с': "c",
	'м': "v",
	'и': "b",
	'т': "n",
	'ь': "m",
	'б': ",",
	'ю': ".",
	'.': "/",
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyName(msg) {
	case "ctrl+c", "q":
		if m.screen == screenInstall && m.cancelInstall != nil {
			m.cancelInstall()
		}
		if m.screen == screenBootstrap && m.cancelBootstrap != nil {
			m.cancelBootstrap()
		}
		return m, tea.Quit
	}

	switch m.screen {
	case screenWelcome:
		if keyName(msg) == "enter" {
			m.screen = screenModeSelect
			return m, tea.ClearScreen
		}
	case screenModeSelect:
		return m.handleModeSelectKey(msg)
	case screenCatalog:
		return m.handleCatalogKey(msg)
	case screenReview:
		return m.handleReviewKey(msg)
	case screenInstall:
		return m.handleInstallKey(msg)
	case screenBootstrap:
		return m.handleBootstrapKey(msg)
	case screenElevation:
		return m.handleElevationKey(msg)
	case screenBrokenChocolatey:
		return m.handleBrokenChocolateyKey(msg)
	}

	return m, nil
}

func (m Model) handleModeSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyName(msg) {
	case "up", "k", "down", "j":
		if m.modeCursor == 0 {
			m.modeCursor = 1
		} else {
			m.modeCursor = 0
		}
	case "enter":
		if m.modeCursor == 0 {
			m.catalogMode = catalogModeFull
			m.searchFocused = false
			m.searchQuery = ""
		} else {
			m.catalogMode = catalogModeCategories
			m.searchFocused = false
			m.searchQuery = ""
		}
		m.catalogCursor = 0
		m.catalogScroll = 0
		m.catalogPath = nil
		m.notice = ""
		m.screen = screenCatalog
		return m, tea.ClearScreen
	}
	return m, nil
}

func (m Model) handleCatalogKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.notice = ""

	if m.searchFocused {
		switch keyName(msg) {
		case "up", "k":
			m.moveCatalogCursor(-1)
		case "down", "j":
			m.moveCatalogCursor(1)
		case "enter":
			m.searchFocused = false
			m.searchCursor = false
		case "esc":
			m.searchFocused = false
			m.searchCursor = false
			m.searchQuery = ""
			m.catalogCursor = 0
			m.catalogScroll = 0
		case " ":
			m.searchQuery += " "
			m.clampCatalogCursor()
			m.ensureCatalogCursorVisible()
		case "backspace":
			if len(m.searchQuery) > 0 {
				m.searchQuery = dropLastRune(m.searchQuery)
				m.clampCatalogCursor()
				m.ensureCatalogCursorVisible()
			}
		default:
			if text, ok := searchTextInput(msg); ok {
				m.searchQuery += text
				m.clampCatalogCursor()
				m.ensureCatalogCursorVisible()
			}
		}
		return m, nil
	}

	if m.searchActive() && keyName(msg) == "esc" {
		m.searchQuery = ""
		m.catalogCursor = 0
		m.catalogScroll = 0
		return m, nil
	}

	switch keyName(msg) {
	case "tab":
		return m, nil
	case "up", "k":
		m.moveCatalogCursor(-1)
	case "down", "j":
		m.moveCatalogCursor(1)
	case "esc", "backspace", "h":
		if m.catalogMode == catalogModeFull {
			m.screen = screenModeSelect
			m.searchFocused = false
			m.searchQuery = ""
			m.catalogCursor = 0
			m.catalogScroll = 0
			return m, tea.ClearScreen
		}
		wasAtRoot := len(m.catalogPath) == 0
		m.goBackInCatalog()
		if wasAtRoot {
			m.screen = screenModeSelect
		}
		return m, tea.ClearScreen
	case " ":
		m.toggleCurrentApp()
	case "enter":
		if m.catalogMode == catalogModeCategories {
			m.openCurrentCategory()
			return m, tea.ClearScreen
		}
	case "i":
		m.screen = screenReview
		m.reviewScroll = 0
		return m, tea.ClearScreen
	case "/":
		m.searchFocused = true
		m.searchCursor = true
		m.searchQuery = ""
		m.catalogCursor = 0
		m.catalogScroll = 0
		return m, searchCursorTickCmd()
	}

	return m, nil
}

func (m Model) handleReviewKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyName(msg) {
	case "up", "k":
		m.moveReviewScroll(-1)
	case "down", "j":
		m.moveReviewScroll(1)
	case "pgup":
		m.moveReviewScroll(-m.reviewVisibleRows())
	case "pgdown":
		m.moveReviewScroll(m.reviewVisibleRows())
	case "home":
		m.reviewScroll = 0
	case "end":
		m.reviewScroll = maxInt(0, len(m.selectedApps())-m.reviewVisibleRows())
	case "b", "esc":
		m.notice = ""
		m.screen = screenCatalog
		return m, tea.ClearScreen
	case "enter":
		apps := m.selectedApps()
		if len(apps) == 0 {
			m.notice = "No apps selected. Go back and choose at least one app."
			return m, nil
		}
		if installer.HasBrokenPackageManagerInstall() {
			m.notice = ""
			m.screen = screenBrokenChocolatey
			m.brokenBack = screenReview
			m.brokenError = ""
			m.brokenRunning = false
			return m, nil
		}
		if !installer.HasPackageManager() {
			m.notice = ""
			if !installer.IsElevated() {
				m.screen = screenElevation
				m.elevationError = ""
				m.elevationRunning = false
				m.elevationArgs = m.selectedArgs()
				return m, nil
			}
			m.screen = screenBootstrap
			m.bootstrapBack = screenReview
			m.bootstrapLog = nil
			m.bootstrapStatus = "Chocolatey was not found on this system."
			m.showBootstrapLog = false
			m.bootstrapRunning = false
			return m, nil
		}
		if !installer.IsElevated() {
			m.screen = screenElevation
			m.elevationError = ""
			m.elevationRunning = false
			m.elevationArgs = m.selectedArgs()
			return m, nil
		}
		m.notice = ""
		return m.startInstall(apps)
	}

	return m, nil
}

func (m *Model) moveReviewScroll(delta int) {
	m.reviewScroll += delta
	m.clampReviewScroll()
}

func (m *Model) clampReviewScroll() {
	selected := len(m.selectedApps())
	visible := m.reviewVisibleRows()
	maxScroll := selected - visible
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.reviewScroll < 0 {
		m.reviewScroll = 0
	}
	if m.reviewScroll > maxScroll {
		m.reviewScroll = maxScroll
	}
}

func (m Model) reviewVisibleRows() int {
	height := m.height - 15
	if height < 4 {
		return 4
	}
	if height > 14 {
		return 14
	}
	return height
}

func (m Model) handleBootstrapKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyName(msg) {
	case "b", "esc":
		if m.cancelBootstrap != nil {
			m.cancelBootstrap()
		}
		m.bootstrapRunning = false
		m.cancelBootstrap = nil
		m.screen = m.bootstrapBack
		m.notice = "Chocolatey is still missing. Bootstrap Chocolatey or retry detection before installing apps."
	case "l":
		m.showBootstrapLog = !m.showBootstrapLog
		return m, tea.ClearScreen
	case "enter":
		if m.bootstrapRunning {
			return m, nil
		}
		if installer.HasBrokenPackageManagerInstall() {
			m.screen = screenBrokenChocolatey
			m.brokenBack = m.bootstrapBack
			m.brokenError = ""
			m.brokenRunning = false
			return m, nil
		}
		if !installer.IsElevated() {
			m.screen = screenElevation
			m.elevationError = ""
			m.elevationRunning = false
			m.elevationArgs = m.selectedArgs()
			return m, nil
		}
		m.notice = ""
		m.bootstrapRunning = true
		m.bootstrapStatus = "Bootstrapping Chocolatey..."
		m.bootstrapLog = append(m.bootstrapLog, "Running official Chocolatey bootstrap command...")
		ctx, cancel := context.WithCancel(context.Background())
		m.cancelBootstrap = cancel
		m.bootstrapEvents = make(chan installer.BootstrapEvent)
		return m, startBootstrapCmd(ctx, m.bootstrapEvents)
	case "r":
		if m.bootstrapRunning {
			return m, nil
		}
		m.bootstrapStatus = "Retrying Chocolatey detection..."
		m.bootstrapLog = append(m.bootstrapLog, "Retrying Chocolatey detection...")
		return m, retryPackageManagerCmd()
	}

	return m, nil
}

func (m Model) handleBrokenChocolateyKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyName(msg) {
	case "b", "esc":
		m.brokenRunning = false
		m.brokenError = ""
		m.screen = m.brokenBack
		m.notice = "Broken Chocolatey installation must be repaired before bootstrap or installs."
	case "enter":
		if m.brokenRunning {
			return m, nil
		}
		if !installer.IsElevated() {
			m.screen = screenElevation
			m.elevationRunning = false
			m.elevationError = ""
			m.elevationArgs = m.selectedArgs()
			return m, nil
		}
		m.brokenRunning = true
		m.brokenError = ""
		return m, recoverBrokenChocolateyCmd()
	}

	return m, nil
}

func (m Model) handleElevationKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyName(msg) {
	case "enter":
		if m.elevationRunning {
			return m, nil
		}
		m.elevationRunning = true
		m.elevationError = ""
		return m, relaunchElevatedWithArgsCmd(m.elevationArgs)
	}

	return m, nil
}

func (m Model) handleInstallEvent(msg installEventMsg) (tea.Model, tea.Cmd) {
	if !msg.ok {
		m.installDone = true
		return m, nil
	}

	event := msg.event
	switch event.Kind {
	case installer.EventLog:
		m.fullLog = append(m.fullLog, event.Line)
		if isImportantInstallLine(event.Line) {
			m.installLog = append(m.installLog, event.Line)
		}
	case installer.EventAppStarted:
		m.currentApp = event.App
		m.currentCmd = event.Line
		m.currentStep = m.installIndex(event.App) + 1
		m.currentStart = time.Now()
		m.appStatus[event.App.PackageID] = "installing"
		m.installLog = append(m.installLog, "installing "+event.App.Name)
		m.fullLog = append(m.fullLog, "> "+event.Line)
	case installer.EventAppFinished:
		if !m.currentStart.IsZero() && m.currentApp.PackageID == event.App.PackageID {
			m.appElapsed[event.App.PackageID] = time.Since(m.currentStart)
		}
		if event.Success {
			m.appStatus[event.App.PackageID] = "installed"
			m.installLog = append(m.installLog, "success "+event.App.Name)
			m.fullLog = append(m.fullLog, "ok: "+event.App.Name)
		} else if event.Err == installer.ErrInstallSkipped {
			m.appStatus[event.App.PackageID] = "skipped"
			m.installLog = append(m.installLog, "skipped "+event.App.Name)
			m.fullLog = append(m.fullLog, "skipped: "+event.App.Name)
		} else if event.Err == context.DeadlineExceeded {
			m.appStatus[event.App.PackageID] = "failed"
			m.installLog = append(m.installLog, "timed out "+event.App.Name)
			m.fullLog = append(m.fullLog, "timed out: "+event.App.Name)
		} else {
			m.appStatus[event.App.PackageID] = "failed"
			m.installLog = append(m.installLog, "failed "+event.App.Name+" - "+event.Err.Error())
			m.fullLog = append(m.fullLog, "failed: "+event.App.Name+" - "+event.Err.Error())
		}
	case installer.EventSummary:
		if !m.currentStart.IsZero() && m.currentApp.PackageID != "" {
			if _, ok := m.appElapsed[m.currentApp.PackageID]; !ok {
				m.appElapsed[m.currentApp.PackageID] = time.Since(m.currentStart)
			}
		}
		m.results = event.Results
		m.installDone = true
		m.cancelInstall = nil
		m.skipInstall = nil
	}

	return m, waitInstallEventCmd(m.installEvents)
}

func (m Model) handleBootstrapEvent(msg bootstrapEventMsg) (tea.Model, tea.Cmd) {
	if !msg.ok {
		m.bootstrapRunning = false
		m.cancelBootstrap = nil
		return m, nil
	}

	event := msg.event
	switch event.Kind {
	case installer.BootstrapLog:
		m.bootstrapLog = append(m.bootstrapLog, event.Line)
	case installer.BootstrapFinished:
		m.bootstrapRunning = false
		m.cancelBootstrap = nil
		if event.Ready {
			m.bootstrapStatus = "Chocolatey installed successfully."
			m.screen = m.bootstrapBack
			m.notice = "Chocolatey is available now. Press enter to continue."
			return m, nil
		}
		m.bootstrapStatus = "Chocolatey bootstrap failed."
		if event.Err != nil {
			m.bootstrapLog = append(m.bootstrapLog, "Bootstrap failed: "+event.Err.Error())
		}
		m.bootstrapLog = append(m.bootstrapLog, "Chocolatey is still not available. Run freshctl as Administrator, bootstrap again, or install Chocolatey manually.")
		return m, nil
	}

	return m, waitBootstrapEventCmd(m.bootstrapEvents)
}

func (m Model) handleRetryPackageManagerMsg(msg retryPackageManagerMsg) (tea.Model, tea.Cmd) {
	if msg.broken {
		m.screen = screenBrokenChocolatey
		m.brokenBack = m.bootstrapBack
		m.brokenError = ""
		m.brokenRunning = false
		return m, nil
	}
	if msg.ready {
		m.bootstrapStatus = "Chocolatey installed successfully."
		m.screen = m.bootstrapBack
		m.notice = "Chocolatey is available now. Press enter to continue."
		return m, nil
	}

	m.bootstrapStatus = "Chocolatey was not found on this system."
	m.bootstrapLog = append(m.bootstrapLog, "Chocolatey is still not available. Run bootstrap again or install Chocolatey manually.")
	return m, nil
}

func (m Model) handleBrokenRecoveryMsg(msg brokenRecoveryMsg) (tea.Model, tea.Cmd) {
	m.brokenRunning = false
	if msg.err != nil {
		m.brokenError = "Could not remove broken Chocolatey folder: " + msg.err.Error()
		return m, nil
	}

	m.screen = screenBootstrap
	m.bootstrapBack = m.brokenBack
	m.bootstrapStatus = "Bootstrapping Chocolatey..."
	m.showBootstrapLog = false
	m.bootstrapLog = []string{"Removed broken C:\\ProgramData\\chocolatey folder.", "Running official Chocolatey bootstrap command..."}
	m.bootstrapRunning = true
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelBootstrap = cancel
	m.bootstrapEvents = make(chan installer.BootstrapEvent)
	return m, startBootstrapCmd(ctx, m.bootstrapEvents)
}

func (m Model) handleElevationMsg(msg elevationMsg) (tea.Model, tea.Cmd) {
	m.elevationRunning = false
	if msg.err != nil {
		m.elevationError = "Could not relaunch as administrator: " + msg.err.Error()
		return m, nil
	}
	return m, tea.Quit
}

func (m Model) handleInstallKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyName(msg) {
	case "up", "k":
		m.moveInstallScroll(-1)
	case "down", "j":
		m.moveInstallScroll(1)
	case "pgup":
		m.moveInstallScroll(-m.installSummaryVisibleRows(0))
	case "pgdown":
		m.moveInstallScroll(m.installSummaryVisibleRows(0))
	case "home":
		m.installScroll = 0
	case "end":
		m.installScroll = maxInt(0, len(m.installApps)-m.installSummaryVisibleRows(0))
	case "l":
		m.showFullLog = !m.showFullLog
		m.clampInstallScroll()
		return m, tea.ClearScreen
	case "s":
		if !m.installDone && m.skipInstall != nil && m.currentApp.PackageID != "" {
			m.installLog = append(m.installLog, "skipping "+m.currentApp.Name+"...")
			m.fullLog = append(m.fullLog, "skip requested for "+m.currentApp.Name)
			m.appStatus[m.currentApp.PackageID] = "skipping"
			select {
			case m.skipInstall <- struct{}{}:
			default:
			}
		}
	}
	return m, nil
}

func (m *Model) moveInstallScroll(delta int) {
	m.installScroll += delta
	m.clampInstallScroll()
}

func (m *Model) clampInstallScroll() {
	visible := m.installSummaryVisibleRows(0)
	maxScroll := len(m.installApps) - visible
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.installScroll < 0 {
		m.installScroll = 0
	}
	if m.installScroll > maxScroll {
		m.installScroll = maxScroll
	}
}

func (m Model) handleInstallTick() (tea.Model, tea.Cmd) {
	if m.screen != screenInstall || m.installDone {
		return m, nil
	}
	m.spinnerFrame++
	if m.showFullLog {
		return m, tea.Batch(tea.ClearScreen, installTickCmd())
	}
	return m, installTickCmd()
}

func (m Model) handleSearchCursorTick() (tea.Model, tea.Cmd) {
	if m.screen != screenCatalog || !m.searchFocused {
		return m, nil
	}
	m.searchCursor = !m.searchCursor
	return m, searchCursorTickCmd()
}

func (m Model) startInstall(apps []catalog.Package) (tea.Model, tea.Cmd) {
	m.screen = screenInstall
	m.installApps = apps
	m.installLog = nil
	m.fullLog = nil
	m.showFullLog = false
	m.results = nil
	m.appStatus = make(map[string]string, len(apps))
	m.appElapsed = make(map[string]time.Duration, len(apps))
	for _, app := range apps {
		m.appStatus[app.PackageID] = "pending"
	}
	m.currentApp = catalog.Package{}
	m.currentStep = 0
	m.currentCmd = ""
	m.currentStart = time.Time{}
	m.installScroll = 0
	m.spinnerFrame = 0
	m.installDone = false
	m.installEvents = make(chan installer.Event)
	m.skipInstall = make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelInstall = cancel
	return m, tea.Batch(tea.ClearScreen, startInstallCmd(ctx, apps, m.installEvents, m.skipInstall), installTickCmd())
}

func (m Model) currentApps() []catalog.Package {
	return m.currentNode().Apps
}

func (m Model) currentCategories() []catalog.Category {
	return m.currentNode().Categories
}

func (m Model) currentNode() catalog.Category {
	node := catalog.Category{Categories: m.categories}
	for _, index := range m.catalogPath {
		if index < 0 || index >= len(node.Categories) {
			return catalog.Category{Categories: m.categories}
		}
		node = node.Categories[index]
	}
	return node
}

func (m Model) currentBreadcrumb() string {
	names := []string{"Catalog"}
	node := catalog.Category{Categories: m.categories}
	for _, index := range m.catalogPath {
		if index < 0 || index >= len(node.Categories) {
			break
		}
		node = node.Categories[index]
		names = append(names, node.Name)
	}
	return strings.Join(names, " > ")
}

func (m *Model) moveCatalogCursor(delta int) {
	count := m.catalogItemCount()
	if count == 0 {
		m.catalogCursor = 0
		return
	}

	m.catalogCursor += delta
	if m.catalogCursor < 0 {
		m.catalogCursor = 0
	}
	if m.catalogCursor >= count {
		m.catalogCursor = count - 1
	}
	m.ensureCatalogCursorVisible()
}

func (m *Model) openCurrentCategory() {
	categories := m.currentCategories()
	if m.catalogCursor >= len(categories) {
		return
	}
	m.catalogPath = append(m.catalogPath, m.catalogCursor)
	m.catalogCursor = 0
	m.catalogScroll = 0
	m.searchFocused = false
	m.searchQuery = ""
}

func (m *Model) goBackInCatalog() {
	if len(m.catalogPath) == 0 {
		return
	}
	m.catalogPath = m.catalogPath[:len(m.catalogPath)-1]
	m.catalogCursor = 0
	m.catalogScroll = 0
	m.searchFocused = false
	m.searchQuery = ""
}

func (m *Model) toggleCurrentApp() {
	app, ok := m.currentPackageSelection()
	if !ok {
		return
	}

	m.selected[app.PackageID] = !m.selected[app.PackageID]
}

func (m Model) catalogItemCount() int {
	if m.catalogMode == catalogModeFull || m.searchActive() {
		return len(m.filteredFullCatalogItems())
	}
	return len(m.currentCategories()) + len(m.currentApps())
}

func (m *Model) clampCatalogCursor() {
	count := m.catalogItemCount()
	if count == 0 {
		m.catalogCursor = 0
		return
	}
	if m.catalogCursor >= count {
		m.catalogCursor = count - 1
	}
	if m.catalogCursor < 0 {
		m.catalogCursor = 0
	}
}

func (m *Model) ensureCatalogCursorVisible() {
	height := m.catalogVisibleRows()
	if height <= 0 {
		m.catalogScroll = 0
		return
	}

	cursorLine := m.catalogCursor
	if m.catalogMode == catalogModeFull {
		cursorLine = m.catalogCursor
	}
	if cursorLine < m.catalogScroll {
		m.catalogScroll = cursorLine
	}
	if cursorLine >= m.catalogScroll+height {
		m.catalogScroll = cursorLine - height + 1
	}
	if m.catalogScroll < 0 {
		m.catalogScroll = 0
	}
}

func (m Model) catalogVisibleRows() int {
	height := m.height - 14
	if height < 6 {
		return 6
	}
	if height > 16 {
		return 16
	}
	return height
}

func (m Model) currentPackageSelection() (catalog.Package, bool) {
	if m.catalogMode == catalogModeFull || m.searchActive() {
		items := m.filteredFullCatalogItems()
		if m.catalogCursor < 0 || m.catalogCursor >= len(items) {
			return catalog.Package{}, false
		}
		return items[m.catalogCursor].Package, true
	}

	apps := m.currentApps()
	appIndex := m.catalogCursor - len(m.currentCategories())
	if appIndex < 0 || appIndex >= len(apps) {
		return catalog.Package{}, false
	}
	return apps[appIndex], true
}

type fullCatalogItem struct {
	Package catalog.Package
	Path    string
}

func (m Model) allCatalogItems() []fullCatalogItem {
	items := make([]fullCatalogItem, 0)
	collectCatalogItems(m.categories, nil, &items)
	sort.SliceStable(items, func(i, j int) bool {
		left := strings.ToLower(items[i].Package.Name)
		right := strings.ToLower(items[j].Package.Name)
		if left == right {
			return items[i].Package.PackageID < items[j].Package.PackageID
		}
		return left < right
	})
	return items
}

func collectCatalogItems(categories []catalog.Category, path []string, items *[]fullCatalogItem) {
	for _, category := range categories {
		nextPath := append(append([]string{}, path...), category.Name)
		collectCatalogItems(category.Categories, nextPath, items)
		for _, app := range category.Apps {
			*items = append(*items, fullCatalogItem{
				Package: app,
				Path:    strings.Join(nextPath, " > "),
			})
		}
	}
}

func (m Model) filteredFullCatalogItems() []fullCatalogItem {
	items := m.allCatalogItems()
	query := strings.TrimSpace(strings.ToLower(m.searchQuery))
	if query == "" {
		return items
	}

	filtered := make([]fullCatalogItem, 0, len(items))
	for _, item := range items {
		if matchesPackageSearch(item, query) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func (m Model) searchActive() bool {
	return m.searchFocused || strings.TrimSpace(m.searchQuery) != ""
}

func matchesPackageSearch(item fullCatalogItem, query string) bool {
	if query == "" {
		return true
	}

	fields := []string{
		item.Package.Name,
		item.Package.PackageID,
		item.Package.Description,
	}
	for _, field := range fields {
		text := strings.ToLower(field)
		if strings.Contains(text, query) || strings.Contains(compactSearchText(text), query) || isSubsequence(query, text) {
			return true
		}
	}
	return strings.Contains(packageInitials(item.Package.Name), query)
}

func compactSearchText(text string) string {
	var builder strings.Builder
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func packageInitials(name string) string {
	var builder strings.Builder
	start := true
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			if start {
				builder.WriteRune(r)
				start = false
			}
			continue
		}
		start = true
	}
	return builder.String()
}

func isSubsequence(query, text string) bool {
	if len(query) < 2 {
		return false
	}
	index := 0
	for _, r := range text {
		if rune(query[index]) == r {
			index++
			if index == len(query) {
				return true
			}
		}
	}
	return false
}

func (m Model) selectedApps() []catalog.Package {
	apps := make([]catalog.Package, 0)
	m.collectSelectedApps(m.categories, &apps)
	return apps
}

func (m Model) collectSelectedApps(categories []catalog.Category, apps *[]catalog.Package) {
	for _, category := range categories {
		m.collectSelectedApps(category.Categories, apps)
		for _, app := range category.Apps {
			if m.selected[app.PackageID] {
				*apps = append(*apps, app)
			}
		}
	}
}

func (m Model) selectedArgs() []string {
	selected := m.selectedApps()
	if len(selected) == 0 {
		return nil
	}

	ids := make([]string, 0, len(selected))
	for _, app := range selected {
		ids = append(ids, app.PackageID)
	}
	args := make([]string, 0, 2)
	if len(ids) > 0 {
		args = append(args, "--selected="+strings.Join(ids, ","))
	}
	return args
}

func selectedFromArgs(args []string) map[string]bool {
	for _, arg := range args {
		if !strings.HasPrefix(arg, "--selected=") {
			continue
		}
		selected := make(map[string]bool)
		for _, id := range strings.Split(strings.TrimPrefix(arg, "--selected="), ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				selected[id] = true
			}
		}
		return selected
	}
	return nil
}

func (m Model) installIndex(app catalog.Package) int {
	for i, candidate := range m.installApps {
		if candidate.PackageID == app.PackageID {
			return i
		}
	}
	return 0
}

func startInstallCmd(ctx context.Context, apps []catalog.Package, events chan installer.Event, skips <-chan struct{}) tea.Cmd {
	return func() tea.Msg {
		go installer.InstallApps(ctx, apps, events, skips)
		event, ok := <-events
		return installEventMsg{event: event, ok: ok}
	}
}

func waitInstallEventCmd(events chan installer.Event) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-events
		return installEventMsg{event: event, ok: ok}
	}
}

func startBootstrapCmd(ctx context.Context, events chan installer.BootstrapEvent) tea.Cmd {
	return func() tea.Msg {
		go installer.BootstrapPackageManager(ctx, events)
		event, ok := <-events
		return bootstrapEventMsg{event: event, ok: ok}
	}
}

func waitBootstrapEventCmd(events chan installer.BootstrapEvent) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-events
		return bootstrapEventMsg{event: event, ok: ok}
	}
}

func retryPackageManagerCmd() tea.Cmd {
	return func() tea.Msg {
		return retryPackageManagerMsg{
			ready:  installer.HasPackageManager(),
			broken: installer.HasBrokenPackageManagerInstall(),
		}
	}
}

func relaunchElevatedWithArgsCmd(args []string) tea.Cmd {
	return func() tea.Msg {
		return elevationMsg{err: installer.RelaunchElevated(args)}
	}
}

func recoverBrokenChocolateyCmd() tea.Cmd {
	return func() tea.Msg {
		return brokenRecoveryMsg{err: installer.RemoveBrokenPackageManagerInstall()}
	}
}

func installTickCmd() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(time.Time) tea.Msg {
		return installTickMsg{}
	})
}

func searchCursorTickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return searchCursorTickMsg{}
	})
}

func isImportantInstallLine(line string) bool {
	text := strings.ToLower(line)
	needles := []string{
		"download",
		"installing",
		"installed",
		"success",
		"successful",
		"complete",
		"failed",
		"failure",
		"error",
	}
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}
