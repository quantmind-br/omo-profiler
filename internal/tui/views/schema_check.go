package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/schema"
)

type schemaCheckState int

const (
	stateSchemaCheckLoading schemaCheckState = iota
	stateSchemaCheckResult
	stateSchemaCheckError
)

type schemaCheckKeyMap struct {
	Esc   key.Binding
	Retry key.Binding
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
	}
}

type SchemaCheck struct {
	state    schemaCheckState
	result   *schema.CompareResult
	spinner  spinner.Model
	errorMsg string
	width    int
	height   int
	keys     schemaCheckKeyMap
}

type NavToSchemaCheckMsg struct{}
type SchemaCheckBackMsg struct{}

type schemaCheckResultMsg struct {
	result *schema.CompareResult
	err    error
}

func NewSchemaCheck() SchemaCheck {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	return SchemaCheck{
		state:   stateSchemaCheckLoading,
		spinner: s,
		keys:    newSchemaCheckKeyMap(),
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
		s.state = stateSchemaCheckResult
		return s, nil

	case spinner.TickMsg:
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keys.Esc):
			return s, func() tea.Msg { return SchemaCheckBackMsg{} }
		case key.Matches(msg, s.keys.Retry):
			if s.state == stateSchemaCheckError {
				s.state = stateSchemaCheckLoading
				s.errorMsg = ""
				return s, tea.Batch(
					s.spinner.Tick,
					fetchSchemaCompareCmd,
				)
			}
		}
	}

	return s, nil
}

func (s SchemaCheck) View() string {
	var content string

	switch s.state {
	case stateSchemaCheckLoading:
		content = fmt.Sprintf("\n %s Checking Schema Updates...\n\n Press Esc to cancel", s.spinner.View())

	case stateSchemaCheckResult:
		if s.result.Identical {
			content = successStyle.Render("\n ✓ Schemas are identical\n\n No updates required.")
		} else {
			content = warningStyle.Render("\n ⚠ Differences found between embedded and upstream schema\n\n Preview:\n")
			content += fmt.Sprintf("\n%s", s.result.Diff)
		}
		content += "\n\n Press Esc to go back"

	case stateSchemaCheckError:
		content = errorStyle.Render(fmt.Sprintf("\n Error: %s", s.errorMsg))
		content += "\n\n [r] retry  [esc] back"
	}

	title := titleStyle.Render("Schema Update Check")
	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		content,
	)
}
