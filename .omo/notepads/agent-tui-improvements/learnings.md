# Learnings

## 2026-04-08 Session Start
- `effortLevels` defined in `wizard_categories.go:21`, shared by both wizard steps (9 refs total)
- `getLineForField` uses hardcoded offsets and form height constant `43` — must update when adding fields
- Sub-editor pattern: boolean flag → take over View/Update → handle input → emit result (see `handleSaveCustomModel` at line 1234)
- `providerOptions` currently preserved via implicit struct reuse in Apply — needs explicit logic
- `permission.bash` preserves objects via `originalBash` passthrough
- `fallback_models` schema: `anyOf[string, array[string|ModelObject]]`
- `allow_non_gpt_model` is hephaestus-only in the agents wizard; skip it in generic field cycling and keep Apply output nil for all other agents
- Inline validation hints can stay display-only by appending colored text to the rendered field value, without blocking submission
- `getLineForField` needs per-agent form height awareness when one agent gets an extra conditional row
- Expanding shared dropdown arrays is safe when SetConfig matches by value; the main risk is tests that assert old indices.
- `reasoningEffort` round-trips now cover the new schema values `none` and `minimal` in both agents and categories views.
- Bash perm editor follows exact same pattern as providerOptions: boolean flag → take over Update/View → handle input → Esc to exit
- `bashRuleKeys` + `bashRulePermIdx` parallel arrays replace `originalBash` passthrough in Apply
- `originalBash` kept for reference but no longer used in Apply — editor state is authoritative
- Object mode: Enter opens editor; String mode: Enter cycles dropdown — determined by `len(bashRuleKeys) > 0`
- getLineForField dynamic content limitation (same as providerOptions): offsets after bash section are inaccurate when editor is expanded, but editor takes over input so user can't navigate past it

## 2026-04-08 F3 Real Manual QA
- `reasoningEffort` cycles through all 7 UI states in the live TUI: `(none)`, `none`, `minimal`, `low`, `medium`, `high`, `xhigh`
- `allow_non_gpt_model` is hidden on `build` and rendered after compaction fields on `hephaestus` in the live wizard
- `providerOptions` empty state, add/edit/delete flow, and summary count all worked in the live TUI
- `permission.bash` convert-to-object prompt and per-command rule editor both worked in the live TUI
- Inline validation hints render live for invalid `color`, `temperature`, and `top_p`, and disappear for valid replacements including max values `2.0` and `1.0`
