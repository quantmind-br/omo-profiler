# TUI Architecture

## OVERVIEW

Bubble Tea application with state machine navigation. Main model in `app.go`.

## STATE MACHINE

```
stateDashboard ←→ stateList ←→ stateWizard
      ↓              ↓
  stateModels    stateDiff
```

## KEY FILES

| File | Purpose |
|------|---------|
| `app.go` | Main model, state transitions, message routing |
| `tui.go` | `Run()` entry point, tea.Program setup |
| `styles.go` | All lipgloss styles (Purple, Cyan, etc.) |
| `keys.go` | Global keybindings (Quit, Help, Back) |

## PATTERNS

- **Navigation**: Views emit `NavTo*Msg`, `app.go` handles state change
- **Async ops**: Return `tea.Cmd`, handle result via typed message
- **Toast**: `showToast(text, type, duration)` → `toastMsg` → auto-clear
- **Loading**: Set `loading=true`, show spinner, process in background

## ADDING A VIEW

1. Create `views/myview.go` with `Model`, `Init`, `Update`, `View`
2. Add state constant in `app.go` (e.g., `stateMyView`)
3. Add field to `App` struct
4. Handle `NavToMyViewMsg` in `Update`
5. Add case in `View()` switch
