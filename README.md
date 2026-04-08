# omo-profiler

TUI profile manager for `oh-my-openagent` configuration files. Built with Go 1.25.6, Bubble Tea, and Cobra.

## Installation

```bash
go install github.com/diogenes/omo-profiler/cmd/omo-profiler@latest
```

Or build from source:

```bash
git clone https://github.com/diogenes/omo-profiler.git
cd omo-profiler
make install
```

## Quick Start

```bash
# Launch the interactive TUI
omo-profiler

# List all profiles
omo-profiler list

# Switch to a profile
omo-profiler switch my-profile

# Check which profile is active
omo-profiler current
```

## CLI Reference

| Command | Description |
|---------|-------------|
| `omo-profiler` | Launch interactive TUI (default) |
| `omo-profiler list` | List all profiles; active profile marked with `*` |
| `omo-profiler current` | Show active profile name; shows `(custom - unsaved)` if orphaned |
| `omo-profiler switch <name>` | Activate a profile; creates timestamped backup first |
| `omo-profiler import <file>` | Import profile from JSON file; validates against schema |
| `omo-profiler export <name> <path>` | Export profile to JSON file |
| `omo-profiler create <name>` | Create profile headlessly; requires `--from <template>` flag |
| `omo-profiler models list` | List registered models, grouped by provider |
| `omo-profiler models add` | Add model interactively; prompts for name, ID, provider |
| `omo-profiler models edit <id>` | Edit existing model with interactive prompts |
| `omo-profiler models delete <id>` | Delete a model with confirmation prompt |
| `omo-profiler schema-check` | Compare embedded vs upstream schema; requires `--output/-o` flag for diff file |

## TUI Features

The TUI provides 10 application states:

| State | Purpose |
|-------|---------|
| **Dashboard** | Active profile overview with navigation menu |
| **Profile List** | Browse, filter, switch, edit, and delete profiles |
| **Profile Wizard** | 6-step creation: Name → Categories → Agents → Hooks → Other → Review |
| **Profile Editing** | Re-uses wizard flow pre-populated with existing config |
| **Diff View** | Side-by-side profile comparison with dual viewports |
| **Import/Export** | File-based profile import/export within TUI |
| **Model Registry** | Local model CRUD with in-place form editing |
| **Model Import** | Async fetch from models.dev API with fuzzy filtering and multi-select |
| **Template Select** | Profile template picker for wizard initialization |
| **Schema Check** | Upstream schema diff viewer with save-to-file |

## Architecture

```
omo-profiler/
├── cmd/omo-profiler/          # Entry point → cli.Execute()
├── internal/
│   ├── cli/                   # Cobra root + cmd/ subcommands
│   │   └── cmd/               # 8 command files
│   ├── config/                # Schema authority (types.go, ~37 fields) + path resolution
│   ├── profile/               # CRUD, switching, naming, sparse serialization
│   ├── tui/                   # Bubble Tea app (state machine + router)
│   │   ├── views/             # 18 view files + 10 test files
│   │   └── layout/            # Terminal width helpers, min-size constants
│   ├── schema/                # Embedded JSON schema + gojsonschema validator
│   ├── models/                # LLM model registry + models.dev API client
│   ├── diff/                  # Side-by-side + unified diff
│   ├── backup/                # Timestamped backup rotation + restore/clean
│   └── testdata/              # JSON fixtures for cross-package tests
├── template/                  # Default oh-my-openagent.json template
└── Makefile
```

## Development

### Make Targets

| Target | Description |
|--------|-------------|
| `make build` | Build binary to `./omo-profiler` |
| `make install` | Install to `~/.local/bin/omo-profiler` |
| `make uninstall` | Remove from `~/.local/bin` |
| `make test` | Run tests with `go test -v ./...` |
| `make lint` | Run golangci-lint (checks for presence first) |
| `make clean` | Remove build artifacts + `go clean` |

### Testing

Tests are co-located as `*_test.go` files (36+ total). The project uses table-driven tests in 14+ files and requires `config.SetBaseDir(t.TempDir())` via the `setupTestEnv` helper for test isolation.

```bash
make test
```

## How It Works

### Profile Storage

Profiles are stored as JSON files in `~/.config/opencode/profiles/<name>.json`. The active configuration lives at `~/.config/opencode/oh-my-openagent.json`.

### Profile Switching

Switching uses a copy-based mechanism (not symlinks) for fsnotify compatibility. Before switching, a timestamped backup is created automatically.

Active profile tracking uses a `.active-profile` sidecar file for O(1) lookup, with a content scan fallback (O(N)) for recovery.

### Sparse Serialization

The application writes only user-selected fields to preserve unknown JSON keys from upstream configurations. This is handled by `internal/profile/sparse.go` via the `MarshalSparse` function.

### Schema Validation

The embedded JSON schema (6067 lines, loaded via `go:embed`) validates all profiles against the upstream schema at:

```
https://raw.githubusercontent.com/code-yeongyu/oh-my-openagent/dev/assets/oh-my-opencode.schema.json
```

Validation uses the gojsonschema library with a singleton validator pattern via `schema.GetValidator()`.

---

Version: 0.1.0
