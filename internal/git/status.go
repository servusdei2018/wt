package git

import (
	"fmt"
	"strings"
)

// Status represents the working-tree state of a worktree.
type Status struct {
	IsDirty   bool
	Staged    int
	Unstaged  int
	Untracked int
}

// Get returns the working-tree status of the worktree at path.
func Get(worktreePath string) (Status, error) {
	out, err := Run(RunOpts{Dir: worktreePath}, "status", "--porcelain")
	if err != nil {
		return Status{}, fmt.Errorf("git status in %s: %w", worktreePath, err)
	}
	return parseStatus(out), nil
}

func parseStatus(out string) Status {
	var s Status
	for _, line := range strings.Split(out, "\n") {
		if len(line) < 2 {
			continue
		}
		s.IsDirty = true
		x, y := line[0], line[1]
		switch {
		case x == '?' && y == '?':
			s.Untracked++
		default:
			if x != ' ' && x != '?' {
				s.Staged++
			}
			if y != ' ' && y != '?' {
				s.Unstaged++
			}
		}
	}
	return s
}
