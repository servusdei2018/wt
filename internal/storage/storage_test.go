package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeduplicateWorktrees(t *testing.T) {
	tempDir := t.TempDir()

	repoRoot := filepath.Join(tempDir, "repo")
	worktreesDir := filepath.Join(repoRoot, ".worktrees")

	wt1 := filepath.Join(worktreesDir, "feature-1")
	wt2 := filepath.Join(worktreesDir, "feature-2")

	// Create node_modules in wt1 and wt2 with identical file content
	heavy1 := filepath.Join(wt1, "node_modules", "package-a")
	heavy2 := filepath.Join(wt2, "node_modules", "package-a")

	if err := os.MkdirAll(heavy1, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(heavy2, 0o755); err != nil {
		t.Fatal(err)
	}

	content := []byte("export const version = '1.0.0';\nconsole.log('heavy package payload');\n")
	file1 := filepath.Join(heavy1, "index.js")
	file2 := filepath.Join(heavy2, "index.js")

	if err := os.WriteFile(file1, content, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, content, 0o644); err != nil {
		t.Fatal(err)
	}

	opts := DedupeOptions{
		RequireReflink: false,
		DryRun:         true,
		HeavyDirs:      []string{"node_modules"},
	}

	// 1. Dry run test
	report, err := DeduplicateWorktrees(repoRoot, worktreesDir, opts)
	if err != nil {
		t.Fatalf("DeduplicateWorktrees dry run failed: %v", err)
	}

	if report.FilesDeduplicated != 1 {
		t.Errorf("expected 1 file deduplicated in dry run, got %d", report.FilesDeduplicated)
	}
	if report.BytesSaved != int64(len(content)) {
		t.Errorf("expected %d bytes saved, got %d", len(content), report.BytesSaved)
	}

	// 2. Real execution test
	opts.DryRun = false
	report, err = DeduplicateWorktrees(repoRoot, worktreesDir, opts)
	if err != nil {
		t.Fatalf("DeduplicateWorktrees execution failed: %v", err)
	}

	// On filesystems supporting CoW reflink (Btrfs, XFS, APFS), FilesDeduplicated is 1.
	// On unsupported filesystems (ext4, tmpfs), execution gracefully skips standard file copying.
	if report.ReflinkCount > 0 && report.FilesDeduplicated != 1 {
		t.Errorf("expected 1 file deduplicated in execution on CoW filesystem, got %d", report.FilesDeduplicated)
	}

	// 3. RequireReflink test on unsupported filesystem
	opts.RequireReflink = true
	_, err = DeduplicateWorktrees(repoRoot, worktreesDir, opts)
	// If reflink is unsupported on tempDir filesystem, RequireReflink must return an error
	if report.ReflinkCount == 0 && err == nil {
		t.Errorf("expected error when RequireReflink is true on unsupported filesystem, got nil")
	}

	// Verify file content intact
	data1, _ := os.ReadFile(file1)
	data2, _ := os.ReadFile(file2)
	if string(data1) != string(content) || string(data2) != string(content) {
		t.Errorf("file contents corrupted after dedupe")
	}
}

func TestSeedWorktree(t *testing.T) {
	tempDir := t.TempDir()

	repoRoot := filepath.Join(tempDir, "repo")
	worktreesDir := filepath.Join(repoRoot, ".worktrees")

	// Main repo root has node_modules
	rootModules := filepath.Join(repoRoot, "node_modules")
	if err := os.MkdirAll(rootModules, 0o755); err != nil {
		t.Fatal(err)
	}
	sampleFile := filepath.Join(rootModules, "react.js")
	if err := os.WriteFile(sampleFile, []byte("module.exports = React;"), 0o644); err != nil {
		t.Fatal(err)
	}

	targetWt := filepath.Join(worktreesDir, "feature-new")
	if err := os.MkdirAll(targetWt, 0o755); err != nil {
		t.Fatal(err)
	}

	stats, err := SeedWorktree(repoRoot, targetWt, []string{"node_modules"}, false)
	if err != nil {
		t.Fatalf("SeedWorktree failed: %v", err)
	}

	if stats.FilesTotal != 1 {
		t.Errorf("expected 1 file seeded, got %d", stats.FilesTotal)
	}

	seededFile := filepath.Join(targetWt, "node_modules", "react.js")
	data, err := os.ReadFile(seededFile)
	if err != nil {
		t.Fatalf("seeded file missing: %v", err)
	}
	if string(data) != "module.exports = React;" {
		t.Errorf("seeded file content mismatch: %s", string(data))
	}
}
