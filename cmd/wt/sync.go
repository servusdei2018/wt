package main

import (
	"fmt"
	"os"

	"github.com/servusdei2018/wt/internal/git"
	"github.com/servusdei2018/wt/internal/ui"

	"github.com/spf13/cobra"
)

func (a *app) syncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Rebase the current worktree onto the latest remote base branch",
		Long: `Aligns the current worktree with the upstream base branch.

Steps:
  1. Fetch all remotes.
  2. Fast-forward the local base branch pointer to match the remote.
  3. Rebase the current branch onto the updated remote base branch.

If rebase conflicts occur, wt pauses and instructs you to resolve them
manually before running 'git rebase --continue'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			remote := a.cfg.Sync.Remote
			base := a.baseBranch(ctx)
			remoteBase := fmt.Sprintf("%s/%s", remote, base)

			fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("Fetching from %s…", remote)))
			if _, err := git.Run(git.RunOpts{Ctx: ctx, Dir: a.repoRoot}, "fetch", remote); err != nil {
				return fmt.Errorf("fetch failed: %w", err)
			}

			fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("Updating local %s…", base)))
			if _, err := git.Run(git.RunOpts{Ctx: ctx, Dir: a.repoRoot},
				"fetch", remote, fmt.Sprintf("%s:%s", base, base)); err != nil {
				// Non-fatal: local branch may not exist yet.
				fmt.Println(ui.StyleWarning.Render(fmt.Sprintf("  note: could not fast-forward local %s", base)))
			}

			fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("Rebasing onto %s…", remoteBase)))
			if _, err := git.Run(git.RunOpts{Ctx: ctx, Dir: cwd}, "rebase", remoteBase); err != nil {
				return fmt.Errorf(
					"rebase conflict detected.\n\n"+
						"Resolve conflicts, then run:\n"+
						"  %s\n\n"+
						"To abort the rebase:\n"+
						"  %s",
					ui.StyleBranch.Render("git rebase --continue"),
					ui.StyleBranch.Render("git rebase --abort"),
				)
			}

			fmt.Println(ui.StyleSuccessBox.Render(
				ui.StyleSuccess.Render("Synced") + "\n" +
					ui.StyleMuted.Render("  rebased onto ") + ui.StyleBranch.Render(remoteBase),
			))
			return nil
		},
	}
}
