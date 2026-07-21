package storage

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/servusdei2018/wt/pkg/fs"
)

// DedupeOptions specifies parameters for deduplication.
type DedupeOptions struct {
	RequireReflink bool
	DryRun         bool
	HeavyDirs      []string
	Verbose        bool
}

// DedupeReport holds statistics on completed deduplication operations.
type DedupeReport struct {
	FilesDeduplicated int64            `json:"files_deduplicated"`
	BytesSaved        int64            `json:"bytes_saved"`
	ReflinkCount      int64            `json:"reflink_count"`
	CopiedCount       int64            `json:"copied_count"`
	Duration          time.Duration    `json:"duration"`
	DirSavings        map[string]int64 `json:"dir_savings"`
}

type fileMeta struct {
	Path     string
	Size     int64
	HeavyDir string
}

// DeduplicateWorktrees scans worktrees for identical files in heavy directories
// and deduplicates them using Copy-On-Write reflinks.
func DeduplicateWorktrees(repoRoot, worktreesDir string, opts DedupeOptions) (*DedupeReport, error) {
	start := time.Now()
	report := &DedupeReport{
		DirSavings: make(map[string]int64),
	}

	heavyMap := make(map[string]bool)
	for _, name := range opts.HeavyDirs {
		heavyMap[name] = true
	}

	// Gather root directories to scan (repo root + all worktree directories)
	scanRoots := []string{repoRoot}

	entries, err := os.ReadDir(worktreesDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				scanRoots = append(scanRoots, filepath.Join(worktreesDir, e.Name()))
			}
		}
	}

	// Group files by size
	sizeGroups := make(map[int64][]fileMeta)

	for _, root := range scanRoots {
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			// Don't traverse .git or .worktrees meta directories
			if d.IsDir() {
				name := d.Name()
				if name == ".git" || (path == filepath.Join(repoRoot, ".worktrees")) {
					return filepath.SkipDir
				}
				return nil
			}

			if !d.Type().IsRegular() {
				return nil
			}

			relToRoot, err := filepath.Rel(root, path)
			if err != nil {
				return nil
			}

			// Check if file is within a heavy directory
			parts := strings.Split(filepath.ToSlash(relToRoot), "/")
			var heavyName string
			for _, part := range parts {
				if heavyMap[part] {
					heavyName = part
					break
				}
			}

			if heavyName == "" {
				return nil
			}

			info, err := d.Info()
			if err != nil || info.Size() == 0 {
				return nil
			}

			sizeGroups[info.Size()] = append(sizeGroups[info.Size()], fileMeta{
				Path:     path,
				Size:     info.Size(),
				HeavyDir: heavyName,
			})
			return nil
		})
	}

	// Hash files in groups with >1 file of same size
	hashGroups := make(map[string][]fileMeta)

	for size, files := range sizeGroups {
		if len(files) < 2 {
			continue
		}
		for _, f := range files {
			h, err := hashFile(f.Path)
			if err != nil {
				continue
			}
			key := fmt.Sprintf("%d:%s", size, h)
			hashGroups[key] = append(hashGroups[key], f)
		}
	}

	// Perform deduplication for matching files
	for _, files := range hashGroups {
		if len(files) < 2 {
			continue
		}

		canonical := files[0]

		for _, dup := range files[1:] {
			if canonical.Path == dup.Path {
				continue
			}

			cStat, err1 := os.Stat(canonical.Path)
			dStat, err2 := os.Stat(dup.Path)
			if err1 == nil && err2 == nil && os.SameFile(cStat, dStat) {
				continue
			}

			if opts.DryRun {
				report.FilesDeduplicated++
				report.BytesSaved += dup.Size
				report.ReflinkCount++
				report.DirSavings[dup.HeavyDir] += dup.Size
				continue
			}

			// Perform atomic CoW reflink replacement
			tmpName := fmt.Sprintf("%s.wt_tmp_%s", dup.Path, randomSuffix())
			cloneOpts := fs.CloneOptions{
				RequireReflink: true, // Deduplication requires reflink CoW to save space
				PreserveAttrs:  true,
			}

			strat, err := fs.CloneFile(canonical.Path, tmpName, cloneOpts)
			if err != nil {
				_ = os.Remove(tmpName)
				if opts.RequireReflink {
					return nil, fmt.Errorf("reflink dedupe failed for %s: %w", dup.Path, err)
				}
				continue
			}

			if err := os.Rename(tmpName, dup.Path); err != nil {
				_ = os.Remove(tmpName)
				continue
			}

			if strat == fs.StrategyReflink {
				report.FilesDeduplicated++
				report.BytesSaved += dup.Size
				report.ReflinkCount++
				report.DirSavings[dup.HeavyDir] += dup.Size
			} else {
				report.CopiedCount++
			}
		}
	}

	report.Duration = time.Since(start)
	return report, nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func randomSuffix() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
