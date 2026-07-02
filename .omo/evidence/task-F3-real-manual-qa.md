# F3 Real Manual QA Evidence

Date: 2026-04-08
Mode: built binary + tmux interactive session

## Build
- `make build` succeeded and produced `./omo-profiler`

## Key live observations

### Scenario 1 — reasoningEffort
- Live cycle observed: `(none)` → `none` → `minimal` → `low` → `medium` → `high` → `xhigh` → `(none)`

### Scenario 2 — allow_non_gpt_model
- `build` form ended at `compVariant` with no `allow_non_gpt` field shown
- `hephaestus` form rendered `allow_non_gpt: [ ] [←/→]` after compaction fields

### Scenario 3 — providerOptions
- Empty state rendered `press 'a' to add`
- Added `alpha=42`
- Added `unicode_key=value/with:specials`
- Deleted first entry and summary updated to `1 options set`

### Scenario 4 — permission.bash
- Enter on string mode opened `Convert to per-command rules? (y/n)`
- Converted to object mode and edited rules live
- Added `npm`, cycled permissions, deleted original `bash` rule
- Summary updated to `1 rules [Enter to edit]`

### Scenario 5 — fallback_models
- Empty state rendered `press 'a' to add`
- Model selector opened successfully
- After selecting a model, editor screen stayed visually empty until another keypress refreshed the viewport
- After refresh, structured entry editing worked (`variant=fast`, `reasoning=none`)

### Scenario 6 — inline validation
- Invalid values rendered live errors:
  - `temperature: 3.0` → `✗ must be 0-2`
  - `top_p: 1.5` → `✗ must be 0-1`
  - `color: #fff` → `✗ invalid hex`
- Valid replacements removed errors:
  - `temperature: 1.5`
  - `top_p: 0.5`
  - `color: #FF6AC1`
- Max valid edge cases also rendered without error:
  - `temperature: 2.0`
  - `top_p: 1.0`
