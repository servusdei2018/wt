package git

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ValidateBranch checks if branch is a valid Git branch name and safe for worktree creation.
func ValidateBranch(branch string) error {
	if branch == "" {
		return fmt.Errorf("branch name cannot be empty")
	}
	if strings.HasPrefix(branch, "-") {
		return fmt.Errorf("branch name cannot start with hyphen: %q", branch)
	}
	if strings.Contains(branch, "..") {
		return fmt.Errorf("invalid branch name (contains '..'): %q", branch)
	}
	_, err := Run(RunOpts{}, "check-ref-format", "--branch", branch)
	if err != nil {
		return fmt.Errorf("invalid branch name %q", branch)
	}
	return nil
}

// DetectBase queries local tracking ref or remote to determine its HEAD branch name (e.g. "main"
// or "master"). Falls back to "main" if detection fails.
func DetectBase(repoRoot string) string {
	// Local symbolic ref
	out, err := Run(RunOpts{Dir: repoRoot}, "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	if err == nil && out != "" {
		b := strings.TrimPrefix(out, "origin/")
		if b != "" && b != out {
			return b
		}
	}

	// Fallback to remote query if local tracking ref is absent
	out, err = Run(RunOpts{Dir: repoRoot}, "remote", "show", "origin")
	if err == nil {
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "HEAD branch:") {
				b := strings.TrimSpace(strings.TrimPrefix(line, "HEAD branch:"))
				if b != "" && b != "(unknown)" {
					return b
				}
			}
		}
	}
	return "main"
}

// IsMerged reports whether branch has been fully merged into into (or origin/into).
func IsMerged(repoRoot, branch, into string) (bool, error) {
	// Check ancestor status against local into and remote into
	targets := []string{into, fmt.Sprintf("origin/%s", into)}
	for _, target := range targets {
		_, err := Run(RunOpts{Dir: repoRoot}, "merge-base", "--is-ancestor", branch, target)
		if err == nil {
			return true, nil
		}
	}

	// Fallback check with branch --merged using machine-readable output
	out, err := Run(RunOpts{Dir: repoRoot}, "branch", "--format=%(refname:short)", "--merged", into)
	if err == nil {
		for _, line := range strings.Split(out, "\n") {
			if strings.TrimSpace(line) == branch {
				return true, nil
			}
		}
	}
	return false, nil
}

// Delete removes the named branch. Use force=true to pass -D.
func Delete(repoRoot, branch string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	if _, err := Run(RunOpts{Dir: repoRoot}, "branch", flag, branch); err != nil {
		return fmt.Errorf("git branch %s %s: %w", flag, branch, err)
	}
	return nil
}

// Age returns the elapsed time since the last commit on branch.
// Returns 0 on error or missing history.
func Age(repoRoot, branch string) (time.Duration, error) {
	out, err := Run(RunOpts{Dir: repoRoot}, "log", "-1", "--format=%ct", branch, "--")
	if err != nil || out == "" {
		return 0, nil
	}
	ts, err := strconv.ParseInt(out, 10, 64)
	if err != nil {
		return 0, nil
	}
	return time.Since(time.Unix(ts, 0)), nil
}
