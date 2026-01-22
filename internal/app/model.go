package app

import (
	"fmt"

	"codeberg.org/JoaoGarcia/Mezzotone/internal/ui"
	"codeberg.org/JoaoGarcia/Mezzotone/internal/ui/screens"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type RootModel struct {
	route   Route
	keys    ui.KeyMap
	screens map[Route]screens.Screen
	help    help.Model
}

func NewRootModel() RootModel {
	return RootModel{
		route: RouteMainMenu,
		keys:  ui.DefaultKeyMap(),
		screens: map[Route]screens.Screen{
			RouteMainMenu: screens.NewMainMenuScreen(),
		},
		help: help.New(),
	}
}

func (m RootModel) Init() tea.Cmd {
	return m.active().Init()
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Back):
			//TODO route stack for back
			if m.route != RouteMainMenu {
				m.route = RouteMainMenu
				return m, m.active().Init()
			}
		}
	}

	active := m.active()
	updated, cmd := active.Update(msg)
	m.screens[m.route] = updated

	return m, cmd
}

func (m RootModel) View() string {
	view := m.active().View()
	helpView := m.help.View(m.keys)
	if helpView == "" {
		return view
	}
	return view + "\n\n" + helpView
}

func (m RootModel) active() screens.Screen {
	s, ok := m.screens[m.route]
	if !ok {
		panic(fmt.Sprintf("missing screen for route: %v", m.route))
	}
	return s
}
