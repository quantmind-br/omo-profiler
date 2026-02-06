package views

import (
	"strings"
	"testing"
)

func TestParseMapStringInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]int
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single pair",
			input:    "key:1",
			expected: map[string]int{"key": 1},
		},
		{
			name:     "multiple pairs",
			input:    "key1:1, key2:2, key3:3",
			expected: map[string]int{"key1": 1, "key2": 2, "key3": 3},
		},
		{
			name:     "with spaces",
			input:    "  key1 : 1 , key2 : 2  ",
			expected: map[string]int{"key1": 1, "key2": 2},
		},
		{
			name:     "ignores invalid pairs - no colon",
			input:    "key1, key2:2",
			expected: map[string]int{"key2": 2},
		},
		{
			name:     "ignores invalid pairs - empty key",
			input:    ":1, key2:2",
			expected: map[string]int{"key2": 2},
		},
		{
			name:     "ignores invalid pairs - non-numeric value",
			input:    "key1:abc, key2:2",
			expected: map[string]int{"key2": 2},
		},
		{
			name:     "zero value",
			input:    "key:0",
			expected: map[string]int{"key": 0},
		},
		{
			name:     "negative values",
			input:    "key1:-1, key2:-10",
			expected: map[string]int{"key1": -1, "key2": -10},
		},
		{
			name:     "only invalid pairs returns nil",
			input:    "abc, :1, key:abc",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseMapStringInt(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("expected %v, got nil", tt.expected)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d items, got %d", len(tt.expected), len(result))
			}

			for k, expectedVal := range tt.expected {
				if val, ok := result[k]; !ok {
					t.Errorf("missing key %q in result", k)
				} else if val != expectedVal {
					t.Errorf("key %q: expected %d, got %d", k, expectedVal, val)
				}
			}
		})
	}
}

func TestSerializeMapStringInt(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]int
		expected string
		contains []string
	}{
		{
			name:     "nil map",
			input:    nil,
			expected: "",
		},
		{
			name:     "empty map",
			input:    map[string]int{},
			expected: "",
		},
		{
			name:     "single item",
			input:    map[string]int{"key": 1},
			expected: "key:1",
		},
		{
			name:     "multiple items",
			input:    map[string]int{"key1": 1, "key2": 2, "key3": 3},
			contains: []string{"key1:1", "key2:2", "key3:3"},
		},
		{
			name:     "zero value",
			input:    map[string]int{"key": 0},
			expected: "key:0",
		},
		{
			name:     "negative values",
			input:    map[string]int{"key1": -1, "key2": -10},
			contains: []string{"key1:-1", "key2:-10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := serializeMapStringInt(tt.input)

			// For exact match tests (nil, empty, single item)
			if tt.expected != "" {
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
				return
			}

			// For empty/nil maps, just check result is empty
			if len(tt.input) == 0 {
				if result != "" {
					t.Errorf("expected empty string, got %q", result)
				}
				return
			}

			// For contains tests (multiple items - map order is not guaranteed)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected result to contain %q, got %q", expected, result)
				}
			}

			// Check comma count (n items = n-1 commas), skip for nil/empty
			if len(tt.input) > 0 {
				expectedCommas := len(tt.input) - 1
				commaCount := strings.Count(result, ", ")
				if commaCount != expectedCommas {
					t.Errorf("expected %d commas, got %d in result %q", expectedCommas, commaCount, result)
				}
			}
		})
	}
}

func TestParseSerializeMapStringIntRoundTrip(t *testing.T) {
	original := map[string]int{
		"key1": 1,
		"key2": 2,
		"key3": 3,
	}

	serialized := serializeMapStringInt(original)
	parsed := parseMapStringInt(serialized)

	if len(parsed) != len(original) {
		t.Errorf("round trip: expected %d items, got %d", len(original), len(parsed))
	}

	for k, v := range original {
		if parsed[k] != v {
			t.Errorf("round trip: key %q: expected %d, got %d", k, v, parsed[k])
		}
	}
}
