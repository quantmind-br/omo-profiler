package tui

import (
	"strings"
	"testing"
)

func TestLayoutFixedSmallWidth(t *testing.T) {
	got := FixedSmallWidth()
	if got != 10 {
		t.Errorf("FixedSmallWidth() = %d, want 10", got)
	}
}

func TestLayoutMediumFieldWidth(t *testing.T) {
	tests := []struct {
		name           string
		availableWidth int
		want           int
	}{
		{"standard 80", 80, 32},      // 80 * 0.4 = 32
		{"wide 120", 120, 48},        // 120 * 0.4 = 48
		{"very wide 200", 200, 80},   // 200 * 0.4 = 80
		{"ultra wide 300", 300, 120}, // 300 * 0.4 = 120, capped at MaxFieldWidth
		{"very large 500", 500, 120}, // 500 * 0.4 = 200, capped at 120
		{"narrow 50", 50, 20},        // 50 * 0.4 = 20
		{"narrow 40", 40, 16},        // 40 * 0.4 = 16
		{"very narrow 20", 20, 10},   // 20 * 0.4 = 8, clamped to 10
		{"zero width", 0, 10},        // 0 * 0.4 = 0, clamped to 10
		{"negative width", -10, 10},  // negative clamped to 10
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MediumFieldWidth(tt.availableWidth)
			if got != tt.want {
				t.Errorf("MediumFieldWidth(%d) = %d, want %d", tt.availableWidth, got, tt.want)
			}
		})
	}
}

func TestLayoutWideFieldWidth(t *testing.T) {
	tests := []struct {
		name           string
		availableWidth int
		padding        int
		want           int
	}{
		{"standard 80 with padding 10", 80, 10, 70},      // 80 - 10 = 70
		{"wide 120 with padding 10", 120, 10, 110},       // 120 - 10 = 110
		{"very wide 200 with padding 10", 200, 10, 120},  // 200 - 10 = 190, capped at 120
		{"ultra wide 300 with padding 10", 300, 10, 120}, // capped at MaxFieldWidth
		{"narrow 30 with padding 10", 30, 10, 20},        // 30 - 10 = 20
		{"narrow 25 with padding 10", 25, 10, 15},        // 25 - 10 = 15
		{"very narrow 15 with padding 10", 15, 10, 10},   // 15 - 10 = 5, clamped to 10
		{"zero width", 0, 10, 10},                        // 0 - 10 = -10, clamped to 10
		{"negative width", -10, 10, 10},                  // clamped to 10
		{"zero padding", 100, 0, 100},                    // 100 - 0 = 100
		{"large padding", 50, 40, 10},                    // 50 - 40 = 10
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WideFieldWidth(tt.availableWidth, tt.padding)
			if got != tt.want {
				t.Errorf("WideFieldWidth(%d, %d) = %d, want %d", tt.availableWidth, tt.padding, got, tt.want)
			}
		})
	}
}

func TestLayoutTruncateWithEllipsis(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxWidth int
		want     string
	}{
		{"short text no truncation", "Short", 10, "Short"},
		{"exact fit", "Hello", 5, "Hello"},
		{"overflow with ellipsis", "Hello World", 8, "Hello..."},
		{"empty string", "", 10, ""},
		{"zero width", "Hello", 0, ""},
		{"negative width", "Hello", -5, ""},
		{"width 1", "Hello", 1, "H"},
		{"width 2", "Hello", 2, "He"},
		{"width 3", "Hello", 3, "Hel"},
		{"width 4 truncates", "Hello", 4, "H..."},
		{"unicode Japanese", "日本語テスト", 10, "日本語..."},
		{"unicode short", "日本", 10, "日本"},
		{"unicode exact", "日本語", 6, "日本語"}, // Each character is 2 wide
		{"unicode overflow", "日本語テスト", 8, "日本..."},
		{"mixed unicode and ASCII", "Hello日本", 10, "Hello日本"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateWithEllipsis(tt.text, tt.maxWidth)
			if got != tt.want {
				t.Errorf("TruncateWithEllipsis(%q, %d) = %q, want %q", tt.text, tt.maxWidth, got, tt.want)
			}
		})
	}
}

func TestLayoutIsBelowMinimumSize(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		want   bool
	}{
		{"below width 39x12", 39, 12, true},
		{"below height 40x11", 40, 11, true},
		{"at minimum 40x12", 40, 12, false},
		{"above minimum 41x13", 41, 13, false},
		{"old minimum still passes 60x25", 60, 25, false},
		{"width zero", 0, 12, true},
		{"height zero", 80, 0, true},
		{"both zero", 0, 0, true},
		{"large terminal", 200, 50, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBelowMinimumSize(tt.width, tt.height)
			if got != tt.want {
				t.Errorf("IsBelowMinimumSize(%d, %d) = %v, want %v", tt.width, tt.height, got, tt.want)
			}
		})
	}
}

func TestLayoutRenderMinimumSizeWarning(t *testing.T) {
	result := RenderMinimumSizeWarning(35, 10)

	if result == "" {
		t.Error("RenderMinimumSizeWarning() returned empty string")
	}

	if !strings.Contains(result, "40") {
		t.Errorf("RenderMinimumSizeWarning() should contain '40' (min width), got: %q", result)
	}
	if !strings.Contains(result, "12") {
		t.Errorf("RenderMinimumSizeWarning() should contain '12' (min height), got: %q", result)
	}

	if !strings.Contains(result, "35") {
		t.Errorf("RenderMinimumSizeWarning() should contain '35' (current width), got: %q", result)
	}
	if !strings.Contains(result, "10") {
		t.Errorf("RenderMinimumSizeWarning() should contain '10' (current height), got: %q", result)
	}
}

func TestLayoutIsCompact(t *testing.T) {
	if !IsCompact(40) {
		t.Error("IsCompact(40) should be true")
	}
	if IsCompact(60) {
		t.Error("IsCompact(60) should be false")
	}
}

func TestLayoutIsShort(t *testing.T) {
	if !IsShort(12) {
		t.Error("IsShort(12) should be true")
	}
	if IsShort(20) {
		t.Error("IsShort(20) should be false")
	}
}

func TestLayoutHelpBarHeight(t *testing.T) {
	if got := HelpBarHeight(12); got != 1 {
		t.Errorf("HelpBarHeight(12) = %d, want 1", got)
	}
	if got := HelpBarHeight(16); got != 2 {
		t.Errorf("HelpBarHeight(16) = %d, want 2", got)
	}
}
