# Decisions

## 2026-04-08 Session Start
- All 6 improvements in one plan, tests-after strategy
- providerOptions: primitives only (string/number/boolean), no nested objects
- fallback_models: structured for model/variant/reasoningEffort, rawJSON for complex
- allow_non_gpt_model: conditional, only shown on hephaestus
- Validation: display-only inline red text, does not block submission
- File: all changes stay in wizard_agents.go (matches project convention)
