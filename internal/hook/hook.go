// Package hook detects and runs post-creation environment setup hooks.
package hook

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/servusdei2018/wt/internal/config"
)

// Hook describes a post-create command to execute inside a new worktree.
type Hook struct {
	Name    string
	Command []string
}

// Detect returns the first applicable Hook for the given worktree, or nil if
// none apply. Detection priority:
//
//  1. Custom script from .wt.toml hooks.post_create
//  2. Specific lockfiles (pnpm, bun, yarn, uv)
//  3. General manifest files (deno, npm, go, pip)
func Detect(worktreePath string, cfg *config.Config) (*Hook, error) {
	if cfg != nil && cfg.Hooks.PostCreate != "" {
		return &Hook{Name: "custom", Command: []string{cfg.Hooks.PostCreate}}, nil
	}

	// Lockfile based detection
	lockfiles := []struct {
		file string
		hook Hook
	}{
		{"pnpm-lock.yaml", Hook{"pnpm install", []string{"pnpm", "install"}}},
		{"bun.lockb", Hook{"bun install", []string{"bun", "install"}}},
		{"bun.lock", Hook{"bun install", []string{"bun", "install"}}},
		{"yarn.lock", Hook{"yarn install", []string{"yarn", "install"}}},
		{"uv.lock", Hook{"uv sync", []string{"uv", "sync"}}},
	}

	for _, c := range lockfiles {
		if fileExists(filepath.Join(worktreePath, c.file)) {
			h := c.hook
			return &h, nil
		}
	}

	checks := []struct {
		file string
		hook Hook
	}{
		{"deno.json", Hook{"deno install", []string{"deno", "install"}}},
		{"deno.jsonc", Hook{"deno install", []string{"deno", "install"}}},
		{"package.json", Hook{"npm install", []string{"npm", "install"}}},
		{"go.mod", Hook{"go mod download", []string{"go", "mod", "download"}}},
		{"requirements.txt", Hook{"pip install", []string{"pip", "install", "-r", "requirements.txt"}}},
		{"pyproject.toml", Hook{"pip install", []string{"pip", "install", "-e", "."}}},
	}
	for _, c := range checks {
		if fileExists(filepath.Join(worktreePath, c.file)) {
			h := c.hook
			return &h, nil
		}
	}
	return nil, nil
}

// Run executes h inside worktreePath, streaming output directly to the terminal.
func Run(h *Hook, worktreePath string) error {
	if len(h.Command) == 0 {
		return nil
	}
	cmd := exec.Command(h.Command[0], h.Command[1:]...)
	cmd.Dir = worktreePath
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("hook %q: %w", h.Name, err)
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
