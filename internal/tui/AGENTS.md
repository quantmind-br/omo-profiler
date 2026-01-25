# TUI Architecture

## OVERVIEW
Bubble Tea application orchestrating navigation, state management, and global UI components.

## STATE MACHINE
Centralized in `App.state` (`app.go`). Transitions via `navigateTo(state)` triggered by messages.

**States:** `stateDashboard` (default), `stateList`, `stateWizard`, `stateDiff`, `stateModels`, `stateModelImport`

## KEY FILES
| File | Role |
|------|------|
| `app.go` | **Main Hub**: State machine, message routing, global layout (Toast, Help). |
| `tui.go` | **Entry**: Initializes `tea.Program` with `NewApp()`. |
| `styles.go` | **Design**: Centralized Lipgloss styles (Palettes, Borders). |
| `keys.go` | **Input**: Global keybindings (`q`, `?`, `esc`) & Help constants. |

## APP STRUCT
The `App` model owns all sub-models and global state:
- **State**: `state`, `prevState`, `ready`, `loading`
- **UI**: `spinner`, `help`, `toast`
- **Views**: `dashboard`, `list`, `wizard`, `diff`, `modelRegistry`

## PATTERNS
- **Navigation**: Views emit `NavTo*Msg` → `App.Update` calls `navigateTo()` → Re-inits view.
- **Async Commands**: Operations (e.g., `doSwitchProfile`) return `tea.Cmd`, results handled via `*DoneMsg`.
- **Toast Notifications**: Global overlay via `showToast()`, auto-clears after duration.
- **Layout**: `App.View()` renders active view + global overlays (Toast, Help) vertically.

## ADDING A VIEW
1. **Create**: `internal/tui/views/myview.go` implementing `Init/Update/View`.
2. **Register**: Add `stateMyView` to `appState` enum in `app.go`.
3. **Embed**: Add `myView views.MyView` to `App` struct.
4. **Route**: Handle `NavToMyViewMsg` in `App.Update` -> `navigateTo(stateMyView)`.
5. **Render**: Add case to `App.View()` switch.
