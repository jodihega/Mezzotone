package main

import (
	"fmt"
	"os"

	"codeberg.org/JoaoGarcia/Mezzotone/internal/app"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(app.NewRootModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
