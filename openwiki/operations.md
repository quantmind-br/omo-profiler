# Operations

This page covers CLI commands, the web server, path resolution, backup/diff utilities, and testing conventions.

## CLI Commands

Source: `/internal/cli/cmd/*.go`

9 commands registered from `rootCmd.init()` (`/internal/cli/root.go` lines 34-43):

| Command | File | Behavior |
|---------|------|----------|
| `omo-profiler` | `root.go` (default Run) | Launches TUI via `tui.Run()` |
| `web` | `web.go` | Launches web server; flags: `--host` (127.0.0.1), `--port` (4747), `--no-open` |
| `list` | `list.go` | Lists profiles from `profile.List()`, marks applied profile with `*` |
| `current` | `current.go` | Prints the profile matching the root of `~/.omo/omo.json` via `profile.GetActive()` |
| `switch` | `switch.go` | `profile.Apply(name)` — substitutes profile keys into the document root with a pre-write backup |
| `import` | `import.go` | Imports profile into the omo document; validates with `ValidateJSONForSave`; backup `OmoFile` first |
| `export` | `export.go` | Exports profile `[opencode]` to JSON file; `--force` to overwrite |
| `create` | `create.go` | Creates a new `profiles.<name>` block; `--from` clones an existing profile name as template. Starter file: `template/opencode-profile.json` |
| `models` | `models.go` | Sub-command group: `list`, `add`, `remove` |
| `schema-check` | `schema_check.go` | Validates schema and checks upstream drift vs `assets/omo.schema.json` |

All commands use `RunE` (returning error) or `Run` (calling `os.Exit` directly). The `profile` package is their primary dependency.

## Web Server (`internal/web/`)

### Starting

```bash
omo-profiler web                  # bind 127.0.0.1:4747, open browser
omo-profiler web --port 8080      # custom port
omo-profiler web --host 0.0.0.0   # override bind
omo-profiler web --no-open        # don't open a browser
```

The web server reuses all business packages (`profile/`, `config/`, `schema/`, `models/`, `backup/`, `diff/`) unchanged. It is the **recommended interface**; the TUI remains the default but is slated for future deprecation.

### API Endpoints

All routes defined in `newMux()` (`/internal/web/server.go` lines 22-53):

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/api/profiles` | `handleListProfiles` | List profiles + active status |
| POST | `/api/profiles` | `handleCreateProfile` | Create (from scratch, template, or clone) |
| GET | `/api/profiles/{name}` | `handleGetProfile` | Load profile + raw JSON |
| PUT | `/api/profiles/{name}` | `handleSaveProfile` | Validate + save into omo document |
| DELETE | `/api/profiles/{name}` | `handleDeleteProfile` | Delete profile block |
| POST | `/api/profiles/{name}/rename` | `handleRenameProfile` | Rename inside document |
| POST | `/api/profiles/{name}/activate` | `handleActivateProfile` | `profile.Apply(name)` — substitutes profile keys into the root (with pre-write backup) |
| GET | `/api/profiles/{name}/export` | `handleExportProfile` | Download `[opencode]` as JSON |
| GET | `/api/active` | `handleGetActive` | Root `[opencode]` config + applied profile name + modified flag |
| GET | `/api/diff` | `handleDiff` | Compare `left` vs `right` (`__active__` for effective) |
| POST | `/api/import` | `handleImport` | Import with auto-naming on collision |
| POST | `/api/validate` | `handleValidate` | `?mode=strict` for full validation; default is "save" mode |
| GET | `/api/schema` | `handleSchema` | Embedded omo document schema bytes |
| GET | `/api/schema-check` | `handleSchemaCheck` | Upstream drift check |
| GET | `/api/models` | `handleListModels` | List all registered models |
| POST | `/api/models` | `handleCreateModel` | Register a model |
| GET | `/api/models/catalog` | `handleModelsCatalog` | Models.dev catalog |
| PUT | `/api/models/{provider}/{modelId}` | `handleUpdateModel` | Update model |
| DELETE | `/api/models/{provider}/{modelId}` | `handleDeleteModel` | Delete model |
| `/` | All other routes | `spaHandler()` | SPA with client-route fallback |

Editor forms should be driven by `schema.GetOpenCodeSchema()` (the flat `[opencode]` sub-schema), not the whole-document schema.

### SPA Embedding (`/internal/web/embed.go`)

The React SPA is embedded via `//go:embed all:frontend/dist`. The `all:` prefix includes dotfiles (`.gitkeep`), so `make build` compiles even without the frontend built.

