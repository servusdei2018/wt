package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/servusdei2018/wt/internal/git"
)

// EnrichedWorktree combines raw worktree data with computed display fields.
type EnrichedWorktree struct {
	git.Worktree
	Age    time.Duration
	Status git.Status
	Index  int
}

// RenderList formats the worktree table into a string.
func RenderList(items []EnrichedWorktree) string {
	var sb strings.Builder

	// Derive column widths dynamically
	maxBranch := lipgloss.Width("Branch")
	for _, it := range items {
		if w := lipgloss.Width(it.Branch); w > maxBranch {
			maxBranch = w
		}
	}

	// Header.
	sb.WriteString("\n  " + StyleHeader.Render("Worktrees") + "\n\n")

	// Index col header is 5 visible chars ("  #  ") to align with "* [1]" or "  [1]"
	colFmt := fmt.Sprintf("  %%-5s  %%-%ds  %%-7s  %%-8s  %%-8s  %%s", maxBranch)
	header := fmt.Sprintf(colFmt, "#", "Branch", "HEAD", "Age", "Status", "Path")
	sb.WriteString(StyleMuted.Render(header))
	sb.WriteString("\n")

	sep := "  " + strings.Repeat("─", 5) + "  " +
		strings.Repeat("─", maxBranch) + "  " +
		strings.Repeat("─", 7) + "  " +
		strings.Repeat("─", 8) + "  " +
		strings.Repeat("─", 8) + "  " +
		strings.Repeat("─", 20)
	sb.WriteString(StyleMuted.Render(sep))
	sb.WriteString("\n")

	// Rows.
	for _, it := range items {
		index := fmt.Sprintf("[%d]", it.Index)
		branch := it.Branch
		head := it.HEAD
		age := FormatAge(it.Age)

		var statusStr string
		switch {
		case it.IsPrunable:
			statusStr = StyleDanger.Render("broken")
			age = StyleDanger.Render("prunable")
		case it.Status.IsDirty:
			statusStr = StyleWarning.Render("dirty")
		default:
			statusStr = StyleSuccess.Render("clean")
		}

		var branchStr, indexStr string
		if it.IsMain {
			indexStr = StyleCurrent.Render("* " + index)
			branchStr = StyleCurrent.Render(branch)
		} else {
			indexStr = StyleMuted.Render("  " + index)
			branchStr = StyleBranch.Render(branch)
		}

		// Pad columns taking ANSI / display width into account
		branchPadded := branchStr + strings.Repeat(" ", max(0, maxBranch-lipgloss.Width(branch)))
		headPadded := head + strings.Repeat(" ", max(0, 7-lipgloss.Width(head)))
		agePadded := age + strings.Repeat(" ", max(0, 8-lipgloss.Width(age)))
		statusPadded := statusStr + strings.Repeat(" ", max(0, 8-lipgloss.Width(statusStr)))

		row := fmt.Sprintf("  %s  %s  %s  %s  %s  %s",
			indexStr, branchPadded, headPadded, agePadded, statusPadded, StyleMuted.Render(it.Path))
		sb.WriteString(row)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

// RunList renders the worktree table and, if prunable entries exist, asks the
// user whether to prune. Returns true if the user confirmed pruning.
func RunList(items []EnrichedWorktree) (bool, error) {
	fmt.Print(RenderList(items))

	hasPrunable := false
	for _, it := range items {
		if it.IsPrunable {
			hasPrunable = true
			break
		}
	}

	if hasPrunable {
		return Confirm("Orphaned worktrees detected.", "Run 'git worktree prune'?")
	}
	return false, nil
}
