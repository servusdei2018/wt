// Package config reads the optional .wt.toml project configuration file.
package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const filename = ".wt.toml"

// HooksConfig holds post-create hook settings.
type HooksConfig struct {
	PostCreate string `toml:"post_create"`
}

// EditorConfig holds editor override settings.
type EditorConfig struct {
	Command string `toml:"command"`
}

// SyncConfig holds remote sync settings.
type SyncConfig struct {
	// BaseBranch, when empty, triggers auto-detection from origin HEAD.
	BaseBranch string `toml:"base_branch"`
	Remote     string `toml:"remote"`
}

// StorageConfig holds shared dependency store settings.
type StorageConfig struct {
	RequireReflink bool     `toml:"require_reflink"`
	DedupeOnCreate bool     `toml:"dedupe_on_create"`
	HeavyDirs      []string `toml:"heavy_dirs"`
}

// SeedConfig holds seed file copy settings.
type SeedConfig struct {
	Include []string `toml:"include"`
}

// Config is the top-level representation of .wt.toml.
type Config struct {
	Hooks   HooksConfig   `toml:"hooks"`
	Editor  EditorConfig  `toml:"editor"`
	Sync    SyncConfig    `toml:"sync"`
	Storage StorageConfig `toml:"storage"`
	Seed    SeedConfig    `toml:"seed"`
}

func defaults() *Config {
	return &Config{
		Sync: SyncConfig{Remote: "origin"},
		Storage: StorageConfig{
			DedupeOnCreate: true,
			RequireReflink: false,
			HeavyDirs: []string{
				"node_modules", "vendor", ".venv", "venv",
				"__pycache__", ".gradle", "target", ".build", ".tox",
			},
		},
		Seed: SeedConfig{
			Include: []string{".env", ".env.local"},
		},
	}
}

// Load reads .wt.toml from repoRoot. Missing file returns defaults.
// Returns an error only on parse failure.
func Load(repoRoot string) (*Config, error) {
	cfg := defaults()
	data, err := os.ReadFile(filepath.Join(repoRoot, filename))
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}
	if _, err := toml.Decode(string(data), cfg); err != nil {
		return nil, err
	}
	if cfg.Sync.Remote == "" {
		cfg.Sync.Remote = "origin"
	}
	return cfg, nil
}
