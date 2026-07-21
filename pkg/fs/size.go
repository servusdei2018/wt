// Package fs provides disk-usage scanning for worktree directories.
package fs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
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
	RelPath string `json:"rel_path"`
	Size    int64  `json:"size"`
}

// DirReport summarises the disk usage of a single worktree directory.
type DirReport struct {
	Path      string     `json:"path"`
	TotalSize int64      `json:"total_size"`
	HeavyDirs []HeavyDir `json:"heavy_dirs"`
}

// CacheEntry holds the cached disk usage info for a worktree directory.
type CacheEntry struct {
	Inode     uint64     `json:"inode"`
	MtimeNs   int64      `json:"mtime_ns"`
	TotalSize int64      `json:"total_size"`
	HeavyDirs []HeavyDir `json:"heavy_dirs"`
}

// CacheStore is a map of directory path -> CacheEntry.
type CacheStore map[string]CacheEntry

// Scan walks worktreesRoot in parallel and returns one DirReport per immediate
// subdirectory (i.e. per worktree). It uses default cache location under repo git dir if accessible.
func Scan(worktreesRoot string, heavyDirs []string) ([]DirReport, error) {
	cacheDir := filepath.Join(worktreesRoot, "..", ".git", "wt", "cache")
	return ScanWithCache(worktreesRoot, cacheDir, heavyDirs)
}

// ScanWithCache walks worktreesRoot in parallel and uses cacheDir to cache reports keyed by inode and mtime.
func ScanWithCache(worktreesRoot, cacheDir string, heavyDirs []string) ([]DirReport, error) {
	entries, err := os.ReadDir(worktreesRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var dirEntries []os.DirEntry
	for _, e := range entries {
		if e.IsDir() {
			dirEntries = append(dirEntries, e)
		}
	}

	if len(dirEntries) == 0 {
		return nil, nil
	}

	var cacheFile string
	cache := make(CacheStore)
	var cacheMu sync.Mutex

	if cacheDir != "" {
		cacheFile = filepath.Join(cacheDir, "disk_usage.json")
		cache = loadCache(cacheFile)
	}

	// Build heavyNames map
	hNames := make(map[string]bool)
	if len(heavyDirs) > 0 {
		for _, name := range heavyDirs {
			hNames[name] = true
		}
	} else {
		hNames = heavyNames
	}

	reports := make([]DirReport, len(dirEntries))
	valid := make([]bool, len(dirEntries))

	workerLimit := runtime.GOMAXPROCS(0) * 2
	if workerLimit < 4 {
		workerLimit = 4
	}

	var g errgroup.Group
	g.SetLimit(workerLimit)

	for i, entry := range dirEntries {
		i, entry := i, entry
		dir := filepath.Join(worktreesRoot, entry.Name())

		g.Go(func() error {
			info, err := os.Stat(dir)
			if err != nil {
				return nil // best effort
			}

			inode, mtimeNs := getFileStat(info)

			cacheMu.Lock()
			cached, hit := cache[dir]
			cacheMu.Unlock()

			if hit && cached.Inode == inode && cached.MtimeNs == mtimeNs {
				reports[i] = DirReport{
					Path:      dir,
					TotalSize: cached.TotalSize,
					HeavyDirs: cached.HeavyDirs,
				}
				valid[i] = true
				return nil
			}

			report, err := scanDirParallel(dir, hNames)
			if err != nil {
				return nil
			}

			reports[i] = report
			valid[i] = true

			if cacheDir != "" {
				cacheMu.Lock()
				cache[dir] = CacheEntry{
					Inode:     inode,
					MtimeNs:   mtimeNs,
					TotalSize: report.TotalSize,
					HeavyDirs: report.HeavyDirs,
				}
				cacheMu.Unlock()
			}
			return nil
		})
	}

	_ = g.Wait()

	if cacheDir != "" {
		saveCache(cacheFile, cache)
	}

	var result []DirReport
	for i, ok := range valid {
		if ok {
			result = append(result, reports[i])
		}
	}
	return result, nil
}

func scanDirParallel(root string, heavyNames map[string]bool) (DirReport, error) {
	report := DirReport{Path: root}

	type heavyTask struct {
		relPath string
		path    string
	}
	var heavyTasks []heavyTask

	var totalSize atomic.Int64

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable paths
		}
		if d.IsDir() && path != root && heavyNames[d.Name()] {
			rel, _ := filepath.Rel(root, path)
			heavyTasks = append(heavyTasks, heavyTask{relPath: rel, path: path})
			return filepath.SkipDir
		}
		if !d.IsDir() {
			if info, err := d.Info(); err == nil {
				totalSize.Add(info.Size())
			}
		}
		return nil
	})
	if err != nil {
		return report, err
	}

	if len(heavyTasks) > 0 {
		heavyDirs := make([]HeavyDir, len(heavyTasks))
		var g errgroup.Group
		g.SetLimit(runtime.GOMAXPROCS(0) * 2)

		for idx, task := range heavyTasks {
			idx, task := idx, task
			g.Go(func() error {
				size, _ := dirSizeParallel(task.path)
				heavyDirs[idx] = HeavyDir{
					RelPath: task.relPath,
					Size:    size,
				}
				totalSize.Add(size)
				return nil
			})
		}
		_ = g.Wait()
		report.HeavyDirs = heavyDirs
	}

	report.TotalSize = totalSize.Load()
	return report, nil
}

func dirSizeParallel(dir string) (int64, error) {
	var total atomic.Int64
	err := filepath.WalkDir(dir, func(_ string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if info, err := d.Info(); err == nil {
			total.Add(info.Size())
		}
		return nil
	})
	return total.Load(), err
}

func loadCache(cacheFile string) CacheStore {
	store := make(CacheStore)
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return store
	}
	_ = json.Unmarshal(data, &store)
	return store
}

func saveCache(cacheFile string, cache CacheStore) {
	if cacheFile == "" {
		return
	}
	dir := filepath.Dir(cacheFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return
	}

	tmpFile := cacheFile + fmt.Sprintf(".tmp.%d_%d", os.Getpid(), time.Now().UnixNano())
	if err := os.WriteFile(tmpFile, data, 0o644); err != nil {
		return
	}
	if err := os.Rename(tmpFile, cacheFile); err != nil {
		_ = os.Remove(tmpFile)
	}
}
