package app

import (
	"fmt"
	"os"
	"strings"

	"codeberg.org/JoaoGarcia/Mezzotone/internal/services"
	"codeberg.org/JoaoGarcia/Mezzotone/internal/ui"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ExperimentalModel struct {
	keys ui.KeyMap

	filePicker filepicker.Model
	renderView viewport.Model
	leftColumn viewport.Model

	style Styles

	width  int
	height int

	err error
}

type Styles struct {
	windowMarginVertical int
	leftColumnWidth      int
}

func NewExperimentalModel() ExperimentalModel {
	windowStyles := Styles{
		windowMarginVertical: 2,
	}

	fp := filepicker.New()
	fp.AllowedTypes = []string{".png", ".jpg", ".jpeg", ".bmp", ".webp", ".tiff"}
	fp.CurrentDirectory, _ = os.UserHomeDir()
	fp.ShowPermissions = false
	fp.ShowSize = true

	// viewports will be sized on first WindowSizeMsg
	renderView := viewport.New(0, 0)
	renderView.SetContent("Placeholder Text")
	leftColumn := viewport.New(0, 0)

	return ExperimentalModel{
		keys:       ui.DefaultKeyMap(),
		filePicker: fp,
		renderView: renderView,
		style:      windowStyles,
		leftColumn: leftColumn,
	}
}

func (m ExperimentalModel) Init() tea.Cmd {
	return m.filePicker.Init()
}

func (m ExperimentalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.filePicker, cmd = m.filePicker.Update(msg)
	if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
		m.filePicker.FileSelected = path
		_ = services.Logger().Info(fmt.Sprintf("Selected File: %s", m.filePicker.FileSelected))

		services.Shared().Set("selectedFile", m.filePicker.FileSelected)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

		m.renderView.Width = m.width / 7 * 5
		m.renderView.Height = m.height - m.style.windowMarginVertical

		m.style.leftColumnWidth = m.width / 7 * 2

		m.filePicker.SetHeight((m.height - m.style.windowMarginVertical) / 2)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

			//case "enter":
			//	path := strings.TrimSpace(m.fileInput.FilePicker.FileSelected)
			//	if path == "" {
			//		m.err = fmt.Errorf("please enter a file path")
			//		return m, nil
			//	}
			//
			//	// For now: load and display file contents.
			//	// Later: decode image and render ASCII into the viewport.
			//	b, err := os.ReadFile(path)
			//	if err != nil {
			//		m.err = err
			//		return m, nil
			//	}
			//
			//	m.err = nil
			//	m.vp.SetContent(string(b))
			//	m.vp.GotoTop()
			//	return m, nil
		}
	}

	// Update both: viewport handles scrolling keys, textinput handles typing.
	m.renderView, cmd = m.renderView.Update(msg)
	cmds = append(cmds, cmd)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m ExperimentalModel) View() string {
	filePickerTitleStyle := lipgloss.NewStyle().
		SetString("Pick a file to convert.\nAllowed types are " + strings.Join(m.filePicker.AllowedTypes, ", "))
	filePickerTitleRender := filePickerTitleStyle.Render()
	filePickerStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		Width(m.style.leftColumnWidth)
	filePickerRender := filePickerStyle.Render(filePickerTitleRender + "\n\n" + m.filePicker.View())

	lefColumnRender := lipgloss.JoinVertical(lipgloss.Left, filePickerRender)

	renderViewStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder())
	renderViewRender := renderViewStyle.Render(m.renderView.View())

	return lipgloss.JoinHorizontal(lipgloss.Left, lefColumnRender, renderViewRender)
}
