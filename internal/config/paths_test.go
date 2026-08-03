package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOmoDir(t *testing.T) {
	defer ResetBaseDir()

	tmpDir := t.TempDir()
	SetBaseDir(tmpDir)

	got := OmoDir()
	want := filepath.Join(tmpDir, ".omo")
	if got != want {
		t.Errorf("OmoDir() = %s, want %s", got, want)
	}
}

func TestModelsFile(t *testing.T) {
	defer ResetBaseDir()

	tmpDir := t.TempDir()
	SetBaseDir(tmpDir)

	got := ModelsFile()
	want := filepath.Join(tmpDir, ".omo", "models.json")
	if got != want {
		t.Errorf("ModelsFile() = %s, want %s", got, want)
	}
}

func TestOmoFile(t *testing.T) {
	tests := []struct {
		name  string
		setup func(dir string)
		want  string // basename expected under .omo/
	}{
		{
			name:  "fresh install defaults to omo.json",
			setup: func(dir string) {},
			want:  OmoBasename,
		},
		{
			name: "only omo.json returns omo.json",
			setup: func(dir string) {
				if err := os.WriteFile(filepath.Join(dir, OmoBasename), []byte("{}"), 0644); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
			},
			want: OmoBasename,
		},
		{
			name: "only omo.jsonc returns omo.jsonc",
			setup: func(dir string) {
				if err := os.WriteFile(filepath.Join(dir, OmoBasenameJSONC), []byte("{}"), 0644); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
			},
			want: OmoBasenameJSONC,
		},
		{
			name: "both present prefers omo.jsonc",
			setup: func(dir string) {
				if err := os.WriteFile(filepath.Join(dir, OmoBasename), []byte("{}"), 0644); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
				if err := os.WriteFile(filepath.Join(dir, OmoBasenameJSONC), []byte("{}"), 0644); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
			},
			want: OmoBasenameJSONC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer ResetBaseDir()
			tmpDir := t.TempDir()
			SetBaseDir(tmpDir)

			omoDir := filepath.Join(tmpDir, OmoDirname)
			if err := os.MkdirAll(omoDir, 0755); err != nil {
				t.Fatalf("MkdirAll: %v", err)
			}
			tt.setup(omoDir)

			got := OmoFile()
			want := filepath.Join(omoDir, tt.want)
			if got != want {
				t.Errorf("OmoFile() = %s, want %s", got, want)
			}
		})
	}
}

func TestLegacyConfigDir(t *testing.T) {
	defer ResetBaseDir()

	tmpDir := t.TempDir()
	SetBaseDir(tmpDir)

	got := LegacyConfigDir()
	want := filepath.Join(tmpDir, ".config", "opencode")
	if got != want {
		t.Errorf("LegacyConfigDir() = %s, want %s", got, want)
	}
}

func TestLegacyProfilesDir(t *testing.T) {
	defer ResetBaseDir()

	tmpDir := t.TempDir()
	SetBaseDir(tmpDir)

	got := LegacyProfilesDir()
	want := filepath.Join(tmpDir, ".config", "opencode", "profiles")
	if got != want {
		t.Errorf("LegacyProfilesDir() = %s, want %s", got, want)
	}
}

func TestLegacyConfigFile(t *testing.T) {
	tests := []struct {
		name  string
		setup func(dir string)
		want  string // empty or basename under legacy dir
	}{
		{
			name:  "absent returns empty string",
			setup: func(dir string) {},
			want:  "",
		},
		{
			name: "openagent file present returns its path",
			setup: func(dir string) {
				if err := os.WriteFile(filepath.Join(dir, LegacyOpenagentBasename), []byte("{}"), 0644); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
			},
			want: LegacyOpenagentBasename,
		},
		{
			name: "only opencode legacy present returns its path",
			setup: func(dir string) {
				if err := os.WriteFile(filepath.Join(dir, LegacyOpencodeBasename), []byte("{}"), 0644); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
			},
			want: LegacyOpencodeBasename,
		},
		{
			name: "both present prefers openagent basename",
			setup: func(dir string) {
				if err := os.WriteFile(filepath.Join(dir, LegacyOpenagentBasename), []byte("{}"), 0644); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
				if err := os.WriteFile(filepath.Join(dir, LegacyOpencodeBasename), []byte("{}"), 0644); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
			},
			want: LegacyOpenagentBasename,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer ResetBaseDir()
			tmpDir := t.TempDir()
			SetBaseDir(tmpDir)

			legacyDir := filepath.Join(tmpDir, ".config", "opencode")
			if err := os.MkdirAll(legacyDir, 0755); err != nil {
				t.Fatalf("MkdirAll: %v", err)
			}
			tt.setup(legacyDir)

			got := LegacyConfigFile()
			var want string
			if tt.want != "" {
				want = filepath.Join(legacyDir, tt.want)
			}
			if got != want {
				t.Errorf("LegacyConfigFile() = %q, want %q", got, want)
			}
		})
	}
}

func TestEnsureDirs(t *testing.T) {
	defer ResetBaseDir()

	tmpDir := t.TempDir()
	SetBaseDir(tmpDir)

	omoDir := OmoDir()
	if _, err := os.Stat(omoDir); !os.IsNotExist(err) {
		t.Fatalf("OmoDir should not exist before EnsureDirs")
	}

	if err := EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs() failed: %v", err)
	}

	info, err := os.Stat(omoDir)
	if err != nil {
		t.Fatalf("OmoDir should exist after EnsureDirs: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("OmoDir should be a directory")
	}

	if err := EnsureDirs(); err != nil {
		t.Errorf("EnsureDirs() should be idempotent: %v", err)
	}
}

func TestHomeDir_UsesBaseDir(t *testing.T) {
	defer ResetBaseDir()

	tmpDir := t.TempDir()
	SetBaseDir(tmpDir)
	if got := HomeDir(); got != tmpDir {
		t.Errorf("HomeDir() = %s, want %s", got, tmpDir)
	}
}

func TestResetBaseDir(t *testing.T) {
	tmpDir := t.TempDir()
	SetBaseDir(tmpDir)
	ResetBaseDir()

	home, _ := os.UserHomeDir()
	want := filepath.Join(home, OmoDirname)
	if got := OmoDir(); got != want {
		t.Errorf("OmoDir() after ResetBaseDir = %s, want %s", got, want)
	}
}

func TestDefaultSchema(t *testing.T) {
	want := "https://raw.githubusercontent.com/code-yeongyu/oh-my-openagent/dev/assets/omo.schema.json"
	if DefaultSchema != want {
		t.Errorf("DefaultSchema = %s, want %s", DefaultSchema, want)
	}
}
