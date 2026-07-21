package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	cfg, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() unexpected error = %v", err)
	}
	if cfg.Sync.Remote != "origin" {
		t.Errorf("cfg.Sync.Remote = %q, want %q", cfg.Sync.Remote, "origin")
	}
}

func TestLoadCustomConfig(t *testing.T) {
	tmpDir := t.TempDir()
	content := `
[hooks]
post_create = "./setup.sh"

[editor]
command = "code"

[sync]
base_branch = "develop"
remote = "upstream"
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".wt.toml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() unexpected error = %v", err)
	}

	if cfg.Hooks.PostCreate != "./setup.sh" {
		t.Errorf("cfg.Hooks.PostCreate = %q, want %q", cfg.Hooks.PostCreate, "./setup.sh")
	}
	if cfg.Editor.Command != "code" {
		t.Errorf("cfg.Editor.Command = %q, want %q", cfg.Editor.Command, "code")
	}
	if cfg.Sync.BaseBranch != "develop" {
		t.Errorf("cfg.Sync.BaseBranch = %q, want %q", cfg.Sync.BaseBranch, "develop")
	}
	if cfg.Sync.Remote != "upstream" {
		t.Errorf("cfg.Sync.Remote = %q, want %q", cfg.Sync.Remote, "upstream")
	}
}

func TestLoadInvalidToml(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, ".wt.toml"), []byte("invalid toml [["), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(tmpDir)
	if err == nil {
		t.Error("Load() expected error for invalid TOML, got nil")
	}
}
