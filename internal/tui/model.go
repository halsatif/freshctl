package tui

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsatif/freshctl/internal/catalog"
	"github.com/halsatif/freshctl/internal/installer"
)

type screen int

const (
	screenWelcome screen = iota
	screenCatalog
	screenReview
	screenInstall
	screenBootstrap
	screenElevation
	screenBrokenChocolatey
)

type focus int

const (
	focusCategories focus = iota
	focusApps
)

type Model struct {
	screen screen

	width  int
	height int

	categories     []catalog.Category
	categoryCursor int
	appCursor      int
	focus          focus
	selected       map[string]bool
	notice         string

	installEvents chan installer.Event
	skipInstall   chan struct{}
	cancelInstall context.CancelFunc
	installApps   []catalog.App
	installLog    []string
	fullLog       []string
	showFullLog   bool
	results       []installer.Result
	appStatus     map[string]string
	appElapsed    map[string]time.Duration
	currentApp    catalog.App
	currentStep   int
	currentCmd    string
	currentStart  time.Time
	spinnerFrame  int
	installDone   bool

	bootstrapBack    screen
	bootstrapEvents  chan installer.BootstrapEvent
	bootstrapLog     []string
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

func NewModel(args []string) Model {
	initialScreen := screenWelcome
	bootstrapLog := []string(nil)
	selected := selectedFromArgs(args)
	if selected == nil {
		selected = make(map[string]bool)
	}
	if installer.HasBrokenPackageManagerInstall() {
		initialScreen = screenBrokenChocolatey
	} else if !installer.HasPackageManager() {
		if installer.IsElevated() {
			initialScreen = screenBootstrap
			bootstrapLog = []string{"choco was not found. freshctl uses Chocolatey to install apps."}
		} else {
			initialScreen = screenElevation
		}
	} else if len(selected) > 0 {
		initialScreen = screenReview
	}

	return Model{
		screen:        initialScreen,
		categories:    catalog.Default(),
		focus:         focusApps,
		selected:      selected,
		bootstrapBack: screenWelcome,
		bootstrapLog:  bootstrapLog,
		brokenBack:    screenWelcome,
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
	}

	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case screenWelcome:
		return m.viewWelcome()
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

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
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
		if msg.String() == "enter" {
			m.screen = screenCatalog
		}
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

func (m Model) handleCatalogKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.notice = ""

	switch msg.String() {
	case "tab":
		if m.focus == focusCategories {
			m.focus = focusApps
		} else {
			m.focus = focusCategories
		}
	case "up", "k":
		if m.focus == focusCategories && m.categoryCursor > 0 {
			m.categoryCursor--
			m.appCursor = 0
		} else if m.focus == focusApps && m.appCursor > 0 {
			m.appCursor--
		}
	case "down", "j":
		if m.focus == focusCategories && m.categoryCursor < len(m.categories)-1 {
			m.categoryCursor++
			m.appCursor = 0
		} else if m.focus == focusApps && m.appCursor < len(m.currentApps())-1 {
			m.appCursor++
		}
	case " ":
		apps := m.currentApps()
		if len(apps) == 0 {
			return m, nil
		}
		if m.focus == focusCategories {
			m.toggleCurrentCategory()
			return m, nil
		}
		app := apps[m.appCursor]
		m.selected[app.ID] = !m.selected[app.ID]
	case "enter":
		m.screen = screenReview
	}

	return m, nil
}

func (m Model) handleReviewKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "b", "esc":
		m.notice = ""
		m.screen = screenCatalog
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
			m.bootstrapLog = []string{"choco was not found. freshctl uses Chocolatey to install apps."}
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

func (m Model) handleBootstrapKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "b", "esc":
		if m.cancelBootstrap != nil {
			m.cancelBootstrap()
		}
		m.bootstrapRunning = false
		m.cancelBootstrap = nil
		m.screen = m.bootstrapBack
		m.notice = "Chocolatey is still missing. Bootstrap Chocolatey or retry detection before installing apps."
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
		m.bootstrapLog = append(m.bootstrapLog, "Running official Chocolatey bootstrap command...")
		ctx, cancel := context.WithCancel(context.Background())
		m.cancelBootstrap = cancel
		m.bootstrapEvents = make(chan installer.BootstrapEvent)
		return m, startBootstrapCmd(ctx, m.bootstrapEvents)
	case "r":
		if m.bootstrapRunning {
			return m, nil
		}
		m.bootstrapLog = append(m.bootstrapLog, "Retrying Chocolatey detection...")
		return m, retryPackageManagerCmd()
	}

	return m, nil
}

func (m Model) handleBrokenChocolateyKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
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
	switch msg.String() {
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
		m.appStatus[event.App.ID] = "installing"
		m.installLog = append(m.installLog, "installing "+event.App.Name)
		m.fullLog = append(m.fullLog, "> "+event.Line)
	case installer.EventAppFinished:
		if !m.currentStart.IsZero() && m.currentApp.ID == event.App.ID {
			m.appElapsed[event.App.ID] = time.Since(m.currentStart)
		}
		if event.Success {
			m.appStatus[event.App.ID] = "installed"
			m.installLog = append(m.installLog, "success "+event.App.Name)
			m.fullLog = append(m.fullLog, "ok: "+event.App.Name)
		} else if event.Err == installer.ErrInstallSkipped {
			m.appStatus[event.App.ID] = "skipped"
			m.installLog = append(m.installLog, "skipped "+event.App.Name)
			m.fullLog = append(m.fullLog, "skipped: "+event.App.Name)
		} else if event.Err == context.DeadlineExceeded {
			m.appStatus[event.App.ID] = "failed"
			m.installLog = append(m.installLog, "timed out "+event.App.Name)
			m.fullLog = append(m.fullLog, "timed out: "+event.App.Name)
		} else {
			m.appStatus[event.App.ID] = "failed"
			m.installLog = append(m.installLog, "failed "+event.App.Name+" - "+event.Err.Error())
			m.fullLog = append(m.fullLog, "failed: "+event.App.Name+" - "+event.Err.Error())
		}
	case installer.EventSummary:
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
			m.screen = m.bootstrapBack
			m.notice = "Chocolatey is available now. Press enter to continue."
			return m, nil
		}
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
		m.screen = m.bootstrapBack
		m.notice = "Chocolatey is available now. Press enter to continue."
		return m, nil
	}

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
	switch msg.String() {
	case "l":
		m.showFullLog = !m.showFullLog
		return m, tea.ClearScreen
	case "s":
		if !m.installDone && m.skipInstall != nil && m.currentApp.ID != "" {
			m.installLog = append(m.installLog, "skipping "+m.currentApp.Name+"...")
			m.fullLog = append(m.fullLog, "skip requested for "+m.currentApp.Name)
			m.appStatus[m.currentApp.ID] = "skipping"
			select {
			case m.skipInstall <- struct{}{}:
			default:
			}
		}
	}
	return m, nil
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

func (m Model) startInstall(apps []catalog.App) (tea.Model, tea.Cmd) {
	m.screen = screenInstall
	m.installApps = apps
	m.installLog = nil
	m.fullLog = nil
	m.showFullLog = false
	m.results = nil
	m.appStatus = make(map[string]string, len(apps))
	m.appElapsed = make(map[string]time.Duration, len(apps))
	for _, app := range apps {
		m.appStatus[app.ID] = "pending"
	}
	m.currentApp = catalog.App{}
	m.currentStep = 0
	m.currentCmd = ""
	m.currentStart = time.Time{}
	m.spinnerFrame = 0
	m.installDone = false
	m.installEvents = make(chan installer.Event)
	m.skipInstall = make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelInstall = cancel
	return m, tea.Batch(tea.ClearScreen, startInstallCmd(ctx, apps, m.installEvents, m.skipInstall), installTickCmd())
}

func (m Model) currentApps() []catalog.App {
	if len(m.categories) == 0 {
		return nil
	}
	return m.categories[m.categoryCursor].Apps
}

func (m Model) toggleCurrentCategory() {
	apps := m.currentApps()
	if len(apps) == 0 {
		return
	}

	allSelected := true
	for _, app := range apps {
		if !m.selected[app.ID] {
			allSelected = false
			break
		}
	}

	for _, app := range apps {
		m.selected[app.ID] = !allSelected
	}
}

func (m Model) selectedApps() []catalog.App {
	apps := make([]catalog.App, 0)
	for _, category := range m.categories {
		for _, app := range category.Apps {
			if m.selected[app.ID] {
				apps = append(apps, app)
			}
		}
	}
	return apps
}

func (m Model) selectedArgs() []string {
	selected := m.selectedApps()
	if len(selected) == 0 {
		return nil
	}

	ids := make([]string, 0, len(selected))
	for _, app := range selected {
		ids = append(ids, app.ID)
	}
	return []string{"--selected=" + strings.Join(ids, ",")}
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

func (m Model) installIndex(app catalog.App) int {
	for i, candidate := range m.installApps {
		if candidate.ID == app.ID {
			return i
		}
	}
	return 0
}

func startInstallCmd(ctx context.Context, apps []catalog.App, events chan installer.Event, skips <-chan struct{}) tea.Cmd {
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
