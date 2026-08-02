package web

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/diogenes/omo-profiler/internal/profile"
	"github.com/diogenes/omo-profiler/internal/schema"
	"github.com/stretchr/testify/require"
)

func setupTestEnv(t *testing.T) {
	t.Helper()
	config.SetBaseDir(t.TempDir())
	require.NoError(t, config.EnsureDirs())
	t.Cleanup(config.ResetBaseDir)
}

func do(t *testing.T, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	rec := httptest.NewRecorder()
	newMux().ServeHTTP(rec, req)
	return rec
}

// seedProfile writes profiles.<name>.[opencode] into the omo document.
func seedProfile(t *testing.T, name, openCodeJSON string) {
	t.Helper()
	doc, err := config.LoadDocument()
	require.NoError(t, err)
	block, err := json.Marshal(map[string]json.RawMessage{
		config.OpenCodeKey: json.RawMessage(openCodeJSON),
	})
	require.NoError(t, err)
	require.NoError(t, doc.SetProfileBlock(name, block))
	doc.EnsureSchema()
	require.NoError(t, doc.Save())
}

func readProfileOpenCode(t *testing.T, name string) map[string]any {
	t.Helper()
	raw, err := profile.ExportOpenCode(name)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(raw, &m))
	return m
}

func hasBakInOmoDir(t *testing.T) bool {
	t.Helper()
	entries, err := os.ReadDir(config.OmoDir())
	require.NoError(t, err)
	for _, e := range entries {
		if strings.Contains(e.Name(), ".bak.") {
			return true
		}
	}
	return false
}

// activeResponse mirrors the JSON shape of activeJSON in handlers.go.
type activeResponse struct {
	DocumentExists bool   `json:"documentExists"`
	ProfileName    string `json:"profileName"`
	Modified       bool   `json:"modified"`
}

type profilesResponse struct {
	Active   activeResponse `json:"active"`
	Profiles []struct {
		Name   string `json:"name"`
		Active bool   `json:"active"`
	} `json:"profiles"`
}

func listProfiles(t *testing.T) profilesResponse {
	t.Helper()
	rec := do(t, "GET", "/api/profiles", "")
	require.Equal(t, 200, rec.Code)
	var resp profilesResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	return resp
}

// With no profile applied, the API reports no active profile.
func TestListProfilesReportsNoActive(t *testing.T) {
	setupTestEnv(t)
	seedProfile(t, "dev", `{"telemetry":false}`)
	seedProfile(t, "prod", `{"telemetry":true}`)

	resp := listProfiles(t)
	require.True(t, resp.Active.DocumentExists)
	require.Empty(t, resp.Active.ProfileName, "no profile is active")
	require.False(t, resp.Active.Modified)

	byName := map[string]bool{}
	for _, p := range resp.Profiles {
		byName[p.Name] = p.Active
	}
	require.Contains(t, byName, "dev")
	require.Contains(t, byName, "prod")
	require.False(t, byName["dev"], "no profile is active")
	require.False(t, byName["prod"])
}

// After applying a profile, the API reports it as active.
func TestListProfilesReportsApplied(t *testing.T) {
	setupTestEnv(t)
	seedProfile(t, "dev", `{"telemetry":false}`)
	seedProfile(t, "prod", `{"telemetry":true}`)

	require.Equal(t, 200, do(t, "POST", "/api/profiles/dev/activate", "").Code)

	resp := listProfiles(t)
	require.Equal(t, "dev", resp.Active.ProfileName)
	require.False(t, resp.Active.Modified)

	for _, p := range resp.Profiles {
		require.Equal(t, p.Name == "dev", p.Active)
	}
}

// A root that matches no profile is flagged as modified.
func TestListProfilesFlagsModifiedRoot(t *testing.T) {
	setupTestEnv(t)
	seedProfile(t, "dev", `{"telemetry":false}`)

	// Hand-edit the root so it no longer matches any profile.
	doc, err := config.LoadDocument()
	require.NoError(t, err)
	doc.SetRaw(config.OpenCodeKey, json.RawMessage(`{"telemetry":true}`))
	require.NoError(t, doc.Save())

	resp := listProfiles(t)
	require.True(t, resp.Active.Modified)
	require.Empty(t, resp.Active.ProfileName)
	for _, p := range resp.Profiles {
		require.False(t, p.Active)
	}
}

