// Package fs provides disk-usage scanning for worktree directories.
package fs

import (
	"os"
	"path/filepath"
)

// heavyNames is the set of directory names considered dependency caches.
var heavyNames = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	".venv":        true,
	"venv":         true,
	"__pycache__":  true,
	".gradle":      true,
	"target":       true, // Rust / Java / Maven
	".build":       true, // Swift
	".tox":         true, // Python tox
}

// HeavyDir is a notable dependency-cache subdirectory within a worktree.
type HeavyDir struct {
	RelPath string
	Size    int64
}

// DirReport summarises the disk usage of a single worktree directory.
type DirReport struct {
	Path      string
	TotalSize int64
	HeavyDirs []HeavyDir
}

// Scan walks worktreesRoot and returns one DirReport per immediate
// subdirectory (i.e. per worktree).
func Scan(worktreesRoot string) ([]DirReport, error) {
	entries, err := os.ReadDir(worktreesRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var reports []DirReport
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(worktreesRoot, entry.Name())
		report, err := scanDir(dir)
		if err != nil {
			continue // best-effort; skip unreadable dirs
		}
		reports = append(reports, report)
	}
	return reports, nil
}

func scanDir(root string) (DirReport, error) {
	report := DirReport{Path: root}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable paths
		}
		if d.IsDir() && path != root && heavyNames[d.Name()] {
			size, _ := dirSize(path)
			rel, _ := filepath.Rel(root, path)
			report.HeavyDirs = append(report.HeavyDirs, HeavyDir{RelPath: rel, Size: size})
			report.TotalSize += size
			return filepath.SkipDir
		}
		if !d.IsDir() {
			if info, err := d.Info(); err == nil {
				report.TotalSize += info.Size()
			}
		}
		return nil
	})
	return report, err
}

// dirSize returns the total byte size of all files under dir.
func dirSize(dir string) (int64, error) {
	var total int64
	err := filepath.WalkDir(dir, func(_ string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if info, err := d.Info(); err == nil {
			total += info.Size()
		}
		return nil
	})
	return total, err
}
