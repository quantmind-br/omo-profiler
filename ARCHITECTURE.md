# omo-profiler Architecture

**Generated:** 2026-05-11
**Source:** GitNexus Knowledge Graph (4,249 nodes, 17,129 edges, 102 clusters, 236 flows)

---

## Overview

`omo-profiler` is a TUI profile manager for `oh-my-openagent` configuration. Built in Go 1.25.6, it combines a Bubble Tea terminal UI with a Cobra CLI, managing profile CRUD inside the unified `~/.omo/omo.json` (or `omo.jsonc`) document, in-document activation via verbatim substitution, and JSON schema validation against upstream `assets/omo.schema.json`.

The architecture follows a **layered, message-driven design** with three primary layers:

1. **Presentation** (`internal/tui/` + `internal/cli/`) — Bubble Tea views and Cobra commands
2. **Business Logic** (`internal/profile/`, `internal/models/`, `internal/schema/`) — CRUD, env activation, validation, model registry
3. **Infrastructure** (`internal/config/`, `internal/diff/`, `internal/backup/`) — Document/paths, diff computation, backup rotation

---

## Functional Areas (Knowledge Graph Communities)

| Area | Symbols | Cohesion | Role |
|--------|---------|----------|------|
| **Views** | ~400+ | 0.38–0.92 | 18 Bubble Tea sub-views (wizard steps, dashboard, list, diff, import, export, models, schema check) |
| **Profile** | 33 | 0.69–0.76 | Profile CRUD in omo document, in-document activation, naming validation, sparse-field detection |
| **Config** | 17 | 0.76 | `[opencode]` `Config` (46 fields) + `Document`, path resolution (`OmoDir`/`OmoFile`), `SetBaseDir` test isolation |
| **Schema** | 15 | 0.88 | Embedded omo document schema + `[opencode]` sub-schema validator, upstream drift detection |
| **Tui** | 14 | 0.90 | Root `App` model — state machine, message router, global overlays (toast, help, spinner) |
| **Diff** | 14 | 0.96 | Side-by-side + unified diff computation (`go-diff` wrapper) |
| **Backup** | 14 | 0.76 | Timestamped backup rotation before mutating omo writes |
| **Models** | 14+ | 0.82 | Local model registry CRUD + `models.dev` API client |
| **Cmd** | 12 | 0.85 | Cobra subcommands (`list`, `switch`, `import`, `export`, `create`, `models`, `schema-check`) |
| **Cli** | 8 | 0.80 | Root command registration, TUI-as-default `Run` behavior |

> **Note:** `Views` is heavily fragmented into ~20 sub-communities because the 18 view files each have tight internal cohesion but loose coupling between each other. This is intentional — each view is an independent Bubble Tea model.

---

## Module Dependency Graph

```mermaid
graph TD
    subgraph Presentation
        CLI[internal/cli/cmd<br/>8 subcommands]
        TUI[internal/tui/app.go<br/>State machine + router]
        VIEWS[internal/tui/views<br/>18 sub-views]
    end

    subgraph Business
        PROFILE[internal/profile<br/>CRUD, env activate, naming]
        MODELS[internal/models<br/>Registry + API client]
        SCHEMA[internal/schema<br/>Validator + upstream sync]
    end

    subgraph Infrastructure
        CONFIG[internal/config<br/>Types + Document + paths]
        DIFF[internal/diff<br/>Side-by-side diff]
        BACKUP[internal/backup<br/>Timestamped backups]
    end

    TUI --> VIEWS
    TUI --> PROFILE
    TUI --> MODELS
    TUI --> SCHEMA
    TUI --> DIFF
    TUI --> BACKUP

    CLI --> PROFILE
    CLI --> BACKUP
    CLI --> MODELS
    CLI --> SCHEMA

    VIEWS --> PROFILE
    VIEWS --> MODELS
    VIEWS --> CONFIG
    VIEWS --> SCHEMA
    VIEWS --> DIFF

    PROFILE --> CONFIG
    PROFILE --> DIFF
    MODELS --> CONFIG
    SCHEMA --> DIFF
    SCHEMA --> CONFIG
    BACKUP --> CONFIG

    style TUI fill:#e1f5fe
    style VIEWS fill:#e1f5fe
    style CLI fill:#e1f5fe
    style CONFIG fill:#fff3e0
    style PROFILE fill:#f3e5f5
    style SCHEMA fill:#f3e5f5
    style MODELS fill:#f3e5f5
```

**Key dependency patterns:**
- **All business logic depends on `config`** — `Config` is the `[opencode]` contract; `Document` is the omo file
- **TUI is the orchestration hub** — `App` dispatches to all business packages via `tea.Cmd`
- **CLI is thin** — commands delegate directly to `profile`/`backup`/`models`/`schema`
- **`diff` is a utility** — used by both `profile` (sparse detection) and `schema` (upstream drift)

