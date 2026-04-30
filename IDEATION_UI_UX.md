# UI/UX Ideation Report — omo-profiler

**App:** omo-profiler — a terminal UI (TUI) profile manager for `oh-my-openagent`
**Tech Stack:** Go + Bubbletea (Elm-architecture TUI) + Lipgloss (styling) + Bubbles (components)
**Analysis Date:** 2026-04-13

---

## Executive Summary

omo-profiler is a well-structured Bubbletea application with a coherent Catppuccin-adjacent color palette (purple `#7D56F4`, magenta `#FF6AC1`, cyan `#78DCE8`, Catppuccin Mocha surface colors). The architecture is clean with proper message-passing, but several UI/UX patterns limit usability, accessibility, and visual clarity. This report identifies 31 concrete, actionable findings across 6 categories — all anchored to specific files and lines.

---

## Category 1: Usability

### 1.1 Style definitions are duplicated across every view file — no shared token system

**Severity:** High | **Files:** `internal/tui/styles.go`, `internal/tui/views/dashboard.go:69-101`, `internal/tui/views/diff.go:17-35`, `internal/tui/views/wizard_other.go:15-27`, `internal/tui/views/wizard_review.go:17-33`, `internal/tui/views/wizard_name.go:14-26`, `internal/tui/views/wizard_categories.go:41-54`

There are 93 separate `lipgloss.Color(...)` calls across 19 files. Every view redeclares its own `titleStyle`, `errorStyle`, `grayStyle` etc. from scratch. The color hex codes (`#7D56F4`, `#6C7086`, `#CDD6F4`, `#F38BA8`, `#A6E3A1`) are literally copy-pasted verbatim into each file.

`dashboard.go` defines `titleStyle`, `subtitleStyle`, `successStyle`, `grayStyle`, `selectedStyle`, `normalStyle`, `accentStyle`, `errorStyle`, `errorIconStyle` locally (lines 69–101). Those exact same styles already exist in `tui/styles.go` as package-level vars: `TitleStyle`, `SubtitleStyle`, `SuccessStyle`, `Gray`, etc. — but the `views` package cannot import them since they live in the parent `tui` package (circular import).

**Proposed fix:** Extract all shared styles into a new package `internal/tui/theme` (no imports from `views` or `tui`). Both `tui` and all `views` files import from `theme`. This eliminates ~80 duplicate style declarations and makes palette changes a single-file operation.

```
internal/tui/theme/
  colors.go     -- lipgloss.Color constants
  styles.go     -- shared Style vars
  layout.go     -- FixedSmallWidth etc (currently in layout package, already clean)
```

---

### 1.2 Dashboard menu does not visually distinguish dangerous vs safe actions

**Severity:** Medium | **File:** `internal/tui/views/dashboard.go:269-284`

All 9 menu items render identically. "Export Profile" and "Delete" (via List) have the same visual weight as "Create New" and "Switch Profile". There is no visual grouping, sectioning, or icon differentiation.

**Proposed fix:** Add icon prefixes per item and a faint separator between action groups. Group items: Status (`Switch`, `Edit`), Creation (`Create New`, `Create from Template`), Comparison (`Compare`), Management (`Manage Models`), I/O (`Import`, `Export`), Maintenance (`Check Schema Updates`).

```go
var menuItemIcons = []string{
    "⇄ ", "✦ ", "⧉ ", "✎ ", "⊞ ", "◈ ", "↓ ", "↑ ", "⬡ ",
}
```

Render a dim separator line between the Status group and Creation group.

---

### 1.3 Dashboard active profile status has no quick-action affordance

**Severity:** Medium | **File:** `internal/tui/views/dashboard.go:229-245`

The "Active: `profile-name`" status line is purely informational. Users who want to edit the active profile must navigate to "Edit Current" in the menu. There is no keyboard shortcut or inline hint to jump directly to edit.

**Proposed fix:** Add `[e]` shortcut on the dashboard that directly triggers `NavToEditorMsg`. Display the hint inline: `Active: my-profile  [e] edit`. This is already partially implemented — `dashboardKeyMap` has an `Export` binding — but no edit shortcut exists at the dashboard level.

---

### 1.4 Wizard step indicator uses fixed `→` separators that overflow on narrow terminals

**Severity:** Medium | **File:** `internal/tui/views/wizard.go:508-518`

