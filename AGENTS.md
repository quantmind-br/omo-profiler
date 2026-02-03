# PROJECT KNOWLEDGE BASE

**Generated:** 2026-02-01
**Commit:** a4a4353
**Branch:** main

## OVERVIEW

TUI profile manager for `oh-my-opencode` configuration files. Go 1.25.6 + Bubble Tea + Cobra CLI. Manages profile CRUD, active state switching (copy-based, not symlink), and schema validation.

## STRUCTURE

```
omo-profiler/
├── cmd/omo-profiler/     # Entry point
├── internal/
│   ├── cli/              # Cobra commands
│   ├── config/           # Schema definitions & paths
│   ├── profile/          # CRUD & persistence
│   ├── tui/              # Bubble Tea app & state machine
│   │   └── views/        # Wizard, List, Dashboard, etc.
│   ├── schema/           # JSON schema validation
│   ├── models/           # LLM model registry
│   ├── diff/             # Profile comparison
│   ├── backup/           # Profile backups
│   └── testdata/         # JSON fixtures
└── Makefile              # Build automation
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| **Add CLI Command** | `internal/cli/cmd/` | Register in `internal/cli/root.go` |
| **Modify Config Schema** | `internal/config/types.go` | **CRITICAL**: Must match upstream schema |
| **Add TUI View** | `internal/tui/views/` | Register state in `internal/tui/app.go` |
| **Change Profile Logic** | `internal/profile/` | Load/Save/Switch/Delete operations |
| **Update Styles** | `internal/tui/styles.go` | Shared Lipgloss definitions |
| **Add Wizard Step** | `internal/tui/views/wizard_*.go` | Follow implicit interface pattern |

## CODE MAP

| Symbol | Type | Location | Role |
|--------|------|----------|------|
| `App` | struct | `internal/tui/app.go` | Root TUI model & state machine |
| `Profile` | struct | `internal/profile/profile.go` | Core data model with persistence |
| `Config` | struct | `internal/config/types.go` | Root config struct (matches JSON schema) |
| `Wizard` | struct | `internal/tui/views/wizard.go` | Multi-step profile creation |
| `SetBaseDir` | func | `internal/config/paths.go` | Test isolation hook |

## CONVENTIONS

- **Go Version**: Strictly `1.25.6` (go.mod)
- **TUI Architecture**: Pure MVU (Model-View-Update), message-driven navigation
- **Testing**: Tests co-located (`*_test.go`), mandatory `config.SetBaseDir(t.TempDir())`
- **Error Handling**: Bubble Tea commands return `Msg` with error field
- **JSON Tags**: Use `omitempty`, pointers for booleans to distinguish false/missing

## ANTI-PATTERNS (THIS PROJECT)

- **Schema Divergence**: Never add `Config` fields without upstream schema support
- **Direct State Mutation**: Views must NOT modify `App` state; emit messages
- **Global Config Mutation**: Wizard steps must NOT modify `Config` directly; use `Apply()`
- **Hardcoded Paths**: Never use `~/.config` literals; use `config.Paths` helpers
- **Raw Styles**: Avoid hex codes in views; use `internal/tui/styles.go`
- **Symlinking**: Profile switching uses COPY, not symlinks (fsnotify compatibility)

## COMMANDS

```bash
make build       # Build binary to ./omo-profiler
make install     # Install to ~/.local/bin/omo-profiler
make test        # Run tests (race detector enabled)
make lint        # Run golangci-lint
make clean       # Remove build artifacts
```

## NOTES

- Profile storage: `~/.config/opencode/profiles/`
- Active config: `~/.config/opencode/oh-my-opencode.json`
- State tracking: `.active-profile` sidecar file + content fallback
- Wizard steps: Name → Categories → Agents → Hooks → Other → Review
