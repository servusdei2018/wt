package git

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DetectBase queries the remote to determine its HEAD branch name (e.g. "main"
// or "master"). Falls back to "main" if detection fails.
func DetectBase(repoRoot string) string {
	out, err := Run(RunOpts{Dir: repoRoot}, "remote", "show", "origin")
	if err != nil {
		return "main"
	}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "HEAD branch:") {
			b := strings.TrimSpace(strings.TrimPrefix(line, "HEAD branch:"))
			if b != "" && b != "(unknown)" {
				return b
			}
		}
	}
	return "main"
}

// IsMerged reports whether branch has been fully merged into into.
func IsMerged(repoRoot, branch, into string) (bool, error) {
	out, err := Run(RunOpts{Dir: repoRoot}, "branch", "--merged", into)
	if err != nil {
		return false, fmt.Errorf("git branch --merged: %w", err)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(strings.TrimPrefix(line, "* ")) == branch {
			return true, nil
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
// Returns 0 on error or missing history (treated as no age).
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
