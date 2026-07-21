package git

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ValidateBranch checks if branch is a valid Git branch name and safe for worktree creation.
func ValidateBranch(ctx context.Context, branch string) error {
	if branch == "" {
		return fmt.Errorf("branch name cannot be empty")
	}
	if strings.HasPrefix(branch, "-") {
		return fmt.Errorf("branch name cannot start with hyphen: %q", branch)
	}
	if strings.Contains(branch, "..") {
		return fmt.Errorf("invalid branch name (contains '..'): %q", branch)
	}
	_, err := Run(RunOpts{Ctx: ctx}, "check-ref-format", "--branch", branch)
	if err != nil {
		return fmt.Errorf("invalid branch name %q", branch)
	}
	return nil
}

// DetectBase queries local tracking ref or local HEAD branch name to determine base branch
// (e.g. "main" or "master"). Performs no network calls. Falls back to "main" if detection fails.
func DetectBase(ctx context.Context, repoRoot string) string {
	// Local symbolic ref for origin/HEAD
	out, err := Run(RunOpts{Ctx: ctx, Dir: repoRoot}, "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	if err == nil && out != "" {
		b := strings.TrimPrefix(out, "origin/")
		if b != "" && b != out {
			return b
		}
	}

	// Fallback to checking existing local branches (main, master)
	for _, candidate := range []string{"main", "master"} {
		_, err := Run(RunOpts{Ctx: ctx, Dir: repoRoot}, "rev-parse", "--verify", candidate)
		if err == nil {
			return candidate
		}
	}
	return "main"
}

// IsMerged reports whether branch has been fully merged into into (or origin/into).
func IsMerged(ctx context.Context, repoRoot, branch, into string) (bool, error) {
	// Check ancestor status against local into and remote into
	targets := []string{into, fmt.Sprintf("origin/%s", into)}
	for _, target := range targets {
		_, err := Run(RunOpts{Ctx: ctx, Dir: repoRoot}, "merge-base", "--is-ancestor", branch, target)
		if err == nil {
			return true, nil
		}
	}

	// Fallback check with branch --merged using machine-readable output
	out, err := Run(RunOpts{Ctx: ctx, Dir: repoRoot}, "branch", "--format=%(refname:short)", "--merged", into)
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
func Delete(ctx context.Context, repoRoot, branch string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	if _, err := Run(RunOpts{Ctx: ctx, Dir: repoRoot}, "branch", flag, branch); err != nil {
		return fmt.Errorf("git branch %s %s: %w", flag, branch, err)
	}
	return nil
}

// Age returns the elapsed time since the last commit on branch.
// Returns 0 on error or missing history.
func Age(ctx context.Context, repoRoot, branch string) (time.Duration, error) {
	out, err := Run(RunOpts{Ctx: ctx, Dir: repoRoot}, "log", "-1", "--format=%ct", branch, "--")
	if err != nil || out == "" {
		return 0, nil
	}
	ts, err := strconv.ParseInt(out, 10, 64)
	if err != nil {
		return 0, nil
	}
	return time.Since(time.Unix(ts, 0)), nil
}
