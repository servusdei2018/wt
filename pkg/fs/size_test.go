package fs

import (
	"os"
	"path/filepath"
	"testing"
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
