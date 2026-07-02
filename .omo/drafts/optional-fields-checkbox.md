# Draft: Optional Fields with Checkbox Toggle

## Requirements (confirmed)
- ALL config fields are optional — none is required
- Checkbox per individual field (including within categories and agents)
- Unconfigured fields must NOT appear in the output JSON file
- Output file should contain ONLY explicitly configured fields
- oh-my-openagent uses hardcoded defaults for missing fields
- Remove all schema validation in review step — just show JSON preview
- New profiles start with ALL checkboxes unchecked (opt-in philosophy)
- Edit/template profiles: checkboxes reflect what's in the existing file

## Research Findings

### oh-my-openagent Default Handling
- Empty config `{}` is FULLY VALID — Zod's `.default()` fills everything
- Only `git_master` is "required" in JSON schema, but has inline defaults
- Deep merge: objects recursively merged, arrays replaced
- All 14 agents + 8 categories have hardcoded fallback chains
- **Conclusion: omo-profiler can safely produce minimal/empty configs**

### omo-profiler Wizard Architecture
- 6 steps: Name → Categories → Agents → Hooks → Other → Review
- Apply()/SetConfig() pattern: steps hold local state, flush to Config on Next
- Config struct: ALL fields use `omitempty` + pointer types (`*bool`, `*float64`)
- Nil pointers and empty maps/slices are already omitted from JSON
- Review step: schema validation blocks save on errors → TO BE REMOVED

### Field Inventory (100+ configurable fields)
- Step 2 (Categories): Dynamic list, 17 fields per category → per-field checkbox
- Step 3 (Agents): 14 agents, 35+ fields per agent → per-field checkbox
- Step 4 (Hooks): 48 hooks toggle (disabled_hooks array) → section-level include checkbox
- Step 5 (Other): 21 sections, ~50 fields → per-field checkbox
- Step 6 (Review): JSON preview only (validation removed)

## Technical Decisions (confirmed)
- Checkbox granularity: PER-FIELD within categories and agents
- Default state: ALL UNCHECKED for new profiles
- Validation: REMOVED entirely from review step
- Hooks step: top-level "include disabled_hooks" checkbox + existing toggle behavior
- Edit mode: checkboxes checked for fields present in loaded profile
- Template mode: checkboxes checked for fields defined in template

## Scope Boundaries
- INCLUDE: Checkbox toggles for all configurable fields in wizard steps 2-5
- INCLUDE: Modified Apply() logic to only write checked fields to Config
- INCLUDE: Modified SetConfig() to set checkboxes based on existing data
- INCLUDE: Removal of schema validation in review step
- INCLUDE: JSON output only contains explicitly configured fields
- EXCLUDE: Changes to oh-my-openagent itself
- EXCLUDE: New config fields
- EXCLUDE: Changes to CLI commands (non-TUI)
- EXCLUDE: Changes to profile import/export CLI logic
