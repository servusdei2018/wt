package main

import (
	"fmt"

	"github.com/servusdei2018/wt/internal/git"
	"github.com/servusdei2018/wt/internal/ui"

	"github.com/spf13/cobra"
)

func (a *app) refreshCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "refresh",
		Short: "Fetch remotes and update the local base branch pointer",
		Long: `A non-intrusive background update.

Runs git fetch --all and then fast-forwards the local base branch pointer to
match the remote, without touching your current worktree or requiring a
checkout. You stay exactly where you are.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			remote := a.cfg.Sync.Remote
			base := a.baseBranch()

			fmt.Println(ui.StyleMuted.Render("Fetching all remotes…"))
			if _, err := git.Run(git.RunOpts{Dir: a.repoRoot}, "fetch", "--all"); err != nil {
				return fmt.Errorf("git fetch --all: %w", err)
			}

			fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("Updating local %s → %s/%s…", base, remote, base)))
			if _, err := git.Run(git.RunOpts{Dir: a.repoRoot},
				"fetch", remote, fmt.Sprintf("%s:%s", base, base)); err != nil {
				fmt.Println(ui.StyleWarning.Render(
					fmt.Sprintf("  note: could not fast-forward local %s (may already be up-to-date or checked out)", base),
				))
			}

			fmt.Println(ui.StyleSuccessBox.Render(
				ui.StyleSuccess.Render("Refreshed") + "\n" +
					ui.StyleMuted.Render("  remote  ") + ui.StyleBranch.Render(remote) + "\n" +
					ui.StyleMuted.Render("  base    ") + ui.StyleBranch.Render(base),
			))
			return nil
		},
	}
}
