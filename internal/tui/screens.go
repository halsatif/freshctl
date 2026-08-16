package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/halsatif/freshctl/internal/catalog"
	presetpkg "github.com/halsatif/freshctl/internal/presets"
)

func (m Model) viewWelcome() string {
	body := strings.Join([]string{
		titleStyle.Render("freshctl"),
		subtitleStyle.Render("fresh windows setup, but not painful"),
		"",
		"Choose apps from a small catalog and install them with Chocolatey.",
		"Nothing runs until you confirm the review screen.",
		"",
		m.footerFor(screenWelcome),
	}, "\n")

	return place(body, m.width, m.height)
}

func (m Model) viewModeSelect() string {
	options := []string{"Full catalog with search", "Categories"}
	lines := []string{
		titleStyle.Render("choose catalog mode"),
		"",
	}
	for i, option := range options {
		line := "  " + option
		if i == m.modeCursor {
			line = activeItemStyle.Render("> " + option)
		}
		lines = append(lines, line)
	}

	lines = append(lines, "")
	if m.modeCursor == 0 {
		lines = append(lines,
			"Full catalog with search:",
			mutedStyle.Render("Browse all apps in one flat list. Best when you already know what you need."),
		)
	} else {
		lines = append(lines,
			"Categories:",
			mutedStyle.Render("Browse apps grouped by purpose. Best for discovering tools."),
		)
	}

	lines = append(lines, "", m.footerFor(screenModeSelect))
	return place(strings.Join(lines, "\n"), m.width, m.height)
}

func (m Model) viewCatalog() string {
	contentWidth := pageWidth(m.width)
	panelHeight := m.catalogPanelHeight()
	leftWidth, rightWidth := catalogPaneWidths(contentWidth)
	itemLines := m.catalogListLines(leftWidth)
	if len(itemLines) == 0 {
		itemLines = append(itemLines, "  "+mutedStyle.Render("No packages found."), "  "+mutedStyle.Render("Try a different search term."))
	}

	itemLines = m.visibleCatalogLines(itemLines, panelHeight)
	itemLines = fitCatalogListLines(itemLines, leftWidth)
	itemLines = padLines(itemLines, panelHeight)
	left := borderStyle.Width(leftWidth).Height(panelHeight).Render(strings.Join(itemLines, "\n"))
	right := borderStyle.Width(rightWidth).Height(panelHeight).Render(m.catalogDetailsPanel(rightWidth, panelHeight))
	content := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	if contentWidth < 72 {
		content = strings.Join([]string{left, "", right}, "\n")
	}

	parts := []string{
		titleStyle.Render("freshctl"),
		mutedStyle.Render(fmt.Sprintf("%d selected", len(m.selectedApps()))),
		mutedStyle.Render(m.catalogHeaderLine()),
	}
	if source := m.selectionSourceLine(); source != "" {
		parts = append(parts, mutedStyle.Render(source))
	}
	if m.searchActive() {
		parts = append(parts,
			mutedStyle.Render("Search: "+m.searchInputText()),
			mutedStyle.Render(fmt.Sprintf("Results: %d packages", len(m.filteredFullCatalogItems()))),
		)
	}
	parts = append(parts,
		"",
		content,
	)
	if m.notice != "" {
		parts = append(parts, "", errorStyle.Render(m.notice))
	}
	parts = append(parts, "", m.footerFor(screenCatalog))

	return place(strings.Join(parts, "\n"), m.width, m.height)
}

func (m Model) viewPresetPicker() string {
	contentWidth := pageWidth(m.width)
	panelHeight := m.presetPanelHeight()
	leftWidth, rightWidth := catalogPaneWidths(contentWidth)

	listLines := m.presetListLines(leftWidth)
	listLines = padLines(fitPresetLines(listLines, leftWidth), panelHeight)
	left := borderStyle.Width(leftWidth).Height(panelHeight).Render(strings.Join(listLines, "\n"))
	right := borderStyle.Width(rightWidth).Height(panelHeight).Render(m.presetPreviewPanel(rightWidth, panelHeight))
	content := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	if contentWidth < 72 {
		content = strings.Join([]string{left, "", right}, "\n")
	}

	lines := []string{
		titleStyle.Render("Presets"),
		"",
		content,
		"",
		m.footerFor(screenPresetPicker),
	}
	return place(strings.Join(lines, "\n"), m.width, m.height)
}

