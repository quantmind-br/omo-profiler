package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/diogenes/omo-profiler/internal/backup"

	"github.com/diogenes/omo-profiler/internal/config"
)

// Every profile lives in one document, so concurrent mutations are whole-file
// overwrites: without serialization the last writer silently discards the
// others. The web server handles requests on separate goroutines, so this is a
// real lost-update path, not a theoretical one.
func TestConcurrentSavesDoNotLoseProfiles(t *testing.T) {
	setupTestEnv(t)

	const n = 24
	var wg sync.WaitGroup
	errs := make(chan error, n)

	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			name := fmt.Sprintf("p%02d", i)
			if err := SaveOpenCodeBlock(name, json.RawMessage(`{"telemetry":false}`)); err != nil {
				errs <- fmt.Errorf("%s: %w", name, err)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent save failed: %v", err)
	}

	names, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) != n {
		t.Fatalf("got %d profiles, want %d — concurrent saves lost updates: %v", len(names), n, names)
	}
}

// Deletes race the same way: each one rewrites the whole document.
func TestConcurrentDeletesRemoveExactlyTheirOwn(t *testing.T) {
	setupTestEnv(t)

	const n = 16
	for i := range n {
		seedProfile(t, fmt.Sprintf("p%02d", i), `{}`)
	}
	seedProfile(t, "survivor", `{}`)
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			name := fmt.Sprintf("p%02d", i)
			if err := Delete(name); err != nil {
				errs <- fmt.Errorf("%s: %w", name, err)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent delete failed: %v", err)
	}

	names, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) != 1 || names[0] != "survivor" {
		t.Fatalf("concurrent deletes corrupted the document: got %v, want [survivor]", names)
	}
}

// Create resolves the name collision inside the transaction, so exactly one of
// N racing creates may win. A precheck outside the lock would let several
// through and let them overwrite each other.
func TestConcurrentCreateSameNameYieldsExactlyOneWinner(t *testing.T) {
	setupTestEnv(t)

	const n = 16
	var wg sync.WaitGroup
	var mu sync.Mutex
	var wins, conflicts int

	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := Create("contested", config.Config{})
			mu.Lock()
			defer mu.Unlock()
			var exists *ExistsError
			switch {
			case err == nil:
				wins++
			case errors.As(err, &exists):
				conflicts++
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if wins != 1 {
		t.Fatalf("wins = %d, want exactly 1", wins)
	}
	if conflicts != n-1 {
		t.Fatalf("conflicts = %d, want %d", conflicts, n-1)
	}
}

// Imports of the same base name must each land on their own suffix. Choosing
// the suffix outside the write would let two imports pick the same one.
func TestConcurrentImportsGetDistinctNames(t *testing.T) {
	setupTestEnv(t)

	const n = 16
	var wg sync.WaitGroup
	var mu sync.Mutex
	seen := map[string]bool{}

	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			name, _, err := CreateAvailable("imported", json.RawMessage(`{}`))
			if err != nil {
				t.Errorf("CreateAvailable: %v", err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if seen[name] {
				t.Errorf("duplicate name handed out: %s", name)
			}
			seen[name] = true
		}()
	}
	wg.Wait()

	names, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) != n {
		t.Fatalf("got %d profiles, want %d — imports overwrote each other: %v", len(names), n, names)
	}
}

