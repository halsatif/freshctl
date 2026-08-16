package tui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

const (
	footerPriorityPrimary = iota
	footerPriorityNavigation
	footerPrioritySecondary
	footerPriorityTertiary
)

type keyHint struct {
	Key         string
	Description string
	Priority    int
	Aliases     []string
}

func hint(key, description string, priority int, aliases ...string) keyHint {
	return keyHint{
		Key:         key,
		Description: description,
		Priority:    priority,
		Aliases:     aliases,
	}
}

func (m Model) footer(hints ...keyHint) string {
	return renderFooter(pageWidth(m.width), hints)
}

func renderFooter(width int, hints []keyHint) string {
	visible := append([]keyHint(nil), hints...)
	for len(visible) > 0 {
		rendered := renderFooterHints(visible)
		if width <= 0 || ansi.StringWidth(rendered) <= width {
			return rendered
		}

		if len(visible) == 1 {
			keyOnly := footerKeyStyle.Render(visible[0].Key)
			if ansi.StringWidth(keyOnly) <= width {
				return keyOnly
			}
			return ""
		}

		remove := 0
		for index := 1; index < len(visible); index++ {
			if visible[index].Priority >= visible[remove].Priority {
				remove = index
			}
		}
		visible = append(visible[:remove], visible[remove+1:]...)
	}
	return ""
}

func renderFooterHints(hints []keyHint) string {
	parts := make([]string, 0, len(hints))
	for _, item := range hints {
		parts = append(parts, footerKeyStyle.Render(item.Key)+" "+footerDescriptionStyle.Render(item.Description))
	}
	return strings.Join(parts, footerSeparatorStyle.Render("  •  "))
}

func helpKeyLabel(item keyHint) string {
	keys := append([]string{item.Key}, item.Aliases...)
	return strings.Join(keys, " / ")
}

func (m Model) footerFor(target screen) string {
	return m.footer(m.footerHints(target)...)
}

func (m Model) footerHints(target screen) []keyHint {
	help := hint("?", "help", footerPriorityTertiary)
	switch target {
	case screenWelcome:
		return []keyHint{
			hint("enter", "continue", footerPriorityPrimary),
			help,
			hint("q", "quit", footerPriorityTertiary),
		}
	case screenModeSelect:
		return []keyHint{
			hint("↑↓", "move", footerPriorityNavigation, "j/k"),
			hint("enter", "confirm", footerPriorityPrimary),
			hint("esc", "back", footerPrioritySecondary),
			help,
		}
	case screenCatalog:
		if m.searchFocused {
			return []keyHint{
				hint("↑↓", "move", footerPriorityNavigation, "j/k"),
				hint("enter", "done", footerPriorityPrimary),
				hint("esc", "clear", footerPrioritySecondary),
				help,
			}
		}
		enterDescription := "review"
		if m.catalogMode == catalogModeCategories {
			enterDescription = "open/review"
		}
		escapeDescription := "back"
		if m.searchActive() {
			escapeDescription = "clear"
		}
		return []keyHint{
			hint("↑↓", "move", footerPriorityNavigation, "j/k"),
			hint("space", "select", footerPriorityNavigation),
			hint("/", "search", footerPrioritySecondary),
			hint("enter", enterDescription, footerPriorityPrimary),
			hint("esc", escapeDescription, footerPrioritySecondary),
			help,
		}
	case screenPresetPicker:
		return []keyHint{
			hint("↑↓", "move", footerPriorityNavigation, "j/k"),
			hint("enter", "apply", footerPriorityPrimary),
			hint("esc", "back", footerPrioritySecondary),
			help,
		}
	case screenReview:
		return []keyHint{
			hint("↑↓", "scroll", footerPriorityNavigation, "j/k"),
			hint("enter", "install", footerPriorityPrimary),
			hint("esc", "back", footerPrioritySecondary),
			help,
		}
	case screenInstall:
		hints := []keyHint{hint("l", "logs", footerPriorityPrimary)}
		if !m.installDone && m.skipInstall != nil && m.currentApp.ID != "" {
			hints = append(hints, hint("s", "skip", footerPriorityNavigation))
		}
		if m.installDone {
			hints = append(hints, hint("esc", "catalog", footerPriorityNavigation))
		}
		return append(hints, help)
	case screenInstallLogs:
		return []keyHint{
			hint("↑↓", "scroll", footerPriorityNavigation, "j/k"),
			hint("end", "latest", footerPrioritySecondary),
			hint("esc", "back", footerPriorityPrimary),
			help,
		}
	case screenBootstrap:
		hints := make([]keyHint, 0, 5)
		if !m.bootstrapRunning {
			hints = append(hints, hint("enter", "bootstrap", footerPriorityPrimary))
		}
		hints = append(hints, hint("l", "logs", footerPriorityNavigation))
		if !m.bootstrapRunning {
			hints = append(hints, hint("r", "retry", footerPrioritySecondary))
		}
		return append(hints, hint("esc", "back", footerPrioritySecondary), help)
	case screenElevation:
		hints := []keyHint{}
		if !m.elevationRunning {
			hints = append(hints, hint("enter", "relaunch", footerPriorityPrimary))
			hints = append(hints, hint("esc", "back", footerPrioritySecondary))
		}
		return append(hints, help)
	case screenBrokenChocolatey:
		hints := []keyHint{}
		if !m.brokenRunning {
			hints = append(hints, hint("enter", "repair", footerPriorityPrimary))
		}
		return append(hints, hint("esc", "back", footerPrioritySecondary), help)
	default:
		return []keyHint{help}
	}
}

