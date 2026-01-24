# PROJECT KNOWLEDGE BASE

**Generated:** 2026-01-19
**Commit:** 5639a48
**Branch:** main

## OVERVIEW

TUI profile manager for oh-my-opencode configuration files. Go 1.25 + Bubble Tea + Cobra CLI.

## STRUCTURE

```
omo-profiler/
├── cmd/omo-profiler/     # Entry point (main.go → cli.Execute())
├── internal/
│   ├── cli/              # Cobra CLI setup
│   │   └── cmd/          # Subcommands (list, switch, export, import, current, models)
│   ├── tui/              # Bubble Tea TUI (see internal/tui/AGENTS.md)
│   │   └── views/        # View components (see internal/tui/views/AGENTS.md)
│   ├── profile/          # Profile CRUD: Load, Save, Delete, List, Exists
│   ├── config/           # Config types matching oh-my-opencode.json schema
│   ├── schema/           # JSON schema validation (gojsonschema)
│   ├── backup/           # Profile backup before switch
│   ├── diff/             # Profile comparison
│   ├── models/           # Model registry management
│   └── testdata/         # Test fixtures (valid-config.json, minimal-config.json)
├── Makefile              # build, install, test, lint
└── go.mod                # Module: github.com/diogenes/omo-profiler
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add CLI command | `internal/cli/cmd/` | Create file, register in `internal/cli/root.go` |
| Add TUI view | `internal/tui/views/` | Follow existing view pattern, add state to `app.go` |
| Modify config types | `internal/config/types.go` | Must match oh-my-opencode JSON schema |
| Profile operations | `internal/profile/` | Load, Save, Delete, List, Exists |
| Schema validation | `internal/schema/validator.go` | Use `GetValidator()` singleton |
| Model registry | `internal/models/` | models.go (prod), modelsdev.go (dev) |

## CODE MAP

| Symbol | Type | Location | Role |
|--------|------|----------|------|
| `App` | struct | `internal/tui/app.go` | Main TUI model, state machine |
| `Profile` | struct | `internal/profile/profile.go` | Profile data + CRUD operations |
| `Config` | struct | `internal/config/types.go` | oh-my-opencode config schema (18 nested structs) |
| `Wizard` | struct | `internal/tui/views/wizard.go` | Multi-step profile creation/edit orchestrator |
| `Execute()` | func | `internal/cli/root.go` | CLI entry point |
| `NewApp()` | func | `internal/tui/app.go` | TUI constructor |

## CONVENTIONS

- **Bubble Tea pattern**: Model-View-Update (Init, Update, View methods)
- **View navigation**: Emit typed messages (e.g., `NavToListMsg`, `SwitchProfileMsg`), `app.go` routes
- **Styles**: Centralized in `internal/tui/styles.go` using lipgloss
- **Keys**: Global keybindings in `internal/tui/keys.go`
- **Tests**: `*_test.go` alongside source, use `t.TempDir()` + `config.SetBaseDir()` for I/O mocking
- **Fixtures**: JSON test data in `internal/testdata/`

## ANTI-PATTERNS

- **DO NOT** modify `internal/config/types.go` without checking oh-my-opencode schema compatibility
- **DO NOT** use raw strings for colors—use style constants from `styles.go`
- **DO NOT** handle navigation in views—emit messages, let `app.go` handle routing
- **DO NOT** use deprecated validator entry point—use `GetValidator()` singleton

## COMMANDS

```bash
make build      # Build binary → ./omo-profiler
make install    # Install to ~/.local/bin
make test       # Run all tests (go test -v ./...)
make lint       # Run golangci-lint
make clean      # Remove build artifacts
```

## NOTES

- Profiles stored in `~/.config/opencode/profiles/*.json`
- Active config: `~/.config/opencode/oh-my-opencode.json`
- Root command (no args) launches TUI; subcommands run CLI mode
- Import/Export TUI views not yet implemented (placeholder)
- No CI/CD configured yet
- `wizard_hooks.go` has pending TODO: "todo-continuation-enforcer"