The full-width wizard header at line 508 hard-codes 6 step names joined with ` → `. At 60 chars wide (the `IsCompact` threshold), the step breadcrumb still renders with full names. The compact branch (line 484) uses dots `●○`, which is good — but the threshold for switching is the same `IsCompact(width)` which triggers at `width < 60`. A 65-column terminal renders "Name → Categories → Agents → Hooks → Other Settings → Review & Save" which is 57 chars minimum plus padding — it clips or wraps badly.

**Proposed fix:** Compute the rendered width of the progress string before rendering; fall back to dot mode if it exceeds `w.width - 4`. Use `lipgloss.Width()` on the constructed string first.

---

### 1.5 Template selection does not show any profile metadata

**Severity:** Low | **File:** `internal/tui/views/template_select.go:125-145`

The template list shows only raw profile names with `> ` cursor. There is no indication of what each profile contains — no description, no creation date, no active indicator. A user with 10+ profiles cannot make an informed template choice.

**Proposed fix:** Load profile metadata during `Init()` and show a secondary line per item (profile name + active indicator + first category name if available). This matches the pattern already used in `list.go`'s `profileItem.Description()` method.

---

### 1.6 List view does not show profile count or position indicator

**Severity:** Low | **File:** `internal/tui/views/list.go:262-278`

The bubbles `list.Model` provides a status bar (`SetShowStatusBar(true)` is set at line 109), but the active/total item count is generic. When filtering, there is no "N matches" counter displayed in the view. The Bubbles list component does support `l.SetStatusBarItemName()` (singular/plural).

**Proposed fix:** Call `l.list.SetStatusBarItemName("profile", "profiles")` in `NewList()`. This gives users "1 profile" / "5 profiles" in the status bar automatically.

---

### 1.7 Confirmation dialogs are inline text — easy to miss

**Severity:** Medium | **Files:** `internal/tui/layout/layout.go:11-18`, `internal/tui/views/list.go:272-274`, `internal/tui/views/model_registry.go:636-654`

`RenderConfirmDialog` (layout.go line 11) produces a single line with a yellow/gray background. This is appended below the list content. On tall terminals, this dialog can appear far down the screen, out of the user's focal area. It also uses `[y/n]` which has no visual affordance for which is "dangerous".

**Proposed fix:** Render the confirm dialog as a centered overlay using `lipgloss.Place()` at 50% width, with a red-bordered box for destructive actions. Show the target name prominently inside the box. Make `y` render as `ErrorStyle` and `n` render as `SuccessStyle` to reinforce the danger/safety contrast.

```go
// In layout.go
func RenderDestructiveConfirm(target, width, height int) string {
    box := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#F38BA8")).
        Padding(1, 3)
    content := lipgloss.JoinVertical(lipgloss.Center,
        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F38BA8")).Render("Delete "+target+"?"),
        "",
        lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8")).Render("[y] Delete") +
        "  " +
        lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Render("[n] Cancel"),
    )
    return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box.Render(content))
}
```

---

### 1.8 Model Registry form uses implicit Tab cycling with no field labels styled as focused/unfocused

**Severity:** Medium | **File:** `internal/tui/views/model_registry.go:598-634`

The Add/Edit Model form renders three fields: `Display Name: [input]`, `Model ID: [input]`, `Provider: [input]`. The focused field is indicated only by the text cursor inside the input — there is no label highlighting, no active border, no visual cue that Tab switches focus between fields (line 598–634). A user cannot easily tell which field is active at a glance.

**Proposed fix:** Apply `wizOtherSelectedStyle` (bold + purple) to the label of the focused field. Add a `>` prefix to the focused line:

```go
focusedLabel := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
unfocusedLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))

// For each field:
prefix := "  "
labelStyle := unfocusedLabel
if m.focusedField == 0 {
    prefix = "> "
    labelStyle = focusedLabel
}
fmt.Sprintf("%sDisplay Name: %s", prefix, m.displayNameInput.View())
```

---

## Category 2: Accessibility

### 2.1 Minimum size warning has no terminal resize guidance on all platforms

**Severity:** Low | **File:** `internal/tui/layout.go:42-53`

`RenderMinimumSizeWarning` shows "Need: 40x12  Now: WxH" which is useful. However the third line "Resize or q to quit" is generic. On some terminal emulators, users might not know *how* to resize. This is a minor UX gap but could mention the window title bar drag or keyboard shortcut.

