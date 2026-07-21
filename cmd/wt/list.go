package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/servusdei2018/wt/internal/git"
	"github.com/servusdei2018/wt/internal/ui"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
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
			ctx := cmd.Context()
			worktrees, err := git.List(ctx, a.repoRoot)
			if err != nil {
				return err
			}

			enriched := make([]ui.EnrichedWorktree, len(worktrees))
			for i, wt := range worktrees {
				enriched[i] = ui.EnrichedWorktree{
					Worktree: wt,
					Index:    i + 1,
				}
			}

			var g errgroup.Group
			g.SetLimit(runtime.GOMAXPROCS(0) * 4)

			for i, wt := range worktrees {
				if wt.IsPrunable || wt.Branch == "" || wt.Branch == "(detached)" {
					continue
				}
				i, wt := i, wt
				g.Go(func() error {
					age, _ := git.Age(ctx, a.repoRoot, wt.Branch)
					status, _ := git.Get(ctx, wt.Path)
					enriched[i].Age = age
					enriched[i].Status = status
					return nil
				})
			}
			_ = g.Wait()

			pruneConfirmed, err := ui.RunList(enriched)
			if err != nil {
				return err
			}

			if pruneConfirmed {
				fmt.Println(ui.StyleMuted.Render("Running git worktree prune…"))
				if err := git.Prune(ctx, a.repoRoot); err != nil {
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