---

## Key Execution Flows

### 1. Profile Switching (`doSwitchProfile`)

**Trigger:** User selects "Switch" in List view → `App` emits `doSwitchProfile` `tea.Cmd`

```
App.doSwitchProfile (tui/app.go)
  → profile.Apply (profile/active.go)
      → doc.ProfileBlock(name) — load the profile block
      → ActiveName(doc) — is the root already this profile?
      → if no match and declared key is non-empty at root:
          snapshot root as profiles.base (collides to base-1, …)
      → doc.SetRaw(key, value) for each declared key
      → backup.CreateOmoIfPresent (pre-save hook)
      → doc.EnsureSchema()
  → UI shows success toast (no shell command)
```

**Cross-community:** Tui → Profile → Config
**Critical constraint:** Activation is **in-document**. `Apply` substitutes verbatim — no merge, no env var, no shell command. The snapshot step (`profiles.base`) is what makes a sparse root recoverable.

### 1b. Sparse Wizard Save (`MarshalSparse` → `WriteOpenCodeBlockInto`)

**Trigger:** Wizard Review confirms save of selected fields

```
wizard save (tui/views/wizard.go)
  → profile.MarshalSparse(cfg, selection, preservedUnknown)
  → config.MutateWithPreSave(backup.CreateOmoIfPresent, fn)   — one transaction
      fn: profile.WriteOpenCodeBlockInto(doc, name, data)
      pre-save snapshot + doc.Save() run under the same lock
```

**Critical constraint:** Do **not** route sparse payloads through `Profile.Save` / `WriteInto` — those marshal `Config` with `omitempty` and drop explicitly selected zero values. `SaveOpenCodeBlock` is the one-shot variant of the same path.

---

### 2. Active Profile Detection (`loadActiveProfile` / `LoadProfiles`)

**Trigger:** Dashboard or List view initializes → loads active profile info

```
Dashboard.loadActiveProfile / List.LoadProfiles (views/)
  → profile.GetActive (profile/active.go)
      → ActiveName(doc) — first profile whose every declared key
        canonicalizes equal to the root
      → decode root [opencode] straight into Config (no merge)
      → ActiveConfig{Exists, Config, ProfileName, Modified}
  → config.OmoFile / LoadDocument (config/)
```

**Cross-community:** Views → Profile → Config
**No sidecar authority:** The active profile is detected by root comparison. `Modified` is true when the root matches no profile.
---

### 3. Upstream Schema Drift Detection (`schema_check`)

**Trigger:** User runs `schema-check` command or TUI view

```
SchemaCheck.fetchSchemaCheckCmd (views/schema_check.go)
  → schema.CompareSchemas (schema/compare.go)
      → schema.FetchUpstreamSchema (schema/compare.go)  — HTTP GET assets/omo.schema.json
      → schema.GetEmbeddedSchema (schema/validator.go)  — read embedded omo document bytes
      → diff.ComputeUnifiedDiff (diff/diff.go)          — generate diff if drift
  → schema.SaveDiff (schema/compare.go)                  — persist .diff report
```

**Cross-community:** Views → Schema → Diff  
**Upstream URL:** `assets/omo.schema.json` on the oh-my-openagent monorepo (`packages/omo-config-core/src/schema/`). Migration transform: `packages/omo-opencode/src/config-migration/`. Forms use `GetOpenCodeSchema()`.

---

### 4. Profile Diff Visualization

**Trigger:** User selects "Compare" in Dashboard → `App` navigates to Diff view

```
Diff.computeDiff (views/diff.go)
  → diff.ComputeDiff (diff/diff.go)
      → buildDiffResult (diff/diff.go)
      → splitLines (diff/diff.go)
  → diff.DiffResult (diff/diff.go)  — typed {Left, Right} with DiffLine slices
```

**Cross-community:** Views → Diff (intra-community, tight coupling)  
**Render:** Dual viewport side-by-side with line numbers and color-coded `DiffAdded`/`DiffRemoved`/`DiffEqual`.

---

### 5. Wizard Model Save (`handleSaveCustomModel`)

**Trigger:** User adds a custom model in wizard Categories or Agents step

```
WizardCategories.handleSaveCustomModel / WizardAgents.handleSaveCustomModel
  → models.ModelsRegistry.Add (models/models.go)
  → models.ModelsRegistry.Save (models/models.go)
      → config.EnsureDirs (config/paths.go)
      → config.ModelsFile / OmoDir (config/paths.go)
```

**Cross-community:** Views → Models → Config  
**Persistence:** `~/.omo/models.json` with auto `.bak` recovery on parse failure. Writes are serialized (`models.Mutate`) and atomic (temp file + rename).

---

## State Machine

The TUI has **10 states** managed by `App` (`internal/tui/app.go`):

