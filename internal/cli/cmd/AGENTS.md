# CLI COMMANDS KNOWLEDGE BASE

## OVERVIEW
Implementation of Cobra CLI commands for headless profile management. Bridge between terminal input and `internal/profile` logic.

## COMMANDS
- **list**: Tabulates profiles; marks active with `*` and `(active)`.
- **current**: Prints active profile name or `(custom - unsaved)` if orphaned.
- **switch**: Triggers `backup.Create` then `profile.SetActive` (copy-based).
- **import**: Loads JSON file and saves to profile directory.
- **export**: Saves existing profile to a target JSON file.
- **create**: Headless cloning via `--from`; wizard logic remains in TUI.
- **models**: Lists supported LLM providers and models from registry.

## CONVENTIONS
- **Exported Vars**: All commands must be exported (e.g., `SwitchCmd`) for registration in `root.go`.
- **Thin Wrappers**: Commands should contain minimal logic; delegate to `internal/profile` or `internal/backup`.
- **StdErr for Errors**: Print error messages to `os.Stderr` and exit with code 1.
- **RunE vs Run**: Prefer `RunE` to return errors to Cobra for consistent reporting.
- **Args Validation**: Use `cobra.ExactArgs(n)` or `cobra.MaximumNArgs(n)` for usage enforcement.

## ANTI-PATTERNS
- **Fat Commands**: Do not implement business logic or FS operations directly in `Run`.
- **Interactive Prompts**: CLI commands must be non-interactive; use TUI for wizards.
- **Global State**: Avoid modifying global variables; use flags and arguments.
- **Raw Paths**: Never hardcode `~/.config`; use `config.Paths` helpers.
