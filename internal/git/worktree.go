package git

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// Worktree represents a single git worktree entry.
type Worktree struct {
	Path       string
	Branch     string
	HEAD       string // 7-char short SHA
	IsBare     bool
	IsMain     bool // true for the primary checkout
	IsPrunable bool // path no longer exists on disk
}

// Add creates a new worktree at path, branching from fromRef.
func Add(ctx context.Context, repoRoot, branch, path, fromRef string) error {
	_, err := Run(RunOpts{Ctx: ctx, Dir: repoRoot}, "worktree", "add", "-b", branch, path, fromRef)
	if err != nil {
		return fmt.Errorf("git worktree add: %w", err)
	}
	return nil
}

// Remove removes the worktree at path. Pass force=true to allow removal of dirty trees.
func Remove(ctx context.Context, path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)
	if _, err := Run(RunOpts{Ctx: ctx}, args...); err != nil {
		return fmt.Errorf("git worktree remove: %w", err)
	}
	return nil
}

// List parses `git worktree list --porcelain` and returns all worktrees.
func List(ctx context.Context, repoRoot string) ([]Worktree, error) {
	out, err := Run(RunOpts{Ctx: ctx, Dir: repoRoot}, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("git worktree list: %w", err)
	}
	return parseWorktreeList(out), nil
}

// Prune removes stale worktree administrative files.
func Prune(ctx context.Context, repoRoot string) error {
	if _, err := Run(RunOpts{Ctx: ctx, Dir: repoRoot}, "worktree", "prune"); err != nil {
		return fmt.Errorf("git worktree prune: %w", err)
	}
	return nil
}

func parseWorktreeList(out string) []Worktree {
	var (
		results []Worktree
		current Worktree
		isFirst = true
	)
	flush := func() {
		if current.Path == "" {
			return
		}
		current.IsMain = isFirst
		isFirst = false
		current.IsPrunable = !pathExists(current.Path)
		results = append(results, current)
		current = Worktree{}
	}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			flush()
			continue
		}
		switch {
		case strings.HasPrefix(line, "worktree "):
			current.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "HEAD "):
			h := strings.TrimPrefix(line, "HEAD ")
			if len(h) > 7 {
				h = h[:7]
			}
			current.HEAD = h
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "bare":
			current.IsBare = true
		case line == "detached":
			current.Branch = "(detached)"
		}
	}
	flush()
	return results
}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
