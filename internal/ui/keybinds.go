package ui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Quit    key.Binding
	Back    key.Binding
	Confirm key.Binding
	Help    key.Binding
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("ctrl+c", "Quit"),
			key.WithHelp("q", "Quit"),
		),
		Back: key.NewBinding(
			key.WithKeys("backspace", "esc"),
			key.WithHelp("backspace", "Back"),
			key.WithHelp("esc", "Back"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "Confirm"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "Help"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("up", "Up"),
			key.WithHelp("k", "Up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("down", "Down"),
			key.WithHelp("j", "Down"),
		),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.Confirm,
		k.Back,
		k.Quit,
		k.Help,
	}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Confirm, k.Back, k.Quit, k.Help},
	}
}
