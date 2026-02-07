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

func TestMediumFieldWidth(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{name: "80 columns", in: 80, want: 32},
		{name: "120 columns", in: 120, want: 48},
		{name: "200 columns", in: 200, want: 80},
		{name: "300 columns capped", in: 300, want: 120},
		{name: "narrow clamped", in: 40, want: 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MediumFieldWidth(tt.in); got != tt.want {
				t.Fatalf("MediumFieldWidth(%d) = %d, want %d", tt.in, got, tt.want)
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
		{name: "clamped minimum", available: 25, padding: 10, want: 20},
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
		{name: "exact minimum", width: 60, height: 25, want: false},
		{name: "below width", width: 59, height: 25, want: true},
		{name: "below height", width: 60, height: 24, want: true},
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
