package schema

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFetchUpstreamSchema_Success(t *testing.T) {
	// Mock server returns valid JSON schema
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"type": "object", "properties": {}}`))
	}))
	defer server.Close()

	// Override URL for testing
	originalURL := UpstreamSchemaURL
	UpstreamSchemaURL = server.URL
	defer func() { UpstreamSchemaURL = originalURL }()

	ctx := context.Background()
	data, err := FetchUpstreamSchema(ctx)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty response")
	}

	expected := `{"type": "object", "properties": {}}`
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestFetchUpstreamSchema_Timeout(t *testing.T) {
	// Mock server that delays response beyond timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	// Override URL for testing
	originalURL := UpstreamSchemaURL
	UpstreamSchemaURL = server.URL
	defer func() { UpstreamSchemaURL = originalURL }()

	// Use a context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := FetchUpstreamSchema(ctx)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	// Should be context deadline exceeded or client timeout
	if !strings.Contains(err.Error(), "context deadline exceeded") &&
		!strings.Contains(err.Error(), "Client.Timeout") {
		t.Errorf("expected timeout-related error, got: %v", err)
	}
}

func TestFetchUpstreamSchema_NetworkError(t *testing.T) {
	// Override URL to invalid address
	originalURL := UpstreamSchemaURL
	UpstreamSchemaURL = "http://localhost:99999/nonexistent"
	defer func() { UpstreamSchemaURL = originalURL }()

	ctx := context.Background()
	_, err := FetchUpstreamSchema(ctx)
	if err == nil {
		t.Fatal("expected network error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to fetch") {
		t.Errorf("expected 'failed to fetch' in error, got: %v", err)
	}
}

func TestFetchUpstreamSchema_HTTPError(t *testing.T) {
	// Mock server returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	originalURL := UpstreamSchemaURL
	UpstreamSchemaURL = server.URL
	defer func() { UpstreamSchemaURL = originalURL }()

	ctx := context.Background()
	_, err := FetchUpstreamSchema(ctx)
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}

	if !strings.Contains(err.Error(), "status 500") {
		t.Errorf("expected 'status 500' in error, got: %v", err)
	}
}

func TestCompareSchemas_Identical(t *testing.T) {
	// Mock server returns the same schema as embedded
	embedded := GetEmbeddedSchema()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(embedded)
	}))
	defer server.Close()

	originalURL := UpstreamSchemaURL
	UpstreamSchemaURL = server.URL
	defer func() { UpstreamSchemaURL = originalURL }()

	result, err := CompareSchemas()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !result.Identical {
		t.Error("expected schemas to be identical")
	}

	if result.Diff != "" {
		t.Errorf("expected empty diff for identical schemas, got: %s", result.Diff)
	}
}

func TestCompareSchemas_Different(t *testing.T) {
	// Mock server returns a different schema
	differentSchema := `{"type": "object", "properties": {"newField": {"type": "string"}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(differentSchema))
	}))
	defer server.Close()

	originalURL := UpstreamSchemaURL
	UpstreamSchemaURL = server.URL
	defer func() { UpstreamSchemaURL = originalURL }()

	result, err := CompareSchemas()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result.Identical {
		t.Error("expected schemas to be different")
	}

	if result.Diff == "" {
		t.Error("expected non-empty diff for different schemas")
	}

	// Verify diff contains unified diff markers
	if !strings.Contains(result.Diff, "---") || !strings.Contains(result.Diff, "+++") {
		t.Errorf("expected unified diff format, got: %s", result.Diff)
	}
}

func TestSaveDiff_CreatesFile(t *testing.T) {
	tempDir := t.TempDir()
	diffContent := "--- embedded\n+++ upstream\n@@ -1,2 +1,2 @@\n-old\n+new"

	filePath, err := SaveDiff(tempDir, diffContent)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("expected file to exist at %s", filePath)
	}

	// Verify filename format: schema-diff-<timestamp>.diff
	filename := filepath.Base(filePath)
	if !strings.HasPrefix(filename, "schema-diff-") {
		t.Errorf("expected filename to start with 'schema-diff-', got: %s", filename)
	}
	if !strings.HasSuffix(filename, ".diff") {
		t.Errorf("expected filename to end with '.diff', got: %s", filename)
	}

	// Verify content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != diffContent {
		t.Errorf("expected content %q, got %q", diffContent, string(content))
	}
}

func TestSaveDiff_InvalidDirectory(t *testing.T) {
	nonexistentDir := "/nonexistent/path/that/does/not/exist"
	diffContent := "some diff content"

	_, err := SaveDiff(nonexistentDir, diffContent)
	if err == nil {
		t.Fatal("expected error for invalid directory, got nil")
	}

	if !strings.Contains(err.Error(), "failed to save diff") {
		t.Errorf("expected 'failed to save diff' in error, got: %v", err)
	}
}
