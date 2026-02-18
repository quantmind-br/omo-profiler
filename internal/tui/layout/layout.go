package layout

import "github.com/mattn/go-runewidth"

const (
	MinTerminalWidth  = 40
	MinTerminalHeight = 12
	MaxFieldWidth     = 120
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