// A failing transaction must not write: the document stays exactly as it was.
func TestMutateAbortsWithoutWriting(t *testing.T) {
	setupTestEnv(t)
	seedProfile(t, "dev", `{"telemetry":false}`)

	doc, err := config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	before, err := doc.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	sentinel := errors.New("abort")
	err = config.Mutate(func(d *config.Document) error {
		if err := d.SetProfileBlock("ghost", json.RawMessage(`{"[opencode]":{}}`)); err != nil {
			return err
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("Mutate error = %v, want the callback's error", err)
	}

	if Exists("ghost") {
		t.Fatal("aborted transaction was persisted")
	}
	doc, err = config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	after, err := doc.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("aborted transaction modified the document:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// CreateFrom must reproduce the stored block byte-for-byte in meaning: a clone
// that round-trips through the typed Config drops explicitly present zero
// values, because every field is `omitempty`. TestCloneAs covers a different
// public path and would not catch a regression here.
func TestCreateFromCopiesBlockVerbatim(t *testing.T) {
	setupTestEnv(t)

	block := json.RawMessage(`{"[opencode]":{"default_run_agent":"","disabled_mcps":[],"telemetry":false},"[senpi]":{"keep":"me"}}`)
	seedProfileBlock(t, "src", block)

	if err := CreateFrom("dst", "src"); err != nil {
		t.Fatalf("CreateFrom: %v", err)
	}

	doc, err := config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	got, ok, err := doc.ProfileBlock("dst")
	if err != nil || !ok {
		t.Fatalf("ProfileBlock(dst): ok=%v err=%v", ok, err)
	}

	var want, have any
	if err := json.Unmarshal(block, &want); err != nil {
		t.Fatalf("unmarshal want: %v", err)
	}
	if err := json.Unmarshal(got, &have); err != nil {
		t.Fatalf("unmarshal got: %v", err)
	}
	if !reflect.DeepEqual(want, have) {
		t.Fatalf("clone is not verbatim:\n want: %s\n  got: %s", block, got)
	}
}

// The same guarantee for imports: the file the user hands over is what lands.
func TestCreateAvailableStoresBlockVerbatim(t *testing.T) {
	setupTestEnv(t)

	openCode := json.RawMessage(`{"default_run_agent":"","disabled_mcps":[],"telemetry":false}`)
	name, collided, err := CreateAvailable("imported", openCode)
	if err != nil {
		t.Fatalf("CreateAvailable: %v", err)
	}
	if collided {
		t.Fatal("unexpected collision on an empty document")
	}

	doc, err := config.LoadDocument()
	if err != nil {
		t.Fatalf("LoadDocument: %v", err)
	}
	got, ok, err := doc.ProfileBlock(name)
	if err != nil || !ok {
		t.Fatalf("ProfileBlock(%s): ok=%v err=%v", name, ok, err)
	}

	var block map[string]json.RawMessage
	if err := json.Unmarshal(got, &block); err != nil {
		t.Fatalf("unmarshal block: %v", err)
	}

	var want, have any
	if err := json.Unmarshal(openCode, &want); err != nil {
		t.Fatalf("unmarshal want: %v", err)
	}
	if err := json.Unmarshal(block[config.OpenCodeKey], &have); err != nil {
		t.Fatalf("unmarshal got: %v", err)
	}
	if !reflect.DeepEqual(want, have) {
		t.Fatalf("import dropped explicit zero values:\n want: %s\n  got: %s", openCode, block[config.OpenCodeKey])
	}
}

// Export reads the stored block, not a re-marshalled Config, so an
// export/import round-trip is lossless. Marshalling the typed Config drops
// explicit zero values and any key this build does not model.
func TestExportOpenCodeIsVerbatimAndRoundTrips(t *testing.T) {
	setupTestEnv(t)

	openCode := `{"default_run_agent":"","disabled_mcps":[],"future_key":123,"telemetry":false}`
	seedProfileBlock(t, "src", json.RawMessage(`{"[opencode]":`+openCode+`,"[senpi]":{"keep":"me"}}`))

	exported, err := ExportOpenCode("src")
	if err != nil {
		t.Fatalf("ExportOpenCode: %v", err)
	}

	var want, have any
	if err := json.Unmarshal([]byte(openCode), &want); err != nil {
		t.Fatalf("unmarshal want: %v", err)
	}
	if err := json.Unmarshal(exported, &have); err != nil {
		t.Fatalf("unmarshal exported: %v", err)
	}
	if !reflect.DeepEqual(want, have) {
		t.Fatalf("export is lossy:\n want: %s\n  got: %s", openCode, exported)
	}

	// Import what was exported: the payload must survive the full trip.
	name, _, err := CreateAvailable("dst", exported)
	if err != nil {
		t.Fatalf("CreateAvailable: %v", err)
	}
	reimported, err := ExportOpenCode(name)
	if err != nil {
		t.Fatalf("ExportOpenCode(%s): %v", name, err)
	}
	if err := json.Unmarshal(reimported, &have); err != nil {
		t.Fatalf("unmarshal reimported: %v", err)
	}
	if !reflect.DeepEqual(want, have) {
		t.Fatalf("export/import round-trip is lossy:\n want: %s\n  got: %s", openCode, reimported)
	}
}

// A missing profile is an error, not an empty export.
func TestExportOpenCodeMissingProfile(t *testing.T) {
	setupTestEnv(t)

	if _, err := ExportOpenCode("ghost"); err == nil {
		t.Fatal("expected an error for a missing profile")
	} else {
		var notFound *NotFoundError
		if !errors.As(err, &notFound) {
			t.Fatalf("expected *NotFoundError, got %v", err)
		}
	}
}

// Each write must leave behind a snapshot of exactly the state it replaced.
// Taking the snapshot before the lock breaks that: concurrent writers all
// capture the same state, so the intermediate versions are unrecoverable.
// Second-precision filenames break it too, by overwriting each other.
func TestEachWriteLeavesItsOwnPreImage(t *testing.T) {
	setupTestEnv(t)

	if err := Create("p0", config.Config{}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	const n = 8
	var wg sync.WaitGroup
	for i := 1; i <= n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := Create(fmt.Sprintf("p%d", i), config.Config{}); err != nil {
				t.Errorf("create p%d: %v", i, err)
			}
		}()
	}
	wg.Wait()

	backups, err := backup.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	// n writes after the seed, each snapshotting the document it replaced.
	if len(backups) != n {
		t.Fatalf("got %d backups, want %d — snapshots collided or were shared", len(backups), n)
	}

	// The pre-images must be the distinct states 1..n, not n copies of one.
	sizes := map[int]bool{}
	for _, b := range backups {
		data, err := os.ReadFile(b.Path)
		if err != nil {
			t.Fatalf("read %s: %v", b.Path, err)
		}
		var doc struct {
			Profiles map[string]json.RawMessage `json:"profiles"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatalf("parse %s: %v", b.Path, err)
		}
		sizes[len(doc.Profiles)] = true
	}
	if len(sizes) != n {
		t.Fatalf("backups captured %d distinct states, want %d — a write's pre-image was lost", len(sizes), n)
	}
}
