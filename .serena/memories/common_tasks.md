# Common Tasks & Reference

## Working with Profiles

### Load a Profile
```go
p, err := profile.Load("profile-name")
if err != nil {
    // handle error
}
// p.Name, p.Config, p.Path, p.HasLegacyFields, p.LegacyFieldsWarning available
```

### Save a Profile
```go
p := &profile.Profile{
    Name:   "my-profile",
    Config: cfg,  // config.Config struct
}
err := profile.Save(p)
// or: err := p.Save()
```

### Check if Profile Exists
```go
if profile.Exists("profile-name") {
    // profile exists
}
```

### List All Profiles
```go
profiles, err := profile.List()
// Returns []string of profile names (no .json extension)
```

### Switch Active Profile
```go
err := profile.SetActive("profile-name")
// Copies profile content to oh-my-opencode.json
// Updates .active-profile sidecar for O(1) lookup
```

### Get Current Active Profile
```go
active, err := profile.GetActive()
// Returns *ActiveConfig with:
// - Exists: bool — whether oh-my-opencode.json exists
// - Config: config.Config — the active config
// - ProfileName: string — name of matching profile (or "(custom)" if orphan)
// - IsOrphan: bool — true if active config doesn't match any profile
```

### Compare Profile to Active Config
```go
matches := p.MatchesConfig(&cfg)
// Normalizes both (strips $schema) before byte-for-byte comparison
```

## Path Resolution

### Get Config Paths
```go
config.ConfigDir()      // ~/.config/opencode/
config.ProfilesDir()    // ~/.config/opencode/profiles/
config.ConfigFile()     // ~/.config/opencode/oh-my-opencode.json
config.ModelsFile()     // ~/.config/opencode/models.json
config.DefaultSchema    // const: upstream schema URL
```

### Ensure Directories Exist
```go
err := config.EnsureDirs()
// Creates config and profiles directories
```

### Test Path Isolation
```go
config.SetBaseDir(t.TempDir())  // Redirect all paths to temp dir
defer config.ResetBaseDir()     // Restore real paths
```

## Schema Validation

### Validate Config
```go
validator, err := schema.GetValidator()
if err != nil {
    // handle error
}

errors, err := validator.Validate(&cfg)  // From config.Config struct
// or
errors, err := validator.ValidateJSON(jsonData)  // From raw JSON bytes

if len(errors) > 0 {
    for _, e := range errors {
        fmt.Printf("%s: %s\n", e.Path, e.Message)
    }
}
```

### Compare Embedded vs Upstream Schema
```go
result, err := schema.CompareSchemas()
if result.Identical {
    // schemas match
} else {
    fmt.Println(result.Diff)  // unified diff output
    path, err := schema.SaveDiff("/some/dir", result.Diff)  // save to file
}
```

### Get Embedded Schema
```go
rawBytes := schema.GetEmbeddedSchema()
```

## Backup Management

### Create Backup
```go
backupPath, err := backup.Create(config.ConfigFile())
// Creates: oh-my-opencode.json.bak.YYYY-MM-DD-HHMMSS
```

### List Backups
```go
backups, err := backup.List()
// Returns []BackupInfo sorted by timestamp (most recent first)
// Each: Path, Timestamp, Name
```

### Restore Backup
```go
err := backup.Restore(backupPath)
// Overwrites oh-my-opencode.json with backup content
```

### Clean Old Backups
```go
err := backup.Clean(5) // Keep only 5 most recent
```

## Diff Computation

### Side-by-Side Diff
```go
result, err := diff.ComputeDiff(json1, json2)
// result.Left, result.Right — []DiffLine
// Each DiffLine: Text, Type (DiffEqual/DiffAdded/DiffRemoved), LineNum
```

### Unified Diff
```go
output := diff.ComputeUnifiedDiff("old-name", "new-name", oldBytes, newBytes)
// Returns unified diff format string
```

## Model Registry

### Load Registry
```go
registry, err := models.Load()
// Auto-recovers from corrupted JSON (backs up to .bak)
```

### CRUD Operations
```go
registry.Add(model)           // Add RegisteredModel
registry.Update(model)        // Update by ModelID
registry.Delete(modelID)      // Remove by ModelID
registry.List()               // Get all models (copy)
registry.ListByProvider()     // Get []ProviderGroup (sorted)
registry.Save()               // Persist to models.json
```

### models.dev API
```go
response, err := models.FetchModelsDevRegistry()
providers := response.ListProviders()      // []ProviderWithCount
models := response.GetProviderModels("openai")
registered := externalModel.ToRegisteredModel()
```

## Testing Helpers

### Setup Test Environment
```go
func setupTestEnv(t *testing.T) func() {
    t.Helper()
    tmpDir := t.TempDir()
    config.SetBaseDir(tmpDir)
    return func() {
        config.ResetBaseDir()
    }
}

// Usage:
func TestSomething(t *testing.T) {
    cleanup := setupTestEnv(t)
    defer cleanup()
    // ... test code
}
```

## Profile Naming

### Validate Profile Name
```go
err := profile.ValidateName("my-profile")
// Returns ErrEmptyName or ErrInvalidName if invalid
// Valid: alphanumeric, underscores, hyphens only (regex: ^[a-zA-Z0-9_-]+$)
```

### Sanitize Profile Name
```go
name := profile.SanitizeName("My Profile!@#")
// Returns: "MyProfile" (strips invalid chars, trims leading/trailing separators)
```

## TUI Message Types

### Navigation Messages (Views → App)
```go
// Dashboard navigation
NavToListMsg, NavToWizardMsg, NavToEditorMsg, NavToDiffMsg
NavToImportMsg, NavToExportMsg, NavToModelsMsg, NavToTemplateSelectMsg

// List navigation
NavigateToDashboardMsg, NavigateToWizardMsg

// Schema/Model navigation
NavToSchemaCheckMsg, NavToModelImportMsg, NavToWizardFromTemplateMsg
```

### Wizard Messages
```go
WizardNextMsg, WizardBackMsg, WizardSaveMsg, WizardCancelMsg
```

### Operation Messages
```go
SwitchProfileMsg{Name: "profile-name"}
EditProfileMsg{Name: "profile-name"}
DeleteProfileMsg{Name: "profile-name"}
ImportDoneMsg{Path: "/path/to/file"}
ExportDoneMsg{Path: "/path/to/export"}
```

### Back/Cancel Messages
```go
ImportCancelMsg, ExportCancelMsg
ModelRegistryBackMsg, ModelImportBackMsg
TemplateSelectCancelMsg, SchemaCheckBackMsg
```

### Model Messages
```go
ModelSavedMsg, ModelDeletedMsg
ModelSelectedMsg, ModelSelectorCancelMsg, PromptSaveCustomMsg
ModelImportDoneMsg
```
