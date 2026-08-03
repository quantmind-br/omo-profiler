# Project Overview: omo-profiler

## Purpose
TUI/CLI/Web profile manager for `oh-my-openagent` configuration. Profiles live as `profiles.<name>` entries inside the unified `~/.omo/omo.json` (or `omo.jsonc`) document; the editable payload is `profiles.<name>.[opencode]`. Activation is in-document (`Apply`). Provides creation, editing, import/export, schema validation, and upstream schema comparison.

## Version
`0.1.0` (defined in `internal/cli/root.go`)

## Key Features
- **Dashboard**: Overview of active profile with quick actions
- **Profile Wizard**: Step-by-step profile creation/editing (6 steps)
- **Profile List**: View, switch (emit export command), edit, and delete profiles
- **Diff View**: Side-by-side profile comparison
- **Import/Export**: JSON profile exchange with validation
- **Model Registry**: Manage custom AI models (local CRUD + models.dev API import)
- **Model Selector**: Pick models from registry when configuring agents
- **Template Select**: Start wizard from pre-configured templates
- **Schema Validation**: Validates against omo document / `[opencode]` schemas
- **Schema Check**: Compare embedded schema against upstream `assets/omo.schema.json` (monorepo: `packages/omo-config-core/src/schema/`; migration: `packages/omo-opencode/src/config-migration/`)

## Technology Stack
- **Language**: Go 1.25.6
- **TUI Framework**: Bubble Tea v1.3.10 (Charmbracelet)
- **CLI Framework**: Cobra v1.10.2
- **UI Styling**: Lipgloss v1.1.0
- **UI Components**: Bubbles v0.21.0
- **Schema Validation**: gojsonschema v1.2.0
- **Diffing**: go-diff v1.4.0
- **Testing**: Standard Go testing + testify v1.11.1

## CLI Commands
| Command | File | Description |
|---------|------|-------------|
| `omo-profiler` | (root) | Launch TUI |
| `omo-profiler list` | `cmd/list.go` | List all profiles |
| `omo-profiler current` | `cmd/current.go` | Show active profile (from env) |
| `omo-profiler switch <name>` | `cmd/switch.go` | Substitute profile keys into document root |
| `omo-profiler create` | `cmd/create.go` | Create profile from CLI |
| `omo-profiler import <file>` | `cmd/import.go` | Import profile into omo document |
| `omo-profiler export <name> <path>` | `cmd/export.go` | Export profile `[opencode]` to file |
| `omo-profiler models` | `cmd/models.go` | Manage model registry |
| `omo-profiler schema-check` | `cmd/schema_check.go` | Compare embedded vs upstream schema |

## Project Structure
```
omo-profiler/
├── cmd/omo-profiler/          # Entry point (main.go)
├── omo.schema.json            # Root omo document schema (mirrors embedded)
├── internal/
│   ├── cli/                   # Cobra CLI commands
│   │   ├── cmd/               # Individual commands (incl. web)
│   │   └── root.go            # Root command registration
│   ├── config/                # Document + [opencode] types & path resolution
│   │   ├── types.go           # Config struct (= [opencode] block)
│   │   ├── document.go        # Document load/save + profile blocks
│   │   ├── jsonc.go           # StripJSONC
│   │   └── paths.go           # OmoDir/OmoFile/ModelsFile/legacy migration helpers
│   ├── profile/               # CRUD logic & env activation
│   │   ├── profile.go         # Core Profile struct with Load/Save/Delete/List/Exists
│   │   ├── active.go          # Apply, ActiveName, GetActive, canonicalJSON
│   │   └── naming.go          # Profile name validation/sanitization
│   ├── schema/                # JSON schema validation & upstream comparison
│   │   ├── validator.go       # Singleton validator + GetOpenCodeSchema
│   │   ├── compare.go         # FetchUpstreamSchema, CompareSchemas, SaveDiff
│   │   └── schema.json        # Embedded omo document schema (go:embed)
│   ├── backup/                # Backup management (Create, List, Restore, Clean)
│   ├── diff/                  # ComputeDiff (side-by-side), ComputeUnifiedDiff
│   ├── models/                # Model registry management
│   │   ├── models.go          # ModelsRegistry CRUD, provider grouping
│   │   └── modelsdev.go       # models.dev API client
│   └── tui/                   # Bubble Tea TUI implementation
│       ├── app.go             # Root App model & state machine (10 states, ~840 lines)
│       ├── tui.go             # Run() entry point
│       ├── styles.go          # Shared Lipgloss styles
│       ├── keys.go            # Key bindings
│       ├── layout/layout.go   # Terminal size constraints & responsive helpers
│       └── views/             # Individual TUI views (18 view files)
│           ├── dashboard.go       # Main dashboard
│           ├── list.go            # Profile list
│           ├── wizard.go          # Wizard orchestrator
│           ├── wizard_name.go     # Step 1: Profile naming
│           ├── wizard_categories.go # Step 2: Category config
│           ├── wizard_agents.go   # Step 3: Agent config
│           ├── wizard_hooks.go    # Step 4: Event triggers
│           ├── wizard_other.go    # Step 5: Other settings (~2460 lines)
│           ├── wizard_review.go   # Step 6: Review & save
│           ├── step.go            # WizardStep interface
│           ├── diff.go            # Profile diff view
│           ├── import.go          # Import view
│           ├── export.go          # Export view
│           ├── model_registry.go  # Model CRUD view
│           ├── model_import.go    # models.dev API import view
│           ├── model_selector.go  # Model picker for agent forms
│           ├── template_select.go # Template-based wizard start
│           └── schema_check.go    # Upstream schema comparison view
└── internal/testdata/         # JSON fixtures for testing
```

## Important Files
- `internal/config/types.go`: `[opencode]` type authority — must stay in sync with the sub-schema
- `internal/config/document.go`: Unified omo document read/write
- `internal/tui/app.go`: Root TUI state machine & router (~840 lines)
- `internal/schema/compare.go`: Upstream schema drift detection (`assets/omo.schema.json`)
- `internal/profile/profile.go` / `active.go`: Persistence + env activation
- `Makefile`: Build, test, install automation
