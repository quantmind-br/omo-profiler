# Keybinding conventions

Used in Phase 5 to design the target keymap and in Phase 6 to flag every change
against the current map. In a refactor, a key change is a cost — list it.

## Reserved keys (never repurpose)

| Key     | Meaning                                  |
| ------- | ---------------------------------------- |
| `?`     | Open help overlay                        |
| `q`     | Quit (or back out one level; confirm if unsaved) |
| `esc`   | Cancel / close modal / clear search      |
| `/`     | Search / filter (incremental)            |
| `:`     | Command palette (if the app has one)     |
| `Enter` | Activate / drill into / confirm          |
| `Tab` / `Shift-Tab` | Cycle focus forward / back   |

## Navigation — always provide BOTH

| Action        | Vim | Arrow      |
| ------------- | --- | ---------- |
| Up / Down     | `k`/`j` | `↑`/`↓` |
| Left / Right  | `h`/`l` | `←`/`→` |
| Page up/down  | `Ctrl-b`/`Ctrl-f` | `PgUp`/`PgDn` |
| Top / Bottom  | `g`/`G` | `Home`/`End` |

Never ship vim-only or arrow-only.

## Action-verb vocabulary (consistent across the app)

| Verb        | Key  | Notes                                        |
| ----------- | ---- | -------------------------------------------- |
| Add / new   | `a` / `n` | pick one, use it everywhere             |
| Edit        | `e`  |                                              |
| Delete      | `d` / `x` | always confirms; never default-focus delete |
| Rename      | `r`  |                                              |
| Refresh     | `R` / `Ctrl-r` |                                    |
| Filter      | `/`  | incremental                                  |
| Toggle      | `Space` | select / expand / check                   |
| Confirm     | `Enter` |                                            |
| Cancel      | `esc` |                                             |
| Help        | `?`  |                                              |
| Quit        | `q` / `Ctrl-c` |                                    |

## Contextual overrides

A key may mean different things in different **contexts** only when the context
is unambiguous (e.g. inside a text input, letter keys type rather than trigger).
Document every such override in `04-keybindings.md`; never silently overload a
global key.

## Conflict audit

Build a table of (context × key). Flag:
- the same key bound to two actions in one context;
- a global key shadowed by a screen with a different meaning (allowed only if
  documented and clearly contextual);
- a reserved key doing something non-standard.

## Refactor change log

In `04-keybindings.md`, add a **Changes vs current** table: `key | was | now |
reason | muscle-memory cost`. Phase 4 decides whether each change ships.