**Proposed fix:** The current implementation is solid. Consider adding the exact chars needed: `"Resize to +Nx+M or q to quit"` where N and M are the deficit amounts. This gives users a precise target.

```go
wNeed := layout.MinTerminalWidth - width
hNeed := layout.MinTerminalHeight - height
if wNeed < 0 { wNeed = 0 }
if hNeed < 0 { hNeed = 0 }
hint := fmt.Sprintf("Needs +%d cols, +%d rows", wNeed, hNeed)
```

---

### 2.2 Help bar disappears entirely when toast is showing on small terminals

**Severity:** Medium | **File:** `internal/tui/app.go:699`

`showHelpBar := toastView == "" || a.height >= 16` — when a toast appears on a terminal shorter than 16 rows, both the help bar AND the contextual hints vanish simultaneously. The user loses all navigation context exactly when an action just completed (success/error toast).

**Proposed fix:** Keep one line of help visible at all times. Instead of suppressing the help bar entirely, suppress only the full help and show a single ultra-compact line: `"[?] help  [q] quit"`. Reserve 1 line always.

```go
// In app.go View()
showHelpBar := true
showFullHelp := toastView == "" || a.height >= 16
```

---

### 2.3 Keyboard hint format is inconsistent — square brackets vs no brackets vs angle brackets

**Severity:** Low | **Files:** Multiple view files

- `app.go:733-759`: `"[↑↓] navigate"`, `"[Enter] select"`, `"[?] help"` — square brackets
- `wizard_other.go:680`: `"[Enter/→] expand  [←] collapse"` — square brackets
- `schema_check.go:238`: `" Press Esc to cancel"` — no brackets, "Press" prefix
- `schema_check.go:244`: `" Press Esc to go back"` — no brackets
- `schema_check.go:247`: `" [enter] save  [esc] cancel"` — lowercase, square brackets
- `wizard_review.go:217-219`: `"Enter to save • Shift+Tab to go back"` — no brackets, bullet separator
- `template_select.go:145`: `"[↑↓] navigate  [Enter] select  [Esc] cancel"` — square brackets

**Proposed fix:** Standardize on `[key] action` format with lowercase key names and sentence-case action. Create a `theme.KeyHint(key, action string) string` helper that renders consistently. Replace all inline hint strings with calls to this helper.

---

### 2.4 Error messages do not provide resolution paths

**Severity:** Medium | **Files:** `internal/tui/app.go:395`, `internal/tui/views/model_registry.go:628`

Toast error messages like `"Switch failed: <err>"`, `"Import failed: <err>"`, and `"Delete failed: <err>"` show the raw Go error string. For example `"Import failed: validation failed"` gives no hint about *what* to fix or *where*. The `importErrorStyle.Render("✗ " + i.err.Error())` pattern in `import.go` line 152 is better (inline, persistent) but still shows raw errors.

**Proposed fix:** Map common error categories to user-friendly messages with action hints:
- Validation failure → `"Config invalid — return to wizard to fix fields"`
- File not found → `"File not found — check path and permissions"`
- Name collision → already handled well in `doImportProfile` at line 299-303

---

### 2.5 No visual feedback during async profile load on dashboard

**Severity:** Low | **File:** `internal/tui/views/dashboard.go:234-235`

When `d.activeProfile == nil`, the status shows `grayStyle.Render("Loading...")` — a static text. There is no spinner or animation to indicate the load is in progress vs. genuinely empty. The spinner (`spinner.Dot`) already exists in `App` struct (app.go line 90) for heavy operations, but dashboard data-loading does not use it.

**Proposed fix:** Either use the dashboard's own spinner component (import `bubbles/spinner` into `dashboard.go`) or emit a `spinnerStartMsg` as part of `dashboardInit` and a `spinnerStopMsg` in `profileLoadedMsg` handler. A simple `…` rotating with tick would suffice.

---

## Category 3: Performance Perception

### 3.1 Dashboard re-creates itself on every navigation return, losing cursor position

**Severity:** Medium | **File:** `internal/tui/app.go:499-501`, `app.go:397-401`

