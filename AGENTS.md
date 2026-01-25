# PROJECT KNOWLEDGE BASE

**Generated:** 2026-01-25
**Commit:** 97d1ab9
**Branch:** main

## OVERVIEW

TUI profile manager for oh-my-opencode configuration files, built with Go 1.25, Bubble Tea, and Cobra CLI.

## STRUCTURE

```
omo-profiler/
├── cmd/omo-profiler/     # Entry point (main.go → cli.Execute())
├── internal/
│   ├── cli/              # Cobra CLI setup & subcommands
│   ├── tui/              # Bubble Tea TUI (state machine, views)
│   ├── profile/          # Profile CRUD logic (Load, Save, Delete)
│   ├── config/           # Data structures for oh-my-opencode.json schema
│   ├── schema/           # JSON schema validation
│   ├── models/           # Model registry management
│   └── testdata/         # Test fixtures
├── Makefile              # Build, install, test, lint commands
└── go.mod                # Go module definition
```

## WHERE TO LOOK

| Task                  | Location                | Notes                                        |
| --------------------- | ----------------------- | -------------------------------------------- |
| Add a CLI command     | `internal/cli/cmd/`     | Register in `internal/cli/root.go`         |
| Add a TUI view        | `internal/tui/views/`   | Add state to `app.go` and handle navigation  |
| Modify profile logic  | `internal/profile/`     | Filesystem operations, active profile switch |
| Change config schema  | `internal/config/types.go` | **Must** match upstream oh-my-opencode schema |
| Adjust UI styles      | `internal/tui/styles.go`  | Centralized lipgloss styling               |

## CODE MAP

| Symbol       | Type   | Location                  | Role                                     |
| ------------ | ------ | ------------------------- | ---------------------------------------- |
| `App`        | struct | `internal/tui/app.go`     | Main TUI model, state machine, router    |
| `Profile`    | struct | `internal/profile/profile.go` | Profile data and CRUD operations         |
| `Config`     | struct | `internal/config/types.go`  | Root of the 19 nested config structs     |
| `Wizard`     | struct | `internal/tui/views/wizard.go` | Orchestrator for multi-step profile forms |
| `Execute()`  | func   | `internal/cli/root.go`    | CLI entry point                          |

## CONVENTIONS

- **Go Version**: 1.25
- **TUI Pattern**: Follows Bubble Tea's Model-View-Update (MVU) architecture.
- **Navigation**: TUI navigation is message-driven. Views emit `NavTo*Msg` messages, which are handled by the main `App` router.
- **Testing**: Tests are co-located with source files. Filesystem is mocked using `t.TempDir()` and `config.SetBaseDir()`.

## ANTI-PATTERNS (THIS PROJECT)

- **DO NOT** handle navigation logic inside views.
- **DO NOT** use raw color strings; use constants from `internal/tui/styles.go`.
- **DO NOT** modify `internal/config/types.go` without checking for schema compatibility.
- **CI/CD**: The project currently lacks any automated CI/CD pipeline.

## COMMANDS

```bash
make build      # Build binary to ./omo-profiler
make install    # Install to ~/.local/bin
make test       # Run all tests
make lint       # Run golangci-lint
```