// Cloning must reproduce the whole profile block. Sibling harness blocks
// ([senpi], [codex]) live next to `[opencode]` and are easy to drop.
func TestCreateFromProfilePreservesSiblingBlocks(t *testing.T) {
	setupTestEnv(t)

	doc, err := config.LoadDocument()
	require.NoError(t, err)
	require.NoError(t, doc.SetProfileBlock("dev", json.RawMessage(
		`{"[opencode]":{"telemetry":false},"[senpi]":{"keep":"me"},"[codex]":{"sibling":"preserved"}}`,
	)))
	doc.EnsureSchema()
	require.NoError(t, doc.Save())

	require.Equal(t, 201, do(t, "POST", "/api/profiles", `{"name":"clone","from":"dev"}`).Code)

	doc, err = config.LoadDocument()
	require.NoError(t, err)
	block, ok, err := doc.ProfileBlock("clone")
	require.NoError(t, err)
	require.True(t, ok)

	var got map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(block, &got))
	require.JSONEq(t, `{"keep":"me"}`, string(got["[senpi]"]), "[senpi] must survive a clone")
	require.JSONEq(t, `{"sibling":"preserved"}`, string(got["[codex]"]), "[codex] must survive a clone")
	require.JSONEq(t, `{"telemetry":false}`, string(got["[opencode]"]))
}

func TestSaveProfileValidRoundTrips(t *testing.T) {
	setupTestEnv(t)
	// PUT updates; it does not create. New profiles come from POST.
	seedProfile(t, "dev", `{}`)

	rec := do(t, "PUT", "/api/profiles/dev", `{"telemetry":false}`)
	require.Equal(t, 200, rec.Code)

	rec = do(t, "GET", "/api/profiles/dev", "")
	require.Equal(t, 200, rec.Code)

	var got struct {
		Config map[string]any
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, false, got.Config["telemetry"])
}

