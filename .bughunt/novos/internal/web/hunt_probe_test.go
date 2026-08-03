package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/diogenes/omo-profiler/internal/profile"
	"github.com/stretchr/testify/require"
)

// crossOrigin issues the request a hostile page can make from a browser with no
// CORS preflight: a "simple" request (GET/POST, text/plain body).
func crossOrigin(t *testing.T, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	rec := httptest.NewRecorder()
	newMux().ServeHTTP(rec, req)
	return rec
}

// The server binds localhost with no auth, no CSRF token and no Origin check.
// Any page the user visits can therefore drive its mutating endpoints.
func TestHuntCrossOriginWriteIsRejected(t *testing.T) {
	setupTestEnv(t)
	seedProfile(t, "victim", `{"default_run_agent":"keep"}`)

	rec := crossOrigin(t, "POST", "/api/profiles", `{"name":"attacker"}`)
	require.NotEqual(t, http.StatusCreated, rec.Code,
		"cross-origin POST created a profile (body=%s)", rec.Body.String())
	require.False(t, profile.Exists("attacker"),
		"a page on https://evil.example created profiles.attacker in the user's omo.json")
}

func TestHuntCrossOriginImportIsRejected(t *testing.T) {
	setupTestEnv(t)
	rec := crossOrigin(t, "POST", "/api/import",
		`{"name":"pwned","config":{"default_run_agent":"attacker"}}`)
	require.NotEqual(t, http.StatusOK, rec.Code,
		"cross-origin POST /api/import succeeded (body=%s)", rec.Body.String())
}

// Concurrent saves of distinct profiles must all survive: the document is one
// file, so a lost update destroys a profile the user just wrote.
func TestHuntConcurrentSavesAllSurvive(t *testing.T) {
	setupTestEnv(t)
	const n = 8
	for i := range n {
		seedProfile(t, fmt.Sprintf("p%d", i), `{}`)
	}

	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			body := fmt.Sprintf(`{"default_run_agent":"agent-%d"}`, i)
			do(t, "PUT", fmt.Sprintf("/api/profiles/p%d", i), body)
		}()
	}
	wg.Wait()

	for i := range n {
		got := readProfileOpenCode(t, fmt.Sprintf("p%d", i))
		require.Equal(t, fmt.Sprintf("agent-%d", i), got["default_run_agent"],
			"profile p%d lost its concurrent update", i)
	}
}

// M2 round-trip: what /api/profiles/{name}/export hands the user must import
// back as the same value.
func TestHuntExportImportRoundTrip(t *testing.T) {
	setupTestEnv(t)
	const payload = `{"default_run_agent":"","disabled_mcps":[],"auto_update":false,"unknown_future_key":{"a":1}}`
	seedProfile(t, "src", payload)

	exported := do(t, "GET", "/api/profiles/src/export", "")
	require.Equal(t, http.StatusOK, exported.Code)

	body, err := json.Marshal(map[string]any{
		"name":   "dst",
		"config": json.RawMessage(exported.Body.Bytes()),
	})
	require.NoError(t, err)

	rec := do(t, "POST", "/api/import", string(body))
	require.Equal(t, http.StatusOK, rec.Code, "re-importing an export failed: %s", rec.Body.String())

	var want, got map[string]any
	require.NoError(t, json.Unmarshal([]byte(payload), &want))
	require.NoError(t, json.Unmarshal(exported.Body.Bytes(), &got))
	require.Equal(t, want, got, "export is not lossless")

	var resp struct {
		Name string `json:"name"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, want, readProfileOpenCode(t, resp.Name), "import is not lossless")
}