`navigateTo(stateDashboard)` always calls `views.NewDashboard()` (line 499), resetting `cursor` to 0. After switching a profile (line 397-401), the user returns to the dashboard with the menu cursor reset to item 0 ("Switch Profile"). If they wanted to immediately export, they must navigate down 6 items again.

**Proposed fix:** Store the last cursor position in `App` struct and restore it when navigating back:

```go
// In App struct:
dashboardCursor int

// In navigateTo(stateDashboard):
a.dashboard = views.NewDashboard()
a.dashboard.SetCursor(a.dashboardCursor)  // add SetCursor method

// In Dashboard.Update() on cursor change:
// emit CursorChangedMsg{pos} which App stores in dashboardCursor
```

Alternatively, never reconstruct `dashboard` on back-navigation; only reconstruct on explicit "refresh" actions (profile load). The `views.NewDashboard()` call in `navigateTo` could be conditionally skipped.

---

### 3.2 Schema check has no timeout or progress update

**Severity:** Low | **File:** `internal/tui/views/schema_check.go:94-98`, `schema_check.go:237`

`fetchSchemaCompareCmd` is a blocking network call. The spinner runs (`spinner.Tick` is started), but if the network is slow or the host is down, the user sees `"Checking Schema Updates..."` indefinitely with no timeout indication. There is no elapsed timer shown.

**Proposed fix:** Add a timeout to `fetchSchemaCompareCmd` using `context.WithTimeout`. Show elapsed seconds in the spinner line: `"Checking Schema Updates... 3s"`. Implement with a `time.Tick` that updates a counter field in `SchemaCheck` struct.

---

### 3.3 WizardOther viewport re-renders entire content on every keystroke

**Severity:** Medium | **File:** `internal/tui/views/wizard_other.go:671-673`, `wizard_other_render.go:12`

`w.refreshView()` calls `w.viewport.SetContent(w.renderContent())` which rebuilds all category/section content from scratch on every key event. `renderContent()` in `wizard_other_render.go` iterates all 28 sections including `renderSubSection()` calls for every expanded section. With many sections expanded, this is O(n) string allocations on every keystroke.

**Proposed fix:** Track a `dirty bool` flag. Only call `refreshView()` when state actually changes (cursor move, toggle, expand). In Update, set `dirty = true` when cursor/state changes, then conditionally call `refreshView()`. Alternatively, memoize the rendered content with a content hash.

---

### 3.4 List view re-initializes on every profile switch, losing scroll position

**Severity:** Low | **File:** `internal/tui/app.go:427-431`

After deleting a profile, `a.list = views.NewList()` is called (line 428). After switching, the list is reset via `a.state = stateDashboard` (going away), so this is less critical. However the pattern of always re-constructing views loses all UI state (scroll offset, filter text) unnecessarily.

**Proposed fix:** Add a `Refresh() tea.Cmd` method to `List` that reloads profile data without resetting the list model. Call `a.list.Refresh()` instead of `views.NewList()` in deleteProfileDoneMsg handler.

---

## Category 4: Visual Polish

### 4.1 Dashboard title has no visual separation or ASCII art identity

**Severity:** Medium | **File:** `internal/tui/views/dashboard.go:225-267`

"OMO-Profiler" renders as a bold purple string followed immediately by the subtitle. In a terminal app, the home screen is the first thing users see. There is no visual anchoring — no separator, no logo mark, no distinctive header treatment.

**Proposed fix:** Add a thin separator under the title using box-drawing characters. Optionally add a one-line "wordmark" with the app name in a distinct style. The separator should adapt to width:

```go
// In dashboard.go View():
separator := strings.Repeat("─", min(a.width, 40))
dimSep := lipgloss.NewStyle().Foreground(lipgloss.Color("#313244")).Render(separator)

header = strings.Join([]string{
    "",
    titleStyle.Render("OMO-Profiler") + "  " + subtitleStyle.Render("oh-my-openagent profiles"),
    dimSep,
    "",
    profileStatus,
    statsLine,
    "",
}, "\n")
```

---

### 4.2 Selected menu item uses full-width background highlight but no padding

**Severity:** Low | **File:** `internal/tui/views/dashboard.go:273-278`

`selectedStyle` applies `Background(Purple)` but without width padding, the highlight only covers the text width — it does not extend to a consistent column. On terminals this creates a jagged look where some selected items appear wider than others due to varying name length.

