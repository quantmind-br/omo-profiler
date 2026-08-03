# Data & Schema

This page covers the core data model (`config.Config` as the `[opencode]` block), the omo document, the embedded JSON schema, validation modes, the profile model, in-document activation, and sparse serialization.

## Config Struct (`internal/config/types.go`)

The `Config` struct is the **`[opencode]` data contract** — the flat 46-field harness block that omo-profiler edits. It is **not** the whole omo file. All business packages depend on it for the editable payload. It has ~46 top-level fields and ~30+ nested struct types.

### Key Fields

```
Agents              map[string]*AgentConfig    — per-agent settings (24 fields each)
Categories          map[string]*CategoryConfig — per-category settings (15 fields)
AgentOrder          []string                   — ordering of agents
Skills              json.RawMessage            — raw JSON passthrough (preserves original shape)
RuntimeFallback     json.RawMessage            — raw JSON passthrough
TeamModeConfig      *TeamModeConfig
MonitorConfig       *MonitorConfig
CodegraphConfig     *CodegraphConfig
ExperimentalConfig  *ExperimentalConfig
BackgroundTaskConfig *BackgroundTaskConfig
TmuxConfig          *TmuxConfig
SisyphusConfig      *SisyphusConfig
OpenclawConfig      *OpenclawConfig
ClaudeCodeConfig    *ClaudeCodeConfig
MCPEnvAllowlist     []string
DisabledMCPs        []string
DisabledAgents      []string
DisabledTools       []string
DisabledProviders   []string
Migrations          []migration
```

### Critical Field Patterns

| Pattern | Rule | Why |
|---------|------|-----|
| **Pointer-to-bool** | `*bool` | Distinguishes `false` from "not set" (tri-state) |
| **All JSON tags** | `json:"...,omitempty"` | Prevents dirty config files |
| **`interface{}`** | For `FallbackModels`, `CommitFooter`, `PermissionConfig.Bash` | Fields that accept string or object |
| **Pure data containers** | No methods on Config types | Validation is in the `schema/` package |
| **Matching upstream** | Field names must match the `[opencode]` sub-schema exactly | Schema-driven validation |

### Config Sub-types

Each nested pointer field (`*TeamModeConfig`, etc.) has its own struct with `json:"...,omitempty"` tags and `*bool` for optional booleans. The full type list is in `internal/config/types.go`.

## Omo Document (`internal/config/document.go`)

```go
type Document struct {
    Path   string
    Exists bool
    // raw top-level keys are private
}
```

Top-level keys (upstream omo schema): `$schema`, shared typed keys (`categories`, `agents`, `codegraph`, `task`, `teams`, `models`), harness blocks `[opencode]` / `[senpi]` / `[codex]`, `profiles`, migration bookkeeping.

| Method | Behavior |
|--------|----------|
| `LoadDocument()` / `LoadDocumentFrom` | Missing file → `Exists:false`, not an error |
| `ProfileNames` / `ProfileBlock` / `SetProfileBlock` / `DeleteProfileBlock` / `HasProfile` | CRUD on `profiles.<name>` |
| `EnsureSchema()` | Sets `$schema` to `config.DefaultSchema` when absent |
| `Bytes()` / `Save()` | Canonical indented JSON, sorted keys; creates `~/.omo` as needed |

JSONC comments are **not** preserved across write.

## JSON Schema

The project embeds the **omo document** schema for whole-file validation and extracts the `[opencode]` sub-schema for forms and flat-config validation.

### Files

| File | Purpose |
|------|---------|
| `internal/schema/schema.json` | **Embedded** omo document schema (`//go:embed`) |
| `omo.schema.json` (repo root) | Same document schema, checked in at root |
| Upstream | `assets/omo.schema.json` on oh-my-openagent (`packages/omo-config-core/src/schema/` in the monorepo) |

The old root `oh-my-opencode.schema.json` was deleted. Migration transform lives in upstream `packages/omo-opencode/src/config-migration/`.

### Upstream Sync

Source: `/internal/schema/compare.go`

```go
var UpstreamSchemaURL = "https://raw.githubusercontent.com/code-yeongyu/oh-my-openagent/dev/assets/omo.schema.json"
```

Also `config.DefaultSchema` — written into new documents.

