# Architecture & Patterns

## TUI Architecture (Bubble Tea MVU)
The TUI follows the Model-View-Update pattern:

### State Machine
- **Centralized State**: `App` struct in `internal/tui/app.go` holds all state
- **State Enum**: `appState` defines all view states (stateDashboard, stateList, etc.)
- **Transitions**: Via `navigateTo(state)` which re-initializes target views

### Message-Driven Navigation
Views NEVER mutate App state directly. They emit messages:
```go
// View emits message
case key.Matches(msg, Keys.Enter):
    return m, func() tea.Msg { return NavToWizardMsg{} }

// App intercepts and handles navigation
case views.NavToWizardMsg:
    a.wizard = views.NewWizard()
    return a.navigateTo(stateWizard)
```

### Global UI Components
- **Toast System**: `showToast(text, type, duration)` - auto-clearing notifications
- **Help Bubble**: Context-aware help (short vs full) based on current state
- **Loading Spinner**: Overlay during async operations

## Wizard Pattern
Multi-step form with implicit interface:
```go
// Each step implements:
SetConfig(*config.Config)  // Load state from config
Apply(*config.Config)      // Save state to config
SetSize(w, h)              // Responsive layout
Init() tea.Cmd             // Lifecycle hook
```

Steps: Name → Categories → Agents → Hooks → Other → Review

## Config Schema Safety
`internal/config/types.go` is the schema authority:
- JSON tags must match upstream schema exactly
- Use pointers (`*bool`) to distinguish false from missing
- Keep structs pure - no methods, only data
- Use `json.RawMessage` for flexible fields like `skills`

## Testing Patterns
- Tests co-located with source (`*_test.go`)
- **MANDATORY**: Use `config.SetBaseDir(t.TempDir())` for filesystem isolation
- Always call `defer config.ResetBaseDir()` or use cleanup helper
- Example:
```go
func setupTestEnv(t *testing.T) func() {
    tmpDir := t.TempDir()
    config.SetBaseDir(tmpDir)
    return func() { config.ResetBaseDir() }
}
```

## Profile Persistence
- Storage: `~/.config/opencode/profiles/<name>.json`
- Active config: `~/.config/opencode/oh-my-opencode.json`
- Switching: **COPY** content, don't symlink (for fsnotify compatibility)
- State tracking: Hidden `.active-profile` sidecar file + content fallback

## Anti-Patterns (NEVER DO)
- Direct state mutation in views (always use messages)
- Blocking I/O in Update() (use tea.Cmd)
- Hardcoded paths (use `config.Paths` helpers)
- Raw hex colors in views (use `internal/tui/styles.go`)
- Modifying config.Config directly in wizard steps (use Apply())
