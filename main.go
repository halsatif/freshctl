package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsatif/freshctl/internal/console"
	"github.com/halsatif/freshctl/internal/tui"
)

func main() {
	console.CenterWindow()

	p := tea.NewProgram(tui.NewModel(os.Args[1:]), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "freshctl failed: %v\n", err)
		os.Exit(1)
	}
}
