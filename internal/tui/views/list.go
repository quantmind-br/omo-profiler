package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/profile"
)

type SwitchProfileMsg struct{ Name string }
type EditProfileMsg struct{ Name string }
type DeleteProfileMsg struct{ Name string }
type ConfirmDeleteMsg struct{ Confirmed bool }
type NavigateToWizardMsg struct{}
type NavigateToDashboardMsg struct{}

type profileItem struct {
	name     string
	isActive bool
}

func (i profileItem) Title() string {
	if i.isActive {
		return "* " + i.name + " (active)"
	}
	return "  " + i.name
}

func (i profileItem) Description() string {
	if i.isActive {
		return "Currently active profile"
	}
	return "Press enter to switch"
}

func (i profileItem) FilterValue() string {
	return i.name
}

type listKeyMap struct {
	Switch key.Binding
	Edit   key.Binding
	Delete key.Binding
	New    key.Binding
	Search key.Binding
	Back   key.Binding
}

func newListKeyMap() listKeyMap {
	return listKeyMap{
		Switch: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "switch profile"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new profile"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}

type List struct {
	list             list.Model
	keys             listKeyMap
	width            int
	height           int
	confirmingDelete bool
	deleteTarget     string
	err              error
}

func NewList() List {
	keys := newListKeyMap()

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#7D56F4")).
		BorderForeground(lipgloss.Color("#7D56F4"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#6C7086"))

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Profiles"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	return List{
		list: l,
		keys: keys,
	}
}

func (l List) Init() tea.Cmd {
	return l.loadProfiles
}

func (l List) loadProfiles() tea.Msg {
	return listProfilesLoadedMsg{}
}

type listProfilesLoadedMsg struct{}

func (l *List) LoadProfiles() error {
	names, err := profile.List()
	if err != nil {
		l.err = err
		return err
	}

	active, err := profile.GetActive()
	if err != nil {
		l.err = err
		return err
	}

	items := make([]list.Item, len(names))
	for i, name := range names {
		isActive := active != nil && !active.IsOrphan && active.ProfileName == name
		items[i] = profileItem{
			name:     name,
			isActive: isActive,
		}
	}

	l.list.SetItems(items)
	l.err = nil
	return nil
}

func (l *List) SetSize(width, height int) {
	l.width = width
	l.height = height
	l.list.SetSize(width, height-4)
}

func (l List) Update(msg tea.Msg) (List, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case listProfilesLoadedMsg:
		_ = l.LoadProfiles()
		return l, nil

	case tea.WindowSizeMsg:
		l.SetSize(msg.Width, msg.Height)
		return l, nil

	case tea.KeyMsg:
		if l.confirmingDelete {
			switch msg.String() {
			case "y", "Y":
				l.confirmingDelete = false
				target := l.deleteTarget
				l.deleteTarget = ""
				return l, func() tea.Msg {
					return DeleteProfileMsg{Name: target}
				}
			case "n", "N", "esc":
				l.confirmingDelete = false
				l.deleteTarget = ""
				return l, nil
			}
			return l, nil
		}

		if l.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, l.keys.Back):
			return l, func() tea.Msg {
				return NavigateToDashboardMsg{}
			}

		case key.Matches(msg, l.keys.Switch):
			if item, ok := l.list.SelectedItem().(profileItem); ok {
				if !item.isActive {
					return l, func() tea.Msg {
						return SwitchProfileMsg{Name: item.name}
					}
				}
			}

		case key.Matches(msg, l.keys.Edit):
			if item, ok := l.list.SelectedItem().(profileItem); ok {
				return l, func() tea.Msg {
					return EditProfileMsg{Name: item.name}
				}
			}

		case key.Matches(msg, l.keys.Delete):
			if item, ok := l.list.SelectedItem().(profileItem); ok {
				l.confirmingDelete = true
				l.deleteTarget = item.name
				return l, nil
			}

		case key.Matches(msg, l.keys.New):
			return l, func() tea.Msg {
				return NavigateToWizardMsg{}
			}
		}

	case ConfirmDeleteMsg:
		if msg.Confirmed && l.deleteTarget != "" {
			return l, func() tea.Msg {
				return DeleteProfileMsg{Name: l.deleteTarget}
			}
		}
		l.confirmingDelete = false
		l.deleteTarget = ""
		return l, nil
	}

	l.list, cmd = l.list.Update(msg)
	cmds = append(cmds, cmd)

	return l, tea.Batch(cmds...)
}

func (l List) View() string {
	var content string

	if l.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8"))
		content = errorStyle.Render(fmt.Sprintf("Error: %v", l.err))
	} else {
		content = l.list.View()
	}

	if l.confirmingDelete {
		confirmStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F9E2AF")).
			Background(lipgloss.Color("#45475A")).
			Padding(0, 1)
		confirmText := confirmStyle.Render(fmt.Sprintf("Delete '%s'? [y/n]", l.deleteTarget))
		return lipgloss.JoinVertical(lipgloss.Left, content, "", confirmText)
	}

	return content
}

func (l List) SelectedProfile() string {
	if item, ok := l.list.SelectedItem().(profileItem); ok {
		return item.name
	}
	return ""
}

func (l List) IsConfirmingDelete() bool {
	return l.confirmingDelete
}
