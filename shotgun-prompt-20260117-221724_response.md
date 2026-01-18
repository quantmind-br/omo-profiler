# Performance Optimizations - omo-profiler

> Analyzed and prioritized. Implementation order reflects dependencies and effort-to-value ratio.

---

## Phase 1: Quick Wins (Implement First)

### 1. Remove unused imports from production binary
**ID**: perf-003 | **Impact**: Low | **Effort**: Trivial

**Problem**: `main.go` contains blank imports including test dependencies (`stretchr/testify/assert`) that compile into the production binary.

**Affected Files**:
- `cmd/omo-profiler/main.go`

**Implementation**:
Remove ALL blank imports from main.go:
```go
// DELETE these lines:
_ "github.com/charmbracelet/bubbles/list"
_ "github.com/charmbracelet/bubbletea"
_ "github.com/charmbracelet/lipgloss"
_ "github.com/sergi/go-diff/diffmatchpatch"
_ "github.com/stretchr/testify/assert"   // <-- TEST DEPENDENCY IN PROD!
_ "github.com/xeipuuv/gojsonschema"
```

These packages are already used in internal packages and don't need blank imports for `go mod` preservation.

**Expected Result**: Cleaner binary, ~0.5-1MB size reduction

---

### 2. Implement Singleton pattern for JSON Schema Validator
**ID**: perf-002 | **Impact**: Medium | **Effort**: Small

**Problem**: `NewValidator()` parses the 64KB embedded schema JSON every time it's called (3 call sites in production).

**Affected Files**:
- `internal/schema/validator.go`

**Implementation**:
```go
var (
    validatorInstance *Validator
    validatorOnce     sync.Once
    validatorErr      error
)

// GetValidator returns the singleton validator instance
func GetValidator() (*Validator, error) {
    validatorOnce.Do(func() {
        loader := gojsonschema.NewBytesLoader(schemaJSON)
        schema, err := gojsonschema.NewSchema(loader)
        if err != nil {
            validatorErr = err
            return
        }
        validatorInstance = &Validator{schema: schema}
    })
    return validatorInstance, validatorErr
}

// NewValidator is deprecated, use GetValidator()
func NewValidator() (*Validator, error) {
    return GetValidator()
}
```

Then update call sites to use `GetValidator()`:
- `internal/cli/cmd/import.go:42`
- `internal/tui/views/wizard_review.go:112`
- `internal/tui/views/editor.go:546`

**Expected Result**: Zero allocation after first use, near-instant validator access

---

## Phase 2: High Impact (Implement After Phase 1)

### 3. Optimize active profile resolution with state file
**ID**: perf-001 | **Impact**: High | **Effort**: Medium

**Problem**: `GetActive()` iterates through ALL profiles in `profiles/` directory, reading and parsing each file to find the active one. This is O(N) file reads + O(N) JSON unmarshals + O(2N) JSON marshals for comparison.

**Affected Files**:
- `internal/profile/active.go`

**Implementation**:

1. Create state file at `~/.config/opencode/.active-profile` storing:
```json
{
  "name": "profile-name",
  "hash": "sha256-of-config-content"
}
```

2. Modify `SetActive()` to update state file when switching profiles

3. Modify `GetActive()`:
```go
func GetActive() (*ActiveConfig, error) {
    // 1. Check if config exists
    configPath := config.ConfigFile()
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        return &ActiveConfig{Exists: false}, nil
    }
    
    // 2. Read current config
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, err
    }
    
    var cfg config.Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    
    // 3. Try fast path: check state file
    if state, err := loadActiveState(); err == nil {
        if profile, err := Load(state.Name); err == nil {
            if profile.MatchesConfig(&cfg) {
                return &ActiveConfig{
                    Exists:      true,
                    Config:      cfg,
                    ProfileName: state.Name,
                    IsOrphan:    false,
                }, nil
            }
        }
    }
    
    // 4. Fallback: full scan (state file missing/stale)
    // ... existing scan logic ...
}
```

**Tradeoffs**: Requires state file synchronization on profile switch

**Expected Result**: O(1) lookup in common case, O(N) fallback only when state is invalid

---

## Deferred (Re-evaluate After Phase 2)

### 4. Optimize configuration comparison logic
**ID**: perf-004 | **Impact**: Low (after perf-001) | **Effort**: Small

**Status**: DEFERRED - Implement perf-001 first, then reassess.

**Rationale**: After perf-001, config comparison happens at most 1-2 times (not N times). The benefit becomes negligible.

**If Still Needed**:
- Use `reflect.DeepEqual` or `github.com/google/go-cmp/cmp` instead of JSON marshal comparison
- Temporarily unset Schema field on copies before comparing

---

## Discarded

### ~~Memoize TUI View Rendering~~
**ID**: perf-005 | **Status**: DISCARDED

**Reason**: Premature optimization. 
- Agent list has only 15 items
- Settings view has ~14 sections
- Lipgloss rendering for these scales is sub-millisecond
- Cache invalidation complexity would exceed any performance benefit
- This is the standard Bubble Tea pattern - cheap re-renders are expected

If lag is observed, profile first to identify the actual bottleneck.

---

## Implementation Checklist

- [ ] Phase 1.1: Remove blank imports from main.go
- [ ] Phase 1.2: Implement validator singleton
- [ ] Phase 1.2: Update 3 call sites to use GetValidator()
- [ ] Phase 2: Add active profile state file
- [ ] Phase 2: Update SetActive() to write state
- [ ] Phase 2: Update GetActive() with fast path
- [ ] Run tests after each phase
- [ ] Verify binary size reduction after Phase 1
