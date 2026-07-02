# Refactor planning — how to write the plan + JSON schemas

Used in Phase 3 (write `findings.json`) and Phase 6 (write the plan). READ THIS
BEFORE WRITING EITHER JSON FILE.

## `keybindings.json` schema (Phase 2)

A JSON array. One object per current keybinding found in source or live capture:

```json
[
  {
    "key": "q",
    "action": "quit",
    "context": "global",
    "source": "internal/ui/keys.go:42",
    "visible_in_help": true,
    "notes": "also exits detail view"
  }
]
```

Use normalized key names from `tui-validator` when live capture is available
(`C-h`, `Escape`, `BTab`, arrows as `Up`/`Down`/`Left`/`Right`). If the same key
does different things in different contexts, keep one object per context.

## `findings.json` schema (Phase 3)

A JSON array. One object per design smell:

```json
[
  {
    "id": "F001",
    "severity": "blocker | major | minor | cosmetic",
    "principle": "short name of the violated principle, e.g. 'color-only state'",
    "title": "one-line description",
    "evidence": {
      "files": ["internal/ui/view.go:142"],
      "screenshot": "before/profiles-80x24.png"
    },
    "impact": "what this does to the user",
    "fix_direction": "one sentence; full detail goes in the plan item",
    "plan_items": ["P003"]
  }
]
```

`plan_items` links the smell to the plan items that resolve it — used in Phase 7
to prove every finding is addressed (or deferred).

## `plan-items.json` schema (Phase 6)

The plan in `04-refactor-plan.md` is human-readable markdown. Back each item
with a matching object in `plan-items.json` so nothing is hand-wavy:

```json
{
  "id": "P003",
  "milestone": "M1 — design-system foundation",
  "title": "Centralize colors into a theme module with semantic roles",
  "what": "Replace scattered literal colors with named roles (info/success/warning/danger/muted/accent/focus).",
  "where": {
    "files": ["internal/ui/styles.go", "internal/ui/view.go"],
    "symbols": ["lipgloss.Style literals"]
  },
  "how": "Create theme.go exposing semantic styles with truecolor/256/16 tiers; replace literals with theme references.",
  "preserves": ["all current behavior", "all keybindings"],
  "changes_keybindings": false,
  "effort": "M",
  "risk": "low",
  "validation": "build; run tui-validator; diff before/ vs after — colors render with fallbacks at each capability tier.",
  "resolves": ["F004", "F011"]
}
```

`plan-items.json` is a JSON array of these objects, ordered in the same sequence
as the markdown plan.

## Milestones

Group items so each milestone leaves the TUI **buildable and runnable**.
Recommended default order:

1. **M1 — Design-system foundation.** Theme/semantic colors, glyph set,
   capability tiers, border/spacing rhythm. Invisible to workflows; unblocks the
   rest. Low risk.
2. **M2 — Navigation & feedback shell.** Persistent status bar, `?` help
   overlay, focus indicators, loading/empty/error states, async un-blocking.
3. **M3 — Per-screen rework.** Apply target wireframes screen by screen, highest
   traffic first.
4. **M4 — Keybinding consolidation** (only if Phase 4 approved key changes).
   Unify the keymap; ship the change log.

Order within a milestone so shared components land before their consumers.

## Plan-writing rules

- **No source code.** Describe the change precisely; don't write the diff. This
  skill stops at the plan.
- **Every item maps to real files** from `meta.json`. No "somewhere in the UI".
- **Every item is independently mergeable** where possible — small PRs.
- **Flag every keybinding/flag/output change** (`changes_keybindings: true`) so
  the user sees workflow churn up front.
- **Validation routes through tui-validator** when available: the refactor is
  "done" for an item when a fresh audit + a `before/after` screenshot diff
  confirm the intended change and no regression.
- Note in each item what is explicitly **out of scope** per Phase 4.
