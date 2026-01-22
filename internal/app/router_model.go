package app

import (
	"fmt"

	"codeberg.org/JoaoGarcia/Mezzotone/internal/navigation"
	"codeberg.org/JoaoGarcia/Mezzotone/internal/ui"
	"codeberg.org/JoaoGarcia/Mezzotone/internal/ui/screens"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type RouterModel struct {
	route    navigation.Route
	keys     ui.KeyMap
	screens  map[navigation.Route]screens.Screen
	help     help.Model
	quitting bool
}

func NewRouterModel() RouterModel {
	return RouterModel{
		route: navigation.RouteMainMenu,
		keys:  ui.DefaultKeyMap(),
		screens: map[navigation.Route]screens.Screen{
			navigation.RouteMainMenu:         screens.NewMainMenuScreen(),
			navigation.RouteConvertImageMenu: screens.NewConvertImageMenuScreen(),
		},
		help: help.New(),
	}
}

func (m RouterModel) Init() tea.Cmd {
	return m.active().Init()
}

func (m RouterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case navigation.NavigateMsg:
		if msg.To != m.route {
			m.route = msg.To
			return m, m.active().Init()
		}
		return m, nil
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
			if m.route != navigation.RouteMainMenu {
				m.route = navigation.RouteMainMenu
				return m, m.active().Init()
			}
		}
	}

	active := m.active()
	updated, cmd := active.Update(msg)
	m.screens[m.route] = updated

	return m, cmd
}

func (m RouterModel) View() string {
	view := m.active().View()
	helpView := m.help.View(m.keys)
	if helpView == "" {
		return view
	}
	return view + "\n\n" + helpView
}

func (m RouterModel) active() screens.Screen {
	s, ok := m.screens[m.route]
	if !ok {
		panic(fmt.Sprintf("missing screen for route: %v", m.route))
	}
	return s
}