- **Without frontend**: serves a `"Web UI not built. Run: make web-build\n"` placeholder
- **With frontend**: serves static files, falls back to `index.html` for client-side routes
- API paths are never handled by the SPA handler

### Build Requirements

```bash
make web-deps    # cd internal/web/frontend && npm install   (requires Node)
make web-build   # npm run build in the frontend directory
make build-web   # web-build + go build (binary with SPA embedded)
```

## Path Resolution (`internal/config/paths.go`)

All filesystem paths go through `config.*` helpers — never hardcode `~/.omo`:

| Function | Resolves to |
|----------|-------------|
| `OmoDir()` | `~/.omo/` |
| `OmoFile()` | `~/.omo/omo.jsonc` if present, else `~/.omo/omo.json` |
| `ModelsFile()` | `~/.omo/models.json` |
| `EnsureDirs()` | Creates `~/.omo` with 0755 permissions |
| `LegacyConfigDir()` | Pre-unification OpenCode config dir — migration detection only |
| `LegacyConfigFile()` | Legacy flat file if present, else `""` |
| `LegacyProfilesDir()` | Legacy file-per-profile dir — migration detection only |
| `DefaultSchema` | Upstream `assets/omo.schema.json` URL |

Removed: `ConfigDir`, `ConfigFile`, `ProfilesDir`, `ConfigBasename`, `LegacyConfigBasename`.

### Test Isolation

```go
func setupTestEnv(t *testing.T) func() {
    tmpDir := t.TempDir()
    config.SetBaseDir(tmpDir)
    return func() { config.ResetBaseDir() }
}
```

`SetBaseDir(path)` redirects all paths under `<tmp>/.omo/`. Seed profiles into the document (`SetProfileBlock` + `Save`), not as separate files. Always pair with `ResetBaseDir()` via `defer`.

## Backup (`internal/backup/backup.go`)

Mutating writes to the omo document (save/delete/import) snapshot `config.OmoFile()` from inside the write lock, so each backup is that write's exact pre-image. Switch does not mutate and needs no backup.

```
omo.json.bak.2006-01-02-150405.000000000
```

Nanosecond timestamps keep same-second writes apart, and names are claimed with `O_CREATE|O_EXCL` (advancing a nanosecond on collision) so equal clock readings — including from a second process — cannot overwrite a pre-image. Older second-precision names remain listable and restorable.

| Function | Purpose |
|----------|---------|
| `Create(configPath)` | Creates a `.bak.<timestamp>` copy under `OmoDir()`, preserving the source mode |
| `List()` | Scans omo dir for backup files, sorted most recent first |
| `Restore(backupPath)` | Writes a backup over the current omo file |
| `Clean(keepLast)` | Rotation — prunes beyond the N most recent |

File matching: `omo.json` / `omo.jsonc` backups, plus legacy openagent/opencode basename leftovers.

## Diff Engine (`internal/diff/diff.go`)

Two modes using `github.com/sergi/go-diff`:

| Function | Output | Use Case |
|----------|--------|----------|
| `ComputeDiff(json1, json2)` | `DiffResult{Left, Right}` with aligned `DiffLine` slices | TUI side-by-side comparison |
| `ComputeUnifiedDiff(oldName, newName, old, new)` | Unified diff string (`---`/`+++` headers) | Schema drift report |

