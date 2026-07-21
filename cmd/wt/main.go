// Package main is the entry point for the wt Git worktree manager.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/servusdei2018/wt/internal/config"
	"github.com/servusdei2018/wt/internal/exclude"
	"github.com/servusdei2018/wt/internal/git"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

// app holds shared state derived during PersistentPreRunE and referenced by
// all subcommand closures.
type app struct {
	repoRoot string
	cfg      *config.Config
}

// baseBranch returns the effective base branch: config override or auto-detected
// from the remote HEAD.
func (a *app) baseBranch() string {
	if a.cfg.Sync.BaseBranch != "" {
		return a.cfg.Sync.BaseBranch
	}
	return git.DetectBase(a.repoRoot)
}

// worktreesDir returns the absolute path of the .worktrees/ directory.
func (a *app) worktreesDir() string {
	return filepath.Join(a.repoRoot, ".worktrees")
}

// worktreePath returns the canonical path for a named worktree and prevents path traversal.
func (a *app) worktreePath(branch string) (string, error) {
	wtDir := a.worktreesDir()
	path := filepath.Clean(filepath.Join(wtDir, branch))
	rel, err := filepath.Rel(wtDir, path)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		return "", fmt.Errorf("invalid branch path outside worktrees directory: %q", branch)
	}
	return path, nil
}

func isNoRepoCommand(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		switch c.Name() {
		case "help", "completion", "man":
			return true
		}
	}
	return false
}

var version = "dev"

func main() {
	a := &app{}

	root := &cobra.Command{
		Use:          "wt",
		Short:        "Git worktree manager",
		Long:         `wt manages Git worktrees.`,
		Version:      version,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip repo checks for commands that don't need a repo context.
			if isNoRepoCommand(cmd) {
				return nil
			}

			root, err := git.RepoRoot()
			if err != nil {
				return fmt.Errorf("not inside a git repository")
			}
			a.repoRoot = root

			if err := os.MkdirAll(filepath.Join(root, ".worktrees"), 0o755); err != nil {
				return fmt.Errorf("could not create .worktrees/: %w", err)
			}
			if err := exclude.Ensure(root, ".worktrees/*"); err != nil {
				return fmt.Errorf("could not update .git/info/exclude: %w", err)
			}

			cfg, err := config.Load(root)
			if err != nil {
				return fmt.Errorf("config: %w", err)
			}
			a.cfg = cfg
			return nil
		},
	}

	root.AddCommand(
		a.newCmd(),
		a.doneCmd(),
		a.syncCmd(),
		a.refreshCmd(),
		a.listCmd(),
		a.sizeCmd(),
	)

	if err := fang.Execute(context.Background(), root); err != nil {
		os.Exit(1)
	}
}
