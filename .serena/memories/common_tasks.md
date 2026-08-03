# Common Tasks & Reference

## Working with Profiles

### Load a Profile
```go
p, err := profile.Load("profile-name")
if err != nil {
    // handle error (os.IsNotExist / NotFoundError when missing)
}
// p.Name, p.Config (== profiles.<name>.[opencode]), p.Path (omo document),
// p.HasLegacyFields, p.LegacyFieldsWarning available
```

### Save a Profile
```go
p := &profile.Profile{
    Name:   "my-profile",
    Config: cfg,  // config.Config = [opencode] block
}
err := profile.Save(p)
// or: err := p.Save()  // read-modify-write of ~/.omo/omo.json
// Note: omitempty drops explicit zeros — not for sparse wizard saves.
```

### Sparse Save (wizard / selected fields)
```go
data, err := profile.MarshalSparse(&cfg, selection, preservedUnknown)
if err != nil { /* ... */ }
// Persist pre-marshalled [opencode] payload; preserves profile sibling keys
err = profile.SaveOpenCodeBlock("my-profile", data)
// or stage then save:
// err = profile.WriteOpenCodeBlockInto(doc, "my-profile", data)
// err = doc.Save()
```

### Check if Profile Exists
```go
if profile.Exists("profile-name") {
    // profiles.<name> present in the omo document
}
```

### List All Profiles
```go
profiles, err := profile.List()
// Returns []string of profile names from the document (sorted)
```

### Switch / Emit Activation
```go
applied, err := profile.Apply("profile-name")
// Applies profile to document root
// Returns: export profile via `Apply`profile-name
// Does NOT copy/symlink or mutate the omo document
```

### Get Current Active Profile
```go
active, err := profile.GetActive()
// Returns *ActiveConfig with:
// - Exists: bool — whether an omo document exists on disk
// - Config: config.Config — effective [opencode] (base ∪ active override)
// - ProfileName: string — active profile from env, or ""
// - Source: ActivationSource
// - SelectionHint / HintOnly / MissingProfile
```

### Compare Profile to a Config
```go
name, err := profile.ActiveName(doc)
// Normalizes both (strips $schema) before byte-for-byte comparison
```

## Path Resolution

### Get Config Paths
```go
config.OmoDir()            // ~/.omo/
config.OmoFile()           // ~/.omo/omo.jsonc if present, else ~/.omo/omo.json
config.ModelsFile()        // ~/.omo/models.json
config.OmoFile() // ~/.omo/omo.json
config.DefaultSchema       // const: upstream assets/omo.schema.json URL
config.LegacyConfigDir()   // migration detection only
```

### Ensure Directories Exist
```go
err := config.EnsureDirs()
// Creates ~/.omo
```

### Test Path Isolation
```go
config.SetBaseDir(t.TempDir())  // Redirect all paths under <tmp>/.omo/
defer config.ResetBaseDir()     // Restore real paths
// Seed profiles via Document.SetProfileBlock + Save — not as separate files
```

## Schema Validation

### Validate Config ([opencode])
```go
validator, err := schema.GetValidator()
if err != nil {
    // handle error
}

errors, err := validator.Validate(&cfg)  // From config.Config struct
// or
errors, err := validator.ValidateJSON(jsonData)  // From raw [opencode] JSON bytes

if len(errors) > 0 {
    for _, e := range errors {
        fmt.Printf("%s: %s\n", e.Path, e.Message)
    }
}
```

### Validate Whole Document
```go
errors, err := validator.ValidateDocument(omoBytes)
errors, err := validator.ValidateDocumentForSave(omoBytes)
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

### Get Schemas
```go
rawDocument := schema.GetEmbeddedSchema()       // full omo document schema
opencode, err := schema.GetOpenCodeSchema()     // [opencode] sub-schema for forms
```

## Backup Management

### Create Backup
```go
backupPath, err := backup.Create(config.OmoFile())
// Creates: omo.json.bak.YYYY-MM-DD-HHMMSS (under OmoDir)
// Use before profile save/delete/import — not for switch
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
// Overwrites current OmoFile with backup content
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
registry.Save()               // Persist to ~/.omo/models.json
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
    // seed profiles into the document, then exercise code
}

// Canonical seeder (profiles are document blocks, not files):
func writeProfile(t *testing.T, name, openCodeJSON string) {
    t.Helper()
    doc, err := config.LoadDocument()
    require.NoError(t, err)
    require.NoError(t, doc.SetProfileBlock(name, json.RawMessage(`{"[opencode]":`+openCodeJSON+`}`)))
    require.NoError(t, doc.Save())
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