func (m Model) helpHintsFor(target screen) []keyHint {
	var hints []keyHint
	switch target {
	case screenWelcome:
		hints = []keyHint{hint("enter", "continue", footerPriorityPrimary)}
	case screenModeSelect:
		hints = []keyHint{
			hint("↑↓", "move", footerPriorityNavigation, "j/k"),
			hint("enter", "confirm catalog mode", footerPriorityPrimary),
			hint("esc", "back to welcome", footerPrioritySecondary),
		}
	case screenCatalog:
		if m.searchFocused {
			hints = []keyHint{
				hint("↑↓", "move through results", footerPriorityNavigation, "j/k"),
				hint("type", "update search query", footerPriorityNavigation),
				hint("backspace", "edit search query", footerPrioritySecondary),
				hint("enter", "finish search input", footerPriorityPrimary),
				hint("esc", "clear search", footerPrioritySecondary),
			}
		} else {
			enterDescription := "open review"
			if m.catalogMode == catalogModeCategories {
				enterDescription = "open folder or review package"
			}
			hints = []keyHint{
				hint("↑↓", "move", footerPriorityNavigation, "j/k"),
				hint("space", "toggle package selection", footerPriorityNavigation),
				hint("enter", enterDescription, footerPriorityPrimary),
				hint("/", "search packages", footerPrioritySecondary),
				hint("esc", "clear search or go back", footerPrioritySecondary),
				hint("i", "open review", footerPrioritySecondary),
				hint("p", "open presets", footerPriorityTertiary),
				hint("o", "import profile", footerPriorityTertiary),
			}
		}
	case screenPresetPicker:
		hints = []keyHint{
			hint("↑↓", "move", footerPriorityNavigation, "j/k"),
			hint("enter", "apply highlighted preset", footerPriorityPrimary),
			hint("esc", "back to catalog", footerPrioritySecondary),
		}
	case screenReview:
		hints = []keyHint{
			hint("↑↓", "scroll selection", footerPriorityNavigation, "j/k"),
			hint("pgup/pgdn", "scroll one page", footerPrioritySecondary),
			hint("home/end", "first or last package", footerPrioritySecondary),
			hint("enter", "start installation", footerPriorityPrimary),
			hint("e", "export profile", footerPriorityTertiary),
			hint("esc", "back to catalog", footerPrioritySecondary),
		}
	case screenInstall:
		hints = []keyHint{
			hint("↑↓", "scroll summary", footerPriorityNavigation, "j/k"),
			hint("pgup/pgdn", "scroll one page", footerPrioritySecondary),
			hint("home/end", "first or last result", footerPrioritySecondary),
			hint("l", "open installation logs", footerPriorityPrimary),
		}
		if !m.installDone && m.skipInstall != nil && m.currentApp.ID != "" {
			hints = append(hints, hint("s", "skip current application", footerPriorityNavigation))
		}
		if m.installDone {
			hints = append(hints, hint("esc", "back to catalog", footerPrioritySecondary))
		}
	case screenInstallLogs:
		hints = []keyHint{
			hint("↑↓", "scroll logs", footerPriorityNavigation, "j/k"),
			hint("pgup/pgdn", "scroll one page", footerPrioritySecondary),
			hint("home/end", "oldest or latest output", footerPrioritySecondary),
			hint("l", "back to installation", footerPrioritySecondary),
			hint("esc", "back to installation", footerPriorityPrimary),
		}
	case screenBootstrap:
		hints = []keyHint{hint("l", "show or hide bootstrap logs", footerPriorityNavigation)}
		if !m.bootstrapRunning {
			hints = append(hints,
				hint("enter", "bootstrap Chocolatey", footerPriorityPrimary),
				hint("r", "retry Chocolatey detection", footerPrioritySecondary),
			)
		}
		hints = append(hints, hint("esc", "back", footerPrioritySecondary))
	case screenElevation:
		if !m.elevationRunning {
			hints = []keyHint{
				hint("enter", "relaunch as administrator", footerPriorityPrimary),
				hint("esc", "back", footerPrioritySecondary),
			}
		}
	case screenBrokenChocolatey:
		if !m.brokenRunning {
			hints = []keyHint{hint("enter", "remove and reinstall Chocolatey", footerPriorityPrimary)}
		}
		hints = append(hints, hint("esc", "back", footerPrioritySecondary))
	}

	if m.canQuitSafely() {
		hints = append(hints, hint("q", "quit", footerPriorityTertiary))
	}
	return append(hints, hint("ctrl+c", "emergency exit", footerPriorityTertiary))
}

func screenName(value screen) string {
	switch value {
	case screenWelcome:
		return "welcome"
	case screenModeSelect:
		return "catalog mode"
	case screenCatalog:
		return "catalog"
	case screenPresetPicker:
		return "presets"
	case screenReview:
		return "review"
	case screenInstall:
		return "install"
	case screenInstallLogs:
		return "logs"
	case screenBootstrap:
		return "Chocolatey bootstrap"
	case screenElevation:
		return "administrator relaunch"
	case screenBrokenChocolatey:
		return "Chocolatey repair"
	default:
		return "freshctl"
	}
}
