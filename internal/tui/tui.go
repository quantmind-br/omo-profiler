package tui

import tea "github.com/charmbracelet/bubbletea"

func Run() error {
	app := NewApp()
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithoutSignalHandler())
	_, err := p.Run()
	return err
}
