# PROJECT KNOWLEDGE BASE

**Generated:** 2026-01-18
**Commit:** 1625626
**Branch:** main

## OVERVIEW

TUI profile manager for oh-my-opencode configuration files. Go + Bubble Tea + Cobra CLI.

## STRUCTURE

```
omo-profiler/
├── cmd/omo-profiler/     # Entry point (main.go)
├── internal/
│   ├── cli/              # Cobra CLI setup
│   │   └── cmd/          # Subcommands (list, switch, export, import, current, models)
│   ├── tui/              # Bubble Tea TUI (see internal/tui/AGENTS.md)
│   │   └── views/        # View components (see internal/tui/views/AGENTS.md)
│   ├── profile/          # Profile CRUD operations
│   ├── config/           # Config types matching oh-my-opencode.json schema
│   ├── schema/           # JSON schema validation
│   ├── backup/           # Profile backup before switch
│   ├── diff/             # Profile comparison
│   ├── models/           # Model registry management
│   └── testdata/         # Test fixtures
├── Makefile              # build, install, test, lint
└── go.mod                # Module: github.com/diogenes/omo-profiler
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add CLI command | `internal/cli/cmd/` | Create file, register in `internal/cli/root.go` |
| Add TUI view | `internal/tui/views/` | Follow existing view pattern, add state to `app.go` |
| Modify config types | `internal/config/types.go` | Must match oh-my-opencode JSON schema |
| Profile operations | `internal/profile/` | Load, Save, Delete, List, SetActive, GetActive |
| Schema validation | `internal/schema/validator.go` | Uses gojsonschema |

## CODE MAP

| Symbol | Type | Location | Role |
|--------|------|----------|------|
| `App` | struct | `internal/tui/app.go` | Main TUI model, state machine |
| `Profile` | struct | `internal/profile/profile.go` | Profile data + operations |
| `Config` | struct | `internal/config/types.go` | oh-my-opencode config schema |
| `Wizard` | struct | `internal/tui/views/wizard.go` | Multi-step profile creation/edit |
| `Execute()` | func | `internal/cli/root.go` | CLI entry point |

## CONVENTIONS

- **Bubble Tea pattern**: Model-View-Update (Init, Update, View methods)
- **View navigation**: Use typed messages (e.g., `NavToListMsg`, `SwitchProfileMsg`)
- **Styles**: Centralized in `internal/tui/styles.go` using lipgloss
- **Keys**: Global keybindings in `internal/tui/keys.go`
- **Tests**: `*_test.go` alongside source, use testify/assert

## ANTI-PATTERNS

- **DO NOT** modify config types without checking oh-my-opencode schema compatibility
- **DO NOT** use raw strings for colors—use style constants from `styles.go`
- **DO NOT** handle navigation in views—emit messages, let `app.go` handle routing

## COMMANDS

```bash
make build      # Build binary
make install    # Install to ~/.local/bin
make test       # Run all tests
make lint       # Run golangci-lint
```

## NOTES

- Profiles stored in `~/.config/opencode/profiles/*.json`
- Active config: `~/.config/opencode/oh-my-opencode.json`
- Import/Export TUI views not yet implemented (placeholder)
- No CI/CD configured yet
