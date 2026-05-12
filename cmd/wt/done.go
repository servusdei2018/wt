package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/servusdei2018/wt/internal/git"
	"github.com/servusdei2018/wt/internal/ui"

	"github.com/spf13/cobra"
)

func (a *app) doneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "done [branch]",
		Short: "Tear down a worktree and delete its branch",
		Long: `Safely removes a worktree and its associated branch.

If run from inside a worktree, the current branch is used when no argument is
given. If run from the main worktree, a branch name must be supplied explicitly.

Checks are performed before any destructive action:
  - Dirty working tree triggers an interactive prompt (stash / push / force).
  - Unmerged branch triggers a confirmation prompt.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branch, err := a.resolveDoneBranch(args)
			if err != nil {
				return err
			}

			path := a.worktreePath(branch)
			base := a.baseBranch()
			force := false

			// --- Dirty check --------------------------------------------------
			status, err := git.Get(path)
			if err != nil {
				return fmt.Errorf(
					"FATAL: could not read git status for %q — handle this directory manually.\n%w",
					path, err,
				)
			}

			if status.IsDirty {
				action, err := ui.AskDirtyAction(branch)
				if err != nil {
					return err
				}
				switch action {
				case ui.ActionAbort:
					fmt.Println(ui.StyleMuted.Render("Aborted."))
					return nil
				case ui.ActionStash:
					if _, err := git.Run(git.RunOpts{Dir: path}, "stash"); err != nil {
						return fmt.Errorf("git stash: %w", err)
					}
				case ui.ActionPush:
					if _, err := git.Run(git.RunOpts{Dir: path}, "push", a.cfg.Sync.Remote, branch); err != nil {
						return fmt.Errorf("git push: %w", err)
					}
				case ui.ActionForceDelete:
					force = true
				}
			}

			// --- Merge check --------------------------------------------------
			if !force {
				merged, err := git.IsMerged(a.repoRoot, branch, base)
				if err != nil {
					return fmt.Errorf("FATAL: could not check merge status: %w", err)
				}
				if !merged {
					ok, err := ui.Confirm(
						fmt.Sprintf("Branch %q is not merged into %q.", branch, base),
						"Continuing will permanently delete unmerged work.",
					)
					if err != nil || !ok {
						fmt.Println(ui.StyleMuted.Render("Aborted."))
						return nil
					}
				}
			}

			// --- Teardown -----------------------------------------------------
			fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("Removing worktree %q…", branch)))

			if err := git.Remove(path, force); err != nil {
				return fmt.Errorf("FATAL: git worktree remove failed: %w\nHandle %q manually", err, path)
			}
			if err := git.Delete(a.repoRoot, branch, force); err != nil {
				return fmt.Errorf("FATAL: git branch delete failed: %w", err)
			}
			// Remove any residual files.
			_ = os.RemoveAll(path)

			relPath, _ := filepath.Rel(a.repoRoot, path)
			fmt.Println(ui.StyleSuccessBox.Render(
				ui.StyleSuccess.Render(ui.IndicatorClean+" Done") + "\n" +
					ui.StyleMuted.Render("  branch  ") + ui.StyleBranch.Render(branch) + "\n" +
					ui.StyleMuted.Render("  removed ") + relPath,
			))
			return nil
		},
	}
	return cmd
}

// resolveDoneBranch determines the target branch for `wt done`.
//
// From inside a worktree the current branch is used when no arg is given.
// From the main worktree an explicit arg is required.
func (a *app) resolveDoneBranch(args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	worktreesDir := a.worktreesDir()
	if !strings.HasPrefix(cwd, worktreesDir) {
		return "", fmt.Errorf(
			"not inside a worktree — provide a branch name explicitly:\n  wt done <branch>",
		)
	}

	return git.CurrentBranch(cwd)
}
