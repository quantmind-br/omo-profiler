# TUI ARCHITECTURE

## OVERVIEW

Built with **Bubble Tea** using **Model-View-Update (MVU)** pattern. Root `App` model (`app.go`) acts as state container, router, and composition layer for all sub-views.

## STATE MACHINE

Application state centralized in `App` struct and `appState` enum:
- **State Enum**: `stateDashboard`, `stateList`, `stateWizard`, etc.
- **Transitions**: Controlled via `navigateTo(state)` which:
  1. Updates `App.state`
  2. Re-initializes target view model
  3. Returns view's `Init()` command

## GLOBAL COMPONENTS

Global UI elements overlay active view in `App.View()`:

1. **Toast System**:
   - Triggered via `showToast(text, type, duration)` cmd
   - Renders at bottom of viewport
   - Auto-clears via `tea.Tick` command

2. **Help Bubble**:
   - Managed globally by `help.Model`
   - Context-aware: `renderShortHelp()` vs `renderFullHelp()` based on `App.state`

3. **Loading Spinner**:
   - Activated by setting `App.loading = true`
   - Replaces content with spinner overlay during async operations

## NAVIGATION FLOW

Navigation is **message-driven** to decouple views from router:
1. **View** emits `NavTo*Msg` (e.g., `NavToWizardMsg`)
2. **App.Update** intercepts message
3. **App** calls `navigateTo(newState)`
4. **App** updates `activeView` field

## ANTI-PATTERNS

- **NO Direct State Mutation**: Sub-views must NEVER modify `App` state; always return `Msg`
- **NO Blocking Operations**: File I/O must happen in `tea.Cmd`, never in `Update()`
- **NO Persisted View State**: Views are re-created on navigation; don't rely on state persistence
