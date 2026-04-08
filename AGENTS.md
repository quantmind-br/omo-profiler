# PROJECT KNOWLEDGE BASE

**Generated:** 2026-04-08
**Commit:** ffb9a49
**Branch:** main

## OVERVIEW

TUI profile manager for `oh-my-openagent` configuration files. Go 1.25.6 + Bubble Tea + Cobra CLI. Manages profile CRUD, active state switching (copy-based, not symlink), sparse field serialization, and JSON schema validation against upstream.

## STRUCTURE

```
omo-profiler/
├── cmd/omo-profiler/     # Entry point → cli.Execute()
├── internal/
│   ├── cli/              # Cobra root + cmd/ subcommands (8 commands + models sub-subcommands)
│   ├── config/           # Schema authority (types.go, ~37 top-level fields) + path resolution
│   ├── profile/          # CRUD, switching, naming, sparse serialization (sparse.go)
│   ├── tui/              # Bubble Tea app (state machine + router, app.go ~872 lines)
│   │   ├── views/        # 18 view files + 10 test files: wizard steps, dashboard, diff, models
│   │   └── layout/       # Terminal width helpers, min-size constants
│   ├── schema/           # Embedded JSON schema (go:embed) + gojsonschema validator (singleton)
│   ├── models/           # LLM model registry + models.dev API client
│   ├── diff/             # Side-by-side + unified diff (go-diff wrapper)
│   ├── backup/           # Timestamped backup rotation + restore/clean
│   └── testdata/         # JSON fixtures for cross-package tests (4 files)
├── template/             # Default oh-my-openagent.json template
└── Makefile              # build, install, test, lint, clean
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| **Add CLI Command** | `internal/cli/cmd/` | Export var, register in `cli/root.go` `init()` |
| **Modify Config Schema** | `internal/config/types.go` | **CRITICAL**: Must match upstream JSON schema 1:1 |
| **Add TUI View** | `internal/tui/views/` | Add `appState` const + register in `app.go` `navigateTo()` |
| **Add Wizard Step** | `internal/tui/views/wizard_*.go` | Implement `WizardStep` interface (`Init`/`SetSize`/`View`) + implicit `SetConfig`/`Apply` |
| **Change Profile Logic** | `internal/profile/` | Load/Save/Switch/Delete + naming in `naming.go`, sparse serialization in `sparse.go` |
| **Sparse Serialization** | `internal/profile/sparse.go` | `MarshalSparse` — writes only user-selected fields, preserving unknown keys |
| **Update Styles** | `internal/tui/styles.go` | Shared Lipgloss palette; views MUST import from here |
| **Update Models** | `internal/models/` | `models.go` (local CRUD) / `modelsdev.go` (API client) |
| **Schema Validation** | `internal/schema/validator.go` | Singleton via `GetValidator()`, `//go:embed schema.json` |
| **Layout Constants** | `internal/tui/layout/layout.go` | `MinTerminalWidth/Height`, responsive field widths |

## CODE MAP

| Symbol | Type | Location | Role |
|--------|------|----------|------|
| `App` | struct | `tui/app.go` | Root TUI model, state machine, message router, toast/spinner overlays |
| `appState` | enum | `tui/app.go` | 10 states: Dashboard→List→Wizard→Diff→Import→Export→Models→ModelImport→TemplateSelect→SchemaCheck |
| `Wizard` | struct | `tui/views/wizard.go` | Multi-step profile creation orchestrator |
| `WizardStep` | interface | `tui/views/step.go` | `Init`, `SetSize`, `View` (explicit) + `SetConfig`/`Apply` (implicit) |
| `Config` | struct | `config/types.go` | Root config (~37 top-level fields, 33+ nested struct types) |
| `Profile` | struct | `profile/profile.go` | Wraps `Config` + metadata (Name, Path, PreservedUnknown, FieldPresence, LegacyFields) |
| `MarshalSparse` | func | `profile/sparse.go` | Selective field serialization — only writes user-selected fields |
| `SetBaseDir` | func | `config/paths.go` | Test isolation: redirects all paths to temp dir |
| `ModelsRegistry` | struct | `models/models.go` | Thread-safe local model CRUD with JSON persistence |
| `Validator` | struct | `schema/validator.go` | Singleton JSON schema validator (`sync.Once` + `go:embed`) |
| `ComputeDiff` | func | `diff/diff.go` | Side-by-side + unified diff with `DiffLine` typed results |
| `Create`/`Restore`/`Clean` | funcs | `backup/backup.go` | Timestamped backup rotation, restore, and cleanup |

