# SCHEMA KNOWLEDGE BASE

## OVERVIEW

JSON schema validation for the unified omo document and its `[opencode]` harness block. Singleton validator with embedded omo document schema (`schema.json`), plus upstream drift detection against `assets/omo.schema.json`.

## FILES

| File | Role |
|------|------|
| `schema.json` | Embedded **omo document** schema (`//go:embed`) — same contract as root `omo.schema.json` |
| `validator.go` | Singleton `Validator`, `GetEmbeddedSchema`, `GetOpenCodeSchema`, document + `[opencode]` validate paths |
| `compare.go` | `FetchUpstreamSchema`, `CompareSchemas`, `SaveDiff` |
| `validator_test.go` | Save vs strict validation path tests |
| `compare_test.go` | Upstream fetch + diff save tests |
| `accessor_test.go` | Schema accessor helpers |

Root `omo.schema.json` mirrors the embedded document schema. The old root `oh-my-opencode.schema.json` was deleted. Upstream monorepo schema source: `packages/omo-config-core/src/schema/`; migration transform: `packages/omo-opencode/src/config-migration/`.

## KEY TYPES

- `Validator`: Holds both the full document schema and the extracted `[opencode]` sub-schema; singleton via `sync.Once`
- `ValidationError`: `{Path, Message}` pair for schema violations
- `CompareResult`: `{Identical bool, Diff string}` from upstream comparison

## ACCESSORS

```go
GetEmbeddedSchema() []byte           // FULL omo document schema
GetOpenCodeSchema() ([]byte, error)  // self-contained [opencode] sub-schema (flat 46-field config)
GetValidator() (*Validator, error)   // singleton
```

**Anything rendering an editor form for the config the user edits must use `GetOpenCodeSchema()`, not `GetEmbeddedSchema()`.** The `[opencode]` sub-schema is byte-for-byte the old flat legacy schema, so 46-field validation behaves as before.

## VALIDATION MODES

| Method | Target | Required Fields |
|--------|--------|----------------|
| `Validate` / `ValidateJSON` | Flat `[opencode]` (`config.Config`) | Enforced |
| `ValidateForSave` / `ValidateJSONForSave` | Flat `[opencode]` — wizard/profile save | Ignored |
| `ValidateDocument` / `ValidateDocumentForSave` | Whole omo.json document | Same strict/permissive split |

`ValidateForSave` is the default for wizard review and profile save — sparse configs are intentional.

## UPSTREAM SYNC

```go
var UpstreamSchemaURL = "https://raw.githubusercontent.com/code-yeongyu/oh-my-openagent/dev/assets/omo.schema.json"
```

Also available as `config.DefaultSchema` (written into new documents via `Document.EnsureSchema()`).

`FetchUpstreamSchema()` → `CompareSchemas()` → `SaveDiff()`:
1. Downloads from `UpstreamSchemaURL` (oh-my-openagent monorepo, `assets/omo.schema.json`)
2. Compares embedded vs upstream bytes
3. Generates unified diff if drift detected
4. `.upstream-sha` tracks last synced commit

**Never hand-edit `schema.json`** — re-sync from upstream, then keep root `omo.schema.json` aligned.

## ANTI-PATTERNS

- **Direct `gojsonschema` usage**: Always use `GetValidator()` — never instantiate loaders directly
- **Forms on the document schema**: Use `GetOpenCodeSchema()` for editor forms; `GetEmbeddedSchema()` is the whole-file schema
- **Calling `Validate` for saves**: Use `ValidateForSave` to allow sparse configs
- **Manual schema updates**: Re-sync from upstream, never invent fields in `schema.json`
- **Referring to `oh-my-opencode.schema.json`**: That root file is gone; the project schema is `omo.schema.json`
