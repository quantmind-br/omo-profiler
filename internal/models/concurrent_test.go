package models

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

// The registry is one file rewritten in full, so a load-edit-save cycle that is
// not serialized loses every concurrent change but the last. The web server
// handles model requests on separate goroutines.
func TestConcurrentAddsDoNotLoseModels(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	const n = 24
	var wg sync.WaitGroup
	errs := make(chan error, n)

	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m := RegisteredModel{ModelID: fmt.Sprintf("m%02d", i), Provider: "p"}
			if err := Add(m); err != nil {
				errs <- fmt.Errorf("%s: %w", m.ModelID, err)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent add failed: %v", err)
	}

	reg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(reg.Models) != n {
		t.Fatalf("got %d models, want %d — concurrent adds lost updates", len(reg.Models), n)
	}
}

// The duplicate check and the write share the transaction, so exactly one of N
// racing adds of the same identity may win.
func TestConcurrentAddSameModelYieldsOneWinner(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	const n = 16
	var wg sync.WaitGroup
	var mu sync.Mutex
	var wins, conflicts int

	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := Add(RegisteredModel{ModelID: "contested", Provider: "p"})
			mu.Lock()
			defer mu.Unlock()
			var exists *ModelExistsError
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

	if wins != 1 || conflicts != n-1 {
		t.Fatalf("wins=%d conflicts=%d, want 1 and %d", wins, conflicts, n-1)
	}

	reg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(reg.Models) != 1 {
		t.Fatalf("got %d models, want 1", len(reg.Models))
	}
}

// Deletes race the same way as adds.
func TestConcurrentDeletesKeepTheRest(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	const n = 16
	for i := range n {
		if err := Add(RegisteredModel{ModelID: fmt.Sprintf("m%02d", i), Provider: "p"}); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	if err := Add(RegisteredModel{ModelID: "survivor", Provider: "p"}); err != nil {
		t.Fatalf("seed survivor: %v", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := fmt.Sprintf("m%02d", i)
			if err := Delete("p", id); err != nil {
				errs <- fmt.Errorf("%s: %w", id, err)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent delete failed: %v", err)
	}

	reg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(reg.Models) != 1 || reg.Models[0].ModelID != "survivor" {
		t.Fatalf("concurrent deletes corrupted the registry: %+v", reg.Models)
	}
}

// A bulk import is one transaction, so it cannot interleave with another
// writer and it reports duplicates rather than failing.
func TestAddManyIsOneTransaction(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	if err := Add(RegisteredModel{ModelID: "dup", Provider: "p"}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	added, skipped, err := AddMany([]RegisteredModel{
		{ModelID: "a", Provider: "p"},
		{ModelID: "dup", Provider: "p"},
		{ModelID: "b", Provider: "p"},
	})
	if err != nil {
		t.Fatalf("AddMany: %v", err)
	}
	if added != 2 || skipped != 1 {
		t.Fatalf("added=%d skipped=%d, want 2 and 1", added, skipped)
	}

	reg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(reg.Models) != 3 {
		t.Fatalf("got %d models, want 3", len(reg.Models))
	}
}

// A failing transaction must leave the file untouched.
func TestMutateAbortsWithoutWriting(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	if err := Add(RegisteredModel{ModelID: "keep", Provider: "p"}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	sentinel := errors.New("abort")
	err := Mutate(func(r *ModelsRegistry) error {
		if err := r.add(RegisteredModel{ModelID: "ghost", Provider: "p"}); err != nil {
			return err
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("Mutate error = %v, want the callback's error", err)
	}

	if Exists("p", "ghost") {
		t.Fatal("aborted transaction was persisted")
	}
	if !Exists("p", "keep") {
		t.Fatal("aborted transaction damaged the registry")
	}
}
