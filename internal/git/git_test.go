package git

import (
	"testing"
)

func TestValidateBranch(t *testing.T) {
	tests := []struct {
		name    string
		branch  string
		wantErr bool
	}{
		{"valid branch", "feature/my-cool-branch", false},
		{"simple branch", "main", false},
		{"empty branch", "", true},
		{"hyphen prefix", "-rf", true},
		{"path traversal", "../outside", true},
		{"invalid git ref", "head..tail", true},
		{"invalid char", "branch~1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBranch(tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBranch(%q) error = %v, wantErr %v", tt.branch, err, tt.wantErr)
			}
		})
	}
}

func TestParseStatus(t *testing.T) {
	output := ` M file1.txt
M  file2.txt
?? file3.txt
 D file4.txt`

	status := parseStatus(output)
	if !status.IsDirty {
		t.Error("parseStatus() expected IsDirty = true")
	}
	if status.Staged != 1 {
		t.Errorf("parseStatus() Staged = %d, want 1", status.Staged)
	}
	if status.Unstaged != 2 {
		t.Errorf("parseStatus() Unstaged = %d, want 2", status.Unstaged)
	}
	if status.Untracked != 1 {
		t.Errorf("parseStatus() Untracked = %d, want 1", status.Untracked)
	}
}

func TestParseWorktreeList(t *testing.T) {
	output := `worktree /path/to/main
HEAD 1a2b3c4d5e6f7890
branch refs/heads/main

worktree /path/to/.worktrees/feature
HEAD 9f8e7d6c5b4a3210
branch refs/heads/feature

`
	worktrees := parseWorktreeList(output)
	if len(worktrees) != 2 {
		t.Fatalf("parseWorktreeList() returned %d items, want 2", len(worktrees))
	}

	if worktrees[0].Branch != "main" || !worktrees[0].IsMain || worktrees[0].HEAD != "1a2b3c4" {
		t.Errorf("worktrees[0] = %+v", worktrees[0])
	}
	if worktrees[1].Branch != "feature" || worktrees[1].IsMain || worktrees[1].HEAD != "9f8e7d6" {
		t.Errorf("worktrees[1] = %+v", worktrees[1])
	}
}
