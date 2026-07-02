# SCHEMA KNOWLEDGE BASE

## OVERVIEW

JSON schema validation against upstream `oh-my-openagent` spec. Singleton validator with embedded schema (180KB `schema.json`), plus upstream drift detection.

## FILES

| File | Role |
|------|------|
| `schema.json` | Embedded upstream schema (~180KB), fetched via `update-schema.sh` |
| `validator.go` | Singleton `Validator`, `ValidationError`, `GetValidator()` |
| `compare.go` | `FetchUpstreamSchema`, `CompareSchemas`, `SaveDiff` |
| `validator_test.go` | Save vs strict validation path tests |
| `compare_test.go` | Upstream fetch + diff save tests |
| `accessor_test.go` | Schema accessor helpers |

## KEY TYPES

- `Validator`: Wraps `gojsonschema.Schema`; singleton via `sync.Once`
- `ValidationError`: `{Path, Message}` pair for schema violations
- `CompareResult`: `{Identical bool, Diff string}` from upstream comparison

## VALIDATION MODES

| Method | Purpose | Required Fields |
|--------|---------|----------------|
| `Validate` | Strict — full schema compliance | Enforced |
| `ValidateForSave` | Permissive — ignores missing required fields | Ignored |
| `ValidateJSON` / `ValidateJSONForSave` | Raw bytes entry points (same semantics) | — |

`ValidateForSave` is the default for wizard review and profile save — sparse configs are intentional.

## UPSTREAM SYNC

`update-schema.sh` → `FetchUpstreamSchema()` → `CompareSchemas()` → `SaveDiff()`:
1. Downloads from `UpstreamSchemaURL` (oh-my-openagent dev branch)
2. Compares embedded vs upstream bytes
3. Generates unified diff if drift detected
4. `.upstream-sha` tracks last synced commit

## ANTI-PATTERNS

- **Direct `gojsonschema` usage**: Always use `GetValidator()` — never instantiate loaders directly
- **Calling `Validate` for saves**: Use `ValidateForSave` to allow sparse configs
- **Manual schema updates**: Use `update-schema.sh`, never hand-edit `schema.json`
- **Ignoring `.upstream-sha`**: Always update sidecar when syncing schema
