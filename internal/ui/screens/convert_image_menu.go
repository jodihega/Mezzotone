package screens

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
)

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

type ConvertImageMenuScreen struct {
	filepicker   filepicker.Model
	selectedFile string
	err          error
}

const FilePickerHeight = 20

func NewConvertImageMenuScreen() ConvertImageMenuScreen {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".png", ".jpg", ".jpeg", "", ".bmp", ".webp", ".tiff"}
	fp.CurrentDirectory, _ = os.UserHomeDir()
	fp.SetHeight(FilePickerHeight)

	return ConvertImageMenuScreen{
		filepicker: fp,
	}
}

func (m ConvertImageMenuScreen) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m ConvertImageMenuScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg.(type) {
	case tea.WindowSizeMsg:
		m.filepicker.SetHeight(FilePickerHeight - 2)
	case clearErrorMsg:
		m.err = nil
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		m.selectedFile = path
	}

	if didSelect, path := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		m.err = errors.New(path + " is not valid.")
		m.selectedFile = ""
		return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return m, cmd
}

func (m ConvertImageMenuScreen) View() string {
	var s strings.Builder
	s.WriteString("\n  ")
	if m.err != nil {
		s.WriteString(m.filepicker.Styles.DisabledFile.Render(m.err.Error()))
	} else if m.selectedFile == "" {
		s.WriteString("Pick a file:")
	} else {
		s.WriteString("Selected file: " + m.filepicker.Styles.Selected.Render(m.selectedFile))
	}
	s.WriteString("\n\n" + m.filepicker.View() + "\n")
	return s.String()
}
