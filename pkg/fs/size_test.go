package fs

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScanDiskUsage(t *testing.T) {
	tmpDir := t.TempDir()

	// Create worktree 1 with node_modules
	wt1 := filepath.Join(tmpDir, "wt-1")
	if err := os.MkdirAll(filepath.Join(wt1, "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wt1, "node_modules", "pkg.js"), []byte("1234567890"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wt1, "main.js"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	reports, err := Scan(tmpDir)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("Scan() returned %d reports, want 1", len(reports))
	}

	r := reports[0]
	if r.Path != wt1 {
		t.Errorf("Report Path = %q, want %q", r.Path, wt1)
	}
	if len(r.HeavyDirs) != 1 {
		t.Fatalf("HeavyDirs count = %d, want 1", len(r.HeavyDirs))
	}
	if r.HeavyDirs[0].RelPath != "node_modules" {
		t.Errorf("HeavyDir RelPath = %q, want %q", r.HeavyDirs[0].RelPath, "node_modules")
	}
	if r.HeavyDirs[0].Size != 10 {
		t.Errorf("HeavyDir Size = %d, want 10", r.HeavyDirs[0].Size)
	}
	if r.TotalSize != 15 {
		t.Errorf("TotalSize = %d, want 15", r.TotalSize)
	}
}

func TestScanWithCache(t *testing.T) {
	worktreesRoot := t.TempDir()
	cacheDir := t.TempDir()

	wt1 := filepath.Join(worktreesRoot, "wt-1")
	if err := os.MkdirAll(filepath.Join(wt1, "vendor"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wt1, "vendor", "lib.go"), []byte("package vendor"), 0o644); err != nil {
		t.Fatal(err)
	}

	// First scan creates cache
	reports1, err := ScanWithCache(worktreesRoot, cacheDir)
	if err != nil {
		t.Fatalf("First ScanWithCache error = %v", err)
	}
	if len(reports1) != 1 || reports1[0].TotalSize != 14 {
		t.Fatalf("First ScanWithCache reports = %+v", reports1)
	}

	cacheFile := filepath.Join(cacheDir, "disk_usage.json")
	if _, err := os.Stat(cacheFile); err != nil {
		t.Fatalf("Cache file should exist at %s: %v", cacheFile, err)
	}

	// Second scan reads from cache
	reports2, err := ScanWithCache(worktreesRoot, cacheDir)
	if err != nil {
		t.Fatalf("Second ScanWithCache error = %v", err)
	}
	if len(reports2) != 1 || reports2[0].TotalSize != 14 {
		t.Fatalf("Second ScanWithCache reports = %+v", reports2)
	}

	// Modify mtime of wt1 directory to trigger cache update
	newTime := time.Now().Add(5 * time.Second)
	_ = os.Chtimes(wt1, newTime, newTime)

	// Add file to wt1
	if err := os.WriteFile(filepath.Join(wt1, "extra.txt"), []byte("12345"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Ensure directory mtime updates
	_ = os.Chtimes(wt1, time.Now().Add(10*time.Second), time.Now().Add(10*time.Second))

	reports3, err := ScanWithCache(worktreesRoot, cacheDir)
	if err != nil {
		t.Fatalf("Third ScanWithCache error = %v", err)
	}
	if len(reports3) != 1 || reports3[0].TotalSize != 19 {
		t.Fatalf("Third ScanWithCache total size = %d, want 19", reports3[0].TotalSize)
	}

	// Test corrupted cache file handling
	if err := os.WriteFile(cacheFile, []byte("{invalid json}"), 0o644); err != nil {
		t.Fatal(err)
	}
	reports4, err := ScanWithCache(worktreesRoot, cacheDir)
	if err != nil {
		t.Fatalf("ScanWithCache with corrupted cache error = %v", err)
	}
	if len(reports4) != 1 || reports4[0].TotalSize != 19 {
		t.Fatalf("ScanWithCache with corrupted cache total size = %d, want 19", reports4[0].TotalSize)
	}
}
