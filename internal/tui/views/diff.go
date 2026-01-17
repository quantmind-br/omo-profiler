package views

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/diff"
	"github.com/diogenes/omo-profiler/internal/profile"
)

var (
	diffPurple  = lipgloss.Color("#7D56F4")
	diffMagenta = lipgloss.Color("#FF6AC1")
	diffGreen   = lipgloss.Color("#A6E3A1")
	diffRed     = lipgloss.Color("#F38BA8")
	diffGray    = lipgloss.Color("#6C7086")
	diffWhite   = lipgloss.Color("#CDD6F4")
)

var (
	diffTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(diffPurple)
	diffSubtitleStyle = lipgloss.NewStyle().Foreground(diffGray)
	diffErrorStyle    = lipgloss.NewStyle().Foreground(diffRed)
	diffHelpStyle     = lipgloss.NewStyle().Foreground(diffGray)
	diffActiveStyle   = lipgloss.NewStyle().Bold(true).Foreground(diffWhite).Background(diffPurple)
	diffInactiveStyle = lipgloss.NewStyle().Foreground(diffWhite)
	diffAccentStyle   = lipgloss.NewStyle().Foreground(diffMagenta)
	addedStyle        = lipgloss.NewStyle().Foreground(diffGreen)
	removedStyle      = lipgloss.NewStyle().Foreground(diffRed)
	equalStyle        = lipgloss.NewStyle().Foreground(diffWhite)
)

type focusedPane int

const (
	focusLeft focusedPane = iota
	focusRight
)

type Diff struct {
	width         int
	height        int
	ready         bool
	leftViewport  viewport.Model
	rightViewport viewport.Model

	profiles     []string
	leftProfile  string
	rightProfile string
	leftIdx      int
	rightIdx     int

	focused        focusedPane
	selectingLeft  bool
	selectingRight bool

	diffResult *diff.DiffResult
	err        error
}

func NewDiff() Diff {
	return Diff{
		focused: focusLeft,
	}
}

func (d Diff) Init() tea.Cmd {
	return d.loadProfiles
}

type diffProfilesLoadedMsg struct {
	profiles []string
	err      error
}

type diffComputedMsg struct {
	result *diff.DiffResult
	err    error
}

func (d Diff) loadProfiles() tea.Msg {
	profiles, err := profile.List()
	return diffProfilesLoadedMsg{profiles: profiles, err: err}
}

func (d Diff) computeDiff() tea.Msg {
	if d.leftProfile == "" || d.rightProfile == "" {
		return diffComputedMsg{result: nil, err: nil}
	}

	left, err := profile.Load(d.leftProfile)
	if err != nil {
		return diffComputedMsg{err: fmt.Errorf("loading left profile: %w", err)}
	}

	right, err := profile.Load(d.rightProfile)
	if err != nil {
		return diffComputedMsg{err: fmt.Errorf("loading right profile: %w", err)}
	}

	json1, err := json.MarshalIndent(left.Config, "", "  ")
	if err != nil {
		return diffComputedMsg{err: fmt.Errorf("marshaling left profile: %w", err)}
	}

	json2, err := json.MarshalIndent(right.Config, "", "  ")
	if err != nil {
		return diffComputedMsg{err: fmt.Errorf("marshaling right profile: %w", err)}
	}

	result, err := diff.ComputeDiff(json1, json2)
	return diffComputedMsg{result: result, err: err}
}

func (d Diff) Update(msg tea.Msg) (Diff, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
		d.initViewports()
		d.updateViewportContent()

	case diffProfilesLoadedMsg:
		d.profiles = msg.profiles
		d.err = msg.err
		if len(d.profiles) >= 2 {
			d.leftProfile = d.profiles[0]
			d.rightProfile = d.profiles[1]
			d.leftIdx = 0
			d.rightIdx = 1
			return d, d.computeDiff
		}

	case diffComputedMsg:
		d.diffResult = msg.result
		d.err = msg.err
		d.updateViewportContent()

	case tea.KeyMsg:
		if d.selectingLeft || d.selectingRight {
			return d.handleSelectionKeys(msg)
		}
		cmd := d.handleNavigationKeys(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if d.ready {
		var cmd tea.Cmd
		d.leftViewport, cmd = d.leftViewport.Update(msg)
		cmds = append(cmds, cmd)
		d.rightViewport, cmd = d.rightViewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return d, tea.Batch(cmds...)
}

func (d *Diff) handleNavigationKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "up", "k":
		d.scrollBoth(-1)
	case "down", "j":
		d.scrollBoth(1)
	case "tab":
		if d.focused == focusLeft {
			d.focused = focusRight
		} else {
			d.focused = focusLeft
		}
	case "enter":
		if d.focused == focusLeft {
			d.selectingLeft = true
		} else {
			d.selectingRight = true
		}
	case "pgup":
		d.scrollBoth(-d.leftViewport.Height)
	case "pgdown":
		d.scrollBoth(d.leftViewport.Height)
	}
	return nil
}

