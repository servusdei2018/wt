package main

import (
	"fmt"
	"time"

	"github.com/servusdei2018/wt/internal/git"
	"github.com/servusdei2018/wt/internal/ui"

	"github.com/spf13/cobra"
)

func (a *app) listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all worktrees with status and age",
		Long: `Displays a table of all active worktrees.

Each row shows an index shortcut, branch name, HEAD SHA, age since last commit,
dirty/clean status, and the worktree path.

Orphaned worktrees (where the directory no longer exists) are highlighted and
you will be prompted to run 'git worktree prune'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			worktrees, err := git.List(a.repoRoot)
			if err != nil {
				return err
			}

			var enriched []ui.EnrichedWorktree
			for i, wt := range worktrees {
				e := ui.EnrichedWorktree{
					Worktree: wt,
					Index:    i + 1,
				}
				if !wt.IsPrunable && wt.Branch != "" && wt.Branch != "(detached)" {
					e.Age, _ = git.Age(a.repoRoot, wt.Branch)
					e.Status, _ = git.Get(wt.Path)
				}
				enriched = append(enriched, e)
			}

			pruneConfirmed, err := ui.RunList(enriched)
			if err != nil {
				return err
			}

			if pruneConfirmed {
				fmt.Println(ui.StyleMuted.Render("Running git worktree prune…"))
				if err := git.Prune(a.repoRoot); err != nil {
					return err
				}
				fmt.Println(ui.StyleSuccessBox.Render(
					ui.StyleSuccess.Render("Pruned orphaned worktrees."),
				))
			}
			return nil
		},
	}
}

// Ensure time.Duration satisfies the FormatAge interface.
var _ interface{ Hours() float64 } = time.Duration(0)
