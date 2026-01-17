package backup

import (
	"fmt"
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

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("2006-01-02-150405")
	backupName := fmt.Sprintf("oh-my-opencode.json.bak.%s", timestamp)
	backupPath := filepath.Join(config.ConfigDir(), backupName)

	// Write backup
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup: %w", err)
	}

	return backupPath, nil
}

// List returns all backups sorted by timestamp (most recent first)
func List() ([]BackupInfo, error) {
	dir := config.ConfigDir()
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
		if !strings.HasPrefix(name, "oh-my-opencode.json.bak.") {
			continue
		}

		// Parse timestamp from filename
		parts := strings.Split(name, ".bak.")
		if len(parts) != 2 {
			continue
		}
		ts, err := time.Parse("2006-01-02-150405", parts[1])
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
	configPath := config.ConfigFile()
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to restore config: %w", err)
	}

	return nil
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
