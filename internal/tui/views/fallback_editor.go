package views

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/diogenes/omo-profiler/internal/tui/layout"
)

// fallbackModelEntry is a single fallback-model row shared by the agent and
// category wizards.
type fallbackModelEntry struct {
	model           string
	modelDisplay    string
	variant         string
	reasoningEffort string
	rawJSON         string
	isRawJSON       bool
}

// hasAdvancedFallbackFields reports whether a decoded fallback object carries
// keys beyond the ones the structured editor understands, forcing raw-JSON mode.
func hasAdvancedFallbackFields(entry map[string]interface{}) bool {
	for key := range entry {
		if key != "model" && key != "variant" && key != "reasoning" && key != "reasoningEffort" {
			return true
		}
	}
	return false
}

// parseFallbackEntries normalizes any of the accepted fallback_models shapes
// (string, []string, []interface{} of strings/objects) into editor entries.
func parseFallbackEntries(value interface{}) []fallbackModelEntry {
	if value == nil {
		return nil
	}

	var entries []fallbackModelEntry
	appendString := func(model string) {
		entries = append(entries, fallbackModelEntry{model: model, modelDisplay: model})
	}
	appendObject := func(entry map[string]interface{}) {
		fe := fallbackModelEntry{}
		if m, ok := entry["model"].(string); ok {
			fe.model = m
			fe.modelDisplay = m
		}
		if v, ok := entry["variant"].(string); ok {
			fe.variant = v
		}
		if r, ok := entry["reasoning"].(string); ok {
			fe.reasoningEffort = r
		} else if r, ok := entry["reasoningEffort"].(string); ok {
			// Map deprecated none→off for the structured editor.
			if r == "none" {
				r = "off"
			}
			fe.reasoningEffort = r
		}
		if hasAdvancedFallbackFields(entry) {
			fe.isRawJSON = true
			if raw, err := json.Marshal(entry); err == nil {
				fe.rawJSON = string(raw)
			}
		}
		entries = append(entries, fe)
	}

	switch v := value.(type) {
	case string:
		appendString(v)
	case []string:
		for _, item := range v {
			appendString(item)
		}
	case []interface{}:
		for _, item := range v {
			switch entry := item.(type) {
			case string:
				appendString(entry)
			case map[string]interface{}:
				appendObject(entry)
			}
		}
	}

	return entries
}

// formatFallbackEntry renders the one-line collapsed summary for an entry.
func formatFallbackEntry(entry fallbackModelEntry) string {
	if entry.isRawJSON {
		text := entry.rawJSON
		if text == "" {
			text = entry.model
		}
		return fmt.Sprintf("raw %s", text)
	}

	parts := []string{entry.modelDisplay}
	if entry.variant != "" {
		parts = append(parts, "variant="+entry.variant)
	}
	if entry.reasoningEffort != "" {
		parts = append(parts, "reasoning="+entry.reasoningEffort)
	}
	return strings.Join(parts, " • ")
}

// fbStyles bundles the lipgloss styles a host wizard passes into the shared
// editor renderer so the overlay matches the surrounding step's palette.
type fbStyles struct{ dim, sel, cursor, text lipgloss.Style }

// fallbackAction tells the host wizard what to do after handleKey.
type fallbackAction int

const (
	fbNone              fallbackAction = iota // nothing to do
	fbChanged                                 // state changed → re-render viewport
	fbClosed                                  // esc at top level → overlay closed
	fbOpenModelSelector                       // host must open its ModelSelector
)

// fallbackEditor is a self-contained collapsible editor for a fallback_models
// value. Each entry is a single collapsible row: collapsed shows a one-line
// summary, expanded shows its editable sub-fields.
type fallbackEditor struct {
	active       bool // overlay open
	entries      []fallbackModelEntry
	focusedIdx   int
	expandedIdx  int // -1 = none expanded
	subField     int // 0=model 1=variant 2=reasoning 3=raw
	editingText  bool
	pendingAdd   bool // pending ModelSelector result is an add (vs set-model)
	variantInput textinput.Model
	rawInput     textinput.Model
}

func newFallbackEditor() fallbackEditor {
	variantInput := textinput.New()
	variantInput.Placeholder = "variant"
	variantInput.Width = 30

	rawInput := textinput.New()
	rawInput.Placeholder = `{"model":"id"}`
	rawInput.Width = 40

	return fallbackEditor{
		expandedIdx:  -1,
		variantInput: variantInput,
		rawInput:     rawInput,
	}
}

// load parses a fallback_models value into entries and resets navigation state.
func (fe *fallbackEditor) load(v interface{}) {
	fe.entries = parseFallbackEntries(v)
	fe.focusedIdx = 0
	fe.expandedIdx = -1
	fe.editingText = false
	fe.active = false
}

