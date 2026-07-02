# Sparse Field Inclusion - Learnings & Conventions

## Project Conventions

### Test Patterns
- Use `setupTestEnv(t)` pattern from `wizard_hooks_test.go:166`
- Co-located `*_test.go` files
- Use `github.com/stretchr/testify` for assertions
- Test naming: `TestWizardOther{Scenario}{Detail}`

### Key Files
- `internal/tui/views/wizard_other.go` - Main file to fix (toggleSubItem, toggleSection)
- `internal/tui/views/wizard_other_test.go` - Where all tests go
- `internal/profile/sparse.go` - MarshalSparse for JSON verification
- `internal/profile/selection.go` - FieldSelection helpers

### Critical Functions
- `toggleSubItem()` - Line 3362 - needs Row 0 fix for disabled sections
- `toggleSection()` - Line 3333 - needs disabled-list cases
- `topLevelFieldPath()` - Line 425 - returns correct paths for disabled sections
- `fieldSelected()` / `toggleFieldSelection()` - Lines 323-335 - selection helpers
- `emptySliceIfSelected()` - Line 303 - handles nil vs empty slice
- `Apply()` - Lines 1432-1908 - writes view state to config
- `SetConfig()` - Lines 1031-1430 - loads config into view state

### Disabled-List Sections
- `sectionDisabledMcps`
- `sectionDisabledAgents`
- `sectionDisabledSkills`
- `sectionDisabledCommands`
- `sectionDisabledTools`

### Test Helper Pattern (from Task 1)
- `setupWizardOtherWithSelection(t, paths...)` - create wizard with selection
- `applyAndMarshal(t, w, selection)` - Apply→MarshalSparse→return map
- `assertJSONContains/Omits/Equals(t, jsonMap, key, expected)` - JSON assertions

## Guardrails (Must NOT)
- NO changes to `config/types.go`
- NO changes to `selection.go`
- NO changes to `sparse.go`
- NO changes to `wizard_hooks.go`
- NO changes to `profile.go`
- NO per-item inclusion for disabled lists (whole-section only)
- NO label/nomenclature changes

## Task 1: Test Helper Implementation (Completed)

### Added Helpers to wizard_other_test.go

1. **setupWizardOtherWithSelection(t, paths...)** - Returns WizardOther with pre-configured FieldSelection
   - Uses profile.NewBlankSelection() + SetSelected() for each path
   - Returns value type (not pointer) matching NewWizardOther() return type

2. **applyAndMarshal(t, w, selection)** - Applies config and returns JSON map
   - Calls w.Apply(&cfg, selection) to populate config
   - Calls profile.MarshalSparse(&cfg, selection, nil) for sparse JSON
   - Unmarshals to map[string]interface{} for easy assertions
   - Uses testify require.NoError for error handling

3. **assertJSONContains(t, jsonMap, key)** - Dot-notation key existence check
   - Supports nested paths like "experimental.aggressive_truncation"
   - Traverses map structure using strings.Split

4. **assertJSONOmits(t, jsonMap, key)** - Dot-notation key absence check
   - Returns early if parent path doesn't exist (key is effectively omitted)

5. **assertJSONEquals(t, jsonMap, key, expected)** - Dot-notation key value check
   - Combines existence check with value comparison
   - Uses testify assert.Equal for comparison

### Imports Added
- encoding/json (for Unmarshal)
- github.com/stretchr/testify/assert
- github.com/stretchr/testify/require

### Pattern Notes
- Test helpers use t.Helper() for proper error line reporting
- Dot-notation key traversal enables readable test assertions
- testify require for fatal errors, assert for non-fatal checks
- All existing tests continue to pass

## Task 2: Disabled-List Section Inclusion Toggle Fix (Completed)

### Problem
The disabled-list sections (disabled_agents, disabled_skills, disabled_commands, disabled_mcps, disabled_tools) in wizard_other.go had a bug where Row 0 (the inclusion checkbox) was not properly toggling the section's inclusion state. Instead, it was falling through to individual item toggling logic.

### Solution

#### 1. toggleSubItem() Fix (line ~3362)
Added early return check at TOP of function, BEFORE subSectionFieldPath() call:

```go
// Handle Row 0 inclusion toggle for disabled-list sections
if w.subCursor == 0 {
    switch w.currentSection {
    case sectionDisabledMcps, sectionDisabledAgents, sectionDisabledSkills,
        sectionDisabledCommands, sectionDisabledTools:
        w.toggleFieldSelection(w.topLevelFieldPath(w.currentSection))
        return
    }
}
```

This ensures that when user is at Row 0 of an expanded disabled-list section and presses Space, it toggles the inclusion state of the entire section (via topLevelFieldPath) rather than falling through to individual item logic.

