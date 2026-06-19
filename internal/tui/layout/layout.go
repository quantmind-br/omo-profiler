package layout

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// RenderConfirmDialog renders a confirmation dialog with the given target and message.
func RenderConfirmDialog(target, message string) string {
	confirmStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#F9E2AF")).
		Background(lipgloss.Color("#6C7086")).
		Padding(0, 1)
	return confirmStyle.Render(fmt.Sprintf("%s '%s'? [y/n]", message, target))
}

const (
	MinTerminalWidth = 40
	// MinTerminalHeight must fit the dashboard's compact layout without
	// overlap: 3 header lines + 9 menu items + 1 help bar = 13 rows. Use 14
	// for a one-row safety margin. Below this the "Too Small" guard renders
	// instead of a broken, overlapping screen.
	MinTerminalHeight = 14
	MaxFieldWidth     = 120

	// ViewportOverhead constants for calculating viewport height
	// Normal mode: title + help + 2 spacing lines
	ViewportOverheadNormal = 4
	// Short/compact mode: title + help only
	ViewportOverheadShort = 2
)

func FixedSmallWidth() int {
	return 10
}

func MediumFieldWidth(availableWidth int) int {
	w := int(float64(availableWidth) * 0.4)
	if w < 10 {
		w = 10
	}
	if w > MaxFieldWidth {
		w = MaxFieldWidth
	}
	return w
}

func WideFieldWidth(availableWidth, padding int) int {
	w := availableWidth - padding
	if w < 10 {
		w = 10
	}
	if w > MaxFieldWidth {
		w = MaxFieldWidth
	}
	return w
}

func TruncateWithEllipsis(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if runewidth.StringWidth(text) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return runewidth.Truncate(text, maxWidth, "")
	}
	return runewidth.Truncate(text, maxWidth, "...")
}

func IsBelowMinimumSize(width, height int) bool {
	return width < MinTerminalWidth || height < MinTerminalHeight
}

// IsCompact returns true when width is below the comfortable threshold.
// Views should use simpler layouts at compact widths.
func IsCompact(width int) bool {
	return width < 60
}

// IsShort returns true when height is below the comfortable threshold.
// Views should reduce vertical spacing at short heights.
func IsShort(height int) bool {
	return height < 20
}

// HelpBarHeight returns the number of lines to reserve for the help bar.
// At short heights, uses 1 line; at normal heights, uses 2 lines.
func HelpBarHeight(height int) int {
	if height < 16 {
		return 1
	}
	return 2
}

// RenderHintLine joins hint segments with a separator, truncating with
// ellipsis if the total width exceeds the available space. Important bindings
// at the end of the slice are preserved.
func RenderHintLine(hints []string, width int) string {
	sep := "  "
	if len(hints) == 0 {
		return ""
	}
	if width <= 0 {
		return strings.Join(hints, sep)
	}
	if width < 60 {
		sep = " "
	}
	joined := strings.Join(hints, sep)
	if runewidth.StringWidth(joined) <= width {
		return joined
	}
	// Keep first and last hint, ellipsis in the middle
	if len(hints) >= 3 {
		first := hints[0]
		last := hints[len(hints)-1]
		ellipsis := " ... "
		max := width - runewidth.StringWidth(first+sep+last+ellipsis)
		if max > 3 {
			mid := strings.Join(hints[1:len(hints)-1], sep)
			mid = runewidth.Truncate(mid, max, "")
			return first + sep + mid + ellipsis + last
		}
		return first + sep + "..." + sep + last
	}
	return TruncateWithEllipsis(joined, width)
}