func TestSaveProfileInvalidTypeReturns422AndDoesNotWrite(t *testing.T) {
	setupTestEnv(t)
	good := `{"telemetry":false}`
	seedProfile(t, "dev", good)

	rec := do(t, "PUT", "/api/profiles/dev", `{"telemetry":"nope"}`)
	require.Equal(t, 422, rec.Code)

	var resp struct {
		ValidationErrors []map[string]string
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.NotEmpty(t, resp.ValidationErrors)

	require.JSONEq(t, good, mustMarshal(t, readProfileOpenCode(t, "dev")))
}

func mustMarshal(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}

// PUT replaces the `[opencode]` block: the body is stored verbatim. GET hands
// back that same block, so the editor's load-edit-save round-trip carries
// unknown keys through untouched...
func TestSaveProfileRoundTripPreservesUnknownTopLevelKey(t *testing.T) {
	setupTestEnv(t)
	seedProfile(t, "dev", `{"future_key":123,"telemetry":false}`)

	rec := do(t, "GET", "/api/profiles/dev", "")
	require.Equal(t, 200, rec.Code)
	var got struct {
		Config json.RawMessage `json:"config"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Contains(t, string(got.Config), "future_key")

	require.Equal(t, 200, do(t, "PUT", "/api/profiles/dev", string(got.Config)).Code)

	m := readProfileOpenCode(t, "dev")
	require.Contains(t, m, "future_key")
	require.Equal(t, float64(123), m["future_key"])
}

// ...and omitting a key removes it. Merging the stored block into every write
// made unknown keys sticky: the editor could never delete one.
func TestSaveProfileOmittingKeyRemovesIt(t *testing.T) {
	setupTestEnv(t)
	seedProfile(t, "dev", `{"future_key":123,"telemetry":false}`)

	require.Equal(t, 200, do(t, "PUT", "/api/profiles/dev", `{"telemetry":true}`).Code)

	m := readProfileOpenCode(t, "dev")
	require.NotContains(t, m, "future_key", "a key absent from the body must be removed, not resurrected")
	require.Equal(t, true, m["telemetry"])
}

// Explicitly present zero values survive; the typed round-trip dropped them.
func TestSaveProfileKeepsExplicitZeroValues(t *testing.T) {
	setupTestEnv(t)
	seedProfile(t, "dev", `{}`)

	body := `{"default_run_agent":"","disabled_mcps":[],"telemetry":false}`
	require.Equal(t, 200, do(t, "PUT", "/api/profiles/dev", body).Code)

	require.JSONEq(t, body, mustMarshal(t, readProfileOpenCode(t, "dev")))
}

// A stale editor tab saving after a concurrent delete must report the
// deletion, not resurrect the profile.
func TestSaveProfileOnDeletedProfileIs404(t *testing.T) {
	setupTestEnv(t)

	rec := do(t, "PUT", "/api/profiles/ghost", `{"telemetry":true}`)
	require.Equal(t, 404, rec.Code)
	require.False(t, profile.Exists("ghost"), "a rejected save must not create the profile")
}

func TestCreateFromDefaultTemplateContainsAgents(t *testing.T) {
	setupTestEnv(t)

	rec := do(t, "POST", "/api/profiles", `{"name":"dev","from":"__default__"}`)
	require.Equal(t, 201, rec.Code)

	rec = do(t, "GET", "/api/profiles/dev", "")
	require.Equal(t, 200, rec.Code)
	var got struct {
		Config map[string]any
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	agents, ok := got.Config["agents"].(map[string]any)
	require.True(t, ok)
	for _, name := range []string{"sisyphus", "hephaestus", "prometheus", "oracle", "librarian", "explore", "multimodal-looker", "metis", "momus", "atlas", "sisyphus-junior"} {
		agent, ok := agents[name].(map[string]any)
		require.Truef(t, ok, "missing agent %s", name)
		model, _ := agent["model"].(string)
		require.NotEmptyf(t, model, "agent %s must have an explicit model in the default template", name)
	}
	atlas, ok := agents["atlas"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "anthropic/claude-sonnet-5", atlas["model"])
	hephaestus, ok := agents["hephaestus"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Explore thoroughly, then implement. Prefer small, testable changes.", hephaestus["prompt_append"])
	prometheus, ok := agents["prometheus"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Always interview first. Validate scope before planning.", prometheus["prompt_append"])
	sisyphus, ok := agents["sisyphus"].(map[string]any)
	require.True(t, ok)
	ultrawork, ok := sisyphus["ultrawork"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "max", ultrawork["reasoning"])
	_, hasVariant := ultrawork["variant"]
	require.False(t, hasVariant)
	_, hasSchema := got.Config["$schema"]
	require.False(t, hasSchema)
}

func TestDefaultTemplateMatchesRootTemplate(t *testing.T) {
	root, err := os.ReadFile(filepath.Join("..", "..", "template", "opencode-profile.json"))
	require.NoError(t, err)
	// embed.go requires the web asset and root template stay byte-identical.
	require.Equal(t, root, DefaultTemplate())
}

func TestCreateDuplicateReturns409(t *testing.T) {
	setupTestEnv(t)
	require.Equal(t, 201, do(t, "POST", "/api/profiles", `{"name":"dev","from":""}`).Code)
	require.Equal(t, 409, do(t, "POST", "/api/profiles", `{"name":"dev","from":""}`).Code)
}

func TestActivateWritesRootBlock(t *testing.T) {
	setupTestEnv(t)
	seedProfile(t, "dev", `{"telemetry":false}`)

	rec := do(t, "POST", "/api/profiles/dev/activate", "")
	require.Equal(t, 200, rec.Code)
	var r struct {
		Ok       bool   `json:"ok"`
		Name     string `json:"name"`
		Snapshot string `json:"snapshot"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &r))
	require.True(t, r.Ok)
	require.Equal(t, "dev", r.Name)

	// Activate must mutate omo.json: the profile block is now at the root.
	doc, err := config.LoadDocument()
	require.NoError(t, err)
	raw, ok := doc.Raw(config.OpenCodeKey)
	require.True(t, ok, "root [opencode] must be set after apply")
	require.JSONEq(t, `{"telemetry":false}`, string(raw))
}

func TestMutatingRoutesBackupOmoFile(t *testing.T) {
	setupTestEnv(t)

	// First create has nothing to back up.
	require.Equal(t, 201, do(t, "POST", "/api/profiles", `{"name":"dev","from":""}`).Code)
	require.False(t, hasBakInOmoDir(t))

	// Subsequent save backs up ~/.omo/omo.json into OmoDir.
	require.Equal(t, 200, do(t, "PUT", "/api/profiles/dev", `{"telemetry":false}`).Code)
	require.True(t, hasBakInOmoDir(t))

	entries, err := os.ReadDir(config.OmoDir())
	require.NoError(t, err)
	found := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), config.OmoBasename+".bak.") {
			found = true
			require.Equal(t, config.OmoDir(), filepath.Dir(filepath.Join(config.OmoDir(), e.Name())))
		}
	}
	require.True(t, found)
}
func TestRenameUpdatesActiveState(t *testing.T) {
	setupTestEnv(t)
	require.Equal(t, 201, do(t, "POST", "/api/profiles", `{"name":"dev","from":""}`).Code)
	require.Equal(t, 200, do(t, "POST", "/api/profiles/dev/activate", "").Code)

	rec := do(t, "POST", "/api/profiles/dev/rename", `{"newName":"prod"}`)
	require.Equal(t, 200, rec.Code)

	require.True(t, profile.Exists("prod"))
	require.False(t, profile.Exists("dev"))

	// The applied profile follows the rename automatically via comparison.
	active, err := profile.GetActive()
	require.NoError(t, err)
	require.Equal(t, "prod", active.ProfileName)

	var r struct {
		Name string `json:"name"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &r))
	require.Equal(t, "prod", r.Name)
}
func TestRenameProfile(t *testing.T) {
	setupTestEnv(t)
	seedProfile(t, "dev", `{"telemetry":false}`)

	rec := do(t, "POST", "/api/profiles/dev/rename", `{"newName":"prod"}`)
	require.Equal(t, 200, rec.Code)

	var r struct {
		Name string `json:"name"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &r))
	require.Equal(t, "prod", r.Name)

	require.True(t, profile.Exists("prod"))
	require.False(t, profile.Exists("dev"))
}

