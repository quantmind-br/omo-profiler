# Issues

(none yet)

## 2026-04-08 F2 Code Quality Review
- `wizard_agents.go`: `providerOptions` editor is lossy for nested/object/array values because `SetConfig` stringifies with `fmt.Sprintf("%v", ...)` and `Apply` only reparses float/bool/string.
- `wizard_agents.go`: fallback raw editor treats `Esc` the same as `Enter`, so the advertised cancel path still saves edited raw JSON.
- `wizard_agents.go`: bash conversion prompt mutates `permBashIdx` on `n`/`Esc`, so dismissing the prompt changes data.
- `wizard_agents_test.go`: no coverage for inline validation rendering or the new editor cancel paths; current tests focus on round-trip serialization only.
- `wizard_categories_test.go`: no coverage for the newly added `max_prompt_tokens` SetConfig/Apply path.

## 2026-04-08 F3 Real Manual QA
- `fallback_models` model picker selection does not visibly refresh the fallback editor after choosing a model. The selection only becomes visible after a later keypress triggers `viewport.SetContent`, so the first post-select screen still looks empty in real use.
