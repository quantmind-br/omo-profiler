# Common Tasks & Reference

## Working with Profiles

### Load a Profile
```go
p, err := profile.Load("profile-name")
if err != nil {
    // handle error
}
// p.Name, p.Config, p.Path available
```

### Save a Profile
```go
p := &profile.Profile{
    Name:   "my-profile",
    Config: cfg,  // config.Config struct
}
err := profile.Save(p)
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
// Updates .active-profile sidecar
```

### Get Current Active Profile
```go
active, err := profile.GetActive()
// Returns *ActiveProfileInfo with:
// - ProfileName: name of active profile
// - Exists: whether profile file exists
// - IsOrphan: true if active config doesn't match any profile
```

## Path Resolution

### Get Config Paths
```go
config.ConfigDir()      // ~/.config/opencode/
config.ProfilesDir()    // ~/.config/opencode/profiles/
config.ConfigFile()     // ~/.config/opencode/oh-my-opencode.json
config.ModelsFile()     // ~/.config/opencode/models.json
```

### Ensure Directories Exist
```go
err := config.EnsureDirs()
// Creates config and profiles directories
```

## Schema Validation

### Validate Config
```go
validator, err := schema.GetValidator()
if err != nil {
    // handle error
}

errors, err := validator.ValidateJSON(jsonData)
if err != nil {
    // handle error
}
if len(errors) > 0 {
    // validation failed
}
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
// Returns error if name contains invalid characters
// Valid: alphanumeric, underscores, hyphens
```

### Sanitize Profile Name
```go
name := profile.SanitizeName("My Profile!@#")
// Returns: "my-profile"
```

## TUI Message Types

### Navigation Messages (Views → App)
```go
NavToListMsg
NavToWizardMsg
NavToEditorMsg
NavToDiffMsg
NavToImportMsg
NavToExportMsg
NavToModelsMsg
NavToModelImportMsg
NavToTemplateSelectMsg
NavigateToDashboardMsg
```

### Wizard Messages
```go
WizardNextMsg
WizardBackMsg
WizardSaveMsg
WizardCancelMsg
```

### Operation Messages
```go
SwitchProfileMsg{Name: "profile-name"}
EditProfileMsg{Name: "profile-name"}
DeleteProfileMsg{Name: "profile-name"}
ImportDoneMsg{Path: "/path/to/file"}
ExportDoneMsg{Path: "/path/to/export"}
```