func (m Model) presetListLines(width int) []string {
	presets := m.availablePresets()
	if len(presets) == 0 {
		return []string{"  " + mutedStyle.Render("No presets available.")}
	}

	lines := make([]string, 0, len(presets))
	for i, preset := range presets {
		line := preset.Name
		if i == m.presetCursor {
			line = activeItemStyle.Render("> " + line)
		} else {
			line = "  " + line
		}
		lines = append(lines, line)
	}
	return lines
}

func (m Model) presetPreviewPanel(width, height int) string {
	presets := m.availablePresets()
	if len(presets) == 0 {
		return fitDetailsLines([]string{"No preset selected."}, width, height)
	}
	cursor := m.presetCursor
	if cursor < 0 || cursor >= len(presets) {
		cursor = 0
	}
	return fitDetailsLines(m.presetPreviewLines(presets[cursor], width, height), width, height)
}

func (m Model) presetPreviewLines(preset presetpkg.Preset, width, height int) []string {
	wrapWidth := width - 4
	if wrapWidth < 18 {
		wrapWidth = 18
	}
	lines := []string{
		preset.Name,
		"",
	}
	lines = append(lines, wrapText(preset.Description, wrapWidth)...)
	lines = append(lines, "", "Includes:")

	names := m.presetPackageNames(preset)
	availableRows := height - len(lines) - 2
	if availableRows < 1 {
		availableRows = 1
	}
	shown := 0
	for _, name := range names {
		if shown >= availableRows {
			break
		}
		lines = append(lines, "- "+name)
		shown++
	}
	if hidden := len(names) - shown; hidden > 0 {
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("...and %d more", hidden)))
	}
	lines = append(lines, "", mutedStyle.Render(fmt.Sprintf("%d packages", len(preset.Packages))))
	return lines
}

func (m Model) presetPackageNames(preset presetpkg.Preset) []string {
	byID := make(map[string]string)
	for _, app := range collectModelPackages(m.categories) {
		byID[app.ID] = app.Name
	}

	names := make([]string, 0, len(preset.Packages))
	for _, packageID := range preset.Packages {
		if name, ok := byID[packageID]; ok {
			names = append(names, name)
		}
	}
	return names
}

func (m Model) presetPanelHeight() int {
	if m.height <= 0 {
		return 14
	}
	height := m.height - 12
	if height < 8 {
		return 8
	}
	if height > 18 {
		return 18
	}
	return height
}

func fitPresetLines(lines []string, width int) []string {
	innerWidth := width - 4
	if innerWidth < 12 {
		innerWidth = 12
	}
	fitted := make([]string, len(lines))
	for i, line := range lines {
		fitted[i] = fitLine(line, innerWidth)
	}
	return fitted
}

func (m Model) catalogListLines(width int) []string {
	if m.catalogMode == catalogModeFull || m.searchActive() {
		return m.fullCatalogListLines(width)
	}
	return m.categoryCatalogListLines(width)
}

func (m Model) searchInputText() string {
	if !m.searchFocused {
		return m.searchQuery
	}
	cursor := " "
	if m.searchCursor {
		cursor = "|"
	}
	return m.searchQuery + cursor
}

func (m Model) categoryCatalogListLines(width int) []string {
	categories := m.currentCategories()
	apps := m.currentApps()
	itemLines := make([]string, 0, len(categories)+len(apps))

	for i, category := range categories {
		line := category.Name + " >"
		if count := m.selectedInCategory(category); count > 0 {
			line = fmt.Sprintf("%s (%d)", line, count)
		}
		if i == m.catalogCursor {
			line = activeItemStyle.Render("> " + line)
		} else {
			line = "  " + line
		}
		itemLines = append(itemLines, line)
	}

	for i, app := range apps {
		box := "[ ]"
		if m.selected[app.ID] {
			box = selectedStyle.Render("[x]")
		}
		line := m.packageListLine(box, app, packageListContentWidth(width))
		if len(categories)+i == m.catalogCursor {
			line = activeItemStyle.Render("> " + line)
		} else {
			line = "  " + line
		}
		itemLines = append(itemLines, line)
	}

	return itemLines
}

func (m Model) fullCatalogListLines(width int) []string {
	items := m.filteredFullCatalogItems()
	lines := make([]string, 0, len(items))
	for i, item := range items {
		box := "[ ]"
		if m.selected[item.Package.ID] {
			box = selectedStyle.Render("[x]")
		}
		line := m.packageListLine(box, item.Package, packageListContentWidth(width))
		if i == m.catalogCursor {
			line = activeItemStyle.Render("> " + line)
		} else {
			line = "  " + line
		}
		lines = append(lines, line)
	}
	return lines
}

