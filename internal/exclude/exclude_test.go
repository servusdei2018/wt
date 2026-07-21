package exclude

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/servusdei2018/wt/internal/git"
)

func TestEnsureIdempotent(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	pattern := ".worktrees/*"

	// Initialize a git repo so git.CommonGitDir works
	_, err := git.Run(git.RunOpts{Ctx: ctx, Dir: tmpDir}, "init")
	if err != nil {
		t.Skip("git init not available:", err)
	}

	// First call should create directory and file, then append pattern
	if err := Ensure(ctx, tmpDir, pattern); err != nil {
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
	if err := Ensure(ctx, tmpDir, pattern); err != nil {
		t.Fatalf("Ensure() second call error = %v", err)
	}

	dataSecond, _ := os.ReadFile(excludeFile)
	if string(data) != string(dataSecond) {
		t.Errorf("Ensure() was not idempotent. First content:\n%s\nSecond content:\n%s", string(data), string(dataSecond))
	}
}
