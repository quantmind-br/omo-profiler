# INTERNAL/CONFIG

## OVERVIEW
Defines the `Config` struct ecosystem mapping `oh-my-opencode.json` schema and manages centralized configuration file paths.

## CORE RESPONSIBILITIES
- **Schema Authority**: Maintains Go struct representations of the OpenCode configuration JSON schema (`Config` struct).
- **Path Resolution**: Centralizes logic for standard directories (`~/.config/opencode/`) via `paths.go`.
- **Testing Hooks**: Provides `SetBaseDir` to redirect filesystem operations to temporary directories during tests.
- **Serialization**: Defines JSON tags ensuring correct marshaling/unmarshaling of config files.
- **Type Safety**: Enforces type correctness for complex nested structures like `AgentConfig`, `CategoryConfig`, and `PermissionConfig`.

## IMPACT & DEPENDENCIES
- **Central Dependency**: Used by `internal/profile` (loading/saving), `internal/tui` (displaying/editing), and `internal/schema` (validation).
- **High Risk**: Changes in `types.go` ripple through the entire application. Breaking changes here can corrupt user configuration files.
- **External Contract**: Must remain synchronized with the `oh-my-opencode` upstream schema.

## KEY FILES
- `types.go`: **The Source of Truth**. Contains the nested `Config` struct hierarchy.
    - `Config`: Root struct containing global settings and maps for Agents/Categories.
    - `AgentConfig`: Detailed configuration for individual AI agents (model, skills, permissions).
    - `PermissionConfig`: Granular security controls for tools and file access.
- `paths.go`: Path resolvers for `oh-my-opencode.json` and `profiles/` directory. Handles `os.UserHomeDir` abstraction.

## ANTI-PATTERNS
- **Schema Divergence**: Adding fields to `types.go` that do not exist in the official `oh-my-opencode` JSON schema.
- **Hardcoded Paths**: Constructing configuration paths manually instead of using `paths.ConfigFile()` or `paths.ProfilesDir()`.
- **Global State Abuse**: Modifying `baseDir` via `SetBaseDir` in production code (strictly for testing only).
- **Logic Leaks**: Adding business logic or validation methods to these data-only structs (logic belongs in `internal/profile` or `internal/schema`).
