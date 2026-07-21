package exclude

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	pattern := ".worktrees/*"

	// First call should create directory and file, then append pattern
	if err := Ensure(tmpDir, pattern); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	excludeFile := filepath.Join(tmpDir, ".git", "info", "exclude")
	data, err := os.ReadFile(excludeFile)
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}
	if !strings.Contains(string(data), pattern) {
		t.Errorf("Exclude file missing pattern %q, content:\n%s", pattern, string(data))
	}

	// Second call should be a no-op
	if err := Ensure(tmpDir, pattern); err != nil {
		t.Fatalf("Ensure() second call error = %v", err)
	}

	dataSecond, _ := os.ReadFile(excludeFile)
	if string(data) != string(dataSecond) {
		t.Errorf("Ensure() was not idempotent. First content:\n%s\nSecond content:\n%s", string(data), string(dataSecond))
	}
}
