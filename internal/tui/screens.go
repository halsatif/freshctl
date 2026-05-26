package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/halsatif/freshctl/internal/catalog"
	"github.com/halsatif/freshctl/internal/installer"
)

func (m Model) viewWelcome() string {
	body := strings.Join([]string{
		titleStyle.Render("freshctl"),
		subtitleStyle.Render("fresh windows setup, but not painful"),
		"",
		"Choose apps from a small catalog and install them with Chocolatey.",
		"Nothing runs until you confirm the review screen.",
		"",
		hotkeyBar("enter continue", "q quit"),
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

	lines = append(lines, "", hotkeyBar("up/down move", "enter confirm", "q quit"))
	return place(strings.Join(lines, "\n"), m.width, m.height)
}

func (m Model) viewCatalog() string {
	contentWidth := pageWidth(m.width)
	itemLines := m.catalogListLines()
	if len(itemLines) == 0 {
		itemLines = append(itemLines, "  "+mutedStyle.Render("No matches."))
	}

	panelHeight := m.catalogPanelHeight()
	leftWidth, rightWidth := catalogPaneWidths(contentWidth)
	itemLines = m.visibleCatalogLines(itemLines, panelHeight)
	itemLines = fitCatalogListLines(itemLines, leftWidth)
	left := borderStyle.Width(leftWidth).Height(panelHeight).Render(strings.Join(itemLines, "\n"))
	right := borderStyle.Width(rightWidth).Height(panelHeight).Render(m.catalogDetailsPanel(rightWidth))
	content := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	if contentWidth < 72 {
		content = strings.Join([]string{left, "", right}, "\n")
	}

	parts := []string{
		titleStyle.Render("freshctl"),
		mutedStyle.Render(fmt.Sprintf("%d selected", len(m.selectedApps()))),
		mutedStyle.Render(m.catalogHeaderLine()),
	}
	if m.searchFocused {
		query := m.searchQuery
		if query == "" {
			query = "type to search"
		}
		parts = append(parts, mutedStyle.Render("Search: "+query+" (placeholder)"))
	}
	parts = append(parts,
		"",
		content,
	)
	if m.notice != "" {
		parts = append(parts, "", errorStyle.Render(m.notice))
	}
	if m.catalogMode == catalogModeFull {
		parts = append(parts, "", hotkeyBar("up/down move", "/ search", "space select", "i install", "esc back/clear", "q quit"))
	} else {
		parts = append(parts, "", hotkeyBar("up/down move", "enter open", "space select", "esc back", "i install", "q quit"))
	}

	return place(strings.Join(parts, "\n"), m.width, m.height)
}

func (m Model) catalogListLines() []string {
	if m.catalogMode == catalogModeFull {
		return m.fullCatalogListLines()
	}
	return m.categoryCatalogListLines()
}

func (m Model) categoryCatalogListLines() []string {
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
		if m.selected[app.PackageID] {
			box = selectedStyle.Render("[x]")
		}
		line := fmt.Sprintf("%s %s", box, app.Name)
		if len(categories)+i == m.catalogCursor {
			line = activeItemStyle.Render("> " + line)
		} else {
			line = "  " + line
		}
		itemLines = append(itemLines, line)
	}

	return itemLines
}

func (m Model) fullCatalogListLines() []string {
	items := m.filteredFullCatalogItems()
	lines := make([]string, 0, len(items))
	for i, item := range items {
		box := "[ ]"
		if m.selected[item.Package.PackageID] {
			box = selectedStyle.Render("[x]")
		}
		line := fmt.Sprintf("%s %s", box, item.Package.Name)
		if i == m.catalogCursor {
			line = activeItemStyle.Render("> " + line)
		} else {
			line = "  " + line
		}
		lines = append(lines, line)
	}
	return lines
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

func (m Model) viewReview() string {
	selected := m.selectedApps()
	lines := []string{
		titleStyle.Render("review"),
	}

	if len(selected) == 0 {
		lines = append(lines, "", mutedStyle.Render("No apps selected yet. Press b to return to the catalog."))
	} else {
		lines = append(lines, "", "Selected apps:")
		for _, app := range selected {
			lines = append(lines, fmt.Sprintf("  - %s %s", app.Name, mutedStyle.Render(app.PackageID)))
		}

		lines = append(lines, "", "Commands:")
		for _, app := range selected {
			lines = append(lines, "  "+installer.CommandFor(app))
		}
	}

	if m.notice != "" {
		lines = append(lines, "", errorStyle.Render(m.notice))
	}

	lines = append(lines, "", hotkeyBar("enter install", "b/esc back", "q quit"))
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
		elapsed = " " + mutedStyle.Render(formatElapsed(time.Since(m.currentStart)))
	}

	progress := fmt.Sprintf("[%d/%d]", m.currentStep, total)
	if total == 0 {
		progress = "[0/0]"
	}

	spin := " "
	if !m.installDone {
		spin = spinnerFrame(m.spinnerFrame)
	}

	command := m.currentCmd
	if command == "" {
		command = "waiting for Chocolatey..."
	}
	command = fitLine(command, contentWidth)

	lines := []string{
		titleStyle.Render("install"),
		fitLine(fmt.Sprintf("%s current: %s%s %s", spin, currentName, elapsed, mutedStyle.Render(progress)), contentWidth),
		mutedStyle.Render(command),
	}

	if !m.showFullLog {
		lines = append(lines, "", mutedStyle.Render("Logs hidden. Press l to show full logs."))
	} else {
		logLines := tailLines(m.fullLog, installLogLimit(m.height))
		lines = append(lines, "", mutedStyle.Render("full logs"))
		if len(logLines) == 0 {
			lines = append(lines, "  "+mutedStyle.Render("Waiting for output..."))
		}
		for _, line := range logLines {
			lines = append(lines, fitLine("  "+sanitizeLogLine(line), contentWidth))
		}
	}

	if m.installDone {
		lines = append(lines, "", m.installDoneMessage())
	}

	lines = append(lines, "", "Summary:")
	lines = append(lines, m.installSummaryTable(contentWidth)...)
	lines = append(lines, "", hotkeyBar("s skip app", "l show/hide logs", "q quit"))
	return place(strings.Join(lines, "\n"), m.width, m.height)
}

func (m Model) viewBootstrap() string {
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
		lines = append(lines, selectedStyle.Render("Bootstrapping Chocolatey..."))
	} else {
		lines = append(lines, mutedStyle.Render("Press enter to bootstrap, or r to retry detection."))
	}

	logLines := m.bootstrapLog
	maxLines := m.height - 14
	if maxLines < 5 {
		maxLines = 5
	}
	if len(logLines) > maxLines {
		logLines = logLines[len(logLines)-maxLines:]
	}
	if len(logLines) > 0 {
		lines = append(lines, "")
		lines = append(lines, logLines...)
	}

	lines = append(lines, "", hotkeyBar("enter bootstrap", "r retry", "b/esc back", "q quit"))
	return place(strings.Join(lines, "\n"), m.width, m.height)
}

