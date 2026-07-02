# omo-profiler — Agent Guide

Compact guidance for OpenCode sessions. For deeper per-package context, see the `AGENTS.md` files under `internal/*/`. For generated architecture detail, see `ARCHITECTURE.md`.

## Project at a Glance

TUI profile manager for `oh-my-openagent` configuration files. Go 1.25.6 + Bubble Tea + Cobra CLI. Manages profile CRUD, active-state switching, and JSON schema validation against upstream.

Entry flow:

```
cmd/omo-profiler/main.go → cli.Execute() → rootCmd.Execute()
  ├── no subcommand → tui.Run() → tea.NewProgram()
  └── subcommand    → internal/cli/cmd/*.go → profile/backup/models/schema packages
```

## Daily Commands

```bash
make build      # go build -v -o omo-profiler ./cmd/omo-profiler
make test       # go test -v ./...  (race detector NOT enabled)
make lint       # golangci-lint run ./...  (requires local install)
make install    # cp binary to ~/.local/bin/omo-profiler
make clean      # rm binary + go clean
```

Run a single package or test:

```bash
go test -v ./internal/profile/...
go test -v -run TestLoad ./internal/profile/
```

There is **no CI, no `.golangci.yml`, no Docker, and no `update-schema.sh`** in this repo. The `update-schema.sh` script referenced in older docs has been removed; schema sync is done through `schema.CompareSchemas()` / `schema.FetchUpstreamSchema()` or manually.

## Layout and Ownership

| Directory | Owns | Notes |
|-----------|------|-------|
| `cmd/omo-profiler/` | Binary entry point | Just calls `cli.Execute()` |
| `internal/cli/cmd/` | Cobra subcommands | 8 commands; thin wrappers only |
| `internal/config/` | Schema authority + paths | `Config` struct (44 top-level fields) is the source of truth |
| `internal/profile/` | CRUD, switching, naming, sparse-field detection | Switching is **copy-based**, never symlink |
| `internal/schema/` | Embedded JSON schema + singleton validator | ~188 KB `schema.json` fetched from upstream oh-my-openagent dev branch |
| `internal/models/` | Local model registry + models.dev API client | `models.json` with auto `.bak` on corruption |
| `internal/backup/` | Timestamped backup rotation before profile switch | |
| `internal/diff/` | Side-by-side + unified diff | Used by profile compare and schema drift |
| `internal/tui/` | Bubble Tea root `App`, styles, layout | |
| `internal/tui/views/` | 18 sub-views incl. 6-step wizard | Complexity hotspots: `wizard_other.go`, `wizard_agents.go`, `wizard_categories.go` |
| `internal/testdata/` | JSON fixtures | `valid-config.json`, `minimal-config.json`, etc. |

## TUI Architecture

- **Pure MVU**: `App.Update` is the only router. Views emit `tea.Msg`, never mutate `App` state.
- **10 states** in `internal/tui/app.go`: Dashboard, List, Wizard, Diff, Import, Export, Models, ModelImport, TemplateSelect, SchemaCheck.
- **Wizard steps**: Name → Categories → Agents → Hooks → Other → Review.
- **Step lifecycle**: `SetConfig(&cfg, selection)` on activation → user edits local state → `Apply(&cfg, selection)` on exit.
- **All I/O in `tea.Cmd`**: never block inside `Update()` or `View()`.
- **Styles**: import from `internal/tui/styles.go`; do not define raw hex colors in views.
- **Navigation**: views emit `NavTo*Msg`; `App` intercepts and calls `navigateTo(state)`. Views are recreated on navigation.

## Schema / Config Rules

- `internal/config/types.go` must stay 1:1 with the upstream JSON schema.
- Every struct field needs `json:"...,omitempty"`.
- Use `*bool` / `*float64` / `*int` to distinguish `false`/`0` from "missing".
- `skills` and `runtime_fallback` are `json.RawMessage` to preserve original shape.
- Validation modes:
  - `Validate()` — strict, enforces required fields.
  - `ValidateForSave()` — permissive, ignores missing required fields; used for wizard save and profile save.
- Always use `schema.GetValidator()` (singleton via `sync.Once`); don't build loaders directly.

## Testing Conventions

- Co-located `*_test.go`.
- Any test touching the filesystem must redirect paths to a temp dir:

  ```go
  func setupTestEnv(t *testing.T) func() {
      tmpDir := t.TempDir()
      config.SetBaseDir(tmpDir)
      return func() { config.ResetBaseDir() }
  }
  ```

- Use `testify/assert` and `testify/require`.

## Current Known Issues

- None. (The pre-existing `TestModelRegistryGetFilteredModelsWithSearch` false-positive — `fuzzy.Find` returning subsequence-only matches like a Qwen model for `"claude"` — was fixed by filtering to substring membership in `model_registry.go:getFilteredModels`.)

## Path Resolution

```go
config.ConfigDir()    // ~/.config/opencode/
config.ProfilesDir()  // ~/.config/opencode/profiles/
config.ConfigFile()   // canonical oh-my-openagent.json, falls back to legacy oh-my-opencode.json
config.ModelsFile()   // ~/.config/opencode/models.json
```

Never hardcode `~/.config`; always use `config.*` helpers.

## GitNexus

This repo is indexed as **omo-profiler**. If a GitNexus tool says the index is stale, run `npx gitnexus analyze` first.

- Impact analysis before editing a symbol: `gitnexus_impact({target: "SymbolName", direction: "upstream"})`
- Check affected flows before committing: `gitnexus_detect_changes()`
- Explore concepts: `gitnexus_query({query: "..."})`
- Rename safely: `gitnexus_rename({symbol_name: "...", new_name: "..."})`