func packageListContentWidth(width int) int {
	contentWidth := width - 6
	if contentWidth < 12 {
		return 12
	}
	return contentWidth
}

func (m Model) packageListLine(box string, app catalog.Application, width int) string {
	status := m.packageListStatusLabel(app)
	if status == "" {
		nameWidth := width - ansi.StringWidth(box) - 1
		return fmt.Sprintf("%s %s", box, fitLine(app.Name, nameWidth))
	}

	statusText := mutedStyle.Render(status)
	nameWidth := width - ansi.StringWidth(box) - ansi.StringWidth(status) - 2
	if nameWidth < 4 {
		nameWidth = 4
	}
	return fmt.Sprintf("%s %s %s", box, fitLine(app.Name, nameWidth), statusText)
}

func (m Model) packageListStatusLabel(app catalog.Application) string {
	status, ok := m.installedStatus(app)
	if !ok || !status.Checked {
		return ""
	}
	if status.Installed {
		return "OK"
	}
	return "--"
}

func (m Model) visibleCatalogLines(lines []string, height int) []string {
	if height <= 0 || len(lines) <= height {
		return lines
	}
	start := m.catalogScroll
	if start < 0 {
		start = 0
	}
	if start > len(lines)-height {
		start = len(lines) - height
	}
	end := start + height
	return lines[start:end]
}

func fitCatalogListLines(lines []string, width int) []string {
	innerWidth := width - 4
	if innerWidth < 12 {
		innerWidth = 12
	}
	fitted := make([]string, len(lines))
	for i, line := range lines {
		fitted[i] = fitLine(line, innerWidth)
	}
	return fitted
}

func (m Model) catalogHeaderLine() string {
	if m.catalogMode == catalogModeFull {
		return "Mode: Full catalog"
	}
	return "Path: " + m.currentBreadcrumb()
}

func (m Model) selectionSourceLine() string {
	if m.appliedProfile != "" {
		return "Profile: " + m.appliedProfile
	}
	if m.appliedPreset != "" {
		return "Preset: " + m.appliedPreset
	}
	return ""
}

func (m Model) viewReview() string {
	contentWidth := pageWidth(m.width)
	selected := m.selectedApps()
	visibleRows := m.reviewVisibleRows()
	start := m.reviewScroll
	if start > len(selected) {
		start = len(selected)
	}
	end := start + visibleRows
	if end > len(selected) {
		end = len(selected)
	}

	lines := []string{
		titleStyle.Render("review"),
		mutedStyle.Render(fmt.Sprintf("Packages selected: %d", len(selected))),
		mutedStyle.Render("Backend: Chocolatey"),
	}
	if source := m.selectionSourceLine(); source != "" {
		lines = append(lines, mutedStyle.Render(source))
	}

	if len(selected) == 0 {
		lines = append(lines, "", mutedStyle.Render("No apps selected yet. Press esc to return to the catalog."))
	} else {
		rangeLine := fmt.Sprintf("Showing %d-%d of %d", start+1, end, len(selected))
		if len(selected) > visibleRows {
			rangeLine += " (use ↑/↓ or PageUp/PageDown to scroll)"
		}
		lines = append(lines, "", "Selected apps:", mutedStyle.Render(rangeLine))
		for i := start; i < end; i++ {
			app := selected[i]
			line := fmt.Sprintf("  %3d. %-34s %s", i+1, app.Name, mutedStyle.Render(app.ID))
			lines = append(lines, fitLine(line, contentWidth))
		}
		lines = append(lines, "", mutedStyle.Render("Commands will run one by one after confirmation."))
	}

	if m.notice != "" {
		lines = append(lines, "", errorStyle.Render(m.notice))
	}

	lines = append(lines, "", m.footerFor(screenReview))
	return place(strings.Join(lines, "\n"), m.width, m.height)
}

