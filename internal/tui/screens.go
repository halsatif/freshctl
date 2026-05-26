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

func (m Model) viewCatalog() string {
	categories := make([]string, 0, len(m.categories))
	for i, category := range m.categories {
		line := category.Name
		if count := m.selectedInCategory(i); count > 0 {
			line = fmt.Sprintf("%s (%d)", line, count)
		}
		if i == m.categoryCursor {
			if m.focus == focusCategories {
				line = activeItemStyle.Render("> " + line)
			} else {
				line = selectedStyle.Render("> " + line)
			}
		} else {
			line = "  " + line
		}
		categories = append(categories, line)
	}

	appLines := make([]string, 0, len(m.currentApps()))
	for i, app := range m.currentApps() {
		box := "[ ]"
		if m.selected[app.ID] {
			box = selectedStyle.Render("[x]")
		}
		line := fmt.Sprintf("%s %s", box, app.Name)
		if i == m.appCursor {
			if m.focus == focusApps {
				line = activeItemStyle.Render("> " + line)
			} else {
				line = selectedStyle.Render("> " + line)
			}
		} else {
			line = "  " + line
		}
		appLines = append(appLines, line)
	}

	panelHeight := maxInt(len(categories), len(appLines))
	if panelHeight < 6 {
		panelHeight = 6
	}
	left := borderStyle.Width(24).Height(panelHeight).Render(strings.Join(categories, "\n"))
	right := borderStyle.Width(58).Height(panelHeight).Render(strings.Join(appLines, "\n"))
	content := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)

	parts := []string{
		titleStyle.Render("freshctl"),
		mutedStyle.Render(fmt.Sprintf("%d selected", len(m.selectedApps()))),
		"",
		content,
	}
	if m.notice != "" {
		parts = append(parts, "", errorStyle.Render(m.notice))
	}
	parts = append(parts, "", hotkeyBar("up/down move", "tab focus", "space select", "enter review", "q quit"))

	return place(strings.Join(parts, "\n"), m.width, m.height)
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
			lines = append(lines, fmt.Sprintf("  - %s %s", app.Name, mutedStyle.Render(app.ID)))
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

func (m Model) elapsedForApp(app catalog.App) string {
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

func (m Model) selectedInCategory(index int) int {
	if index < 0 || index >= len(m.categories) {
		return 0
	}

	count := 0
	for _, app := range m.categories[index].Apps {
		if m.selected[app.ID] {
			count++
		}
	}
	return count
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
