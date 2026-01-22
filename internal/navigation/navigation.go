package navigation

import tea "github.com/charmbracelet/bubbletea"

type NavigateMsg struct {
	To Route
}

func Navigate(to Route) tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{To: to}
	}
}