func (m Model) viewElevation() string {
	lines := []string{
		titleStyle.Render("administrator privileges required"),
		"",
		"freshctl needs administrator privileges to install Chocolatey and applications.",
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

	lines = append(lines, "", hotkeyBar("enter relaunch as administrator", "q quit"))
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

	lines = append(lines, "", hotkeyBar("enter remove and reinstall", "b/esc back", "q quit"))
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

func (m Model) installSummaryTable(width int) []string {
	if len(m.installApps) == 0 {
		return []string{"  " + mutedStyle.Render("No apps queued.")}
	}

	lines := make([]string, 0, len(m.installApps))
	nameWidth := m.installNameWidth(width)
	for _, app := range m.installApps {
		status := m.appStatus[app.PackageID]
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

func (m Model) catalogDetailsPanel(width int) string {
	if m.catalogMode == catalogModeFull {
		items := m.filteredFullCatalogItems()
		if m.catalogCursor < 0 || m.catalogCursor >= len(items) {
			return fitDetailsLines([]string{"No item selected."}, width)
		}
		item := items[m.catalogCursor]
		selected := "No"
		if m.selected[item.Package.PackageID] {
			selected = "Yes"
		}
		return fitDetailsLines(packageDetailsLines(item.Package, selected), width)
	}

	categories := m.currentCategories()
	if m.catalogCursor < len(categories) {
		category := categories[m.catalogCursor]
		return fitDetailsLines(categoryDetailsLines(category), width)
	}

	appIndex := m.catalogCursor - len(categories)
	apps := m.currentApps()
	if appIndex < 0 || appIndex >= len(apps) {
		return fitDetailsLines([]string{"No item selected."}, width)
	}

	app := apps[appIndex]
	selected := "No"
	if m.selected[app.PackageID] {
		selected = "Yes"
	}
	return fitDetailsLines(packageDetailsLines(app, selected), width)
}

func fitDetailsLines(lines []string, width int) string {
	innerWidth := width - 2
	if innerWidth < 12 {
		innerWidth = 12
	}
	for i, line := range lines {
		lines[i] = fitLine(line, innerWidth)
	}
	return strings.Join(lines, "\n")
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

func packageDetailsLines(app catalog.Package, selected string) []string {
	lines := []string{
		"Package:",
		app.PackageID,
		"",
		"Manager:",
		"Chocolatey",
		"",
		"Description:",
	}
	lines = append(lines, wrapText(app.Description, 28)...)
	lines = append(lines, "", "Selected:", selected)
	return lines
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

func (m Model) elapsedForApp(app catalog.Package) string {
	if elapsed, ok := m.appElapsed[app.PackageID]; ok {
		return formatElapsed(elapsed)
	}
	if m.currentApp.PackageID == app.PackageID && !m.currentStart.IsZero() && !m.installDone {
		return formatElapsed(time.Since(m.currentStart))
	}
	return "--:--"
}

func (m Model) installDoneMessage() string {
	failed := 0
	skipped := 0
	installed := 0
	for _, app := range m.installApps {
		switch m.appStatus[app.PackageID] {
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

func installLogLimit(height int) int {
	if height <= 0 {
		return 12
	}
	limit := height - 18
	if limit < 6 {
		return 6
	}
	if limit > 15 {
		return 15
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
		if m.selected[app.PackageID] {
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
			height = maxInt(height, len(packageDetailsLines(app, "No")))
		}
		height = maxInt(height, maxCatalogDetailsHeight(category.Categories))
	}
	return height
}

func hotkeyBar(parts ...string) string {
	return hotkeyStyle.Render(strings.Join(parts, "  |  "))
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
