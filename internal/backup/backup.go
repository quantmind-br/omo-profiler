package backup

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/diogenes/omo-profiler/internal/config"
)

// BackupInfo contains information about a backup file
type BackupInfo struct {
	Path      string
	Timestamp time.Time
	Name      string // filename without path
}

// timestampLayout names backups with nanosecond precision so rapid successive
// writes get distinct names.
//
// Precision alone is not a uniqueness guarantee: two writers can read the same
// wall clock (coarse clock sources repeat readings, and two omo-profiler
// processes share no lock), and a truncating write would then destroy the
// pre-image the other writer depends on for recovery. Uniqueness comes from
// O_EXCL in createExclusive, not from the clock.
//
// parseLayout deliberately omits the fraction: time.Parse accepts an optional
// fractional second after the seconds field regardless, so this one layout
// reads both the current names and the second-precision names written before
// the change — and keeps the nanoseconds, so ordering stays correct.
const (
	timestampLayout = "2006-01-02-150405.000000000"
	parseLayout     = "2006-01-02-150405"
)

// maxNameAttempts bounds collision retries. Each attempt claims a distinct
// name, so this only caps pathological bursts rather than ordinary contention.
const maxNameAttempts = 1000

// now is a variable so tests can freeze the clock and force a collision.
var now = time.Now

// createExclusive claims an unused backup name at or after ts and returns the
// open file. On collision it advances a nanosecond and retries, so the name
// stays a plain timestamp — List, Restore and sort order need no special case.
func createExclusive(dir, basename string, ts time.Time, perm os.FileMode) (*os.File, string, error) {
	for i := 0; i < maxNameAttempts; i++ {
		path := filepath.Join(dir, fmt.Sprintf("%s.bak.%s", basename, ts.Format(timestampLayout)))

		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
		if err == nil {
			// O_CREATE applies the umask; set the mode explicitly so a 0600
			// source stays 0600 and a 0644 one is not narrowed.
			if err := f.Chmod(perm); err != nil {
				// Cleanup only; the error that matters is already returned.
				_ = f.Close()
				_ = os.Remove(path)
				return nil, "", err
			}
			return f, path, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return nil, "", err
		}
		ts = ts.Add(time.Nanosecond)
	}
	return nil, "", fmt.Errorf("no free backup name for %s after %d attempts", basename, maxNameAttempts)
}

// Create creates a timestamped backup of the config file
// Returns the backup path or error
func Create(configPath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("config file does not exist: %s", configPath)
	}

	// Read original
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config: %w", err)
	}

	// Inherit the source's mode: the document can hold API tokens, so a
	// hard-coded 0644 would widen 0600 every time we back it up.
	perm := os.FileMode(0o600)
	if info, err := os.Stat(configPath); err == nil {
		perm = info.Mode().Perm()
	}

	f, backupPath, err := createExclusive(config.OmoDir(), filepath.Base(configPath), now(), perm)
	if err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		// A partial backup is worse than none: it would look like a valid
		// pre-image. Discard it; the write error is what gets reported.
		_ = f.Close()
		_ = os.Remove(backupPath)
		return "", fmt.Errorf("failed to write backup: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(backupPath)
		return "", fmt.Errorf("failed to write backup: %w", err)
	}

	return backupPath, nil
}

// List returns all backups sorted by timestamp (most recent first)
func List() ([]BackupInfo, error) {
	dir := config.OmoDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupInfo{}, nil
		}
		return nil, err
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !isBackupFile(name) {
			continue
		}

		parts := strings.Split(name, ".bak.")
		if len(parts) != 2 {
			continue
		}
		ts, err := time.Parse(parseLayout, parts[1])
		if err != nil {
			continue
		}

		backups = append(backups, BackupInfo{
			Path:      filepath.Join(dir, name),
			Timestamp: ts,
			Name:      name,
		})
	}

	// Sort by timestamp descending (most recent first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

// Restore restores a backup to the config file
func Restore(backupPath string) error {
	// Read backup
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	// Write to config
	configPath := config.OmoFile()
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to restore config: %w", err)
	}

	return nil
}

func isBackupFile(name string) bool {
	return strings.HasPrefix(name, config.OmoBasename+".bak.") ||
		strings.HasPrefix(name, config.OmoBasenameJSONC+".bak.") ||
		strings.HasPrefix(name, config.LegacyOpenagentBasename+".bak.") ||
		strings.HasPrefix(name, config.LegacyOpencodeBasename+".bak.")
}

// Clean removes old backups, keeping only the N most recent
func Clean(keepLast int) error {
	backups, err := List()
	if err != nil {
		return err
	}

	if len(backups) <= keepLast {
		return nil
	}

	// Remove backups beyond keepLast
	for i := keepLast; i < len(backups); i++ {
		if err := os.Remove(backups[i].Path); err != nil {
			return fmt.Errorf("failed to remove backup %s: %w", backups[i].Name, err)
		}
	}

	return nil
}

// CreateOmoIfPresent snapshots ~/.omo/omo.json(c) before a mutating write.
// A missing document is fine — there is nothing to back up yet.
//
// Every mutating entry point (CLI, TUI, web) goes through this, so "back up
// before you write" is one rule with one implementation.
func CreateOmoIfPresent() error {
	path := config.OmoFile()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	_, err := Create(path)
	return err
}