func TestDeleteProfile(t *testing.T) {
	setupTestEnv(t)
	seedProfile(t, "dev", `{"telemetry":false}`)
	seedProfile(t, "prod", `{"telemetry":true}`)

	rec := do(t, "DELETE", "/api/profiles/dev", "")
	require.Equal(t, 200, rec.Code)
	var r struct {
		Ok bool `json:"ok"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &r))
	require.True(t, r.Ok)

	require.False(t, profile.Exists("dev"))
	require.True(t, profile.Exists("prod"))
}

func TestSchemaEndpointReturnsOpenCodeSchema(t *testing.T) {
	setupTestEnv(t)
	rec := do(t, "GET", "/api/schema", "")
	require.Equal(t, 200, rec.Code)
	want, err := schema.GetOpenCodeSchema()
	require.NoError(t, err)
	require.Equal(t, want, rec.Body.Bytes())
}

func TestValidateStrictVsSave(t *testing.T) {
	setupTestEnv(t)

	rec := do(t, "POST", "/api/validate?mode=strict", `{"telemetry":false}`)
	require.Equal(t, 200, rec.Code)
	var strictR struct{ Valid bool }
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &strictR))
	require.False(t, strictR.Valid)

	rec = do(t, "POST", "/api/validate?mode=save", `{"telemetry":false}`)
	require.Equal(t, 200, rec.Code)
	var saveR struct{ Valid bool }
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &saveR))
	require.True(t, saveR.Valid)
}

func TestModelsCreateDuplicateUpdateDelete(t *testing.T) {
	setupTestEnv(t)

	require.Equal(t, 201, do(t, "POST", "/api/models", `{"displayName":"D","modelId":"m1","provider":"p"}`).Code)
	require.Equal(t, 409, do(t, "POST", "/api/models", `{"displayName":"D","modelId":"m1","provider":"p"}`).Code)

	rec := do(t, "PUT", "/api/models/p/m1", `{"displayName":"D2","modelId":"m2","provider":"p"}`)
	require.Equal(t, 200, rec.Code)

	rec = do(t, "GET", "/api/models", "")
	require.Equal(t, 200, rec.Code)
	require.Contains(t, rec.Body.String(), "m2")
	require.NotContains(t, rec.Body.String(), `"m1"`)

	require.Equal(t, 200, do(t, "DELETE", "/api/models/p/m2", "").Code)

	rec = do(t, "GET", "/api/models", "")
	require.NotContains(t, rec.Body.String(), `"m2"`)
}

// A rename onto a taken identity is a conflict, not a "not found" — the two
// were indistinguishable when every registry error mapped to 404.
func TestModelsUpdateStatusCodes(t *testing.T) {
	setupTestEnv(t)

	require.Equal(t, 201, do(t, "POST", "/api/models", `{"displayName":"A","modelId":"a","provider":"p"}`).Code)
	require.Equal(t, 201, do(t, "POST", "/api/models", `{"displayName":"B","modelId":"b","provider":"p"}`).Code)

	// Renaming `a` onto `b` collides.
	rec := do(t, "PUT", "/api/models/p/a", `{"displayName":"A","modelId":"b","provider":"p"}`)
	require.Equal(t, 409, rec.Code, "rename onto an existing model must be a conflict")

	// A genuinely missing model is still 404.
	rec = do(t, "PUT", "/api/models/p/ghost", `{"displayName":"G","modelId":"g","provider":"p"}`)
	require.Equal(t, 404, rec.Code)

	require.Equal(t, 404, do(t, "DELETE", "/api/models/p/ghost", "").Code)

	// The collision left both models intact.
	rec = do(t, "GET", "/api/models", "")
	require.Contains(t, rec.Body.String(), `"a"`)
	require.Contains(t, rec.Body.String(), `"b"`)
}