- `FetchUpstreamSchema(ctx)` — HTTP GET with 30s timeout
- `CompareSchemas()` — downloads upstream, compares bytes via `bytes.Equal`, computes unified diff if drift detected
- `SaveDiff(dir, diffContent)` — persists `.diff` report
- `.upstream-sha` sidecar tracks last synced commit (base64-encoded)

**Never hand-edit `schema.json`** — always re-sync from upstream and keep root `omo.schema.json` aligned.

### Validator Singleton (`internal/schema/validator.go`)

```go
GetEmbeddedSchema() []byte           // FULL omo document schema
GetOpenCodeSchema() ([]byte, error)  // self-contained [opencode] sub-schema for editor forms
GetValidator() (*Validator, error)   // singleton via sync.Once
```

Validation modes:

| Method | Target | Enforces Required Fields? |
|--------|--------|--------------------------|
| `Validate(cfg)` / `ValidateJSON(data)` | Flat `[opencode]` | Yes |
| `ValidateForSave(cfg)` / `ValidateJSONForSave(data)` | Flat `[opencode]` — wizard/profile saves | No |
| `ValidateDocument` / `ValidateDocumentForSave` | Whole omo document | Same strict/permissive split |

**Always use `ValidateForSave` for saves** — sparse configs are intentional. **Forms must use `GetOpenCodeSchema()`.**

## Profile Model (`internal/profile/`)

### Profile Type

```go
type Profile struct {
    Name                string
    Path                string            // omo document path (informational)
    Config              config.Config     // == profiles.<name>.[opencode]
    PreservedUnknown    map[string]json.RawMessage // unknown keys INSIDE [opencode]
    PreservedBlock      map[string]json.RawMessage // sibling keys of [opencode] in the profile block
    FieldPresence       map[string]bool
    HasLegacyFields     bool
    LegacyFieldsWarning string
}
```

### CRUD Operations

| Function | Behavior |
|----------|----------|
| `Load(name)` / `LoadFromDocument` | Reads `profiles.<name>.[opencode]` from the omo document; preserves unknowns/siblings; builds `FieldPresence` |
| `Save(p)` / `(p).Save()` | Read-modify-write of the omo document; other profiles untouched. Marshals `Config` with `omitempty` — drops explicit zeros |
| `WriteInto(doc)` | Stage without persisting |
| `WriteOpenCodeBlockInto(doc, name, openCode)` | Stage a pre-marshalled `[opencode]` payload; preserve sibling keys of the profile block |
| `SaveOpenCodeBlock(name, openCode)` | Persist a pre-marshalled `[opencode]` payload (the sparse-save path) |
| `Delete(name)` | Removes `profiles.<name>` from the document |
| `List()` | Sorted profile names from the document |
| `Exists(name)` | `doc.HasProfile(name)` |

There is **no** `profiles/` directory and **no** file-per-profile.

Root starter template: `template/opencode-profile.json` (renamed from `template/oh-my-openagent.json`; `$schema` removed — harness blocks do not carry `$schema`; the document root does).

### Activation (`internal/profile/active.go`)

Activation is **in-document** — `Apply` substitutes profile keys verbatim over the root, no merge, no copy, no symlink.

```go
func Apply(name string) (Applied, error)
func ActiveName(doc *config.Document) (string, error)
func GetActive() (*ActiveConfig, error)
func canonicalJSON(raw json.RawMessage) ([]byte, error)
```

`Apply` copies every key in `profiles.<name>` over the matching document root key. If the root matches no profile, it is snapshotted as `profiles.base` (colliding to `base-1`, …) before being overwritten.

`ActiveName` returns the first profile (in sorted order) whose every declared key canonicalizes equal to the corresponding root key. `canonicalJSON` renders raw JSON with sorted keys at every depth and drops `$schema`.

`ActiveConfig` fields: `Exists`, `Config` (root `[opencode]`, no merge), `ProfileName` (detected by root comparison), `Modified` (root matches no profile).

### Sparse Serialization (`internal/profile/sparse.go` + `selection.go`)

Profiles store only selected fields inside `[opencode]`, not the entire config.

**`FieldSelection`**: Tracks which config fields (`agents.*.model`, etc.) are selected via `map[string]bool`.

**Path matching** supports wildcards: `agents.*.model` matches `agents.claude.model`, `agents.build.model`, etc. ~197 known dotted paths are defined.

```go
type FieldSelection struct {
    selected map[string]bool
}

func NewBlankSelection() *FieldSelection
func NewSelectionFromPresence(presence map[string]bool) *FieldSelection
func (s *FieldSelection) IsSelected(path string) bool
func (s *FieldSelection) Toggle(path string)
func (s *FieldSelection) SelectedPaths() []string
```