func (m Model) viewInstall() string {
	contentWidth := pageWidth(m.width)
	total := len(m.installApps)
	currentName := "preparing"
	if m.currentApp.Name != "" {
		currentName = m.currentApp.Name
	}
	elapsed := ""
	if !m.currentStart.IsZero() {
		currentElapsed := formatElapsed(time.Since(m.currentStart))
		if m.installDone {
			if frozen := m.elapsedForApp(m.currentApp); frozen != "--:--" {
				currentElapsed = frozen
			}
		}
		elapsed = " " + mutedStyle.Render(currentElapsed)
	}

	progress := fmt.Sprintf("[%d/%d]", m.currentStep, total)
	if total == 0 {
		progress = "[0/0]"
	}

	spin := " "
	if !m.installDone {
		spin = spinnerFrame(m.spinnerFrame)
	}

	status := m.currentStatus
	if status == "" {
		status = "Preparing installation..."
	}

	lines := []string{
		titleStyle.Render("install"),
	}
	if source := m.selectionSourceLine(); source != "" {
		lines = append(lines, mutedStyle.Render(source))
	}
	lines = append(lines,
		mutedStyle.Render(m.installPlanLine()),
		fitLine(fmt.Sprintf("%s current: %s%s %s", spin, currentName, elapsed, mutedStyle.Render(progress)), contentWidth),
		mutedStyle.Render(fitLine("Status: "+status, contentWidth)),
	)

	if m.installDone {
		lines = append(lines, "", m.installDoneMessage())
	}

	lines = append(lines, "", "Summary:")
	visibleRows := m.installSummaryVisibleRows(len(lines))
	if len(m.installApps) > visibleRows {
		start, end := m.installSummaryRange(visibleRows)
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("Showing %d-%d of %d", start+1, end, len(m.installApps))))
	}
	lines = append(lines, m.installSummaryTable(contentWidth, visibleRows)...)
	if m.installDone {
		lines = append(lines, m.installFailureLines(contentWidth)...)
	}
	lines = append(lines, "", m.footerFor(screenInstall))
	return place(strings.Join(lines, "\n"), m.width, m.height)
}

func (m Model) viewInstallLogs() string {
	contentWidth := pageWidth(m.width)
	appName := "waiting for installation"
	providerName := "-"
	if m.currentApp.Name != "" {
		appName = m.currentApp.Name
		if provider, ok := m.currentApp.PrimaryProvider(); ok {
			providerName = string(provider.Type)
		}
	}

	lines := []string{
		titleStyle.Render("logs"),
		fitLine("application: "+appName, contentWidth),
		fitLine("provider: "+providerName, contentWidth),
		"",
	}

	visibleRows := m.installLogViewportHeight()
	start, end := m.installLogRange()
	for index := start; index < end; index++ {
		lines = append(lines, formatInstallLogEntry(m.installLogs[index], contentWidth))
	}
	if start == end {
		lines = append(lines, mutedStyle.Render("Waiting for output..."))
	}
	for len(lines) < 4+visibleRows {
		lines = append(lines, "")
	}

	lines = append(lines, "", m.footerFor(screenInstallLogs))
	return place(strings.Join(lines, "\n"), m.width, m.height)
}

func (m Model) viewHelp() string {
	contentWidth := pageWidth(m.width)
	hints := m.helpHintsFor(m.helpBack)
	visibleRows := len(hints)
	showHidden := false
	if m.height > 0 {
		availableRows := m.height - 5
		if availableRows < 0 {
			availableRows = 0
		}
		if visibleRows > availableRows {
			visibleRows = availableRows
			if availableRows >= 2 {
				visibleRows--
				showHidden = true
			}
		}
	}

	keyWidth := 12
	for _, item := range hints[:visibleRows] {
		if width := ansi.StringWidth(helpKeyLabel(item)); width > keyWidth {
			keyWidth = width
		}
	}
	if limit := maxInt(8, contentWidth/3); keyWidth > limit {
		keyWidth = limit
	}

	lines := []string{
		titleStyle.Render("help"),
		mutedStyle.Render("shortcuts for " + screenName(m.helpBack)),
		"",
	}
	for _, item := range hints[:visibleRows] {
		key := fitLine(helpKeyLabel(item), keyWidth)
		key = footerKeyStyle.Width(keyWidth).Render(key)
		lines = append(lines, fitLine("  "+key+"  "+footerDescriptionStyle.Render(item.Description), contentWidth))
	}
	if showHidden {
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("  %d more shortcuts hidden at this height", len(hints)-visibleRows)))
	}
	lines = append(lines, "", m.footer(
		hint("esc", "back", footerPriorityPrimary),
		hint("?", "close", footerPrioritySecondary),
	))
	return place(strings.Join(lines, "\n"), m.width, m.height)
}

