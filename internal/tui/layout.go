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

func IsCompact(width int) bool {
	return layout.IsCompact(width)
}

func IsShort(height int) bool {
	return layout.IsShort(height)
}

func HelpBarHeight(height int) int {
	return layout.HelpBarHeight(height)
}

func RenderMinimumSizeWarning(width, height int) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(Yellow).Render("⚠ Too Small")
	body := fmt.Sprintf(
		"Need: %dx%d  Now: %dx%d",
		layout.MinTerminalWidth, layout.MinTerminalHeight, width, height,
	)
	dim := lipgloss.NewStyle().Foreground(Gray)
	quit := dim.Render("Resize or q to quit")

	content := lipgloss.JoinVertical(lipgloss.Center, title, body, quit)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