**Proposed fix:** Use a fixed-width render for menu items. Calculate the longest menu item width and pad all items to that length:

```go
maxLen := 0
for _, item := range menuItems {
    if len(item) > maxLen { maxLen = len(item) }
}
// Then: label := selectedStyle.Width(maxLen + 2).Render(item)
```

This creates a consistent selection bar width across all items.

---

### 4.3 Diff view pane labels use directional arrows inconsistently

**Severity:** Low | **File:** `internal/tui/views/diff.go:406-416`

`renderPaneLabel` uses `"◀ Left Profile"` when left is focused, `"Left Profile"` when not focused, `"Right Profile ▶"` when right is focused. The arrows point inward toward the center — ◀ points right (toward center from left), ▶ points left (toward center from right). This is counter-intuitive. The arrows suggest "navigate left/right" but they actually indicate which pane is active.

**Proposed fix:** Use a consistent indicator that means "active/focused" rather than directional. A filled dot or a simple bracket:

```go
// Focused:   "[Left Profile]" styled bold+purple
// Unfocused: " Left Profile " styled gray
```

Or use `●` prefix for the focused pane:
```go
if isFocused { return focusedStyle.Render("● Left Profile") }
return inactiveStyle.Render("  Left Profile")
```

---

### 4.4 Toast notification has no border or visual container — blends with content

**Severity:** Medium | **File:** `internal/tui/app.go:686-697`

Toast messages render as colored text appended below the content (line 715-717). On a busy view (long list, wizard with many fields), the toast appears at the bottom of the viewport area but without any border or background — it reads as continuation of content rather than an ephemeral notification.

**Proposed fix:** Add a subtle border and background to toasts. Right-align them or float them in the lower-right corner using `lipgloss.Place()` with `lipgloss.Right` alignment. Use a rounded border:

```go
var style lipgloss.Style
switch a.toastType {
case toastSuccess:
    style = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#A6E3A1")).
        Foreground(lipgloss.Color("#A6E3A1")).
        Padding(0, 1)
case toastError:
    style = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#F38BA8")).
        Foreground(lipgloss.Color("#F38BA8")).
        Padding(0, 1)
default:
    style = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#78DCE8")).
        Foreground(lipgloss.Color("#78DCE8")).
        Padding(0, 1)
}
toastView = style.Render("  " + icon + " " + a.toast + "  ")
```

---

### 4.5 WizardOther section content uses raw snake_case field names with no human labels

**Severity:** High | **File:** `internal/tui/views/wizard_other_render.go:196-298`

All sub-section fields render with their JSON field path names as labels: `"aggressive_truncation"`, `"dynamic_context_pruning.turn_protection.enabled"`, `"circuit_breaker.consecutive_threshold"`. These are developer-facing schema names, not user-friendly labels.

**Proposed fix:** Create a `fieldLabels map[string]string` that maps schema paths to human-readable descriptions. At minimum, add short descriptions next to section headers. Use the `otherSectionNames` pattern (which already has human names like "Background Task", "Tmux") as a model and extend it to field level:

```go
var fieldDescriptions = map[string]string{
    "experimental.aggressive_truncation": "Aggressively trim context to save tokens",
    "experimental.auto_resume":           "Auto-resume interrupted sessions",
    "tmux.isolation":                     "How Tmux sessions are isolated",
    // ...
}
```

Render these as dim gray secondary text on the same line or a sub-line.

---

### 4.6 Color palette duplication makes future theme changes expensive

**Severity:** Low | **File:** All view files

The Catppuccin Mocha-adjacent palette (`#7D56F4` purple, `#6C7086` overlay1, `#CDD6F4` text, `#F38BA8` red, `#A6E3A1` green, `#F9E2AF` yellow, `#FF6AC1` pink, `#78DCE8` teal) is excellent and consistent. But it is repeated verbatim 93 times. A single brand change (e.g., switching from purple to teal as the primary) requires editing 19 files.

**Proposed fix:** See 1.1 above — extract to `internal/tui/theme/colors.go` with named constants:

```go
package theme

import "github.com/charmbracelet/lipgloss"

var (
    Primary   = lipgloss.Color("#7D56F4") // purple
    Accent    = lipgloss.Color("#FF6AC1") // magenta
    Highlight = lipgloss.Color("#78DCE8") // cyan
    Success   = lipgloss.Color("#A6E3A1") // green
    Warning   = lipgloss.Color("#F9E2AF") // yellow
    Error     = lipgloss.Color("#F38BA8") // red
    Muted     = lipgloss.Color("#6C7086") // gray
    Text      = lipgloss.Color("#CDD6F4") // white
)
```

