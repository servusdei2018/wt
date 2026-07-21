package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/servusdei2018/wt/internal/storage"
	"github.com/servusdei2018/wt/internal/ui"
	"github.com/spf13/cobra"
)

func (a *app) dedupeCmd() *cobra.Command {
	var (
		dryRun         bool
		requireReflink bool
		verbose        bool
	)

	cmd := &cobra.Command{
		Use:   "dedupe",
		Short: "Retroactively deduplicate heavy build dependencies across worktrees",
		Long: `Scans all worktrees for duplicate dependency files (node_modules, .venv, target, etc.)
and replaces them with OS-native Copy-On-Write (reflink) clones to save disk space.

Hardlinks are not used to prevent inadvertent cross-worktree file mutations.
If reflink is unsupported on the underlying filesystem, deduplication falls back
to copy unless --require-reflink is specified.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := storage.DedupeOptions{
				RequireReflink: requireReflink || a.cfg.Storage.RequireReflink,
				DryRun:         dryRun,
				HeavyDirs:      a.cfg.Storage.HeavyDirs,
				Verbose:        verbose,
			}

			if dryRun {
				fmt.Println(ui.StyleMuted.Render("Scanning worktrees for duplicate dependency files (dry run)…"))
			} else {
				fmt.Println(ui.StyleMuted.Render("Deduplicating shared dependency stores across worktrees via CoW reflink…"))
			}

			report, err := storage.DeduplicateWorktrees(a.repoRoot, a.worktreesDir(), opts)
			if err != nil {
				return fmt.Errorf("deduplication failed: %w", err)
			}

			title := "Deduplication Complete"
			if dryRun {
				title = "Deduplication Dry Run Summary"
			}

			var sb strings.Builder
			sb.WriteString(ui.StyleSuccess.Render(title) + "\n")
			sb.WriteString(ui.StyleMuted.Render("  files deduplicated  ") + fmt.Sprintf("%d", report.FilesDeduplicated) + "\n")
			sb.WriteString(ui.StyleMuted.Render("  disk space saved    ") + ui.StyleBranch.Render(ui.FormatSize(report.BytesSaved)) + "\n")
			sb.WriteString(ui.StyleMuted.Render("  reflink operations  ") + fmt.Sprintf("%d", report.ReflinkCount) + "\n")
			sb.WriteString(ui.StyleMuted.Render("  scan duration       ") + fmt.Sprintf("%v", report.Duration.Round(100*time.Microsecond)))

			if len(report.DirSavings) > 0 {
				sb.WriteString("\n\n" + ui.StyleHeader.Render("Savings Breakdown:"))
				for dirName, bytes := range report.DirSavings {
					fmt.Fprintf(&sb, "\n  %-18s %s", ui.StyleWarning.Render(dirName), ui.FormatSize(bytes))
				}
			}

			fmt.Println(ui.StyleSuccessBox.Render(sb.String()))
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview disk space savings without modifying files")
	cmd.Flags().BoolVar(&requireReflink, "require-reflink", false, "Fail if Copy-On-Write reflink is unsupported on filesystem")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose deduplication output")
	return cmd
}