func (d *Diff) handleSelectionKeys(msg tea.KeyMsg) (Diff, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if d.selectingLeft && d.leftIdx > 0 {
			d.leftIdx--
		} else if d.selectingRight && d.rightIdx > 0 {
			d.rightIdx--
		}
	case "down", "j":
		if d.selectingLeft && d.leftIdx < len(d.profiles)-1 {
			d.leftIdx++
		} else if d.selectingRight && d.rightIdx < len(d.profiles)-1 {
			d.rightIdx++
		}
	case "enter":
		if d.selectingLeft {
			d.leftProfile = d.profiles[d.leftIdx]
			d.selectingLeft = false
		} else {
			d.rightProfile = d.profiles[d.rightIdx]
			d.selectingRight = false
		}
		return *d, d.computeDiff
	case "esc":
		d.selectingLeft = false
		d.selectingRight = false
	}
	return *d, nil
}

func (d *Diff) scrollBoth(delta int) {
	if !d.ready {
		return
	}
	newOffset := d.leftViewport.YOffset + delta
	if newOffset < 0 {
		newOffset = 0
	}
	d.leftViewport.SetYOffset(newOffset)
	d.rightViewport.SetYOffset(newOffset)
}

func (d *Diff) initViewports() {
	if d.width == 0 || d.height == 0 {
		return
	}

	paneWidth := (d.width - 3) / 2
	paneHeight := d.height - 8

	if paneHeight < 1 {
		paneHeight = 1
	}

	if !d.ready {
		d.leftViewport = viewport.New(paneWidth, paneHeight)
		d.rightViewport = viewport.New(paneWidth, paneHeight)
		d.ready = true
	} else {
		d.leftViewport.Width = paneWidth
		d.leftViewport.Height = paneHeight
		d.rightViewport.Width = paneWidth
		d.rightViewport.Height = paneHeight
	}
}

func (d *Diff) updateViewportContent() {
	if !d.ready || d.diffResult == nil {
		return
	}

	leftContent := d.renderDiffPane(d.diffResult.Left, true)
	rightContent := d.renderDiffPane(d.diffResult.Right, false)

	d.leftViewport.SetContent(leftContent)
	d.rightViewport.SetContent(rightContent)
}

func (d Diff) renderDiffPane(lines []diff.DiffLine, isLeft bool) string {
	var sb strings.Builder

	for _, line := range lines {
		var lineStyle lipgloss.Style
		prefix := "  "

		switch line.Type {
		case diff.DiffEqual:
			lineStyle = equalStyle
		case diff.DiffAdded:
			lineStyle = addedStyle
			if !isLeft {
				prefix = "+ "
			}
		case diff.DiffRemoved:
			lineStyle = removedStyle
			if isLeft {
				prefix = "- "
			}
		}

		sb.WriteString(lineStyle.Render(prefix + line.Text))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (d Diff) View() string {
	if d.err != nil {
		return diffErrorStyle.Render(fmt.Sprintf("Error: %v", d.err))
	}

	if len(d.profiles) == 0 {
		return diffSubtitleStyle.Render("No profiles available for comparison")
	}

	if len(d.profiles) < 2 {
		return diffSubtitleStyle.Render("Need at least 2 profiles for comparison")
	}

	title := diffTitleStyle.Render("Profile Diff")

	leftSelector := d.renderSelector(d.leftProfile, d.selectingLeft, d.leftIdx, d.focused == focusLeft)
	rightSelector := d.renderSelector(d.rightProfile, d.selectingRight, d.rightIdx, d.focused == focusRight)

	selectors := lipgloss.JoinHorizontal(lipgloss.Top, leftSelector, " ", rightSelector)

	var content string
	if d.ready && d.diffResult != nil {
		paneWidth := (d.width - 3) / 2

		leftBorder := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(d.borderColor(focusLeft)).
			Width(paneWidth)

		rightBorder := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(d.borderColor(focusRight)).
			Width(paneWidth)

		leftPane := leftBorder.Render(d.leftViewport.View())
		rightPane := rightBorder.Render(d.rightViewport.View())

		content = lipgloss.JoinHorizontal(lipgloss.Top, leftPane, " ", rightPane)
	} else if d.diffResult == nil && d.leftProfile != "" && d.rightProfile != "" {
		content = diffSubtitleStyle.Render("Computing diff...")
	} else {
		content = diffSubtitleStyle.Render("Select profiles to compare")
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		selectors,
		"",
		content,
	)
}

func (d Diff) borderColor(pane focusedPane) lipgloss.Color {
	if d.focused == pane {
		return diffPurple
	}
	return diffGray
}

func (d Diff) renderSelector(selected string, isSelecting bool, idx int, isFocused bool) string {
	paneWidth := (d.width - 3) / 2
	if paneWidth < 20 {
		paneWidth = 20
	}

	var content string
	if isSelecting {
		var items []string
		for i, p := range d.profiles {
			cursor := "  "
			if i == idx {
				cursor = "> "
			}
			style := diffInactiveStyle
			if i == idx {
				style = diffActiveStyle
			}
			items = append(items, style.Render(cursor+p))
		}
		content = strings.Join(items, "\n")
	} else {
		content = diffAccentStyle.Render(selected)
	}

	borderColor := diffGray
	if isFocused {
		borderColor = diffPurple
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(paneWidth).
		Render(content)
}

func (d Diff) ShouldReturn() bool {
	return false
}