---

## Category 5: Interaction Improvements

### 5.1 WizardOther category navigation lacks breadcrumb — users lose context inside deep trees

**Severity:** High | **File:** `internal/tui/views/wizard_other.go:675-703`, `wizard_other_render.go:12-103`

The Other Settings step has 6 categories, each with 3–8 sections, each with 1–20 sub-fields. The current UI shows all categories with expand/collapse but there is no persistent indicator of "you are in: Infrastructure > Tmux > layout". The `inCategory bool` and `inSubSection bool` state exists (lines 362-366) but the only hint is the description line changing (line 683-688 in wizard_other.go).

**Proposed fix:** Render a breadcrumb line below the title when `inCategory` or `inSubSection` is true:

```go
// In wizard_other.go View()
if w.inSubSection {
    catName := otherCategoryNames[int(w.currentCategory)]
    sectionName := otherSectionNames[int(w.currentSection)]
    breadcrumb := dimStyle.Render(catName + " › " + sectionName)
    // show below title
}
```

This is a 3-line addition to the View function with no state changes needed.

---

### 5.2 WizardOther expanded sections do not show a summary of current values when collapsed

**Severity:** Medium | **File:** `internal/tui/views/wizard_other_render.go:88-99`

When a section like "Ralph Loop" is collapsed, it shows only `▶ Ralph Loop`. There is no preview of current values (e.g., "Ralph Loop — enabled, 10 iterations"). The user must expand each section to audit their current configuration.

**Proposed fix:** For boolean-primary sections, show an inline summary when collapsed:

```go
// Example for Ralph Loop:
if !w.sectionExpanded[section] {
    summary := ""
    if section == sectionRalphLoop && w.fieldSelected("ralph_loop.enabled") {
        if w.rlEnabled {
            summary = dimStyle.Render(" (enabled)")
        }
    }
    lines = append(lines, fmt.Sprintf("%s%s%s %s%s", sectionIndent, cursor, expandIcon, labelStyle.Render(name), summary))
}
```

---

### 5.3 Diff view: scroll is synchronized between panes but there is no per-pane scroll option

**Severity:** Medium | **File:** `internal/tui/views/diff.go:227-237`, `diff.go:172-195`

`scrollBoth()` always moves both viewports together. For profiles with very different structures, the left and right sides may have content at different vertical positions. There is no way to scroll only the focused pane independently.

**Proposed fix:** Add a modifier key for independent scroll. `Shift+Up/Down` or `Ctrl+Up/Down` scrolls only the focused pane:

```go
case "ctrl+up":
    if d.focused == focusLeft {
        d.leftViewport.SetYOffset(max(0, d.leftViewport.YOffset - 1))
    } else {
        d.rightViewport.SetYOffset(max(0, d.rightViewport.YOffset - 1))
    }
case "ctrl+down":
    // ... focused pane only
```

Document this in the help bar as `"[↑↓] sync scroll  [ctrl+↑↓] solo scroll"`.

---

### 5.4 Profile list does not support multi-select for batch delete

**Severity:** Low | **File:** `internal/tui/views/list.go`

Users with many legacy profiles must delete them one by one, each with a confirmation dialog. The `bubbles/list` component does not natively support multi-select, but it can be implemented with a custom delegate and a `selectedItems map[string]bool`.

**Proposed fix:** Add Space to toggle selection, `D` to bulk-delete selected items. Show selected count in the status bar. This is a larger feature addition but aligns with power-user workflows typical in TUI profile managers.

---

### 5.5 Import view does not support tab-completion or file browsing

**Severity:** Medium | **File:** `internal/tui/views/import.go:57-75`

The import view is a bare `textinput` for a file path. There is no path completion, no directory listing, no drag-and-drop indication. Users must know the exact path and type it (or paste it).

**Proposed fix:** Implement basic path tab-completion using `os.ReadDir()`:
- On `Tab` keypress, call `filepath.Glob(input.Value()+"*")` 
- If 1 match: auto-complete
- If multiple: show a dropdown below the input with up to 8 matches
- If 0 matches: flash the input border red briefly

