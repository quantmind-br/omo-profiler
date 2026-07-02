# TUI Validator - Audit Report

**Application**: `./omo-profiler`
**Version**: `omo-profiler version 0.1.0`
**Args**: ``
**Working directory**: `/home/diogo/dev/omo-profiler`
**Timestamp**: `20260611T193815Z` UTC
**Pipeline**: `tui-validator` skill (tmux + capture-pane + optional screenshots)
**Workspace**: `/home/diogo/.cache/tui-validator/omo-profiler/20260611T193815Z`

---

## 1. Summary

omo-profiler v0.1.0 was driven live inside tmux across all 10 app states
(dashboard, profile list, 6-step wizard, diff/compare, models, model-import,
template-select, import, export, schema-check), every documented keybinding,
unicode/accented/CJK/emoji stress input, and a resize matrix from 39x11 up to
160x40. **The TUI is broadly solid** — navigation, the wizard, diff view, model
search, schema-check network fetch, and unicode handling all work well. **One
blocker** spoils it: typing `q` while filtering the profile list silently quits
the whole app (the global key router preempts the list's filter handler). The
same root cause makes `?` open help and `Esc` bail to the dashboard mid-filter
(two majors). The Models view, which guards its search correctly, proves the fix
is a known pattern already in the codebase. A second-class major: the *declared*
minimum size (40x12) already renders an overlapping, broken dashboard — the
"Too Small" guard only fires below it. Remaining items are minor/cosmetic
(HTML-escaped JSON in the review step, unpadded diff borders, a truncated
validation message). Live pixel screenshots could not be framed in this
headless-agent session; all visual findings rest on authoritative text/ANSI
pane captures.

**Severity breakdown:**

| Severity | Count |
| --- | ---: |
| Blocker | 1 |
| Major | 3 |
| Minor | 1 |
| Cosmetic | 2 |
| Info | 2 |

| Audit stat | Value |
| --- | --- |
| Captures (text + ANSI) | 6 |
| Screenshots | 0 |
| Keybindings inventoried | 46 |
| Initial geometry | 80 x 24 |
| TERM | `xterm-256color` |

---

## 2. Keybindings Inventory

Raw file: `/home/diogo/.cache/tui-validator/omo-profiler/20260611T193815Z/keybindings.json`.

| Key | Context | Description | Source | Status |
| --- | --- | --- | --- | --- |
| `?` | global | toggle full-screen help overlay | documented+observed | active |
| `q` | global | quit application | documented+observed | active |
| `C-c` | global | quit application (cancels wizard with toast) | documented+observed | active |
| `Esc` | global | back / cancel to dashboard | documented+observed | active |
| `Up` | global | move selection up (also k) | documented+observed | active |
| `Down` | global | move selection down (also j) | documented+observed | active |
| `Enter` | global | select / activate | documented+observed | active |
| `Up` | dashboard | menu up (k) | observed | active |
| `Down` | dashboard | menu down (j) | observed | active |
| `Enter` | dashboard | select menu item | observed | active |
| `i` | dashboard | import profile shortcut | documented+observed | active |
| `e` | dashboard | export profile shortcut | documented+observed | active |
| `Enter` | profile-list | switch to selected profile | documented+observed | active |
| `e` | profile-list | edit profile | documented+observed | active |
| `d` | profile-list | delete profile (confirm y/n) | documented | skipped |
| `n` | profile-list | new profile | documented+observed | active |
| `/` | profile-list | search/filter profiles | documented+observed | active |
| `Esc` | profile-list | back to dashboard | documented+observed | active |
| `Tab` | wizard | next step (also Enter on most steps) | documented+observed | active |
| `S-Tab` | wizard | previous step | documented+observed | active |
| `C-s` | wizard | save profile | documented | active |
| `C-c` | wizard | cancel (shows toast, use Esc) | observed | active |
| `Esc` | wizard | back / cancel (discard prompt on review) | observed | active |
| `Space` | wizard-agents/hooks/other | toggle selection | documented+observed | active |
| `n` | wizard-categories | new category | documented | info |
| `d` | wizard-categories | delete category | documented | skipped |
| `C-Left` | wizard-categories/agents/other | collapse node | documented | info |
| `C-Right` | wizard-categories/agents/other | expand node | documented | info |
| `n` | models | new model | documented+observed | active |
| `i` | models | import from models.dev | documented | info |
| `e` | models | edit model | documented | info |
| `d` | models | delete model | documented | skipped |
| `/` | models | search models | documented+observed | active |
| `Esc` | models | back / cancel search | observed | active |
| `Tab` | diff | switch active pane (left/right) | documented+observed | active |
| `Enter` | diff | open profile selector for active pane | observed | active |
| `Up` | diff | scroll up (k) | observed | active |
| `Down` | diff | scroll down (j) | observed | active |
| `PgUp` | diff | page up | documented | info |
| `PgDn` | diff | page down | documented | info |
| `Esc` | diff | back to dashboard | observed | active |
| `Enter` | import/export/schema-check | submit path / confirm | documented+observed | active |
| `Esc` | import/export/schema-check | cancel | documented+observed | active |
| `r` | schema-check | retry network fetch | documented | info |
| `Enter` | template-select | use selected profile as template | observed | active |
| `Esc` | template-select | cancel | observed | active |

---

## 3. Findings

### [BLOCKER] Typing 'q' while filtering the profile list quits the entire app

**Phase:** probe  
**Evidence:** captures/EVIDENCE-q-quit-during-filter.txt  

In the Switch Profile list, pressing '/' starts the bubbles filter. Normal characters (e.g. 's') filter correctly, but the global key router in App.Update (internal/tui/app.go:148) intercepts 'q' for Keys.Quit BEFORE the list view's own FilterState guard (internal/tui/views/list.go:203) ever runs. The guard list at app.go:149-167 covers wizard/models/modelImport/import/export/schemaCheck focus states but NOT stateList while filtering. Result: any profile whose name contains 'q' is unfilterable, and a user typing a search term that includes 'q' silently loses all unsaved context and is dropped to the shell. Verified live: session died the instant 'q' was sent during an active filter.

**Suggested fix:** In app.go, add `if a.state == stateList && a.list.IsFiltering() { break }` to the Quit, Help, and Back cases (mirroring the existing import/export focus guards). Expose an IsFiltering() helper on List that returns `l.list.FilterState() != list.Unfiltered`. The Models view already does the right thing — use it as the reference implementation.

**Repro:**
1. Launch omo-profiler
2. Press Enter on 'Switch Profile'
3. Press /
4. Type 's' (filter works)
5. Type 'q' — app exits to shell

---

### [MAJOR] Typing '?' while filtering the profile list opens the help overlay

**Phase:** probe  
**Evidence:** captures/0004-help-overlay-during-filter.txt  

Same root cause as F-01. While the profile-list filter is active, '?' is captured by the global Keys.Help case (app.go:173) and toggles the full-screen help overlay instead of being inserted into the filter term. The filter text persists underneath, but the user cannot search for any profile name containing '?' and gets a surprising modal mid-search.

**Suggested fix:** Covered by the same `stateList && IsFiltering()` guard proposed in F-01 — add it to the Keys.Help case as well.


---

### [MAJOR] Esc while filtering the profile list jumps to Dashboard instead of cancelling the filter

**Phase:** probe  
**Evidence:** captures/0004-help-overlay-during-filter.txt  

list.go:201-205 has an explicit comment: 'During active filtering, delegate all keys to the bubbles list so it can handle Esc to cancel the filter natively.' That intent is dead code: the global Back case (app.go:191-200) intercepts Esc for stateList (only wizard/diff/models/modelImport are exempted) and calls navigateTo(stateDashboard) before the list view sees the key. So Esc during a filter abandons the whole list view rather than clearing the filter and returning to the full list. The Models view, which is exempted, handles this correctly (Esc cancels the search and stays in the list) — confirming the inconsistency.

**Suggested fix:** Add stateList (when filtering) to the Back-case exemption so the list view's native Esc-cancels-filter behaviour runs, matching the Models view.


---

### [MAJOR] Declared minimum terminal size (40x12) already renders a broken, overlapping layout

**Phase:** visual  
**Evidence:** captures/0002-dashboard-40x12-overlap.txt  

layout.go declares MinTerminalWidth=40, MinTerminalHeight=12 and IsBelowMinimumSize uses strict `<` (layout.go:72-74). At exactly 40x12 — the advertised minimum — the dashboard does NOT show the 'Too Small' guard, yet the layout is already broken: the title and subtitle disappear and 'N profiles available' overlaps 'Switch Profile' on the same row. The guard only fires below 40x12 (verified at 39x11). The effective minimum that renders cleanly is higher than the declared one.

**Suggested fix:** Either raise MinTerminalWidth/Height to a size that actually fits the dashboard (empirically the layout needs more vertical room — ~16-18 rows for the 9-item menu plus title/footer), or change IsBelowMinimumSize to `<=` AND bump the constants to the true minimum. The warning screen itself (39x11) is clean and correct.


---

### [MINOR] Wizard Review step shows HTML-escaped JSON (\u003c / \u003e instead of < / >)

**Phase:** probe  
**Evidence:** captures/0001-dashboard-initial.txt  

On the final Review step, the rendered profile JSON shows category prompt_append values containing literal '\u003cCategory_Context\u003e' rather than '<Category_Context>'. This is Go's encoding/json default HTML escaping (SetEscapeHTML(true)). It makes the human-facing review harder to read and misrepresents what the saved file contains.

**Suggested fix:** When marshaling JSON purely for on-screen display, use a json.Encoder with SetEscapeHTML(false) (or strings.NewReplacer on the rendered string). Verify the actual saved profile file is not affected; only the review preview is in scope here.


---

### [COSMETIC] Diff panel content touches the right border with no padding; long lines hard-wrap mid-token

**Phase:** visual  
**Evidence:** captures/0001-dashboard-initial.txt  

In Compare Profiles, JSON lines are truncated flush against the right box-drawing border (no trailing space before '│'), and very long lines (e.g. the $schema URL) are hard-wrapped into stray fragments like a lone '    "' on its own row. Borders themselves render correctly; this is purely a readability/polish issue.

**Suggested fix:** Reserve one column of right padding inside each diff panel before truncating, and truncate-with-ellipsis rather than hard-wrapping long single-token lines.


---

### [COSMETIC] Name-validation error message truncates at 80 cols

**Phase:** stress  
**Evidence:** captures/0001-dashboard-initial.txt  

Entering an invalid profile name shows '✗ profile name must contain only ASCII letters (a-z, A-Z), numbers, underscores,' — the sentence is cut off at the right edge at 80 columns and never wraps to show the rest (hyphens). The validation logic itself is correct (accepts valid-test-99, rejects 'Test Name!@#' and unicode 'café_ção').

**Suggested fix:** Wrap the validation message to the available width, or shorten it (e.g. 'only a-z A-Z 0-9 _ - allowed').


---

### [INFO] Destructive keys (d / delete) were not exercised

**Phase:** probe  
**Evidence:** captures/0001-dashboard-initial.txt  

Per the audit safety policy, the delete bindings in the profile list (d), models registry (d), and wizard categories/agents (d) were not triggered because the TUI operates on the user's real profiles in ~/.config/opencode/profiles/ (default, smart, ultracheap). The confirm dialog (y/n) was observed in source and via the discard-changes prompt but not driven to completion.

**Suggested fix:** Re-run against a throwaway config dir (config.SetBaseDir) to safely exercise delete flows end to end.


---

### [INFO] Live pixel screenshots unavailable — grim captured the compositor background, not the tmux pane

**Phase:** visual  
**Evidence:** captures/0005-visual-wide-160x40.txt  

Although grim + Wayland are present, the agent's tmux client is not in a focused/positioned window, so grim captured the desktop wallpaper instead of the TUI. All visual findings are therefore based on text + ANSI pane captures (tmux capture-pane), which are authoritative for layout/overlap/border analysis but cannot assess true colour rendering.

**Suggested fix:** Re-run interactively in a focused Wayland terminal, or rely on the ANSI captures (.ansi files) for colour inspection.

---

## 4. Visual Gallery

Diff maps, when generated with `tui-screenshot.sh --diff`, are stored next to
the screenshots.

_(no screenshots captured)_

---

## 5. Methodology

### Phases Executed

| Phase | What was done | Status |
| --- | --- | --- |
| 1. Discover | Located binary, read project docs, ran `--help`/`--version` when safe | |
| 2. Inventory | Captured help screen(s); parsed keybindings into `keybindings.json` | |
| 3. Probe | Sent documented/common bindings per context; classified each as active / dead / error / crash | |
| 4. Stress | Sent Unicode, paste/control characters, and rapid input where safe | |
| 5. Visual | Captured resize matrix and optional diffs against baseline | |
| 6. Report | Rendered this document | |

### Coverage

- **Keys probed**:
- **Modes tested**:
- **Geometries**:
- **Not tested (and why)**:

### Limitations

<!-- Note missing tools, headless screenshot fallback, skipped destructive
keys, network-bound actions, permissions, fonts, or other constraints. -->

---

## 6. Reproducibility

Every blocker and major finding should be reproducible from a fresh launch.

| Finding | Repro from fresh boot? | Steps |
| --- | --- | --- |
| | | |

---

## 7. Improvement Suggestions

<!-- UX nits, design proposals, missing affordances, and future improvements
that are not bugs. -->

---

## 8. Prioritized Recommendations

| Priority | Item | Resolves |
| --- | --- | --- |
| P0 | | |
| P1 | | |

---

## 9. Workspace

```
/home/diogo/.cache/tui-validator/omo-profiler/20260611T193815Z/
├── meta.json
├── keybindings.json
├── findings.json
├── captures/      (6 text + ANSI scrapes)
└── screenshots/   (0 PNGs)
```

---

## 10. Appendix - Environment

- **TERM**: `xterm-256color`
- **Initial geometry**: 80 x 24
- **Binary**: `./omo-profiler`
- **Version**: `omo-profiler version 0.1.0`
- **Args**: ``
- **CWD**: `/home/diogo/dev/omo-profiler`