// value serializes the current entries back into a fallback_models value:
// nil when empty, a bare string for a single plain entry, else a slice of
// strings / {model,variant?,reasoning?} maps / parsed raw JSON.
func (fe *fallbackEditor) value() interface{} {
	if len(fe.entries) == 0 {
		return nil
	}

	if len(fe.entries) == 1 && fe.entries[0].variant == "" && fe.entries[0].reasoningEffort == "" && !fe.entries[0].isRawJSON {
		return fe.entries[0].model
	}

	arr := make([]interface{}, len(fe.entries))
	for i, fbe := range fe.entries {
		if fbe.isRawJSON {
			var parsed interface{}
			if json.Unmarshal([]byte(fbe.rawJSON), &parsed) == nil {
				arr[i] = parsed
			} else {
				arr[i] = fbe.model
			}
			continue
		}

		if fbe.variant != "" || fbe.reasoningEffort != "" {
			obj := map[string]interface{}{"model": fbe.model}
			if fbe.variant != "" {
				obj["variant"] = fbe.variant
			}
			if fbe.reasoningEffort != "" {
				obj["reasoning"] = fbe.reasoningEffort
			}
			arr[i] = obj
			continue
		}

		arr[i] = fbe.model
	}
	return arr
}

// summaryLabel is the one-line value shown for the collapsed field.
func (fe *fallbackEditor) summaryLabel() string {
	if len(fe.entries) == 0 {
		return "(none) [Enter to edit]"
	}
	return fmt.Sprintf("%d models [Enter to edit]", len(fe.entries))
}

// setWidth resizes the embedded text inputs to the host step's width.
func (fe *fallbackEditor) setWidth(width int) {
	fe.variantInput.Width = layout.MediumFieldWidth(width)
	fe.rawInput.Width = layout.WideFieldWidth(width, 10)
}

// open activates the overlay at the top-level entry list.
func (fe *fallbackEditor) open() {
	fe.active = true
	fe.expandedIdx = -1
	fe.editingText = false
	fe.clampFocus()
}

// clampFocus keeps focusedIdx within [0, len-1] (or 0 when empty).
func (fe *fallbackEditor) clampFocus() {
	if fe.focusedIdx >= len(fe.entries) {
		fe.focusedIdx = len(fe.entries) - 1
	}
	if fe.focusedIdx < 0 {
		fe.focusedIdx = 0
	}
}

// applySelectedModel folds a ModelSelector result into the editor: append when
// the selection was triggered by add, otherwise set the target entry's model.
func (fe *fallbackEditor) applySelectedModel(id, display string) {
	if fe.pendingAdd {
		fe.entries = append(fe.entries, fallbackModelEntry{model: id, modelDisplay: display})
		fe.focusedIdx = len(fe.entries) - 1
		fe.pendingAdd = false
		return
	}
	idx := fe.focusedIdx
	if fe.expandedIdx >= 0 {
		idx = fe.expandedIdx
	}
	if idx >= 0 && idx < len(fe.entries) {
		fe.entries[idx].model = id
		fe.entries[idx].modelDisplay = display
	}
}

