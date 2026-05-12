// Package exclude manages the .git/info/exclude file.
package exclude

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Ensure idempotently appends pattern to .git/info/exclude.
// Creates the file and its parent directory if they do not exist.
// No-ops if the pattern is already present.
func Ensure(repoRoot, pattern string) error {
	path := filepath.Join(repoRoot, ".git", "info", "exclude")

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("exclude: mkdir: %w", err)
	}

	// Check for existing pattern.
	if f, err := os.Open(path); err == nil {
		defer func() {
			_ = f.Close()
		}()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if strings.TrimSpace(sc.Text()) == pattern {
				return nil
			}
		}
	}

	// Append.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("exclude: open: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()
	_, err = fmt.Fprintf(f, "\n# managed by wt\n%s\n", pattern)
	return err
}