## CONVENTIONS

- **Go Version**: `1.25.6` (go.mod)
- **TUI Architecture**: Pure MVU — views emit `tea.Msg`, `App.Update` routes, `tea.Cmd` for async
- **Navigation**: Message-driven — views emit `NavTo*Msg`, App intercepts and calls `navigateTo(state)`
- **Testing**: Co-located `*_test.go` (36 files), mandatory `config.SetBaseDir(t.TempDir())` via `setupTestEnv` helper, `t.Helper()` on all helpers
- **Test Assertions**: Mix of standard `t.Errorf` (predominant) and `github.com/stretchr/testify` (`require`/`assert`) in ~4 files
- **Test Fixtures**: `internal/testdata/` with 4 JSON fixtures (valid-config, minimal-config, complex-permissions, skills-object)
- **Table-Driven Tests**: Standard `[]struct{name string; ...}` + `t.Run(tt.name, ...)` in 14+ files
- **Error Handling**: `tea.Cmd` returns `*Msg` with `err` field; CLI uses `RunE` for Cobra error propagation
- **JSON Tags**: `omitempty` required; `*bool` pointers to distinguish `false` from missing
- **Profile Naming**: Strict regex `^[a-zA-Z0-9_-]+$` enforced by `profile.SanitizeName`
- **No Build Tags**: No `//go:build` or `// +build` directives anywhere
- **No Lint Config**: golangci-lint runs with defaults (no `.golangci.yml`)

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
- **Deprecated**: `NewValidator()` — use `GetValidator()` for singleton access

## COMMANDS

```bash
make build       # Build binary to ./omo-profiler
make install     # Install to ~/.local/bin/omo-profiler
make uninstall   # Remove from ~/.local/bin
make test        # Run tests with go test -v ./...
make lint        # Run golangci-lint (checks for presence first)
make clean       # Remove build artifacts + go clean
```

## EXECUTION FLOW

```
main.go → cli.Execute() → rootCmd.Execute()
  ├── (no subcommand) → tui.Run() → NewApp() → tea.NewProgram(WithAltScreen)
  ├── list             → profile.List()
  ├── current          → profile.GetActive()
  ├── switch <name>    → backup.Create + profile.SetActive
  ├── import <file>    → profile.Save (with schema validation)
  ├── export <name>    → profile.Load + JSON write
  ├── create <name>    → profile.Load + profile.Save (headless, --from flag)
  ├── models           → 4 sub-subcommands: list/add/edit/delete
  └── schema-check     → schema.CompareSchemas()
```

## NOTES

- Profile storage: `~/.config/opencode/profiles/<name>.json`
- Active config: `~/.config/opencode/oh-my-openagent.json`
- State tracking: `.active-profile` sidecar (O(1) lookup) + content scan fallback (O(N))
- Wizard steps: Name → Categories → Agents → Hooks → Other → Review
- Sparse serialization: `sparse.go` writes only fields user modified, preserving unknown JSON keys from upstream
- No CI/CD: No GitHub Actions, no Docker, no Goreleaser
- Embedded schema: `internal/schema/schema.json` (6067 lines, `go:embed`)
- `models.json` corruption: auto-backup to `.bak` on load failure
- Upstream schema URL: `https://raw.githubusercontent.com/code-yeongyu/oh-my-openagent/dev/assets/oh-my-opencode.schema.json`
- HTTP mocking in tests: `httptest.NewServer` pattern for schema compare tests