func (m Model) viewBootstrap() string {
	contentWidth := pageWidth(m.width)
	status := m.bootstrapStatus
	if status == "" {
		status = "Chocolatey was not found on this system."
	}

	lines := []string{
		titleStyle.Render("chocolatey bootstrap"),
		"",
		"Chocolatey was not found on this system.",
		"freshctl uses Chocolatey to install apps.",
		"Press enter to run Chocolatey's official PowerShell bootstrap command.",
		"Administrator privileges may be required.",
		"",
		mutedStyle.Render("Source: https://community.chocolatey.org/install.ps1"),
		"",
	}

	if m.bootstrapRunning {
		lines = append(lines, selectedStyle.Render(status))
	} else {
		lines = append(lines, mutedStyle.Render(status))
	}

	if !m.showBootstrapLog {
		lines = append(lines, "", mutedStyle.Render("Logs hidden. Press l to show full logs."))
	} else {
		logLines := tailLines(m.bootstrapLog, bootstrapLogLimit(m.height))
		lines = append(lines, "", mutedStyle.Render("full logs"))
		if len(logLines) == 0 {
			lines = append(lines, "  "+mutedStyle.Render("Waiting for output..."))
		}
		for _, line := range logLines {
			lines = append(lines, fitLine("  "+sanitizeLogLine(line), contentWidth))
		}
	}

	lines = append(lines, "", m.footerFor(screenBootstrap))
	return place(strings.Join(lines, "\n"), m.width, m.height)
}

func (m Model) viewElevation() string {
	description := "freshctl needs administrator privileges to install selected applications."
	if applicationsUseProvider(m.selectedApps(), catalog.ProviderChocolatey) {
		description = "freshctl needs administrator privileges to install Chocolatey and selected applications."
	}
	lines := []string{
		titleStyle.Render("administrator privileges required"),
		"",
		description,
		"",
	}

	if m.elevationRunning {
		lines = append(lines, selectedStyle.Render("Relaunching as administrator..."))
	} else {
		lines = append(lines, mutedStyle.Render("Press enter to relaunch freshctl as administrator."))
	}

	if m.elevationError != "" {
		lines = append(lines, "", errorStyle.Render(m.elevationError))
	}

	lines = append(lines, "", m.footerFor(screenElevation))
	return place(strings.Join(lines, "\n"), m.width, m.height)
}

func (m Model) viewBrokenChocolatey() string {
	lines := []string{
		titleStyle.Render("Broken Chocolatey installation detected."),
		"",
		"C:\\ProgramData\\chocolatey exists, but C:\\ProgramData\\chocolatey\\bin\\choco.exe is missing.",
		"freshctl will not rerun bootstrap while this partial install exists.",
		"",
	}

	if m.brokenRunning {
		lines = append(lines, selectedStyle.Render("Removing broken folder..."))
	} else {
		lines = append(lines, mutedStyle.Render("Press enter to remove the broken folder and reinstall Chocolatey."))
	}

	if m.brokenError != "" {
		lines = append(lines, "", errorStyle.Render(m.brokenError))
	}

	lines = append(lines, "", m.footerFor(screenBrokenChocolatey))
	return place(strings.Join(lines, "\n"), m.width, m.height)
}

func (m Model) summaryLines() []string {
	if len(m.results) == 0 {
		return []string{mutedStyle.Render("No installs were run.")}
	}

	okCount := 0
	failCount := 0
	lines := []string{"Summary:"}
	for _, result := range m.results {
		if result.Success {
			okCount++
			lines = append(lines, successStyle.Render("  ok     ")+result.App.Name)
		} else {
			failCount++
			errText := "unknown error"
			if result.Err != nil {
				errText = result.Err.Error()
			}
			lines = append(lines, errorStyle.Render("  failed ")+result.App.Name+" - "+errText)
		}
	}
	lines = append(lines, mutedStyle.Render(fmt.Sprintf("%d succeeded, %d failed", okCount, failCount)))
	return lines
}

func (m Model) installSummaryTable(width, visibleRows int) []string {
	if len(m.installApps) == 0 {
		return []string{"  " + mutedStyle.Render("No apps queued.")}
	}

	start, end := m.installSummaryRange(visibleRows)
	lines := make([]string, 0, end-start)
	nameWidth := m.installNameWidth(width)
	for _, app := range m.installApps[start:end] {
		status := m.appStatus[app.ID]
		if status == "" {
			status = "pending"
		}
		info := installStatusInfo(status)
		elapsed := m.elapsedForApp(app)
		line := fmt.Sprintf("  %s %-11s %-*s %s", info.RenderedCode(), info.Label, nameWidth, app.Name, elapsed)
		lines = append(lines, fitLine(line, width))
	}
	return lines
}

