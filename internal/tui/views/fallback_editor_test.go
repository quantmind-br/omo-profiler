package views

import (
	"reflect"
	"testing"
)

// TestFallbackEditorLoadValueRoundTrip pins the load→value() serialization
// contract for each accepted fallback_models shape: a single plain entry
// collapses to a bare string, multiple/attributed entries expand to a slice
// (strings for plain rows, {model,variant?,reasoning?} maps otherwise),
// and an empty editor yields nil.
func TestFallbackEditorLoadValueRoundTrip(t *testing.T) {
	cases := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "single plain string collapses to bare string",
			input:    "m1",
			expected: "m1",
		},
		{
			name:     "multiple plain entries stay a slice of strings",
			input:    []interface{}{"m1", "m2"},
			expected: []interface{}{"m1", "m2"},
		},
		{
			name: "attributed entry serializes with canonical reasoning key",
			input: []interface{}{
				map[string]interface{}{"model": "m2", "variant": "medium", "reasoningEffort": "high"},
			},
			expected: []interface{}{
				map[string]interface{}{"model": "m2", "variant": "medium", "reasoning": "high"},
			},
		},
		{
			name: "reasoning key round-trips as-is",
			input: []interface{}{
				map[string]interface{}{"model": "m3", "reasoning": "auto"},
			},
			expected: []interface{}{
				map[string]interface{}{"model": "m3", "reasoning": "auto"},
			},
		},
		{
			name:     "empty input yields nil",
			input:    nil,
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fe := newFallbackEditor()
			fe.load(tc.input)
			got := fe.value()
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("value() = %#v, want %#v", got, tc.expected)
			}
		})
	}
}

// TestFallbackEditorAddOpensModelSelector asserts that both the "+" and the "a"
// alias, at the top level, ask the host to open the model selector and record
// that the pending selection is an append (not a replace).
func TestFallbackEditorAddOpensModelSelector(t *testing.T) {
	for _, k := range []string{"+", "a"} {
		fe := newFallbackEditor()
		fe.open()

		action, _ := fe.handleKey(keyMsg(k))

		if action != fbOpenModelSelector {
			t.Errorf("key %q: action = %v, want fbOpenModelSelector", k, action)
		}
		if !fe.pendingAdd {
			t.Errorf("key %q: expected pendingAdd to be true", k)
		}
	}
}

// TestFallbackEditorApplySelectedModelAppends verifies that a model chosen
// after a pending add is appended as a new entry, focus moves to it, and the
// pending flag is cleared.
func TestFallbackEditorApplySelectedModelAppends(t *testing.T) {
	fe := newFallbackEditor()
	fe.open()
	fe.handleKey(keyMsg("+")) // arm pendingAdd

	before := len(fe.entries)
	fe.applySelectedModel("x/y", "Y")

	if len(fe.entries) != before+1 {
		t.Fatalf("expected entries to grow by 1 (%d -> %d)", before, len(fe.entries))
	}
	last := fe.entries[len(fe.entries)-1]
	if last.model != "x/y" {
		t.Errorf("appended entry model = %q, want %q", last.model, "x/y")
	}
	if last.modelDisplay != "Y" {
		t.Errorf("appended entry modelDisplay = %q, want %q", last.modelDisplay, "Y")
	}
	if fe.pendingAdd {
		t.Error("expected pendingAdd cleared after applySelectedModel")
	}
	if fe.focusedIdx != len(fe.entries)-1 {
		t.Errorf("focusedIdx = %d, want %d", fe.focusedIdx, len(fe.entries)-1)
	}
}