`DiffLine` has `Text`, `Type` (DiffEqual/DiffAdded/DiffRemoved), and `LineNum` (0 for empty side).

## Testing

### Conventions

- **Co-located tests**: `*_test.go` files alongside source files
- **Assertions**: `github.com/stretchr/testify/assert` and `github.com/stretchr/testify/require`
- **Filesystem isolation**: Any test touching the filesystem must redirect paths via `config.SetBaseDir(tmpDir)` and seed profiles into the omo document
- **No CI**: All validation is local — run `make test && make lint`
- **No Docker**: Tests do not require containers

### Key Test Files

| Test File | What it Covers |
|-----------|----------------|
| `internal/config/types_test.go` | Config struct marshaling, JSON round-trip |
| `internal/config/paths_test.go` | Path resolution, test isolation |
| `internal/profile/profile_test.go` | Profile CRUD against Document, field presence |
| `internal/profile/selection_test.go` | Field selection, path matching, wildcards |
| `internal/profile/sparse_test.go` | Sparse serialization, reflection-based struct building |
| `internal/profile/naming_test.go` | Name validation and sanitization |
| `internal/profile/active_test.go` | In-document activation, `Apply` substitution/snapshot, `ActiveName` detection |
| `internal/schema/validator_test.go` | Validator singleton, strict vs permissive, document paths |
| `internal/schema/compare_test.go` | Schema comparison, upstream drift detection |
| `internal/models/models_test.go` | Model registry CRUD, corruption recovery |
| `internal/models/modelsdev_test.go` | models.dev API parsing |
| `internal/backup/backup_test.go` | Backup creation, listing, rotation |
| `internal/diff/diff_test.go` | Side-by-side and unified diff |
| `internal/tui/app_test.go` | App state machine, navigation, routing |
| `internal/tui/layout_test.go` | Layout system, responsive helpers |
| `internal/tui/views/dashboard_test.go` | Dashboard rendering, menu navigation |
| `internal/tui/views/list_test.go` | Profile list, filtering, switch/delete |
| `internal/tui/views/wizard_*_test.go` | Wizard steps (name, categories, agents, hooks, other, review) |
| `internal/tui/views/diff_test.go` | Diff navigation, pane switching |
| `internal/tui/views/model_registry_test.go` | Model list, search, CRUD |
| `internal/tui/views/model_import_test.go` | Import from models.dev |
| `internal/tui/views/import_test.go` | Profile import |
| `internal/tui/views/export_test.go` | Profile export |
| `internal/tui/views/schema_check_test.go` | Schema check view |
| `internal/tui/views/keybindings_test.go` | Keybinding inventory (46+ bindings) |
| `internal/web/server_test.go` | Web server API handlers |

### Running Tests

```bash
make test                    # go test -v ./... (no race detector)
go test -v ./internal/tui/   # TUI package tests
go test -v ./internal/profile/...  # Profile package tests
go test -v -run TestLoad ./internal/profile/   # Single test
```

## Change Guidance

### CLI Changes
- Add a new subcommand: create file in `/internal/cli/cmd/`, add `rootCmd.AddCommand()` in `root.go`'s `init()`
- Change default behavior: modify `rootCmd.Run` in `root.go`

### Web Server Changes
- Add API endpoint: add `mux.HandleFunc` in `newMux()` in `server.go`, implement handler in `handlers.go` (or `handlers_models.go`)
- Change embedding: update `embed.go` — the SPA is built separately and embedded via `//go:embed`

### Path Changes
- Add path helper: add function in `config/paths.go`, verify `SetBaseDir`/`ResetBaseDir` work correctly
- New config file under `~/.omo`: add a helper; update `backup.go` basename matching if it should be rotatable

### Testing Changes
- Any test using the filesystem: MUST call `config.SetBaseDir(tmpDir)` and `defer config.ResetBaseDir()`, and seed profiles into the document
- New test data: add JSON fixtures to `internal/testdata/`
- Keybinding changes: update `keybindings_test.go` which inventories all bindings
