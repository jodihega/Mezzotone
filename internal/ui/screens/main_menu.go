package screens

import (
	"fmt"
	"io"
	"strings"

	"codeberg.org/JoaoGarcia/Mezzotone/internal/ui"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TODO maybe move to a style package ?
var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	_, _ = fmt.Fprint(w, fn(str))
}

type MainMenuScreen struct {
	list   list.Model
	choice string
}

func NewMainMenuScreen() MainMenuScreen {
	items := []list.Item{
		item("Option 1"),
		item("Option 2"),
		item("Option 3"),
		item("Quit"),
	}

	const defaultWidth = 20
	listHeight := len(items) * 2

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Main Menu"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle

	return MainMenuScreen{
		list: l,
	}
}

func (m MainMenuScreen) Init() tea.Cmd {
	return nil
}

func (m MainMenuScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	keys := ui.DefaultKeyMap()

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Confirm):
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m MainMenuScreen) View() string {
	return "Mezzotone\n\n" + m.list.View()
}