// handleKey drives the editor state machine, returning the action the host must
// take next plus any tea.Cmd from an embedded text input.
func (fe *fallbackEditor) handleKey(msg tea.KeyMsg) (fallbackAction, tea.Cmd) {
	fe.clampFocus()

	// A text sub-field is capturing input.
	if fe.editingText {
		switch msg.String() {
		case "enter":
			if fe.expandedIdx >= 0 && fe.expandedIdx < len(fe.entries) {
				entry := &fe.entries[fe.expandedIdx]
				switch fe.subField {
				case 1:
					entry.variant = strings.TrimSpace(fe.variantInput.Value())
				case 3:
					entry.rawJSON = fe.rawInput.Value()
					entry.isRawJSON = entry.rawJSON != ""
				}
			}
			fe.editingText = false
			fe.variantInput.Blur()
			fe.rawInput.Blur()
			return fbChanged, nil
		case "esc":
			fe.editingText = false
			fe.variantInput.Blur()
			fe.rawInput.Blur()
			return fbChanged, nil
		default:
			var cmd tea.Cmd
			if fe.subField == 1 {
				fe.variantInput, cmd = fe.variantInput.Update(msg)
			} else {
				fe.rawInput, cmd = fe.rawInput.Update(msg)
			}
			return fbChanged, cmd
		}
	}

	// An entry is expanded; navigate/edit its sub-fields.
	if fe.expandedIdx >= 0 {
		if fe.expandedIdx >= len(fe.entries) {
			fe.expandedIdx = -1
			return fbChanged, nil
		}
		entry := &fe.entries[fe.expandedIdx]
		switch msg.String() {
		case "up", "k":
			fe.subField = (fe.subField + 3) % 4
			return fbChanged, nil
		case "down", "j":
			fe.subField = (fe.subField + 1) % 4
			return fbChanged, nil
		case "enter":
			switch fe.subField {
			case 0:
				fe.pendingAdd = false
				return fbOpenModelSelector, nil
			case 1:
				fe.editingText = true
				fe.variantInput.SetValue(entry.variant)
				fe.variantInput.Focus()
				return fbChanged, textinput.Blink
			case 2:
				cur := 0
				for i, level := range effortLevels {
					if level == entry.reasoningEffort {
						cur = i
						break
					}
				}
				entry.reasoningEffort = effortLevels[(cur+1)%len(effortLevels)]
				return fbChanged, nil
			case 3:
				fe.editingText = true
				fe.rawInput.SetValue(entry.rawJSON)
				fe.rawInput.Focus()
				return fbChanged, textinput.Blink
			}
		case "-", "d":
			fe.entries = append(fe.entries[:fe.expandedIdx], fe.entries[fe.expandedIdx+1:]...)
			fe.expandedIdx = -1
			fe.clampFocus()
			return fbChanged, nil
		case "esc", "ctrl+left":
			fe.expandedIdx = -1
			return fbChanged, nil
		}
		return fbNone, nil
	}

	// Top-level entry navigation.
	switch msg.String() {
	case "up", "k":
		if fe.focusedIdx > 0 {
			fe.focusedIdx--
		}
		return fbChanged, nil
	case "down", "j":
		if fe.focusedIdx < len(fe.entries)-1 {
			fe.focusedIdx++
		}
		return fbChanged, nil
	case "+", "a":
		fe.pendingAdd = true
		return fbOpenModelSelector, nil
	case "-", "d":
		if len(fe.entries) > 0 {
			fe.entries = append(fe.entries[:fe.focusedIdx], fe.entries[fe.focusedIdx+1:]...)
			fe.clampFocus()
		}
		return fbChanged, nil
	case "enter", "ctrl+right":
		if len(fe.entries) > 0 {
			fe.expandedIdx = fe.focusedIdx
			fe.subField = 0
		}
		return fbChanged, nil
	case "esc":
		fe.active = false
		return fbClosed, nil
	}
	return fbNone, nil
}

// render returns the overlay's lines, styled with the host's palette.
func (fe *fallbackEditor) render(indent string, s fbStyles) []string {
	var lines []string
	lines = append(lines, indent+s.sel.Render("┌─ Editing Fallback Models ─┐"))

	if len(fe.entries) == 0 {
		lines = append(lines, indent+"  "+s.dim.Render("(empty) press + to add"))
	}

	subCursor := func(field int) string {
		if fe.subField == field {
			return s.cursor.Render("> ")
		}
		return "  "
	}

	for i, entry := range fe.entries {
		marker := "  "
		if i == fe.focusedIdx {
			marker = s.cursor.Render("> ")
		}

		if i != fe.expandedIdx {
			lines = append(lines, indent+marker+s.text.Render("▶ "+formatFallbackEntry(entry)))
			continue
		}

		lines = append(lines, indent+marker+s.text.Render("▼ "+formatFallbackEntry(entry)))
		subIndent := indent + "    "

		modelVal := entry.modelDisplay
		if modelVal == "" {
			modelVal = "[Select model...]"
		}
		lines = append(lines, subIndent+subCursor(0)+s.text.Render("model    : ")+modelVal)

		variantVal := entry.variant
		if fe.editingText && fe.subField == 1 {
			variantVal = fe.variantInput.View()
		} else if variantVal == "" {
			variantVal = "(none)"
		}
		lines = append(lines, subIndent+subCursor(1)+s.text.Render("variant  : ")+variantVal)

		reasoningVal := entry.reasoningEffort
		if reasoningVal == "" {
			reasoningVal = "(none)"
		}
		lines = append(lines, subIndent+subCursor(2)+s.text.Render("reasoning: ")+reasoningVal)

		rawVal := entry.rawJSON
		if fe.editingText && fe.subField == 3 {
			rawVal = fe.rawInput.View()
		} else if rawVal == "" {
			rawVal = "(none)"
		}
		lines = append(lines, subIndent+subCursor(3)+s.text.Render("raw      : ")+rawVal)
	}

	help := "+:add  -:del  ↑↓:nav  enter:expand  esc:done"
	switch {
	case fe.editingText:
		help = "enter:save  esc:cancel"
	case fe.expandedIdx >= 0:
		help = "↑↓:field  enter:edit/pick  -:del  esc:collapse"
	}
	lines = append(lines, indent+"  "+s.dim.Render(help))

	return lines
}
