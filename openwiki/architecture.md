# Architecture

omo-profiler follows a **layered, message-driven** architecture. The system has three primary layers: **Presentation** (TUI + CLI), **Business Logic** (profile CRUD, schema, model registry), and **Infrastructure** (config types, omo document, paths, diff, backup).

## Module Dependency Graph

```
┌──────────────────────────────┐
│        Presentation          │
│  ┌──────────┐ ┌───────────┐  │
│  │ CLI      │ │ TUI App   │  │
│  │ (cobra)  │ │ (bubble   │  │
│  │ 9 cmds   │ │  tea)     │  │
│  └────┬─────┘ └─────┬─────┘  │
│       │             │        │
│  ┌────┴─────────────┴────┐   │
│  │     Views (18)        │   │
│  │  dashboard, wizard,   │   │
│  │  list, diff, models…  │   │
│  └──────────┬────────────┘   │
└─────────────┼────────────────┘
              │
┌─────────────┼────────────────┐
│  Business   │   Logic        │
│  ┌──────────┴────────────┐   │
│  │  profile/             │   │
│  │  (CRUD, env activate, │   │
│  │   sparse, naming)     │   │
│  └──────────┬────────────┘   │
│  ┌──────────┴────────────┐   │
│  │  schema/              │   │
│  │  (validator, compare) │   │
│  └──────────┬────────────┘   │
│  ┌──────────┴────────────┐   │
│  │  models/              │   │
│  │  (registry, api)      │   │
│  └───────────────────────┘   │
└─────────────┬────────────────┘
              │
┌─────────────┼────────────────┐
│ Infrastructure               │
│  ┌──────────┴────────────┐   │
│  │  config/              │   │
│  │  (types, Document,    │   │
│  │   paths)              │   │
│  └──────────┬────────────┘   │
│  ┌──────────┴────────────┐   │
│  │  diff/                │   │
│  │  (side-by-side,       │   │
│  │   unified)            │   │
│  └──────────┬────────────┘   │
│  ┌──────────┴────────────┐   │
│  │  backup/              │   │
│  │  (timestamped         │   │
│  │   rotation)           │   │
│  └───────────────────────┘   │
└──────────────────────────────┘
```

**Key dependency patterns:**

- **All business logic depends on `config/`** — `Config` is the `[opencode]` data contract; `Document` is the omo file
- **TUI is the orchestration hub** — `App.Update` dispatches to all business packages via `tea.Cmd`
- **CLI is thin** — commands delegate directly to `profile/`, `backup/`, `models/`, `schema/`
- **`diff/` is a utility** — used by both `profile/` (sparse detection) and `schema/` (upstream drift)
- **Views are independent** — each view is a self-contained Bubble Tea model with its own state, update loop, and view rendering

## Entry Points

### TUI (default)

```
cmd/omo-profiler/main.go
  → cli.Execute()
    → rootCmd.Execute()
      → no subcommand → tui.Run()
        → tea.NewProgram(&App{...})
```

Source: `/cmd/omo-profiler/main.go`, `/internal/cli/root.go` (lines 21-27), `/internal/tui/tui.go`

### CLI Subcommand

```
cmd/omo-profiler/main.go
  → cli.Execute()
    → rootCmd.Execute() with subcommand
      → e.g., cmd.ListCmd.Run → profile.List() → print
```

Source: `/internal/cli/cmd/*.go`

### Web UI

```
cmd/omo-profiler/main.go
  → cli.Execute()
    → web subcommand
      → Serve(opts)
        → newMux() → http.Serve
```

Source: `/internal/cli/cmd/web.go`, `/internal/web/server.go`

## Package Ownership

