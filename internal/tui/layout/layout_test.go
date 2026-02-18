package layout

import (
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestFixedSmallWidth(t *testing.T) {
	if got := FixedSmallWidth(); got != 10 {
		t.Fatalf("FixedSmallWidth() = %d, want 10", got)
	}
}

func TestIsCompact(t *testing.T) {
	tests := []struct {
		name  string
		width int
		want  bool
	}{
		{name: "compact at 40", width: 40, want: true},
		{name: "compact at 59", width: 59, want: true},
		{name: "not compact at 60", width: 60, want: false},
		{name: "not compact at 80", width: 80, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCompact(tt.width); got != tt.want {
				t.Fatalf("IsCompact(%d) = %v, want %v", tt.width, got, tt.want)
			}
		})
	}
}

func TestIsShort(t *testing.T) {
	tests := []struct {
		name   string
		height int
		want   bool
	}{
		{name: "short at 12", height: 12, want: true},
		{name: "short at 19", height: 19, want: true},
		{name: "not short at 20", height: 20, want: false},
		{name: "not short at 25", height: 25, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsShort(tt.height); got != tt.want {
				t.Fatalf("IsShort(%d) = %v, want %v", tt.height, got, tt.want)
			}
		})
	}
}

func TestHelpBarHeight(t *testing.T) {
	tests := []struct {
		name   string
		height int
		want   int
	}{
		{name: "tiny returns 1", height: 12, want: 1},
		{name: "at 15 returns 1", height: 15, want: 1},
		{name: "at 16 returns 2", height: 16, want: 2},
		{name: "normal returns 2", height: 25, want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HelpBarHeight(tt.height); got != tt.want {
				t.Fatalf("HelpBarHeight(%d) = %d, want %d", tt.height, got, tt.want)
			}
		})
	}
}

func TestWideFieldWidth(t *testing.T) {
	tests := []struct {
		name      string
		available int
		padding   int
		want      int
	}{
		{name: "standard width", available: 80, padding: 10, want: 70},
		{name: "larger width", available: 120, padding: 10, want: 110},
		{name: "capped at max", available: 200, padding: 10, want: 120},
		{name: "clamped minimum", available: 25, padding: 10, want: 15},
		{name: "very narrow clamped", available: 15, padding: 10, want: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WideFieldWidth(tt.available, tt.padding); got != tt.want {
				t.Fatalf("WideFieldWidth(%d, %d) = %d, want %d", tt.available, tt.padding, got, tt.want)
			}
		})
	}
}

func TestTruncateWithEllipsis(t *testing.T) {
	t.Run("short text unchanged", func(t *testing.T) {
		text := "short"
		if got := TruncateWithEllipsis(text, 10); got != text {
			t.Fatalf("TruncateWithEllipsis(%q, 10) = %q, want %q", text, got, text)
		}
	})

	t.Run("overflow uses ellipsis", func(t *testing.T) {
		got := TruncateWithEllipsis("abcdefghij", 7)
		if got != "abcd..." {
			t.Fatalf("TruncateWithEllipsis overflow = %q, want %q", got, "abcd...")
		}
	})

	t.Run("non-positive width returns empty", func(t *testing.T) {
		if got := TruncateWithEllipsis("abcdef", 0); got != "" {
			t.Fatalf("TruncateWithEllipsis(..., 0) = %q, want empty", got)
		}
	})

	t.Run("very small width has no ellipsis suffix", func(t *testing.T) {
		got := TruncateWithEllipsis("abcdef", 3)
		if got != "abc" {
			t.Fatalf("TruncateWithEllipsis(..., 3) = %q, want %q", got, "abc")
		}
	})

	t.Run("unicode width respected", func(t *testing.T) {
		text := "こんにちは世界"
		got := TruncateWithEllipsis(text, 8)
		if runewidth.StringWidth(got) > 8 {
			t.Fatalf("TruncateWithEllipsis unicode width = %d, want <= 8 (value=%q)", runewidth.StringWidth(got), got)
		}
		if got == text {
			t.Fatalf("TruncateWithEllipsis unicode should truncate, got unchanged %q", got)
		}
	})
}

func TestIsBelowMinimumSize(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		want   bool
	}{
		{name: "exact minimum", width: 40, height: 12, want: false},
		{name: "below width", width: 39, height: 12, want: true},
		{name: "below height", width: 40, height: 11, want: true},
		{name: "above minimum", width: 120, height: 40, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBelowMinimumSize(tt.width, tt.height); got != tt.want {
				t.Fatalf("IsBelowMinimumSize(%d, %d) = %v, want %v", tt.width, tt.height, got, tt.want)
			}
		})
	}
}