// TestFallbackEditorDeleteAtTopLevel loads a two-entry value, deletes the
// focused plain entry, and asserts the surviving attributed entry serializes
// as an attributed model map inside the slice.
func TestFallbackEditorDeleteAtTopLevel(t *testing.T) {
	fe := newFallbackEditor()
	fe.load([]interface{}{
		"m1",
		map[string]interface{}{"model": "m2", "variant": "medium"},
	})
	fe.open()
	fe.focusedIdx = 0

	fe.handleKey(keyMsg("-"))

	if len(fe.entries) != 1 {
		t.Fatalf("expected 1 entry after delete, got %d", len(fe.entries))
	}
	got := fe.value()
	want := []interface{}{
		map[string]interface{}{"model": "m2", "variant": "medium"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("value() = %#v, want %#v", got, want)
	}
}

// TestFallbackEditorExpandCollapse verifies enter expands the focused entry and
// esc while expanded collapses it WITHOUT closing the overlay.
func TestFallbackEditorExpandCollapse(t *testing.T) {
	fe := newFallbackEditor()
	fe.load([]interface{}{"m1", "m2"})
	fe.open()
	fe.focusedIdx = 1

	action, _ := fe.handleKey(keyMsg("enter"))
	if action != fbChanged {
		t.Errorf("expand action = %v, want fbChanged", action)
	}
	if fe.expandedIdx != 1 {
		t.Errorf("expandedIdx = %d, want 1", fe.expandedIdx)
	}

	fe.handleKey(keyMsg("esc"))
	if fe.expandedIdx != -1 {
		t.Errorf("expandedIdx after collapse = %d, want -1", fe.expandedIdx)
	}
	if !fe.active {
		t.Error("esc while expanded must collapse, not close the overlay")
	}
}

// TestFallbackEditorReasoningCycle verifies enter on the reasoning sub-field
// advances through effortLevels (canonical: off, not none).
func TestFallbackEditorReasoningCycle(t *testing.T) {
	fe := newFallbackEditor()
	fe.load("m1")
	fe.open()
	fe.handleKey(keyMsg("enter")) // expand entry 0
	fe.subField = 2               // reasoning

	fe.handleKey(keyMsg("enter"))
	if got := fe.entries[0].reasoningEffort; got != "off" {
		t.Errorf("reasoningEffort after 1 cycle = %q, want %q", got, "off")
	}

	fe.handleKey(keyMsg("enter"))
	if got := fe.entries[0].reasoningEffort; got != "minimal" {
		t.Errorf("reasoningEffort after 2 cycles = %q, want %q", got, "minimal")
	}
}

// TestFallbackEditorVariantTextEdit verifies enter on the variant sub-field
// enters text-editing mode (returning a blink cmd), keeps capturing runes, and
// commits the typed value to entry.variant on the next enter.
func TestFallbackEditorVariantTextEdit(t *testing.T) {
	fe := newFallbackEditor()
	fe.load("m1")
	fe.open()
	fe.handleKey(keyMsg("enter")) // expand entry 0
	fe.subField = 1               // variant

	action, cmd := fe.handleKey(keyMsg("enter"))
	if !fe.editingText {
		t.Fatal("expected editingText true after entering variant edit")
	}
	if cmd == nil {
		t.Error("expected a non-nil cmd (textinput.Blink) when entering text edit")
	}
	if action != fbChanged {
		t.Errorf("enter-to-edit action = %v, want fbChanged", action)
	}

	// A rune keeps capturing input rather than falling through to navigation.
	fe.handleKey(keyMsg("x"))
	if !fe.editingText {
		t.Error("expected editingText to remain true while typing")
	}

	// Commit a deterministic value.
	fe.variantInput.SetValue("medium")
	fe.handleKey(keyMsg("enter"))
	if fe.editingText {
		t.Error("expected editingText false after commit")
	}
	if got := fe.entries[0].variant; got != "medium" {
		t.Errorf("committed variant = %q, want %q", got, "medium")
	}
}

// TestFallbackEditorTopLevelEscCloses verifies esc at the top level closes the
// overlay and reports fbClosed to the host.
func TestFallbackEditorTopLevelEscCloses(t *testing.T) {
	fe := newFallbackEditor()
	fe.load("m1")
	fe.open()

	action, _ := fe.handleKey(keyMsg("esc"))
	if action != fbClosed {
		t.Errorf("action = %v, want fbClosed", action)
	}
	if fe.active {
		t.Error("expected active false after top-level esc")
	}
}
