package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/tui/layout"
)

func FixedSmallWidth() int {
	return layout.FixedSmallWidth()
}

func MediumFieldWidth(availableWidth int) int {
	return layout.MediumFieldWidth(availableWidth)
}

func WideFieldWidth(availableWidth, padding int) int {
	return layout.WideFieldWidth(availableWidth, padding)
}

func TruncateWithEllipsis(text string, maxWidth int) string {
	return layout.TruncateWithEllipsis(text, maxWidth)
}

func IsBelowMinimumSize(width, height int) bool {
	return layout.IsBelowMinimumSize(width, height)
}

func RenderMinimumSizeWarning(width, height int) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(Yellow).Render("⚠  Terminal Too Small")
	body := fmt.Sprintf(
		"Minimum required: %dx%d\nCurrent size:     %dx%d",
		layout.MinTerminalWidth, layout.MinTerminalHeight, width, height,
	)
	dim := lipgloss.NewStyle().Foreground(Gray)
	instruction := dim.Render("Please resize your terminal window.")
	quit := dim.Render("Press q or Ctrl+C to quit.")

	content := lipgloss.JoinVertical(lipgloss.Center, title, "", body, "", instruction, quit)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