**`MarshalSparse(cfg, selection, preservedUnknown)`**: Builds JSON with only selected fields:
1. Recursively walks `Config` struct via reflection
2. Filters each field through the selection (wildcard-aware)
3. Deep-merges preserved-unknown keys back
4. Uses `json.NewEncoder` with `SetEscapeHTML(false)` so `<>` renders literally
5. Trims trailing newline

Persist the result with `WriteOpenCodeBlockInto` / `SaveOpenCodeBlock` — **not** `Profile.Save`, which would re-marshal through `omitempty` and drop explicitly selected zero values.

### Naming Validation (`internal/profile/naming.go`)

```go
func ValidateName(name string) error   // ^[a-zA-Z0-9_-]+$
func SanitizeName(name string) string  // strip invalid, trim leading/trailing -_
```

Sentinel errors: `ErrInvalidName`, `ErrEmptyName`.

## Backup (`internal/backup/backup.go`)

Mutating writes to the omo document (profile save/delete/import) snapshot `config.OmoFile()` first, from inside the `config.MutateWithPreSave` lock so the copy is the exact pre-image of that write. Switching does **not** mutate the document and needs no backup.

```
omo.json.bak.2006-01-02-150405.000000000
# or omo.jsonc.bak.… when that variant is the live file
```

Timestamps carry nanoseconds so writes in the same second get distinct files, but uniqueness comes from `O_CREATE|O_EXCL`, not the clock: on collision `Create` advances a nanosecond and retries, so two writers reading the same instant cannot overwrite each other. Second-precision names written by older versions stay listable and restorable.

Key functions:

| Function | Behavior |
|----------|----------|
| `Create(configPath)` | Reads file, writes copy under `OmoDir()` with `.bak.<timestamp>` suffix, preserving the source file's mode |
| `List()` | Scans omo dir for backup files, sorted most recent first |
| `Restore(backupPath)` | Writes a backup over the current omo file |
| `Clean(keepLast)` | Rotation — prunes beyond the N most recent backups |

File matching recognizes `omo.json` / `omo.jsonc` backups plus legacy openagent/opencode basenames for migration leftovers.

## Diff Engine (`internal/diff/diff.go`)

Two modes:

| Function | Output | Used By |
|----------|--------|---------|
| `ComputeDiff(json1, json2)` | `DiffResult` with aligned `Left`/`Right` slices | Profile comparison (side-by-side view) |
| `ComputeUnifiedDiff(oldName, newName, old, new)` | Unified diff string (`---`/`+++` format) | Schema drift detection |

`DiffResult` contains `Left` and `Right` slices of `DiffLine{Text, Type, LineNum}` with types `DiffEqual`, `DiffAdded`, `DiffRemoved`.

## Model Registry (`internal/models/models.go`)

Persisted to `~/.omo/models.json` (`config.ModelsFile()`):

```go
type RegisteredModel struct {
    DisplayName string
    ModelID     string
    Provider    string
}
```

- **Corruption auto-recovery**: if JSON unmarshal fails, corrupted file is backed up to `.bak`, empty registry returned
- **Duplicate detection**: `(Provider, ModelID)` uniqueness
- **Grouped listing**: `ListByProvider()` groups models, sorts within group by `DisplayName`

### Models.dev API Client (`internal/models/modelsdev.go`)

- `FetchModelsDevRegistry()` — HTTP GET to `https://models.dev/api.json` with 30s timeout
- Returns providers and their models with capability metadata (context length, reasoning, tool calling, vision)

## Change Guidance

- **Adding a new config field**: Add to `Config` in `config/types.go`, re-sync `schema.json` / `omo.schema.json` from upstream, update `allFieldPaths` in `profile/selection.go` if it should be selectable for sparse output
- **Changing schema validation**: Prefer `ValidateForSave` for user-facing saves; use `ValidateDocument*` for whole-file checks; forms use `GetOpenCodeSchema()`
- **Changing profile serialization**: `MarshalSparse` must preserve wildcard path matching and deep-merge of preserved-unknown keys; persist via `WriteOpenCodeBlockInto` / `SaveOpenCodeBlock` into the omo document, not loose files or `Profile.Save`
- **Activation changes**: Never invent an authoritative sidecar; keep env precedence aligned with upstream
