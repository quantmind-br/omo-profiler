# UI/UX Improvements for omo-profiler

Analysis of the TUI forms and navigation, with improvements prioritized by implementation order.

## Phase 1: Critical Fixes (Must Fix)

### 1. Fix form navigation scroll trap [HIGH PRIORITY]

**ID**: uiux-001  
**Category**: usability  
**Affected Components**:
- `internal/tui/views/wizard_agents.go`
- `internal/tui/views/wizard_categories.go`
- `internal/tui/views/wizard_other.go`

**Problem**: In wizard steps with forms, keyboard navigation (Up/Down) changes the focused field but does not scroll the viewport. If the form exceeds the screen height, the user cannot see the active field. This is a blocking usability issue.

**Root Cause**: The `Update` method returns early when handling form navigation keys (lines 517-533 in wizard_agents.go), preventing the viewport from receiving scroll commands.

**Solution**: Implement automatic viewport scrolling when the focused field changes.

```go
// Add helper method to calculate field line position
func (w WizardAgents) getLineForField(field agentFormField) int {
    // Calculate based on agent list position + form field offset
    baseLines := w.cursor + 1 // Lines before expanded agent
    fieldOffset := int(field) * 1 // Each field is ~1 line (adjust for textareas)
    return baseLines + fieldOffset + 2 // +2 for spacing
}

// In Update loop after changing focus:
line := w.getLineForField(w.focusedField)
if line < w.viewport.YOffset {
    w.viewport.SetYOffset(line)
} else if line >= w.viewport.YOffset + w.viewport.Height {
    w.viewport.SetYOffset(line - w.viewport.Height + 1)
}
```

**Estimated Effort**: Medium

---

### 2. Enable standard Tab navigation in forms [HIGH PRIORITY]

**ID**: uiux-002  
**Category**: interaction  
**Affected Components**:
- `internal/tui/views/wizard_agents.go`
- `internal/tui/views/wizard_categories.go`
- `internal/tui/views/wizard_other.go`

**Problem**: The `Tab` key does not navigate between fields inside wizard forms. Tab is the universal standard for form field navigation.

**Root Cause**: `Tab` key is caught by the outer wizard navigation logic and ignored when `inForm` is true.

**Solution**: Add Tab/Shift+Tab handling inside the `inForm` logic blocks.

```go
// Inside form handling block (after line 511 in wizard_agents.go):
case "tab":
    w.focusedField++
    if w.focusedField > fieldPermExtDir {
        w.focusedField = fieldModel
    }
    w.updateFieldFocus(ac)
    // Also scroll viewport (use solution from uiux-001)
    return w, nil
case "shift+tab":
    if w.focusedField == fieldModel {
        w.focusedField = fieldPermExtDir
    } else {
        w.focusedField--
    }
    w.updateFieldFocus(ac)
    return w, nil
```

**Estimated Effort**: Small

---

## Phase 2: Usability Improvements

### 3. Add section-level validation in Editor [MEDIUM PRIORITY]

**ID**: uiux-006 (adjusted)  
**Category**: usability  
**Affected Components**:
- `internal/tui/views/editor.go`

**Problem**: The Editor only validates the profile upon saving. Users might make invalid changes and not know until they try to save.

**Solution (Simplified)**: Validate when switching sections, not on every keystroke.

```go
// In handleSidebarKeys, when changing sections:
case key.Matches(msg, e.keys.Up), key.Matches(msg, e.keys.Down):
    previousSection := e.section
    // ... existing navigation logic ...
    
    // Validate previous section before leaving
    if errs := e.validateSection(previousSection); len(errs) > 0 {
        e.sectionWarnings[previousSection] = true
    }

// In renderSidebar, show indicators:
indicator := ""
if e.sectionWarnings[i] {
    indicator = " ⚠"
}
lines = append(lines, style.Render(cursor + name + indicator))
```