| Directory | Owns | Notes |
|-----------|------|-------|
| `cmd/omo-profiler/` | Binary entrypoint | Minimal — calls `cli.Execute()` |
| `internal/cli/cmd/` | Cobra subcommands | 9 commands; thin wrappers delegating to business packages |
| `internal/config/` | `[opencode]` types + `Document` + paths | 46 top-level fields on `Config`; `OmoFile` / `SetBaseDir` for isolation |
| `internal/profile/` | Profile CRUD, in-document activation, naming, sparse | `Apply` substitutes profile keys into the root; `ActiveName` detects the applied profile by root comparison |
| `internal/schema/` | Embedded omo document schema + validator | `GetOpenCodeSchema()` for forms; upstream drift vs `assets/omo.schema.json` |
| `internal/models/` | Model registry + models.dev API | `~/.omo/models.json` with auto `.bak` corruption recovery |
| `internal/backup/` | Timestamped backup rotation | Before mutating omo writes (not for switch) |
| `internal/diff/` | Side-by-side + unified diff | `go-diff` wrapper |
| `internal/web/` | HTTP server + JSON API + embedded React SPA | Reuses all business packages unchanged |
| `internal/tui/` | Bubble Tea root App, styles, layout | 10-state state machine |
| `internal/tui/views/` | 18 sub-views (6-step wizard, etc.) | Complexity hotspots: wizard_other, wizard_agents, wizard_categories |
| `internal/testdata/` | JSON test fixtures | `valid-config.json`, `minimal-config.json`, etc. |

## Key Execution Flows

### 1. Profile Switching

**Trigger:** User selects a profile in List view → `App` emits `doSwitchProfile`

```
App.doSwitchProfile
  → profile.Apply(name)
    → doc.ProfileBlock(name)                   — load the profile block
    → ActiveName(doc)                          — is the root already this profile?
    → if no match and a declared key is non-empty at root:
        snapshot those root keys as profiles.base (collides to base-1, …)
    → doc.SetRaw(key, value) for each declared key
    → doc.EnsureSchema()
    → backup.CreateOmoIfPresent (pre-save hook)
  → UI shows success toast (no shell command)
```

Critical constraint: Activation is **in-document**. `Apply` substitutes profile keys verbatim over the root — no merge, no copy, no symlink. The snapshot step (`profiles.base`) is what makes a sparse root recoverable.

### 2. Active Profile Detection

**Trigger:** Dashboard or List view initializes

```
Dashboard / List
  → profile.GetActive()
    → ActiveName(doc)                          — first profile whose every declared key
                                                  canonicalizes equal to the root
    → decode root [opencode] straight into Config (no merge)
    → Returns ActiveConfig{Exists, Config, ProfileName, Modified}
```

`Modified` is true when the root matches no profile — a hand-edited or never-saved configuration.

### 3. Upstream Schema Drift Detection

**Trigger:** User runs `schema-check` command or TUI view

```
SchemaCheck
  → schema.CompareSchemas()
    → FetchUpstreamSchema()             — HTTP GET assets/omo.schema.json (30s timeout)
    → GetEmbeddedSchema()               — read embedded omo document bytes
    → bytes.Equal → if drift detected → diff.ComputeUnifiedDiff()
  → SaveDiff(dir, content)              — persist .diff report
```

### 4. Wizard Model Save

**Trigger:** User adds a custom model in wizard Categories or Agents step

```
WizardCategories / WizardAgents
  → modelsRegistry.Add(model)           — duplicate (Provider, ModelID) check
  → modelsRegistry.Save()               — marshal to ~/.omo/models.json
```

Corruption handling: if `models.json` fails to parse on load, it's backed up to `.bak` and an empty registry is returned.

### 5. Profile Diff Visualization

**Trigger:** User selects "Compare" in Dashboard

```
Diff.computeDiff(left, right)
  → diff.ComputeDiff(json1, json2)      — side-by-side aligned lines
    → diffmatchpatch.DiffMain
    → DiffCleanupSemantic
    → buildDiffResult with aligned left/right arrays
```

## Design Principles

- **Pure MVU in TUI**: Views emit `tea.Msg`, never mutate App state. All I/O in `tea.Cmd`.
- **Views are recreated on navigation**: No persisted view state between navigations.
- **Config is the `[opencode]` contract**: All business packages depend on `config.Config` for the editable flat block; the whole file is `config.Document`.
- **Schema-driven editor**: Web/TUI forms use `GetOpenCodeSchema()`; upstream additions appear after re-sync.
- **Pointer-to-bool for tri-state**: `*bool` distinguishes `false` from "not set" across all optional Config fields.

## Sparse profile persistence

Selected-field wizard saves use `profile.MarshalSparse` then `profile.WriteOpenCodeBlockInto` / `SaveOpenCodeBlock` to update `profiles.<name>.[opencode]`. Do not persist sparse payloads with `Profile.Save` (`omitempty` drops explicit zeros).
