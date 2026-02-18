package schema

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/diogenes/omo-profiler/internal/diff"
)

var UpstreamSchemaURL = "https://raw.githubusercontent.com/code-yeongyu/oh-my-opencode/dev/assets/oh-my-opencode.schema.json"

type CompareResult struct {
	Identical bool
	Diff      string
}

func FetchUpstreamSchema(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, UpstreamSchemaURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch upstream schema: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

func CompareSchemas() (*CompareResult, error) {
	ctx := context.Background()
	upstream, err := FetchUpstreamSchema(ctx)
	if err != nil {
		return nil, err
	}

	embedded := GetEmbeddedSchema()

	if bytes.Equal(embedded, upstream) {
		return &CompareResult{Identical: true, Diff: ""}, nil
	}

	diffOutput := diff.ComputeUnifiedDiff("embedded", "upstream", embedded, upstream)

	return &CompareResult{Identical: false, Diff: diffOutput}, nil
}

func SaveDiff(dir, diffContent string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("schema-diff-%s.diff", timestamp)
	filePath := filepath.Join(dir, filename)

	if err := os.WriteFile(filePath, []byte(diffContent), 0644); err != nil {
		return "", fmt.Errorf("failed to save diff: %w", err)
	}

	return filePath, nil
}