This pattern is common in TUI file pickers and significantly reduces error rate.

---

### 5.6 No undo for accidental profile deletion

**Severity:** Medium | **File:** `internal/tui/app.go:413-431`

Profile deletion is irreversible. The confirmation dialog (`[y/n]`) is easy to accidentally confirm. There is no undo, no recycle bin, no "deleted X seconds ago" grace period.

**Proposed fix:** Implement a soft-delete system: on delete, move the profile file to a `~/.config/oh-my-openagent/profiles/.trash/` directory with a timestamp suffix. Add a "Recently Deleted" section to the List view showing the last 5 deleted profiles with restore option. Auto-purge after 7 days.

Alternatively, a simpler approach: keep the last deleted profile in a `lastDeleted *profile.Profile` field in `App` struct. If a new toast appears after delete, add `[z] undo` to the hint. The undo window is the toast duration (3 seconds).

---

### 5.7 Wizard navigation allows going to any step directly — but there is no step indicator interactivity

**Severity:** Low | **File:** `internal/tui/views/wizard.go:464-520`

The step indicator is purely visual (rendered steps at the top). Users cannot click/select individual steps. In vim-key mode, there is no jump-to-step shortcut (e.g., `1` for Name, `2` for Categories). Tab/Shift+Tab linearly traverse steps.

**Proposed fix:** Add digit shortcuts `1-6` that jump directly to a step. When jumping forward, validate and apply the current step first (same as `nextStep()` does). This removes friction when editing specific parts of a profile:

```go
case tea.KeyMsg:
    if msg.String() >= "1" && msg.String() <= "6" {
        targetStep, _ := strconv.Atoi(msg.String())
        return w.jumpToStep(targetStep)
    }
```

---

## Category 6: State Handling

### 6.1 App.navigateTo always re-creates the target view, losing all transient state

**Severity:** High | **File:** `internal/tui/app.go:492-531`

Every call to `navigateTo(stateDashboard)` creates a fresh `views.NewDashboard()`. Same for `stateList`, `stateModels`, etc. This means:
- Search filter in List view is lost when navigating to wizard and back
- Model Registry cursor position resets after saving a model
- Export path field is cleared if user navigates away accidentally

The current pattern of always constructing fresh views is intentional for simplicity but has UX cost.

**Proposed fix:** Keep view instances alive in `App` struct (they already are — `a.list`, `a.modelRegistry`, etc.). The problem is that `navigateTo` explicitly reconstructs them. Remove the reconstruction from `navigateTo` and only reconstruct on explicit "fresh start" actions. Add a `view.Reset()` method for cases where fresh state is needed:

```go
// In navigateTo() — remove:
// case stateList:
//     a.list = views.NewList()  // REMOVE THIS

// Instead:
case stateList:
    a.list.SetSize(a.width, a.contentHeight())
    cmd = a.list.Init()  // Init refreshes data but keeps UI state
```

The `Init()` method already handles data refresh for most views — the view reconstruction is redundant.

---

### 6.2 WizardSaveMsg discards the saved profile — callers cannot access the saved data

**Severity:** Low | **File:** `internal/tui/views/wizard.go:213-214`, `internal/tui/app.go:434-441`

`WizardSaveMsg` carries a `*profile.Profile` field (line 21) but the app-level handler at line 434-441 ignores it entirely. The profile is saved to disk, then the dashboard is re-created with a fresh data load. This means after saving, there is a brief "Loading..." state on the active profile status.

**Proposed fix:** Pass the saved profile data directly to the dashboard refresh instead of triggering a full re-load. Or emit a `ProfileSavedMsg{profileName string}` that the dashboard can handle to update its `activeProfile` field without re-fetching from disk.

---

### 6.3 Toast state has no queue — rapid actions drop notifications

**Severity:** Medium | **File:** `internal/tui/app.go:207-218`

The toast system is single-slot: `a.toast string`, `a.toastType toastType`, `a.toastActive bool`. If two operations complete in rapid succession (e.g., import + auto-switch), the second `toastMsg` overwrites the first one. The `clearToastMsg` timer from the first toast then clears the second toast early.

**Proposed fix:** Implement a simple toast queue:

```go
type App struct {
    // Replace:
    // toast       string
    // toastType   toastType  
    // toastActive bool
    
    // With:
    toastQueue []toastItem
}

type toastItem struct {
    text     string
    typ      toastType
    duration time.Duration
}
```

