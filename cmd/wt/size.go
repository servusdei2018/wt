package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/servusdei2018/wt/internal/git"
	"github.com/servusdei2018/wt/internal/ui"
	wtfs "github.com/servusdei2018/wt/pkg/fs"

	"github.com/spf13/cobra"
)

func (a *app) sizeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "size",
		Short: "Report disk usage of all worktrees",
		Long: `Scans the .worktrees/ directory and reports total disk usage per worktree.

Heavy dependency directories (node_modules, vendor, .venv, target, etc.) are
called out specifically so you can identify bloated trees that are safe to
garbage-collect with 'wt done'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			worktreesDir := a.worktreesDir()

			commonGitDir, err := git.CommonGitDir(ctx, a.repoRoot)
			if err != nil {
				commonGitDir = filepath.Join(a.repoRoot, ".git")
			}
			cacheDir := filepath.Join(commonGitDir, "wt", "cache")

			reports, err := wtfs.ScanWithCache(worktreesDir, cacheDir, a.cfg.Storage.HeavyDirs)
			if err != nil {
				return fmt.Errorf("size scan: %w", err)
			}

			if len(reports) == 0 {
				fmt.Println(ui.StyleMuted.Render("No worktrees found in .worktrees/."))
				return nil
			}

			var grandTotal int64
			for _, r := range reports {
				grandTotal += r.TotalSize
				rel, _ := filepath.Rel(a.repoRoot, r.Path)

				header := fmt.Sprintf("%s  %s",
					ui.StyleBranch.Render(rel),
					ui.StyleMuted.Render("("+ui.FormatSize(r.TotalSize)+")"),
				)
				fmt.Println(ui.StyleBox.Render(header + heavyDirsBlock(r.HeavyDirs)))
			}

			fmt.Println(ui.StyleSuccessBox.Render(
				ui.StyleSuccess.Render("Total: ") + ui.StyleBranch.Render(ui.FormatSize(grandTotal)),
			))
			return nil
		},
	}
}

func heavyDirsBlock(dirs []wtfs.HeavyDir) string {
	if len(dirs) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("\n\n")
	sb.WriteString(ui.StyleWarning.Render("  Heavy directories:"))
	for _, d := range dirs {
		fmt.Fprintf(&sb, "\n    %s  %s", ui.StyleWarning.Render(d.RelPath), ui.StyleMuted.Render(ui.FormatSize(d.Size)))
	}
	return sb.String()
}