**Estimated Effort**: Medium

---

### 4. Improve Editor focus indication [LOW PRIORITY]

**ID**: uiux-005 (adjusted)  
**Category**: interaction  
**Affected Components**:
- `internal/tui/views/editor.go`

**Problem**: Visual separation between sidebar and content focus is subtle, causing mode confusion.

**Solution**: Add arrow indicator instead of aggressive dimming (which hurts readability).

```go
// In renderSidebar:
func (e Editor) renderSidebar() string {
    var lines []string
    
    // Add focus indicator at top
    if e.focus == focusSidebar {
        lines = append(lines, editorAccentStyle.Render("► Sections"))
    } else {
        lines = append(lines, editorSubtitleStyle.Render("  Sections"))
    }
    // ... rest of sidebar rendering
}

// In renderContent header:
if e.focus == focusContent {
    header = editorAccentStyle.Render("► " + sectionNames[e.section])
}
```

**Estimated Effort**: Small

---

### 5. Improve help text consistency [LOW PRIORITY]

**ID**: uiux-004 (adjusted)  
**Category**: usability  
**Affected Components**:
- `internal/tui/views/editor.go`
- `internal/tui/views/wizard.go`
- `internal/tui/views/list.go`

**Problem**: The `Esc` key behavior appears inconsistent but is actually logical - users don't understand the pattern.

**Current Pattern** (already implemented correctly):
- Editor content → Esc → focus sidebar → Esc → exit
- Wizard → Esc → previous step → Esc → cancel
- List → Esc → dashboard

**Solution**: Add consistent help footer across all views documenting this behavior.

```go
// Shared help component
func renderEscapeHelp(context string) string {
    return helpStyle.Render(fmt.Sprintf("esc: %s", context))
}

// Usage in each view:
// Editor (content): "esc: sidebar"
// Editor (sidebar): "esc: cancel"  
// Wizard: "esc: back"
// List: "esc: dashboard"
```

**Estimated Effort**: Small

---

## Phase 3: Deferred (v2.0)

### 6. Implement adaptive colors for terminal themes [DEFERRED]

**ID**: uiux-003  
**Category**: accessibility  
**Status**: Deferred to v2.0 or when user reports readability issues

**Problem**: Hardcoded hex colors may have poor contrast on light terminal themes.

**Reason for Deferral**:
- No user complaints reported yet
- Requires extensive testing across terminal themes
- Risk of making things worse with poor light-mode color choices
- Current dark-mode colors work well for majority of users

**When to Implement**:
- When a user reports readability issues
- When adding theme support as a feature
- As part of v2.0 accessibility improvements

**Solution (for future)**:
```go
var (
    Purple = lipgloss.AdaptiveColor{Light: "#6C4AB6", Dark: "#7D56F4"}
    Gray   = lipgloss.AdaptiveColor{Light: "#5C5F77", Dark: "#6C7086"}
    // ... etc for all colors
)
```

---

## Implementation Summary

| Order | ID | Title | Priority | Effort | Status |
|-------|-----|-------|----------|--------|--------|
| 1 | uiux-001 | Fix form navigation scroll trap | HIGH | Medium | TODO |
| 2 | uiux-002 | Enable Tab navigation in forms | HIGH | Small | TODO |
| 3 | uiux-006 | Section-level validation | MEDIUM | Medium | TODO |
| 4 | uiux-005 | Focus arrow indicators | LOW | Small | TODO |
| 5 | uiux-004 | Consistent help text | LOW | Small | TODO |
| 6 | uiux-003 | Adaptive colors | DEFERRED | Medium | v2.0 |

## Notes

- **Dependencies**: uiux-002 should include viewport scrolling from uiux-001
- **Testing**: Focus on forms with many fields (Agents step has 18+ fields)
- **Files to modify**: Primarily `wizard_agents.go`, `wizard_categories.go`, `wizard_other.go`, `editor.go`
