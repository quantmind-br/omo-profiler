package views

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ImportDoneMsg is sent when import is completed
type ImportDoneMsg struct {
	ProfileName string
	Path        string
	Err         error
}

// ImportCancelMsg is sent when user cancels import
type ImportCancelMsg struct{}

type importKeyMap struct {
	Confirm key.Binding
	Cancel  key.Binding
}

func newImportKeyMap() importKeyMap {
	return importKeyMap{
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "import"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// Import is a view for importing profiles from JSON files
type Import struct {
	textInput textinput.Model
	width     int
	height    int
	err       error
	loading   bool
	keys      importKeyMap
}

// NewImport creates a new Import view
func NewImport() Import {
	ti := textinput.New()
	ti.Placeholder = "Path to JSON file..."
	ti.Focus()
	ti.Width = 60

	return Import{
		textInput: ti,
		keys:      newImportKeyMap(),
	}
}

// Init initializes the view
func (i Import) Init() tea.Cmd {
	return textinput.Blink
}

// SetSize sets the view dimensions
func (i *Import) SetSize(width, height int) {
	i.width = width
	i.height = height
	i.textInput.Width = width - 10
}

// Update handles messages and user input
func (i Import) Update(msg tea.Msg) (Import, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, i.keys.Confirm):
			path := i.textInput.Value()
			if path == "" {
				i.err = errors.New("please enter a file path")
				return i, nil
			}

			expandedPath, err := expandPath(path)
			if err != nil {
				i.err = err
				return i, nil
			}

			if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
				i.err = errors.New("file not found")
				return i, nil
			}

			return i, func() tea.Msg {
				return ImportDoneMsg{
					ProfileName: "",
					Path:        expandedPath,
					Err:         nil,
				}
			}

		case key.Matches(msg, i.keys.Cancel):
			return i, func() tea.Msg {
				return ImportCancelMsg{}
			}
		}
	}

	// Update text input
	i.textInput, cmd = i.textInput.Update(msg)
	return i, cmd
}

// View renders the import view
func (i Import) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#CDD6F4")).
		MarginBottom(1)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7086")).
		MarginBottom(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7086"))

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F38BA8"))

	title := titleStyle.Render("Import Profile")
	desc := descStyle.Render("Enter the path to a JSON profile file")
	input := i.textInput.View()
	help := helpStyle.Render("[enter] import  [esc] cancel")

	content := []string{
		"",
		title,
		"",
		desc,
		input,
	}

	if i.err != nil {
		content = append(content, errorStyle.Render("✗ "+i.err.Error()))
	}

	content = append(content, "", help)

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

// GetPath returns the current path value
func (i Import) GetPath() string {
	return i.textInput.Value()
}

// expandPath expands ~ to home directory and converts relative paths to absolute
func expandPath(path string) (string, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
	}

	// Convert to absolute path
	if !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		path = absPath
	}

	return path, nil
}