#### 2. toggleSection() Fix (line ~3333)
Added cases for all 5 disabled-list sections:

```go
case sectionDisabledMcps:
    w.toggleFieldSelection(w.topLevelFieldPath(sectionDisabledMcps))
case sectionDisabledAgents:
    w.toggleFieldSelection(w.topLevelFieldPath(sectionDisabledAgents))
case sectionDisabledSkills:
    w.toggleFieldSelection(w.topLevelFieldPath(sectionDisabledSkills))
case sectionDisabledCommands:
    w.toggleFieldSelection(w.topLevelFieldPath(sectionDisabledCommands))
case sectionDisabledTools:
    w.toggleFieldSelection(w.topLevelFieldPath(sectionDisabledTools))
```

This enables section-level toggling when the section is NOT expanded (user presses Space on the collapsed section header).

### Tests Added
5 new tests verify the fix:
- TestWizardOtherDisabledAgentsInclusionToggle
- TestWizardOtherDisabledSkillsInclusionToggle
- TestWizardOtherDisabledCommandsInclusionToggle
- TestWizardOtherDisabledMcpsInclusionToggle
- TestWizardOtherDisabledToolsInclusionToggle

Each test:
1. Creates WizardOther with blank selection
2. Expands the section and sets subCursor=0
3. Sends Space key twice
4. Verifies selection.IsSelected(path) toggles on then off

### Files Modified
- internal/tui/views/wizard_other.go (toggleSubItem + toggleSection)
- internal/tui/views/wizard_other_test.go (5 new tests)

### Verification
All existing tests pass + 5 new tests pass.

## Task 7: Wizard Other Edit/Template Flow Integration Tests (Completed)

### New Integration Patterns
- Real edit/template flow tests should build `FieldSelection` with `profile.NewSelectionFromPresence(...)` to mirror the production bridge from JSON field presence to wizard inclusion state.
- For `wizard_other`, the correct in-memory integration sequence is `json.Unmarshal -> NewSelectionFromPresence -> SetConfig -> mutate wizard state if needed -> Apply -> profile.MarshalSparse`.
- `SetConfig()` is the point where checkbox/value state is hydrated from config; assertions for edit/template behavior should verify both selection state and loaded widget state before calling `Apply()`.
- Round-trip assertions are simplest when the source fixture is a minimal JSON string and the final sparse output is unmarshaled to `map[string]interface{}` for structural equality.

### Blank Selection Note
- `WizardOther.fieldSelected()` returns `true` when `selection == nil`, so `Apply()` includes all touched fields by default for new-profile scenarios.
- `profile.MarshalSparse` does not treat `nil` selection as "include everything"; tests that verify new-profile sparse output must provide an explicit marshal selection for the fields that `Apply()` populated.

### Cases Covered
- Edit flow preserves original field set for `disabled_agents` + `experimental.aggressive_truncation`.
- Edit flow can modify disabled agent values while keeping inclusion state.
- Edit flow can remove a selected field by toggling inclusion off before sparse marshal.
- Template flow preserves template field selection across `SetConfig -> Apply`.
- New-profile flow with blank selection includes touched values once marshal selection is derived from populated config.

## Task 5: Empty-Array and Field Omission Tests (Completed)

### Three-State Pattern for Disabled-List Fields
Every disabled-list field has 3 distinct JSON states that must be tested:
1. **Unselected** (not in FieldSelection) → field completely absent from JSON
2. **Selected, empty** (in FieldSelection, no items disabled) → field present as `[]`
3. **Selected, with values** (in FieldSelection, items disabled) → field present as `["item1", ...]`

### Test Structure
- **Omission tests**: `TestWizardOtherOmission_Disabled{Section}` — `setupWizardOtherWithSelection(t)` (no paths) → `assertJSONOmits`
- **Empty-array tests**: `TestWizardOtherEmptyArray_Disabled{Section}` — `setupWizardOtherWithSelection(t, path)` → `assertJSONEquals(result, key, []interface{}{})`
- **With-values tests**: `TestWizardOtherWithValues_Disabled{Section}` — setup + set map values → verify array content

### Boolean Three-State Pattern (auto_update)
- **Unselected** → absent from JSON (`assertJSONOmits`)
- **Selected, false** → `"auto_update": false` in JSON (`assertJSONEquals(result, "auto_update", false)`)
- **Selected, true** → `"auto_update": true` in JSON

### Key Insight: `emptySliceIfSelected` Function
`emptySliceIfSelected(selected, values)` at line 303 is the critical function:
- `selected=false` → returns `nil` (field omitted by MarshalSparse)
- `selected=true, values=nil` → returns `[]string{}` (empty array in JSON)
- `selected=true, values=["a","b"]` → returns `["a","b"]`

