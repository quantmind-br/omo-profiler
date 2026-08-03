package backup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/diogenes/omo-profiler/internal/config"
	"github.com/stretchr/testify/require"
)

func huntEnv(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	config.SetBaseDir(dir)
	t.Cleanup(config.ResetBaseDir)
	require.NoError(t, config.EnsureDirs())
	return dir
}

// I2: the omo document can hold API tokens, so no write path may widen a
// tightened mode. Create() and WriteFileAtomic() both honour this; Restore()
// is the third write path to the same file.
func TestHuntRestorePreservesMode(t *testing.T) {
	huntEnv(t)
	cfgPath := config.OmoFile()
	require.NoError(t, os.WriteFile(cfgPath, []byte(`{"profiles":{"a":{}}}`), 0o600))

	bak, err := Create(cfgPath)
	require.NoError(t, err)

	info, err := os.Stat(bak)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "backup must inherit 0600")

	require.NoError(t, os.Remove(cfgPath))
	require.NoError(t, Restore(bak))

	restored, err := os.Stat(cfgPath)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), restored.Mode().Perm(),
		"Restore widened a secrets-bearing document to %v", restored.Mode().Perm())
}

// O6: every mutating write snapshots the document first, so without rotation
// the ~/.omo directory grows one copy of a secrets-bearing file per edit.
func TestHuntBackupsAreRotated(t *testing.T) {
	huntEnv(t)
	cfgPath := config.OmoFile()
	require.NoError(t, os.WriteFile(cfgPath, []byte(`{}`), 0o600))

	const writes = 40
	for range writes {
		require.NoError(t, CreateOmoIfPresent())
	}

	list, err := List()
	require.NoError(t, err)
	require.Less(t, len(list), writes,
		"%d writes left %d backups — rotation never runs", writes, len(list))
}

// O1: Clean is the rotation entry point; a non-positive keep count must not
// index out of range.
func TestHuntCleanNonPositiveKeep(t *testing.T) {
	huntEnv(t)
	cfgPath := config.OmoFile()
	require.NoError(t, os.WriteFile(cfgPath, []byte(`{}`), 0o600))
	for range 3 {
		_, err := Create(cfgPath)
		require.NoError(t, err)
	}

	require.NoError(t, Clean(0), "Clean(0) should drop every backup")
	list, err := List()
	require.NoError(t, err)
	require.Empty(t, list)

	for range 3 {
		_, err := Create(cfgPath)
		require.NoError(t, err)
	}
	require.NotPanics(t, func() { _ = Clean(-1) }, "Clean(-1) panicked")
}

// M2: a backup written by Create must be findable by List — the two agree on
// the name format or rotation silently stops working.
func TestHuntCreateListAgree(t *testing.T) {
	huntEnv(t)
	cfgPath := config.OmoFile()
	require.NoError(t, os.WriteFile(cfgPath, []byte(`{}`), 0o600))

	made := map[string]bool{}
	for range 5 {
		p, err := Create(cfgPath)
		require.NoError(t, err)
		made[filepath.Base(p)] = true
	}

	list, err := List()
	require.NoError(t, err)
	require.Len(t, list, len(made), "List lost backups that Create made")
	for _, b := range list {
		require.True(t, made[b.Name], "List invented %q", b.Name)
	}
	for i := 1; i < len(list); i++ {
		require.False(t, list[i].Timestamp.After(list[i-1].Timestamp), "List not sorted newest first")
	}
}
