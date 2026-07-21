package storage

import (
	"os"
	"path/filepath"

	"github.com/servusdei2018/wt/pkg/fs"
)

// SeedWorktree pre-populates heavy dependency directories in targetWorktree by cloning
// matching heavy directories found in repoRoot or existing worktrees using CoW reflinks.
func SeedWorktree(repoRoot, targetWorktree string, heavyDirs []string, requireReflink bool) (*fs.CloneStats, error) {
	totalStats := &fs.CloneStats{}

	worktreesDir := filepath.Join(repoRoot, ".worktrees")

	// Collect candidate search roots: repoRoot first, then other worktrees
	candidates := []string{repoRoot}
	entries, err := os.ReadDir(worktreesDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				wPath := filepath.Join(worktreesDir, e.Name())
				if wPath != targetWorktree {
					candidates = append(candidates, wPath)
				}
			}
		}
	}

	for _, heavyName := range heavyDirs {
		targetHeavyPath := filepath.Join(targetWorktree, heavyName)
		if _, err := os.Stat(targetHeavyPath); err == nil {
			// Already exists in target worktree
			continue
		}

		// Find first candidate directory containing heavyName
		var srcHeavyPath string
		for _, cand := range candidates {
			candidatePath := filepath.Join(cand, heavyName)
			if info, err := os.Stat(candidatePath); err == nil && info.IsDir() {
				srcHeavyPath = candidatePath
				break
			}
		}

		if srcHeavyPath == "" {
			continue
		}

		opts := fs.CloneOptions{
			RequireReflink: requireReflink,
			PreserveAttrs:  true,
		}

		stats, err := fs.CloneDir(srcHeavyPath, targetHeavyPath, opts)
		if err != nil {
			// Best-effort seeding per directory
			continue
		}

		totalStats.FilesTotal += stats.FilesTotal
		totalStats.FilesReflinked += stats.FilesReflinked
		totalStats.FilesCopied += stats.FilesCopied
		totalStats.BytesTotal += stats.BytesTotal
		totalStats.BytesReflinked += stats.BytesReflinked
		totalStats.BytesCopied += stats.BytesCopied
	}

	return totalStats, nil
}
