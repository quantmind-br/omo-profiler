package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/tui/layout"
)

var (
	exportWhite  = lipgloss.Color("#CDD6F4")
	exportGray   = lipgloss.Color("#6C7086")
	exportRed    = lipgloss.Color("#F38BA8")
	exportPurple = lipgloss.Color("#7D56F4")
)

var (
	exportTitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(exportWhite).MarginBottom(1)
	exportDescStyle    = lipgloss.NewStyle().Foreground(exportGray).MarginBottom(1)
	exportHelpStyle    = lipgloss.NewStyle().Foreground(exportGray)
	exportErrorStyle   = lipgloss.NewStyle().Foreground(exportRed)
	exportProfileStyle = lipgloss.NewStyle().Bold(true).Foreground(exportPurple)
)

// ExportDoneMsg is sent when export is completed
type ExportDoneMsg struct {
	Path string
	Err  error
}

// ExportCancelMsg is sent when user cancels export
type ExportCancelMsg struct{}

type exportKeyMap struct {
	Confirm key.Binding
	Cancel  key.Binding
}

func newExportKeyMap() exportKeyMap {
	return exportKeyMap{
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "export"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// Export is a view for exporting profiles to JSON files
type Export struct {
	textInput   textinput.Model
	profileName string
	width       int
	height      int
	err         error
	keys        exportKeyMap
}

// NewExport creates a new Export view for the given profile
func NewExport(profileName string) Export {
	ti := textinput.New()
	ti.Placeholder = "Export path..."
	ti.Focus()
	ti.Width = 60

	return Export{
		textInput:   ti,
		profileName: profileName,
		keys:        newExportKeyMap(),
	}
}

// Init initializes the view
func (e Export) Init() tea.Cmd {
	return textinput.Blink
}

// SetSize sets the view dimensions
func (e *Export) SetSize(width, height int) {
	e.width = width
	e.height = height
	e.textInput.Width = layout.WideFieldWidth(width, 10)
}

// Update handles messages and user input
func (e Export) Update(msg tea.Msg) (Export, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, e.keys.Confirm):
			path := e.textInput.Value()
			if path == "" {
				e.err = fmt.Errorf("please enter a file path")
				return e, nil
			}

			// Expand path
			expandedPath, err := expandExportPath(path)
			if err != nil {
				e.err = err
				return e, nil
			}

			// Auto-rename if file exists
			expandedPath = autoRenameIfExists(expandedPath)

			return e, func() tea.Msg {
				return ExportDoneMsg{
					Path: expandedPath,
					Err:  nil,
				}
			}

		case key.Matches(msg, e.keys.Cancel):
			return e, func() tea.Msg {
				return ExportCancelMsg{}
			}
		}
	}

	e.textInput, cmd = e.textInput.Update(msg)
	return e, cmd
}

// View renders the export view
func (e Export) View() string {
	title := exportTitleStyle.Render("Export Profile")
	profileInfo := fmt.Sprintf("Exporting profile: %s", exportProfileStyle.Render(e.profileName))
	desc := exportDescStyle.Render("Enter the destination path for the JSON file")
	input := e.textInput.View()
	help := exportHelpStyle.Render("[enter] export  [esc] cancel")

	content := []string{
		"",
		title,
		"",
		profileInfo,
		desc,
		input,
	}

	if e.err != nil {
		content = append(content, exportErrorStyle.Render("✗ "+e.err.Error()))
	}

	content = append(content, "", help)
	if layout.IsShort(e.height) {
		compact := make([]string, 0, len(content))
		for _, line := range content {
			if line != "" {
				compact = append(compact, line)
			}
		}
		content = compact
	}

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

// GetPath returns the current path value
func (e Export) GetPath() string {
	return e.textInput.Value()
}

// GetProfileName returns the profile being exported
func (e Export) GetProfileName() string {
	return e.profileName
}

// expandExportPath expands ~ to home directory and converts relative paths to absolute
func expandExportPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
	}

	if !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		path = absPath
	}

	return path, nil
}

// autoRenameIfExists adds a numeric suffix if the file already exists
func autoRenameIfExists(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)

	for i := 1; ; i++ {
		newPath := filepath.Join(dir, fmt.Sprintf("%s-%d%s", base, i, ext))
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}
