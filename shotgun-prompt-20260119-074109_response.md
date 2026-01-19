# Code Improvements - Analyzed & Prioritized

**Analyzed**: 2026-01-19
**Status**: Ready for implementation

## Phase 1: Core Functionality (High Priority)

### 1. CLI Validate Command
**Effort**: Trivial | **Value**: High

Add a `validate` command to check configuration files against the schema.

**Implementation**:
- Create `internal/cli/cmd/validate.go`
- Read `config.ConfigFile()`, use `schema.GetValidator().ValidateJSON(data)`
- Print success or list errors (pattern: `import.go`)
- Add `--file` flag to validate arbitrary files
- Return exit code 1 on failure for scripting

**Files**: `internal/cli/cmd/validate.go`, `internal/cli/root.go`

**Pattern**: `internal/cli/cmd/import.go`, `internal/cli/cmd/current.go`

---

### 2. TUI Import/Export Implementation
**Effort**: Medium | **Value**: High

Implement the placeholder `stateImport` and `stateExport` views.

**Implementation**:
- Create `views/file_ops.go` with `ImportView` and `ExportView`
- Use `textinput` for file path entry (pattern: `wizard_name.go`)
- ImportView: Read file → validate with `schema.ValidateJSON` → `profile.Save`
- ExportView: Select profile → `profile.Load` → write to specified path
- Add clear error messages for invalid paths/JSON

**Files**: `internal/tui/views/file_ops.go`, `internal/tui/app.go`

**Pattern**: `internal/tui/views/wizard_name.go`, `internal/profile/profile.go`

---

## Phase 2: UX Polish (Medium Priority)

### 3. Diff Against Active Configuration
**Effort**: Small | **Value**: High

Allow comparing stored profiles against the live configuration file.

**Implementation**:
- In `loadProfiles`, call `profile.GetActive()` and prepend `"(Active Config)"` to list
- In `computeDiff`, if selection is `"(Active Config)"`, use `ActiveConfig.Config` instead of `profile.Load()`
- Display `(orphan)` indicator if `IsOrphan == true`

**Files**: `internal/tui/views/diff.go`

**Pattern**: `internal/profile/active.go` (`GetActive` returns `ActiveConfig`)

---

### 4. Model Selector for Category Configuration  
**Effort**: Small | **Value**: Medium

Replace plain textinput with `ModelSelector` for category model field.

**Implementation**:
- Add `selectingModel` state and `modelSelector` field to `categoryConfig` struct
- Handle `enter` on `catFieldModel` to switch to selector mode
- Handle `ModelSelectedMsg` to update the value
- Mirror logic from `wizard_agents.go` (lines 523+)

**Files**: `internal/tui/views/wizard_categories.go`

**Pattern**: `internal/tui/views/wizard_agents.go` (selectingModel handling)

---

## Phase 3: Future Consideration

### 5. Backup Management (CLI-first)
**Effort**: Small | **Value**: Low-Medium | **Status**: Deferred

Instead of TUI view, add CLI commands for backup management.

**Rationale**: Backups are a safety net rarely accessed. TUI has 7 states already. CLI is sufficient for this use case.

**If implemented**:
```bash
omo-profiler backup list          # List backups with timestamps
omo-profiler backup restore <id>  # Restore specific backup
omo-profiler backup clean [--keep=5]  # Rotate old backups
```

**Files**: `internal/cli/cmd/backup.go`

---

### 6. Skills JSON Editor Enhancement (Not New Wizard Step)
**Effort**: Medium | **Value**: Medium | **Status**: Adjusted

**Original proposal**: Create structured wizard step with checkboxes.

**Problem**: `Config.Skills` is `json.RawMessage` that preserves both arrays AND objects. A checkbox UI would break object-based skill configurations.

**Adjusted approach**: Enhance existing JSON editor in `wizard_other.go`:
- Add real-time JSON syntax validation
- Add preset dropdown ("Default", "Minimal", "All Skills")
- Keep raw JSON for full flexibility
- Show validation errors inline

**Files**: `internal/tui/views/wizard_other.go`

---

## Implementation Order

```
1. CLI Validate Command (trivial, immediate value)
      ↓
2. TUI Import/Export (completes existing placeholder)
      ↓
3. Diff Against Active (small, high UX value)
      ↓
4. Model Selector for Categories (consistency fix)
      ↓
[Future] CLI Backup commands
[Future] Skills editor enhancement
```

## Summary

| Feature | Verdict | Priority | Effort |
|---------|---------|----------|--------|
| CLI Validate | ✅ Keep | High | Trivial |
| TUI Import/Export | ✅ Keep | High | Medium |
| Diff Active Config | ✅ Keep | Medium | Small |
| Model Selector Categories | ✅ Keep | Medium | Small |
| TUI Backup Manager | 🔄 Defer → CLI | Low | Small |
| Skills Wizard Step | ⚠️ Adjust → Editor | Low | Medium |

**Total**: 4 features ready, 2 adjusted/deferred