func (m Model) installPlanLine() string {
	alreadyInstalled := len(m.initialSkippedResults)
	remaining := 0
	for _, app := range m.installApps {
		switch m.appStatus[app.ID] {
		case "", "pending", "installing", "skipping":
			remaining++
		}
	}
	return fmt.Sprintf("Selected: %d  Already installed: %d  Remaining: %d", len(m.installApps), alreadyInstalled, remaining)
}

func (m Model) installSummaryRange(visibleRows int) (int, int) {
	if visibleRows <= 0 || visibleRows > len(m.installApps) {
		visibleRows = len(m.installApps)
	}
	start := m.installScroll
	if start < 0 {
		start = 0
	}
	if start > len(m.installApps)-visibleRows {
		start = len(m.installApps) - visibleRows
	}
	if start < 0 {
		start = 0
	}
	end := start + visibleRows
	if end > len(m.installApps) {
		end = len(m.installApps)
	}
	return start, end
}

func (m Model) installSummaryVisibleRows(usedLines int) int {
	if m.height <= 0 {
		return 12
	}
	rows := m.height - usedLines - 4
	if rows < 4 {
		return 4
	}
	if rows > 14 {
		return 14
	}
	return rows
}

func (m Model) installFailureLines(width int) []string {
	if len(m.results) == 0 {
		return nil
	}

	lines := make([]string, 0)
	shown := 0
	hidden := 0
	for _, result := range m.results {
		if result.Success || result.Skipped || result.Err == nil {
			continue
		}
		if len(lines) == 0 {
			lines = append(lines, "", "Failures:")
		}
		if shown >= 3 {
			hidden++
			continue
		}
		lines = append(lines, fitLine(errorStyle.Render("  failed ")+result.App.Name+" ("+result.App.ID+") - "+result.Err.Error(), width))
		shown++
	}
	if hidden > 0 {
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("  %d more failures. Press l to view logs.", hidden)))
	}
	return lines
}

func (m Model) installNameWidth(width int) int {
	maxName := 12
	for _, app := range m.installApps {
		if nameWidth := ansi.StringWidth(app.Name); nameWidth > maxName {
			maxName = nameWidth
		}
	}

	limit := width - 28
	if limit < 12 {
		return 12
	}
	if maxName > limit {
		return limit
	}
	if maxName > 30 {
		return 30
	}
	return maxName
}

func (m Model) catalogDetailsPanel(width, height int) string {
	if m.catalogMode == catalogModeFull || m.searchActive() {
		items := m.filteredFullCatalogItems()
		if m.catalogCursor < 0 || m.catalogCursor >= len(items) {
			return fitDetailsLines([]string{"No item selected."}, width, height)
		}
		item := items[m.catalogCursor]
		selected := "No"
		if m.selected[item.Package.ID] {
			selected = "Yes"
		}
		return fitDetailsLines(packageDetailsLines(item.Package, selected, m.installedStatusLabel(item.Package)), width, height)
	}

	categories := m.currentCategories()
	if m.catalogCursor < len(categories) {
		category := categories[m.catalogCursor]
		return fitDetailsLines(categoryDetailsLines(category), width, height)
	}

	appIndex := m.catalogCursor - len(categories)
	apps := m.currentApps()
	if appIndex < 0 || appIndex >= len(apps) {
		return fitDetailsLines([]string{"No item selected."}, width, height)
	}

	app := apps[appIndex]
	selected := "No"
	if m.selected[app.ID] {
		selected = "Yes"
	}
	return fitDetailsLines(packageDetailsLines(app, selected, m.installedStatusLabel(app)), width, height)
}

func fitDetailsLines(lines []string, width, height int) string {
	innerWidth := width - 2
	if innerWidth < 12 {
		innerWidth = 12
	}
	for i, line := range lines {
		lines[i] = fitLine(line, innerWidth)
	}
	lines = padLines(lines, height)
	return strings.Join(lines, "\n")
}

func padLines(lines []string, height int) []string {
	if height <= 0 {
		return lines
	}
	if len(lines) > height {
		return lines[:height]
	}
	padded := append([]string{}, lines...)
	for len(padded) < height {
		padded = append(padded, "")
	}
	return padded
}

