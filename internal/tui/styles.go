package tui

import "github.com/charmbracelet/lipgloss"

var (
	Purple  = lipgloss.Color("#7D56F4")
	Magenta = lipgloss.Color("#FF6AC1")
	Cyan    = lipgloss.Color("#78DCE8")
	Green   = lipgloss.Color("#A6E3A1")
	Red     = lipgloss.Color("#F38BA8")
	Yellow  = lipgloss.Color("#F9E2AF")
	Gray    = lipgloss.Color("#6C7086")
	White   = lipgloss.Color("#CDD6F4")
)

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Purple)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Gray)

	ActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(White).
			Background(Purple)

	InactiveStyle = lipgloss.NewStyle().
			Foreground(White)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Green)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Yellow)

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Purple).
			Padding(1, 2)

	HelpStyle = lipgloss.NewStyle().
			Foreground(Gray)

	AccentStyle = lipgloss.NewStyle().
			Foreground(Magenta)

	CyanAccentStyle = lipgloss.NewStyle().
			Foreground(Cyan)
)
