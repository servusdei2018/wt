package fs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCloneFileAndDir(t *testing.T) {
	tempDir := t.TempDir()

	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")

	if err := os.MkdirAll(filepath.Join(srcDir, "sub"), 0o755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	file1 := filepath.Join(srcDir, "sub", "test1.txt")
	content1 := []byte("hello world dependency artifact content")
	if err := os.WriteFile(file1, content1, 0o644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}

	file2 := filepath.Join(srcDir, "test2.txt")
	content2 := []byte("another dependency artifact content line")
	if err := os.WriteFile(file2, content2, 0o755); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}

	// Test CloneDir
	opts := CloneOptions{
		RequireReflink: false,
		PreserveAttrs:  true,
	}

	stats, err := CloneDir(srcDir, dstDir, opts)
	if err != nil {
		t.Fatalf("CloneDir failed: %v", err)
	}

	if stats.FilesTotal != 2 {
		t.Errorf("expected 2 total files, got %d", stats.FilesTotal)
	}

	expectedBytes := int64(len(content1) + len(content2))
	if stats.BytesTotal != expectedBytes {
		t.Errorf("expected %d bytes total, got %d", expectedBytes, stats.BytesTotal)
	}

	// Verify content in cloned directory
	dstFile1 := filepath.Join(dstDir, "sub", "test1.txt")
	dstContent1, err := os.ReadFile(dstFile1)
	if err != nil {
		t.Fatalf("failed to read dstFile1: %v", err)
	}
	if string(dstContent1) != string(content1) {
		t.Errorf("content mismatch in dstFile1: got %q, want %q", dstContent1, content1)
	}

	dstFile2 := filepath.Join(dstDir, "test2.txt")
	dstContent2, err := os.ReadFile(dstFile2)
	if err != nil {
		t.Fatalf("failed to read dstFile2: %v", err)
	}
	if string(dstContent2) != string(content2) {
		t.Errorf("content mismatch in dstFile2: got %q, want %q", dstContent2, content2)
	}
}

func TestCloneFileRequireReflinkUnsupported(t *testing.T) {
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "src.txt")
	dstFile := filepath.Join(tempDir, "dst.txt")

	if err := os.WriteFile(srcFile, []byte("data"), 0o644); err != nil {
		t.Fatalf("write src failed: %v", err)
	}

	// If underlying filesystem (e.g. ext4/tmpfs) doesn't support FICLONE, RequireReflink should return error.
	opts := CloneOptions{RequireReflink: true}
	strat, err := CloneFile(srcFile, dstFile, opts)
	if err != nil {
		// Verify dst file was cleaned up on failure
		if _, errExists := os.Stat(dstFile); !os.IsNotExist(errExists) {
			t.Errorf("expected dstFile to be cleaned up on reflink error, but it exists")
		}
	} else {
		// If platform supported reflink (e.g. Btrfs/XFS/APFS), verify strategy was reflink
		if strat != StrategyReflink {
			t.Errorf("expected StrategyReflink, got %s", strat)
		}
	}
}

func TestCloneStatsJSON(t *testing.T) {
	stats := CloneStats{
		FilesTotal:     10,
		FilesReflinked: 8,
		FilesCopied:    2,
		BytesTotal:     1000,
		BytesReflinked: 800,
		BytesCopied:    200,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"files_reflinked":8`) {
		t.Errorf("expected JSON tag 'files_reflinked', got %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"bytes_reflinked":800`) {
		t.Errorf("expected JSON tag 'bytes_reflinked', got %s", jsonStr)
	}
}
