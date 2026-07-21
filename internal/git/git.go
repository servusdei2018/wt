// Package git provides low-level wrappers around git subprocess invocations.
package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunOpts configures a single git invocation.
type RunOpts struct {
	// Ctx is the context for cancellation. If nil, context.Background() is used.
	Ctx context.Context
	// Dir overrides the working directory. Empty inherits the process CWD.
	Dir string
}

// ExitError is returned when git exits with a non-zero status code.
type ExitError struct {
	Code   int
	Stderr string
}

func (e *ExitError) Error() string {
	msg := strings.TrimSpace(e.Stderr)
	if msg != "" {
		return fmt.Sprintf("git exited %d: %s", e.Code, msg)
	}
	return fmt.Sprintf("git exited %d", e.Code)
}

// Run executes git with args, returning trimmed stdout on success.
// On non-zero exit it returns *ExitError.
func Run(opts RunOpts, args ...string) (string, error) {
	ctx := opts.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	if opts.Dir != "" {
		cmd.Dir = opts.Dir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return "", &ExitError{Code: exit.ExitCode(), Stderr: stderr.String()}
		}
		return "", fmt.Errorf("exec git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// RepoRoot returns the absolute path of the repository root from any CWD.
func RepoRoot(ctx context.Context) (string, error) {
	root, err := Run(RunOpts{Ctx: ctx}, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("not inside a git repository")
	}
	return root, nil
}

// CommonGitDir returns the absolute path to the main repository .git directory.
// For secondary worktrees, this returns the shared .git directory of the parent repo.
func CommonGitDir(ctx context.Context, dir string) (string, error) {
	path, err := Run(RunOpts{Ctx: ctx, Dir: dir}, "rev-parse", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("could not determine git common dir: %w", err)
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}
	return filepath.Clean(path), nil
}

// CurrentBranch returns the short branch name checked out in dir.
func CurrentBranch(ctx context.Context, dir string) (string, error) {
	out, err := Run(RunOpts{Ctx: ctx, Dir: dir}, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("could not determine current branch: %w", err)
	}
	return out, nil
}
