# PROJECT KNOWLEDGE BASE

**Generated:** 2026-02-11
**Commit:** cb84c43
**Branch:** main

## OVERVIEW

TUI profile manager for `oh-my-opencode` configuration files. Go 1.25.6 + Bubble Tea + Cobra CLI. Manages profile CRUD, active state switching (copy-based, not symlink), and JSON schema validation against upstream.

## STRUCTURE

```
omo-profiler/
├── cmd/omo-profiler/     # Entry point → cli.Execute()
├── internal/
│   ├── cli/              # Cobra root + cmd/ subcommands
│   ├── config/           # Schema authority (types.go) + path resolution
│   ├── profile/          # CRUD, switching, naming validation
│   ├── tui/              # Bubble Tea app (state machine + router)
│   │   ├── views/        # 18 view files: wizard steps, dashboard, diff, models
│   │   └── layout/       # Terminal width helpers, min-size constants
│   ├── schema/           # Embedded JSON schema + gojsonschema validator (singleton)
│   ├── models/           # LLM model registry + models.dev API client
│   ├── diff/             # Side-by-side + unified diff (go-diff wrapper)
│   ├── backup/           # Timestamped backup rotation before profile switch
│   └── testdata/         # JSON fixtures for cross-package tests
├── Makefile              # build, install, test, lint, clean
└── update-schema.sh      # Downloads upstream schema + generates diff report
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| **Add CLI Command** | `internal/cli/cmd/` | Export var, register in `cli/root.go` `init()` |
| **Modify Config Schema** | `internal/config/types.go` | **CRITICAL**: Must match upstream JSON schema 1:1 |
| **Add TUI View** | `internal/tui/views/` | Add `appState` const + register in `app.go` `navigateTo()` |
| **Add Wizard Step** | `internal/tui/views/wizard_*.go` | Implement `WizardStep` interface + implicit `SetConfig`/`Apply` |
| **Change Profile Logic** | `internal/profile/` | Load/Save/Switch/Delete + naming in `naming.go` |
| **Update Styles** | `internal/tui/styles.go` | Shared Lipgloss palette; views MUST import from here |
| **Update Models** | `internal/models/` | `models.go` (local CRUD) / `modelsdev.go` (API client) |
| **Schema Validation** | `internal/schema/validator.go` | Singleton via `GetValidator()`, embedded schema |
| **Layout Constants** | `internal/tui/layout/layout.go` | `MinTerminalWidth/Height`, responsive field widths |

## CODE MAP

| Symbol | Type | Location | Role |
|--------|------|----------|------|
| `App` | struct | `tui/app.go` | Root TUI model, state machine, message router |
| `appState` | enum | `tui/app.go` | 10 states: Dashboard→List→Wizard→Diff→Import→Export→Models→ModelImport→TemplateSelect→SchemaCheck |
| `Wizard` | struct | `tui/views/wizard.go` | Multi-step profile creation orchestrator |
| `WizardStep` | interface | `tui/views/step.go` | `Init`, `SetSize`, `View` (explicit) + `SetConfig`/`Apply` (implicit) |
| `Config` | struct | `config/types.go` | Root config (~25 top-level fields, deeply nested) |
| `Profile` | struct | `profile/profile.go` | Wraps `Config` + metadata (Name, Path, LegacyFields) |
| `SetBaseDir` | func | `config/paths.go` | Test isolation: redirects all paths to temp dir |
| `ModelsRegistry` | struct | `models/models.go` | Thread-safe local model CRUD with JSON persistence |
| `Validator` | struct | `schema/validator.go` | Singleton JSON schema validator (`sync.Once`) |
| `ComputeDiff` | func | `diff/diff.go` | Side-by-side diff with `DiffLine` typed results |
| `Create` | func | `backup/backup.go` | Timestamped backup before profile switch |

## CONVENTIONS

- **Go Version**: `1.25.6` (go.mod)
- **TUI Architecture**: Pure MVU — views emit `tea.Msg`, `App.Update` routes, `tea.Cmd` for async
- **Navigation**: Message-driven — views emit `NavTo*Msg`, App intercepts and calls `navigateTo(state)`
- **Testing**: Co-located `*_test.go`, mandatory `config.SetBaseDir(t.TempDir())`, `setupTestEnv` helper pattern
- **Error Handling**: `tea.Cmd` returns `*Msg` with `err` field; CLI uses `RunE` for Cobra error propagation
- **JSON Tags**: `omitempty` required; `*bool` pointers to distinguish `false` from missing
- **Assertions**: `github.com/stretchr/testify` for test assertions
- **Profile Naming**: Strict regex `^[a-zA-Z0-9_-]+$` enforced by `profile.SanitizeName`

## ANTI-PATTERNS (THIS PROJECT)

- **Schema Divergence**: Never add `Config` fields without upstream schema support
- **Direct State Mutation**: Views must NOT modify `App` state; always emit `tea.Msg`
- **Global Config Mutation**: Wizard steps must NOT modify `Config` directly; use `Apply()` pattern
- **Hardcoded Paths**: Never use `~/.config` literals; use `config.ConfigDir()` / `config.ProfilesDir()`
- **Raw Styles**: Never define hex colors in views; import from `internal/tui/styles.go`
- **Symlinking**: Profile switching uses COPY, not symlinks (fsnotify compatibility)
- **Blocking in Update**: File I/O must happen in `tea.Cmd`, never in `Update()` or `View()`
- **Fat CLI Commands**: CLI `Run` must delegate to `profile`/`backup` packages; no business logic inline
- **Type Suppression**: No `//nolint`, no type assertion shortcuts in production code

## COMMANDS

```bash
make build       # Build binary to ./omo-profiler
make install     # Install to ~/.local/bin/omo-profiler
make uninstall   # Remove from ~/.local/bin
make test        # Run tests with -v (race detector enabled)
make lint        # Run golangci-lint
make clean       # Remove build artifacts
```

## EXECUTION FLOW

```
main.go → cli.Execute() → rootCmd.Execute()
  ├── (no subcommand) → tui.Run() → NewApp() → tea.NewProgram()
  └── (subcommand) → cmd/*.go → profile/backup packages
```

## NOTES

- Profile storage: `~/.config/opencode/profiles/<name>.json`
- Active config: `~/.config/opencode/oh-my-opencode.json`
- State tracking: `.active-profile` sidecar (O(1) lookup) + content scan fallback (O(N))
- Wizard steps: Name → Categories → Agents → Hooks → Other → Review
- No CI/CD: No GitHub Actions, no Docker, no Goreleaser
- Schema sync: `./update-schema.sh` fetches from `code-yeongyu/oh-my-opencode`
- `models.json` corruption: auto-backup to `.bak` on load failure
 Upstream schema URL: `https://raw.githubusercontent.com/code-yeongyu/oh-my-opencode/dev/assets/oh-my-opencode.schema.json`