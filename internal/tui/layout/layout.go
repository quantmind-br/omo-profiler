package layout

import "github.com/mattn/go-runewidth"

const (
	MinTerminalWidth  = 80
	MinTerminalHeight = 24
	MaxFieldWidth     = 120
)

func FixedSmallWidth() int {
	return 10
}

func MediumFieldWidth(availableWidth int) int {
	w := int(float64(availableWidth) * 0.4)
	if w < 20 {
		w = 20
	}
	if w > MaxFieldWidth {
		w = MaxFieldWidth
	}
	return w
}

func WideFieldWidth(availableWidth, padding int) int {
	w := availableWidth - padding
	if w < 20 {
		w = 20
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