```
stateDashboard ──→ stateList ──→ stateWizard
    │                              ├──→ stateImport
    │                              ├──→ stateExport
    │                              ├──→ stateDiff
    │                              ├──→ stateModels
    │                              ├──→ stateModelImport
    │                              ├──→ stateTemplateSelect
    │                              └──→ stateSchemaCheck
    └──→ (direct jumps from dashboard to any state)
```

**Transitions:** Views emit `NavTo*Msg` → `App.Update` intercepts → `navigateTo(state)` re-initializes target view. Views are **re-created on every navigation** — no persisted state.

---

## Message Protocol

```
View emits tea.Msg ──→ App.Update intercepts ──→ App calls tea.Cmd for async
                                            └──→ navigateTo(newState) for routing
```

Key message types:
- `NavToWizardMsg`, `NavToDiffMsg`, `NavToListMsg`, etc. — routing
- `switchProfileDoneMsg`, `deleteProfileDoneMsg`, `importProfileDoneMsg` — async completion
- `WizardNextMsg`, `WizardBackMsg`, `WizardSaveMsg`, `WizardCancelMsg` — wizard lifecycle
- `toastMsg` / `clearToastMsg` — global toast overlay

---

## Validation Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Strict Mode    │     │  Permissive Mode │     │  Upstream Sync  │
│  Validate()     │     │  ValidateForSave │     │  CompareSchemas │
│  ([opencode])   │     │  (ignore missing)│     │  (diff vs HTTP) │
└─────────────────┘     └──────────────────┘     └─────────────────┘
        │                        │                       │
        └────────────────────────┴───────────────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │   gojsonschema.Schema     │
                    │   document + [opencode]   │
                    │   (singleton, sync.Once)  │
                    └───────────────────────────┘
```

- **`Validate` / `ValidateJSON`** — strict on flat `[opencode]`
- **`ValidateForSave` / `ValidateJSONForSave`** — permissive; default for wizard/profile save
- **`ValidateDocument` / `ValidateDocumentForSave`** — whole omo document
- **`CompareSchemas`** — fetches upstream `assets/omo.schema.json`, compares bytes, generates unified diff
- **Forms** — `GetOpenCodeSchema()`, not `GetEmbeddedSchema()`

---

## Testing Architecture

- **36 test files** (~10,000+ lines of tests)
- **Co-located** `*_test.go` per package
- **Mandatory isolation:** `config.SetBaseDir(t.TempDir())` in every test that touches FS
- **Seed profiles into the document**, not as files under a profiles directory (`writeProfile` / `SetProfileBlock` + `Save`)
- **`setupTestEnv` helper** pattern for cross-package test setup
- **High-coverage packages:** `profile/` (sparse detection), `config/` (round-trip), `schema/` (strict vs permissive)
- **TUI tests** use Bubble Tea's `tea.Program` testing patterns with simulated key messages

---

## Critical Constraints & Invariants

1. **`Config` is the `[opencode]` source of truth** — `internal/config/types.go` must match the `[opencode]` sub-schema 1:1; the whole file is `Document` / `omo.schema.json`
2. **In-document activation** — `Apply` substitutes profile keys verbatim into the root; `ActiveName` detects the applied profile via root comparison. No env vars, no copy, no symlink.
2b. **Sparse persists via raw block APIs** — `MarshalSparse` → `WriteOpenCodeBlockInto` / `SaveOpenCodeBlock`, never `Profile.Save` for selected zeros
3. **No blocking in Update** — all I/O happens in `tea.Cmd`, never in `Update()` or `View()`
4. **Views emit, App routes** — views must never mutate `App` state; navigation is message-driven
5. **Wizard `Apply()` pattern** — steps must not mutate `Config` directly; use `SetConfig`/`Apply` lifecycle
6. **Singleton validator** — `schema.Validator` is initialized once via `sync.Once`; always use `GetValidator()`; forms use `GetOpenCodeSchema()`

---

## File Size Hotspots

| File | Lines | Complexity |
|------|-------|------------|
| `internal/tui/views/wizard_other.go` | ~2,460 | 60+ fields, 33 sections — the catch-all wizard step |
| `internal/tui/views/wizard_agents.go` | ~1,230 | Agent config forms with nested viewport scrolling |
| `internal/tui/views/wizard_categories.go` | ~980 | Category CRUD with dynamic form injection |
| `internal/tui/app.go` | ~840 | Root state machine, message router, overlays |
| `internal/tui/views/model_registry.go` | ~625 | Local model CRUD with in-place form swapping |
| `internal/tui/views/model_import.go` | ~546 | Async models.dev fetcher with fuzzy filtering + multi-select |
| `internal/tui/views/model_selector.go` | ~528 | Reusable searchable model dropdown |

These 7 files account for ~40% of the application code. The `wizard_other.go` monolith is the primary maintenance risk — any config field addition touches this file.
