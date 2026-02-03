# Project Overview: omo-profiler

## Purpose
TUI profile manager for `oh-my-opencode` configuration files. Provides a user-friendly terminal interface for managing configuration profiles, including creation, editing, switching, import/export, and validation.

## Key Features
- **Dashboard**: Overview of active profile with quick actions
- **Profile Wizard**: Step-by-step profile creation/editing (6 steps)
- **Profile List**: View, switch, edit, and delete profiles
- **Diff View**: Side-by-side profile comparison
- **Import/Export**: JSON profile exchange with validation
- **Model Registry**: Manage custom AI models
- **Schema Validation**: Validates against oh-my-opencode schema

## Technology Stack
- **Language**: Go 1.25.6
- **TUI Framework**: Bubble Tea (Charmbracelet)
- **CLI Framework**: Cobra
- **Schema Validation**: gojsonschema
- **Testing**: Standard Go testing + testify

## Project Structure
```
omo-profiler/
├── cmd/omo-profiler/          # Entry point (main.go)
├── internal/
│   ├── cli/                   # Cobra CLI commands
│   │   ├── cmd/               # Individual commands (list, switch, import, etc.)
│   │   └── root.go            # Root command registration
│   ├── config/                # Schema definitions & path resolution
│   │   ├── types.go           # Config struct ecosystem (CRITICAL - matches JSON schema)
│   │   └── paths.go           # Path helpers with test mocking hooks
│   ├── profile/               # CRUD logic & persistence
│   │   ├── profile.go         # Core Profile struct with Load/Save/Delete/List
│   │   ├── active.go          # Active profile switching logic
│   │   └── naming.go          # Profile name validation/sanitization
│   ├── schema/                # JSON schema validation
│   ├── backup/                # Profile backup before switching
│   ├── diff/                  # Profile comparison logic
│   ├── models/                # Model registry management
│   └── tui/                   # Bubble Tea TUI implementation
│       ├── app.go             # Root App model & state machine
│       ├── styles.go          # Shared Lipgloss styles
│       ├── keys.go            # Key bindings
│       └── views/             # Individual TUI views
│           ├── dashboard.go   # Main dashboard
│           ├── list.go        # Profile list
│           ├── wizard.go      # Wizard orchestrator
│           ├── wizard_*.go    # Individual wizard steps
│           ├── diff.go        # Profile diff
│           └── import/export  # Import/export views
└── internal/testdata/         # JSON fixtures for testing

## Important Files
- `internal/config/types.go`: Schema authority - must stay in sync with upstream
- `internal/tui/app.go`: Root TUI state machine & router
- `internal/profile/profile.go`: Core persistence logic
- `Makefile`: Build, test, install automation