### Disabled Section Map Keys
- agents/skills/commands: `map[string]bool` fields (e.g., `w.disabledAgents["sisyphus"] = true`)
- mcps/tools: text input fields using `.SetValue("a, b")` pattern
- All use top-level JSON keys (no nesting): `disabled_agents`, `disabled_skills`, etc.

### Pre-existing Test Failures
- `TestWizardOtherInclusionSeparateFromValue_BoolField` and `_StringField` fail independently (not caused by new tests)
- Run `go test -run "TestWizardOtherEmptyArray|TestWizardOtherOmission"` to verify new tests in isolation

## Task 3: Round-Trip Tests for Disabled-List Sections (Completed)

### Pattern
15 tests added in `TestWizardOtherRoundTrip_*` naming convention:
- `{Section}` — selected with values → JSON has array with items
- `{Section}Empty` — selected, no values → JSON has `[]`
- `{Section}Omitted` — unselected → JSON has no key at all

### Sections tested
1. DisabledAgents (map[string]bool toggles → `disabled_agents`)
2. DisabledSkills (map[string]bool toggles → `disabled_skills`)
3. DisabledCommands (map[string]bool toggles → `disabled_commands`)
4. DisabledMcps (textinput.Value → comma-split → `disabled_mcps`)
5. DisabledTools (textinput.Value → comma-split → `disabled_tools`)

### Key insight: value types differ by section
- **Agents/Skills/Commands**: Set values via `w.disabledXxx["item"] = true` on map[string]bool
- **Mcps/Tools**: Set values via `w.disabledXxx.SetValue("a, b")` on textinput.Model
- Apply() iterates disableableXxx slices (agents/skills/commands) or splits by comma (mcps/tools)
- Order of items in output follows the `disableableXxx` slice order for map-based sections

### JSON output verification
- `applyAndMarshal` helper: `Apply(&cfg, selection)` → `MarshalSparse(cfg, selection, nil)` → `json.Unmarshal` → `map[string]interface{}`
- assertJSONEquals works with `[]interface{}{}` for empty arrays, `[]interface{}{"a", "b"}` for populated
- When field unselected: `MarshalSparse` skips it entirely (returns `include=false`)
- When field selected with nil/empty slice: `marshalLeafValue` converts nil slice to `[]interface{}{}`

### Pre-existing issue
- `TestWizardOtherInclusionSeparateFromValue_StringField` fails (unrelated to this task)
- Failure is in sectionDefaultRunAgent toggle logic, not disabled-list sections

## Task 6: Inclusion/Value Separation Tests (Completed)

### Tests Added
4 tests verifying inclusion and value state independence:

1. **TestWizardOtherInclusionSeparateFromValue_DisabledAgents** — toggles `disabled_agents` inclusion off/on via `subCursor=0` Space press; verifies `disabledAgents` map entries (per-agent bools) survive round-trip; applies and checks `cfg.DisabledAgents` (`[]string`)
2. **TestWizardOtherInclusionSeparateFromValue_BoolField** — toggles `experimental.aggressive_truncation` inclusion via Space at `subCursor=0`, then toggles value via `subValueFocused=true` Space; verifies each property changes independently
3. **TestWizardOtherInclusionSeparateFromValue_SliceField** — toggles `disabled_mcps` inclusion off/on via `subCursor=0` Space press; verifies `disabledMcps` textinput value preserved across inclusion changes
4. **TestWizardOtherInclusionSeparateFromValue_StringField** — uses `toggleFieldSelection()` directly (not Space key) for `default_run_agent` because simple sections toggle inclusion from `toggleSection()`, not `toggleSubItem()`

### Key Patterns
- **Disabled-list sections** (agents, skills, commands, mcps, tools): inclusion toggled at `subCursor=0` in subsection via Space key → `toggleSubItem()` → `toggleFieldSelection(topLevelFieldPath)`
- **Experimental subsection**: inclusion toggled at `subCursor=0` via `subSectionFieldPath()` path; value toggled with `subValueFocused=true`
- **Simple sections** (default_run_agent, auto_update, etc.): no subsection Row 0 inclusion toggle — `subSectionFieldPath` returns `""` and `toggleSubItem` has no case for them. Inclusion is toggled via `toggleSection()` from outside subsection or via `toggleFieldSelection()` directly
- **For re-enabling inclusion in tests**: use `toggleFieldSelection()` directly, not another Space press (which would toggle it off again)
- `cfg.DisabledAgents` is `[]string` (only true entries), `WizardOther.disabledAgents` is `map[string]bool`
