package main

import (
	"fmt"
	"path/filepath"

	"strings"

	"github.com/servusdei2018/wt/internal/editor"
	"github.com/servusdei2018/wt/internal/git"
	"github.com/servusdei2018/wt/internal/hook"
	"github.com/servusdei2018/wt/internal/seed"
	"github.com/servusdei2018/wt/internal/storage"
	"github.com/servusdei2018/wt/internal/ui"

	"github.com/spf13/cobra"
)

func (a *app) newCmd() *cobra.Command {
	var (
		fromRef string
		open    bool
	)

	cmd := &cobra.Command{
		Use:   "new <branch>",
		Short: "Create a new worktree for a feature branch",
		Long: `Provisions a new branch and isolated worktree workspace under .worktrees/.

The branch is created from the HEAD of the remote base branch (auto-detected
or configured via .wt.toml). Use --from to override the starting ref.

A post-creation hook is run automatically if one is detected (npm install,
go mod download, etc.) or configured in .wt.toml.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branch := args[0]
			if err := git.ValidateBranch(branch); err != nil {
				return err
			}

			path, err := a.worktreePath(branch)
			if err != nil {
				return err
			}

			// Resolve the ref to branch from.
			ref := fromRef
			if ref == "" {
				base := a.baseBranch()
				ref = fmt.Sprintf("origin/%s", base)
			}

			fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("Creating worktree for branch %q from %s…", branch, ref)))

			if err := git.Add(a.repoRoot, branch, path, ref); err != nil {
				return fmt.Errorf("worktree add failed: %w", err)
			}

			// Pre-seed heavy dependency directories via CoW reflink if enabled
			if a.cfg.Storage.DedupeOnCreate {
				seedStats, err := storage.SeedWorktree(a.repoRoot, path, a.cfg.Storage.HeavyDirs, a.cfg.Storage.RequireReflink)
				if err == nil && seedStats.FilesTotal > 0 {
					fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("Seeded %d heavy dependency files (%s) via CoW reflink", seedStats.FilesTotal, ui.FormatSize(seedStats.BytesTotal))))
				}
			}

			// Copy local config and secret files with variable interpolation
			fileSeedReport, err := seed.SeedFiles(a.repoRoot, path, branch, a.cfg.Seed)
			if err != nil {
				return fmt.Errorf("seed files failed: %w", err)
			}
			if fileSeedReport.Count() > 0 {
				fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("Seeded %d config file(s) into worktree (%s)", fileSeedReport.Count(), strings.Join(fileSeedReport.FilesCopied, ", "))))
			}

			// Detect and run post-create hook.
			h, err := hook.Detect(path, a.cfg)
			if err != nil {
				return err
			}
			if h != nil {
				fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("Running hook: %s", h.Name)))
				if err := hook.Run(h, path); err != nil {
					return err
				}
			}

			relPath, _ := filepath.Rel(a.repoRoot, path)
			fmt.Println(ui.StyleSuccessBox.Render(
				ui.StyleSuccess.Render("Worktree ready") + "\n" +
					ui.StyleMuted.Render("  branch  ") + ui.StyleBranch.Render(branch) + "\n" +
					ui.StyleMuted.Render("  path    ") + relPath,
			))

			if open {
				ed := editor.Resolve(a.cfg)
				fmt.Println(ui.StyleMuted.Render(fmt.Sprintf("Opening %s in %s…", relPath, ed)))
				return editor.Open(ed, path)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&fromRef, "from", "", "Branch from this ref instead of the remote base branch HEAD")
	cmd.Flags().BoolVar(&open, "open", false, "Open the new worktree in $EDITOR after creation")
	return cmd
}