func catalogPaneWidths(contentWidth int) (int, int) {
	if contentWidth < 72 {
		width := contentWidth - 4
		if width < 36 {
			width = 36
		}
		return width, width
	}

	left := contentWidth / 2
	if left > 42 {
		left = 42
	}
	right := contentWidth - left - 6
	if right < 30 {
		right = 30
	}
	return left, right
}

func categoryDetailsLines(category catalog.Category) []string {
	lines := []string{
		category.Name,
		"",
	}
	lines = append(lines, wrapText(category.Description, 28)...)
	lines = append(lines, "", "Contains:")
	items := categoryContents(category, 6)
	if len(items) == 0 {
		lines = append(lines, "- No packages yet")
	} else {
		for _, item := range items {
			lines = append(lines, "- "+item)
		}
	}
	return lines
}

func packageDetailsLines(app catalog.Application, selected string, installed string) []string {
	manager := "Unavailable"
	if provider, ok := app.PrimaryProvider(); ok {
		manager = string(provider.Type)
	}
	verified := "No"
	if app.Verified {
		verified = "Yes"
	}
	lines := []string{
		"Package:",
		app.Name,
		"",
		"ID:",
		app.ID,
		"",
		"Type: " + packageTypeLabel(app.Type),
		"Manager: " + manager,
		"Verified: " + verified,
	}
	if installed != "" {
		lines = append(lines, "Installed: "+installed)
	}
	lines = append(lines, "", "Description:")
	lines = append(lines, wrapText(app.Description, 28)...)
	lines = append(lines, "", "Selected: "+selected)
	return lines
}

func (m Model) installedStatusLabel(app catalog.Application) string {
	status, ok := m.installedStatus(app)
	if !ok {
		return ""
	}
	if status.Checked && status.Installed {
		return "Yes"
	}
	return "No"
}

func packageTypeLabel(packageType catalog.PackageType) string {
	switch packageType {
	case catalog.PackageTypeCLITool:
		return "CLI Tool"
	case catalog.PackageTypeRuntime:
		return "Runtime"
	case catalog.PackageTypeApplication:
		return "Application"
	default:
		return "Application"
	}
}

func categoryContents(category catalog.Category, limit int) []string {
	items := make([]string, 0, limit)
	for _, child := range category.Categories {
		items = append(items, child.Name)
		if len(items) >= limit {
			return items
		}
	}
	for _, app := range category.Apps {
		items = append(items, app.Name)
		if len(items) >= limit {
			return items
		}
	}
	return items
}

