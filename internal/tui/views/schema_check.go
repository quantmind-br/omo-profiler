package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/schema"
	"github.com/diogenes/omo-profiler/internal/tui/layout"
)

type schemaCheckState int

const (
	stateSchemaCheckLoading schemaCheckState = iota
	stateSchemaCheckResult
	stateSchemaCheckError
	stateSchemaCheckSavePath
	stateSchemaCheckSaved
)

type schemaCheckKeyMap struct {
	Esc   key.Binding
	Retry key.Binding
	Enter key.Binding
}

func newSchemaCheckKeyMap() schemaCheckKeyMap {
	return schemaCheckKeyMap{
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Retry: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "retry"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
	}
}

type SchemaCheck struct {
	state     schemaCheckState
	result    *schema.CompareResult
	spinner   spinner.Model
	errorMsg  string
	width     int
	height    int
	keys      schemaCheckKeyMap
	textInput textinput.Model
	savedPath string
}

type NavToSchemaCheckMsg struct{}
type SchemaCheckBackMsg struct{}

type schemaCheckResultMsg struct {
	result *schema.CompareResult
	err    error
}

type schemaCheckSaveMsg struct {
	path string
	err  error
}

func NewSchemaCheck() SchemaCheck {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	ti := textinput.New()
	ti.Placeholder = "Enter folder path to save diff (e.g., ~/Downloads)"
	ti.CharLimit = 256
	ti.Width = 50

	return SchemaCheck{
		state:     stateSchemaCheckLoading,
		spinner:   s,
		keys:      newSchemaCheckKeyMap(),
		textInput: ti,
	}
}

func (s SchemaCheck) Init() tea.Cmd {
	return tea.Batch(
		s.spinner.Tick,
		fetchSchemaCompareCmd,
	)
}

func fetchSchemaCompareCmd() tea.Msg {
	result, err := schema.CompareSchemas()
	return schemaCheckResultMsg{result: result, err: err}
}

func (s *SchemaCheck) SetSize(w, h int) {
	s.width = w
	s.height = h
	s.textInput.Width = layout.WideFieldWidth(w, 10)
}

// expandSchemaPath expands ~ to home directory and resolves to absolute path
func expandSchemaPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
	}
	return filepath.Abs(path)
}

func (s SchemaCheck) saveDiffToFile(folderPath string) tea.Cmd {
	return func() tea.Msg {
		expandedPath, err := expandSchemaPath(folderPath)
		if err != nil {
			return schemaCheckSaveMsg{err: fmt.Errorf("invalid path: %w", err)}
		}

		// Create directory if it doesn't exist
		if err := os.MkdirAll(expandedPath, 0755); err != nil {
			return schemaCheckSaveMsg{err: fmt.Errorf("failed to create directory: %w", err)}
		}

		// Save diff to file
		filePath := filepath.Join(expandedPath, "schema-diff.md")
		content := fmt.Sprintf("# Schema Diff\n\nDifferences between embedded and upstream schema:\n\n```diff\n%s\n```\n", s.result.Diff)

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return schemaCheckSaveMsg{err: fmt.Errorf("failed to save file: %w", err)}
		}

		return schemaCheckSaveMsg{path: filePath, err: nil}
	}
}

func (s SchemaCheck) Update(msg tea.Msg) (SchemaCheck, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case schemaCheckResultMsg:
		if msg.err != nil {
			s.state = stateSchemaCheckError
			s.errorMsg = msg.err.Error()
			return s, nil
		}
		s.result = msg.result
		if s.result.Identical {
			s.state = stateSchemaCheckResult
		} else {
			s.state = stateSchemaCheckSavePath
			s.textInput.Focus()
			return s, textinput.Blink
		}
		return s, nil

	case schemaCheckSaveMsg:
		if msg.err != nil {
			s.state = stateSchemaCheckError
			s.errorMsg = msg.err.Error()
			return s, nil
		}
		s.savedPath = msg.path
		s.state = stateSchemaCheckSaved
		return s, nil

	case spinner.TickMsg:
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd

	case tea.KeyMsg:
		switch s.state {
		case stateSchemaCheckSavePath:
			switch {
			case key.Matches(msg, s.keys.Esc):
				return s, func() tea.Msg { return SchemaCheckBackMsg{} }
			case key.Matches(msg, s.keys.Enter):
				path := strings.TrimSpace(s.textInput.Value())
				if path == "" {
					s.errorMsg = "Please enter a folder path"
					return s, nil
				}
				return s, s.saveDiffToFile(path)
			default:
				s.textInput, cmd = s.textInput.Update(msg)
				return s, cmd
			}

		case stateSchemaCheckResult, stateSchemaCheckSaved:
			if key.Matches(msg, s.keys.Esc) {
				return s, func() tea.Msg { return SchemaCheckBackMsg{} }
			}

		case stateSchemaCheckError:
			switch {
			case key.Matches(msg, s.keys.Esc):
				return s, func() tea.Msg { return SchemaCheckBackMsg{} }
			case key.Matches(msg, s.keys.Retry):
				s.state = stateSchemaCheckLoading
				s.errorMsg = ""
				return s, tea.Batch(
					s.spinner.Tick,
					fetchSchemaCompareCmd,
				)
			}

		default:
			if key.Matches(msg, s.keys.Esc) {
				return s, func() tea.Msg { return SchemaCheckBackMsg{} }
			}
		}
	}

	if s.state == stateSchemaCheckSavePath {
		s.textInput, cmd = s.textInput.Update(msg)
		return s, cmd
	}

	return s, nil
}

func (s SchemaCheck) View() string {
	var content string

	switch s.state {
	case stateSchemaCheckLoading:
		content = fmt.Sprintf("\n %s Checking Schema Updates...\n\n Press Esc to cancel", s.spinner.View())

	case stateSchemaCheckResult:
		content = successStyle.Render("\n ✓ Schemas are identical\n\n No updates required.")
		content += "\n\n Press Esc to go back"

	case stateSchemaCheckSavePath:
		content = warningStyle.Render("\n ⚠ Differences found between embedded and upstream schema")
		content += "\n\n Enter a folder path to save the diff:\n\n"
		content += " " + s.textInput.View()
		if s.errorMsg != "" {
			content += "\n\n " + errorStyle.Render(s.errorMsg)
		}
		content += "\n\n [enter] save  [esc] cancel"

	case stateSchemaCheckSaved:
		content = successStyle.Render("\n ✓ Diff saved successfully!")
		content += fmt.Sprintf("\n\n Saved to: %s", s.savedPath)
		content += "\n\n Press Esc to go back"

	case stateSchemaCheckError:
		content = errorStyle.Render(fmt.Sprintf("\n Error: %s", s.errorMsg))
		content += "\n\n [r] retry  [esc] back"
	}
	if layout.IsShort(s.height) {
		content = strings.TrimLeft(content, "\n")
	}

	title := titleStyle.Render("Schema Update Check")
	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		content,
	)
}