Process queue in FIFO order. Display the front item, pop it when its timer fires, start the next item's timer. This is a 30-line change with significant reliability improvement.

---

### 6.4 SchemaCheck state machine has an implicit edge: no-diff transitions immediately to save path

**Severity:** Low | **File:** `internal/tui/views/schema_check.go:158-165`

In the `schemaCheckResultMsg` handler (line 158), if `result.Identical` is false, the view immediately transitions to `stateSchemaCheckSavePath` and focuses the text input. The user gets no opportunity to review *what* the differences are before being asked where to save them. The diff content is in `result.Diff` but is never displayed in the UI.

**Proposed fix:** Add an intermediate state `stateSchemaCheckDiffPreview` that shows the diff content in a scrollable viewport. Add a "Save diff? [y/n]" prompt at the bottom. Only on `y` transition to `stateSchemaCheckSavePath`. This lets users see whether the schema updates are significant before deciding to save.

---

### 6.5 Dashboard profile data is not refreshed when returning from List without switching

**Severity:** Low | **File:** `internal/tui/app.go:376-377`, `app.go:492-501`

`navigateTo(stateDashboard)` calls `a.dashboard = views.NewDashboard()` and then `cmd = a.dashboard.Init()`. `Init()` triggers `loadActiveProfile` which re-fetches from disk. This is correct but means every navigation to the dashboard (even just pressing Esc from the list without any change) triggers a disk read.

**Proposed fix:** Track a `profileDataStale bool` flag in `App`. Only set it to true after operations that actually change profile state (switch, delete, import, save). In `navigateTo(stateDashboard)`, only call `dashboard.Init()` (which triggers the load) when `profileDataStale == true`, otherwise use `dashboard.Refresh()` which is a no-op.

---

## Priority Matrix

| # | Finding | Impact | Effort | Priority |
|---|---------|--------|--------|----------|
| 1.1 | Extract shared theme package | High | Medium | P1 |
| 4.5 | Human-readable field labels in WizardOther | High | Medium | P1 |
| 5.1 | Breadcrumb in WizardOther | High | Low | P1 |
| 6.1 | Stop re-creating views on navigation | High | Low | P1 |
| 6.3 | Toast queue system | Medium | Low | P2 |
| 4.4 | Bordered toast notifications | Medium | Low | P2 |
| 1.7 | Centered overlay for confirm dialogs | Medium | Medium | P2 |
| 1.8 | Focused field label highlighting in forms | Medium | Low | P2 |
| 2.2 | Keep help hint visible when toast shows | Medium | Low | P2 |
| 5.2 | Section summary when collapsed | Medium | Low | P2 |
| 5.3 | Per-pane scroll in Diff view | Medium | Medium | P2 |
| 5.5 | Path tab-completion in Import | Medium | High | P3 |
| 3.1 | Preserve dashboard cursor on return | Medium | Low | P2 |
| 6.4 | Show diff content before save prompt | Medium | Low | P2 |
| 1.2 | Menu item icons and grouping | Low | Low | P3 |
| 4.1 | Dashboard title separator | Low | Low | P3 |
| 4.2 | Fixed-width menu selection bar | Low | Low | P3 |
| 2.3 | Consistent keyboard hint format | Low | Medium | P3 |
| 1.6 | Profile count in list status bar | Low | Low | P3 |
| 1.4 | Wizard step indicator overflow fix | Low | Low | P3 |

---

## Implementation Notes

All proposed changes are strictly within the existing Bubbletea/Lipgloss paradigm. No new dependencies are required. The highest-value, lowest-effort changes are:

1. **Finding 5.1** (breadcrumb in WizardOther) — 3 lines added to `wizard_other.go:View()`
2. **Finding 1.6** (profile count) — 1 line in `NewList()`
3. **Finding 4.2** (fixed-width selection) — 5 lines in `renderMenuContent()`
4. **Finding 6.3** (toast queue) — isolated to `app.go`, no view changes needed
5. **Finding 2.2** (help visibility with toast) — 1 line change in `app.go:699`

The **theme extraction (1.1 / 4.6)** is the highest-leverage structural change and should be the first major refactor — it unblocks all future visual polish work and makes consistent changes trivial.
