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
	if !cfg.Storage.DedupeOnCreate {
		t.Errorf("cfg.Storage.DedupeOnCreate = false, want true")
	}
	if len(cfg.Storage.HeavyDirs) == 0 {
		t.Errorf("cfg.Storage.HeavyDirs is empty, expected defaults")
	}
	if len(cfg.Seed.Include) != 2 || cfg.Seed.Include[0] != ".env" || cfg.Seed.Include[1] != ".env.local" {
		t.Errorf("cfg.Seed.Include = %v, want [.env .env.local]", cfg.Seed.Include)
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

[storage]
dedupe_on_create = false
require_reflink = true
heavy_dirs = ["node_modules", "vendor"]

[seed]
include = [".env.local", "config/local.json"]
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
	if cfg.Storage.DedupeOnCreate {
		t.Errorf("cfg.Storage.DedupeOnCreate = true, want false")
	}
	if !cfg.Storage.RequireReflink {
		t.Errorf("cfg.Storage.RequireReflink = false, want true")
	}
	if len(cfg.Storage.HeavyDirs) != 2 || cfg.Storage.HeavyDirs[0] != "node_modules" {
		t.Errorf("cfg.Storage.HeavyDirs = %v, want [node_modules vendor]", cfg.Storage.HeavyDirs)
	}
	if len(cfg.Seed.Include) != 2 || cfg.Seed.Include[0] != ".env.local" || cfg.Seed.Include[1] != "config/local.json" {
		t.Errorf("cfg.Seed.Include = %v, want [.env.local config/local.json]", cfg.Seed.Include)
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