func wrapText(text string, width int) []string {
	if text == "" {
		return []string{"No description."}
	}
	if width < 12 {
		width = 12
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{"No description."}
	}

	lines := []string{}
	current := ""
	for _, word := range words {
		if current == "" {
			current = word
			continue
		}
		if ansi.StringWidth(current)+1+ansi.StringWidth(word) > width {
			lines = append(lines, current)
			current = word
			continue
		}
		current += " " + word
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func (m Model) elapsedForApp(app catalog.Application) string {
	if elapsed, ok := m.appElapsed[app.ID]; ok {
		return formatElapsed(elapsed)
	}
	if m.currentApp.ID == app.ID && !m.currentStart.IsZero() && !m.installDone {
		return formatElapsed(time.Since(m.currentStart))
	}
	return "--:--"
}

func (m Model) installDoneMessage() string {
	failed := 0
	skipped := 0
	installed := 0
	for _, app := range m.installApps {
		switch m.appStatus[app.ID] {
		case "installed":
			installed++
		case "failed":
			failed++
		case "skipped":
			skipped++
		}
	}

	if failed == 0 && skipped == 0 {
		return successStyle.Render("All selected apps were installed.")
	}
	return mutedStyle.Render(fmt.Sprintf("Install finished: %d installed, %d failed, %d skipped.", installed, failed, skipped))
}

func spinnerFrame(frame int) string {
	frames := []string{"|", "/", "-", "\\"}
	return frames[frame%len(frames)]
}

type installStatus struct {
	Code  string
	Label string
	Style lipgloss.Style
}

func installStatusInfo(status string) installStatus {
	switch status {
	case "installed":
		return installStatus{Code: "OK", Label: "installed", Style: successStyle}
	case "failed":
		return installStatus{Code: "FAIL", Label: "failed", Style: errorStyle}
	case "installing":
		return installStatus{Code: "RUN", Label: "installing", Style: selectedStyle}
	case "skipping":
		return installStatus{Code: "SKIP", Label: "skipping", Style: selectedStyle}
	case "skipped":
		return installStatus{Code: "SKIP", Label: "skipped", Style: mutedStyle}
	default:
		return installStatus{Code: "WAIT", Label: "pending", Style: mutedStyle}
	}
}

func (s installStatus) RenderedCode() string {
	return s.Style.Render(fmt.Sprintf("%-4s", s.Code))
}

func (m Model) installLogViewportHeight() int {
	if m.height <= 0 {
		return 12
	}
	height := m.height - 8
	if height < 1 {
		return 1
	}
	return height
}

func (m Model) installLogRange() (int, int) {
	visibleRows := m.installLogViewportHeight()
	start := m.logScroll
	maxScroll := maxInt(0, len(m.installLogs)-visibleRows)
	if start < 0 {
		start = 0
	}
	if start > maxScroll {
		start = maxScroll
	}
	end := start + visibleRows
	if end > len(m.installLogs) {
		end = len(m.installLogs)
	}
	return start, end
}

func formatInstallLogEntry(entry installLogEntry, width int) string {
	line := sanitizeLogLine(entry.Line)
	if entry.Application != "" {
		line = "[" + entry.Application + "] " + line
	}
	return fitLine(line, width)
}

func bootstrapLogLimit(height int) int {
	if height <= 0 {
		return 10
	}
	limit := height - 17
	if limit < 4 {
		return 4
	}
	if limit > 12 {
		return 12
	}
	return limit
}

func tailLines(lines []string, limit int) []string {
	if limit <= 0 || len(lines) <= limit {
		return lines
	}
	return lines[len(lines)-limit:]
}

func formatElapsed(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}
	total := int(duration.Round(time.Second).Seconds())
	minutes := total / 60
	seconds := total % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func sanitizeLogLine(line string) string {
	line = ansi.Strip(line)
	line = strings.Map(func(r rune) rune {
		if r == '\t' {
			return ' '
		}
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, line)
	line = strings.TrimRight(line, "\r\n")
	return line
}

func (m Model) selectedInCategory(category catalog.Category) int {
	count := 0
	for _, child := range category.Categories {
		count += m.selectedInCategory(child)
	}
	for _, app := range category.Apps {
		if m.selected[app.ID] {
			count++
		}
	}
	return count
}

func (m Model) catalogPanelHeight() int {
	height := m.catalogVisibleRows()
	height = maxInt(height, maxCatalogDetailsHeight(m.categories))
	if height > 18 {
		return 18
	}
	return height
}

func maxCatalogPanelHeight(categories []catalog.Category) int {
	height := len(categories)
	for _, category := range categories {
		height = maxInt(height, len(category.Categories)+len(category.Apps))
		height = maxInt(height, maxCatalogPanelHeight(category.Categories))
	}
	return height
}

func maxCatalogDetailsHeight(categories []catalog.Category) int {
	height := 0
	for _, category := range categories {
		height = maxInt(height, len(categoryDetailsLines(category)))
		for _, app := range category.Apps {
			height = maxInt(height, len(packageDetailsLines(app, "No", "No")))
		}
		height = maxInt(height, maxCatalogDetailsHeight(category.Categories))
	}
	return height
}

func place(content string, width, height int) string {
	if width <= 0 || height <= 0 {
		return content
	}

	contentWidth := pageWidth(width)
	content = fitContent(content, contentWidth)
	content = lipgloss.NewStyle().Width(contentWidth).Render(content)
	content = lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
	return fillLines(content, width, height)
}

func pageWidth(width int) int {
	if width <= 0 {
		return 80
	}
	contentWidth := width - 6
	if contentWidth > 92 {
		return 92
	}
	if contentWidth < 40 {
		return width
	}
	return contentWidth
}

func fitContent(content string, width int) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = fitLine(line, width)
	}
	return strings.Join(lines, "\n")
}

func fitLine(line string, width int) string {
	if width <= 0 {
		return line
	}
	return ansi.Truncate(line, width, "...")
}

func fillLines(content string, width, height int) string {
	if width <= 0 || height <= 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, "")
	}

	for i, line := range lines {
		lineWidth := ansi.StringWidth(line)
		if lineWidth < width {
			line += strings.Repeat(" ", width-lineWidth)
		}
		lines[i] = line
	}

	return strings.Join(lines, "\n")
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
