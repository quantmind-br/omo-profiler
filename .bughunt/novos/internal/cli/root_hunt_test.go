package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/diogenes/omo-profiler/internal/config"
)

// A runtime failure must report the failure, not the command's flag reference.
func TestHuntRuntimeErrorDoesNotPrintUsage(t *testing.T) {
	home := t.TempDir()
	config.SetBaseDir(home)
	t.Cleanup(config.ResetBaseDir)

	if err := os.MkdirAll(filepath.Join(home, config.OmoDirname), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(config.OmoFile(), []byte("not json"), 0o600); err != nil {
		t.Fatalf("seed corrupt document: %v", err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"list"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})

	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected list to fail on a corrupt document")
	}

	got := out.String()
	if !strings.Contains(got, "failed to list profiles") {
		t.Errorf("real error missing from output:\n%s", got)
	}
	if strings.Contains(got, "Usage:") {
		t.Errorf("usage block printed for a runtime error:\n%s", got)
	}
}
