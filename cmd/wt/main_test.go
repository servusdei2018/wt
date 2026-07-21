package main

import (
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestIsNoRepoCommand(t *testing.T) {
	root := &cobra.Command{Use: "wt"}
	helpCmd := &cobra.Command{Use: "help"}
	compCmd := &cobra.Command{Use: "completion"}
	bashCmd := &cobra.Command{Use: "bash"}
	manCmd := &cobra.Command{Use: "man"}
	newCmd := &cobra.Command{Use: "new"}

	root.AddCommand(helpCmd, compCmd, manCmd, newCmd)
	compCmd.AddCommand(bashCmd)

	if !isNoRepoCommand(helpCmd) {
		t.Error("helpCmd should be no-repo command")
	}
	if !isNoRepoCommand(compCmd) {
		t.Error("compCmd should be no-repo command")
	}
	if !isNoRepoCommand(bashCmd) {
		t.Error("bashCmd (subcommand of completion) should be no-repo command")
	}
	if !isNoRepoCommand(manCmd) {
		t.Error("manCmd should be no-repo command")
	}
	if isNoRepoCommand(newCmd) {
		t.Error("newCmd should require repo")
	}
}

func TestWorktreePathValidation(t *testing.T) {
	tmpDir := t.TempDir()
	a := &app{repoRoot: tmpDir}

	validPath, err := a.worktreePath("feature/test")
	if err != nil {
		t.Fatalf("worktreePath() error = %v", err)
	}
	expected := filepath.Join(tmpDir, ".worktrees", "feature/test")
	if validPath != expected {
		t.Errorf("worktreePath() = %q, want %q", validPath, expected)
	}

	_, err = a.worktreePath("../../outside")
	if err == nil {
		t.Error("worktreePath() expected error for path traversal, got nil")
	}
}
