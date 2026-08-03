# omo-profiler Quickstart

**omo-profiler** is a profile manager for [oh-my-openagent](https://github.com/code-yeongyu/oh-my-openagent). It provides a **Terminal UI** (Bubble Tea), a **CLI** (Cobra), and a **Web UI** (React SPA) to create, edit, compare, import/export, and select profiles stored as `profiles.<name>` entries inside the unified `~/.omo/omo.json` (or `omo.jsonc`) document.

- **Go 1.25.6** project — `github.com/diogenes/omo-profiler`
- **TUI** is the default mode (`omo-profiler` with no subcommand)
- **Web UI** (`omo-profiler web`) is the recommended interface and runs at `http://127.0.0.1:4747`
- Each profile's editable payload is `profiles.<name>.[opencode]` — the flat oh-my-openagent config block

## Quick Start

```bash
# Install
go install github.com/diogenes/omo-profiler/cmd/omo-profiler@latest

# Or build from source
make build

# Launch TUI
omo-profiler

# Launch web UI (recommended)
omo-profiler web

# List profiles from the CLI
omo-profiler list
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `omo-profiler` | Launch TUI (default) |
| `omo-profiler web` | Launch web UI at `http://127.0.0.1:4747` |
| `omo-profiler list` | List all profiles |
| `omo-profiler current` | Show the profile matching the root of `~/.omo/omo.json` |
| `omo-profiler switch <name>` | Apply a profile by substituting its keys into `~/.omo/omo.json` |
| `omo-profiler import <file>` | Import a profile from a JSON file into the omo document |
| `omo-profiler export <name> <path>` | Export a profile's `[opencode]` block to a JSON file |
| `omo-profiler create` | Create a new profile |
| `omo-profiler models` | Manage the model registry (subcommands: `list`, `add`, `remove`) |
| `omo-profiler schema-check` | Validate schema and check for upstream drift |

## Key Directory Layout

```
cmd/omo-profiler/         # Binary entrypoint
internal/
  cli/cmd/                # Cobra subcommands (9 commands)
  config/                 # Config ([opencode] 46 fields) + Document + path resolution
  profile/                # Profile CRUD, env activation, sparse serialization
  schema/                 # Embedded omo schema + validator + upstream drift
  models/                 # Model registry + models.dev API client
  backup/                 # Timestamped backup rotation (mutating omo writes)
  diff/                   # Side-by-side and unified diff engine
  web/                    # HTTP server + embedded React SPA
  tui/                    # Bubble Tea root App + layout + styles
  tui/views/              # 18 Bubble Tea views (dashboard, wizard, list, diff, etc.)
  testdata/               # JSON test fixtures
```

## Important Paths

All paths are resolved through `internal/config/paths.go` helpers — never hardcode `~/.omo`.

| Path | Purpose |
|------|---------|
| `~/.omo/` | User config layer directory |
| `~/.omo/omo.json` / `omo.jsonc` | Unified omo document (`OmoFile()` prefers `.jsonc` when present) |
| `profiles.<name>.[opencode]` | Profile editable payload inside that document (not a file) |
| `~/.omo/models.json` | Model registry (omo-profiler local state) |
| `~/.omo/.omo-profiler-selection` | UI last-selected hint — **not** authoritative activation |
| Legacy user-layer (pre-unification) | Migration detection only (`LegacyConfigDir` / `LegacyConfigFile` / `LegacyProfilesDir`) |

## Web UI

```bash
omo-profiler web                 # bind 127.0.0.1:4747, open browser
omo-profiler web --port 8080     # custom port
omo-profiler web --host 0.0.0.0  # override bind address
omo-profiler web --no-open       # don't open a browser automatically
```

The web UI has API parity with the TUI and a **schema-driven editor** — it renders forms from the `[opencode]` sub-schema (`schema.GetOpenCodeSchema()`), so upstream field additions appear automatically after re-syncing `internal/schema/schema.json`.

Building the web UI requires Node.js (`make build-web`). Without it, a placeholder is served.

## TUI States

The TUI has 10 states navigated from the dashboard:

1. **Dashboard** — main menu (9 items)
2. **List** — profile list with filtering
3. **Wizard** — 6-step profile creation/editing
4. **Diff** — side-by-side profile comparison
5. **Import** — import from JSON
6. **Export** — export to file
7. **Models** — model registry management
8. **Model Import** — batch import from models.dev
9. **Template Select** — pick a template for new profile
10. **Schema Check** — validate schema and detect upstream drift

## Key Technical Concepts

- **In-document activation**: `Apply(name)` substitutes every key in `profiles.<name>` over the matching document root key — verbatim, no merging. `ActiveName` detects the applied profile by comparing the root against stored profiles. If the root matches no profile, it is snapshotted as `profiles.base` before being overwritten.
- **Single-document profiles**: a profile is `profiles.<name>` inside `omo.json`; editable payload is `profiles.<name>.[opencode]`.
- **Sparse serialization**: only selected config fields are written into the profile's `[opencode]` block; `FieldSelection` with wildcard path matching controls this.
- **Two validation modes** (plus document variants): `Validate` (strict) vs `ValidateForSave` (permissive — ignores missing required fields). Forms use `GetOpenCodeSchema()`.
- **Snapshot recoverability**: before `Apply` overwrites the root, unmatched keys are snapshotted as `profiles.base` (colliding to `base-1`, …). This is what makes a sparse profile's verbatim substitution recoverable.
- **Backup rotation**: nanosecond-timestamped `.bak.*` pre-images of `OmoFile()`, taken inside the write lock (save/delete/import/apply).

## Documentation Sections

- [Architecture](architecture.md) — layer model, module dependencies, key execution flows
- [TUI Architecture](tui.md) — state machine, views, wizard steps, keybinding patterns
- [Data & Schema](data-schema.md) — Config struct, JSON schema, validation, profiles, sparse serialization
- [Operations](operations.md) — CLI commands, web server, paths, backup, diff, testing

## Makefile Targets

```bash
make build       # go build -v -o omo-profiler ./cmd/omo-profiler
make test        # go test -v ./... (no race detector)
make lint        # golangci-lint run ./...
make install     # copy binary to ~/.local/bin/
make web-deps    # cd internal/web/frontend && npm install
make web-build   # build the React frontend
make build-web   # build frontend + binary (binary with SPA embedded)
make clean       # rm binary + go clean
```

## Change Guidance

When working in this repository:

1. **Start here** — read the relevant OpenWiki section for your area of change
2. **Read the AGENTS.md** for the specific package you're modifying (e.g., `internal/tui/views/AGENTS.md`)
3. **Follow the patterns**: never hardcode paths, always use `config.*` helpers (`OmoDir`/`OmoFile`/…); never copy/symlink to "activate"; always use `schema.GetValidator()` singleton; forms use `GetOpenCodeSchema()`
4. **For schema changes**: update `internal/schema/schema.json` (never edit by hand — re-sync from upstream `assets/omo.schema.json`) and keep `internal/config/types.go` 1:1 with the `[opencode]` sub-schema; keep root `omo.schema.json` aligned
5. **For tests**: use `config.SetBaseDir(tmpDir)` and seed profiles into the document; use `testify/assert` and `testify/require`
6. **No CI/Docker**: all validation is local — run `make test && make lint` before committing
