package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.Border{
			Top:         "-",
			Bottom:      "-",
			Left:        "|",
			Right:       "|",
			TopLeft:     "+",
			TopRight:    "+",
			BottomLeft:  "+",
			BottomRight: "+",
		}).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	activeItemStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("31"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	hotkeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)
)
