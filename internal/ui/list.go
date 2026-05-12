package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/servusdei2018/wt/internal/git"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EnrichedWorktree combines raw worktree data with computed display fields.
type EnrichedWorktree struct {
	git.Worktree
	Age    time.Duration
	Status git.Status
	Index  int
}

// listModel is a bubbletea model that renders the enriched worktree table.
type listModel struct {
	items       []EnrichedWorktree
	width       int
	hasPrunable bool
	// prune prompt state
	pruneAsked bool
	pruneDone  bool
	PruneYes   bool
}

func newListModel(items []EnrichedWorktree) listModel {
	m := listModel{items: items}
	for _, it := range items {
		if it.IsPrunable {
			m.hasPrunable = true
			break
		}
	}
	return m
}

func (m listModel) Init() tea.Cmd { return nil }

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		if !m.hasPrunable {
			return m, tea.Quit
		}
		if !m.pruneAsked {
			m.pruneAsked = true
			return m, nil
		}
		switch msg.String() {
		case "y", "Y":
			m.PruneYes = true
			m.pruneDone = true
		default:
			m.pruneDone = true
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m listModel) View() string {
	var sb strings.Builder

	// Derive column widths dynamically.
	maxBranch := len("Branch")
	for _, it := range m.items {
		if l := len(it.Branch); l > maxBranch {
			maxBranch = l
		}
	}

	// Header.
	sb.WriteString("\n" + StyleHeader.Render("  Worktrees") + "\n\n")

	colFmt := fmt.Sprintf("  %%3s  %%-%ds  %%-7s  %%-8s  %%-6s  %%s", maxBranch)
	header := fmt.Sprintf(colFmt, "#", "Branch", "HEAD", "Age", "Status", "Path")
	sb.WriteString(StyleMuted.Render(header) + "\n")

	sep := "  " + strings.Repeat("─", 3) + "  " +
		strings.Repeat("─", maxBranch) + "  " +
		strings.Repeat("─", 7) + "  " +
		strings.Repeat("─", 8) + "  " +
		strings.Repeat("─", 6) + "  " +
		strings.Repeat("─", 20)
	sb.WriteString(StyleMuted.Render(sep) + "\n")

	// Rows.
	for _, it := range m.items {
		index := fmt.Sprintf("[%d]", it.Index)
		branch := it.Branch
		head := it.HEAD
		age := FormatAge(it.Age)
		if it.IsPrunable {
			age = StyleDanger.Render("prunable")
		}

		var statusStr string
		switch {
		case it.IsPrunable:
			statusStr = StyleDanger.Render(IndicatorPrunable + " broken")
		case it.Status.IsDirty:
			statusStr = StyleWarning.Render(IndicatorDirty + " dirty ")
		default:
			statusStr = StyleSuccess.Render(IndicatorClean + " clean ")
		}

		var branchStr, indexStr string
		if it.IsMain {
			indexStr = StyleCurrent.Render(IndicatorCurrent + " " + index)
			branchStr = StyleCurrent.Render(branch)
		} else {
			indexStr = StyleMuted.Render("  " + index)
			branchStr = StyleBranch.Render(branch)
		}

		// Pad branch to column width (lipgloss renders ANSI so we manually pad).
		branchPadded := branchStr + strings.Repeat(" ", max(0, maxBranch-len(branch)))

		row := fmt.Sprintf("  %s  %s  %-7s  %-8s  %s  %s",
			indexStr, branchPadded, head, age, statusStr, StyleMuted.Render(it.Path))
		sb.WriteString(row + "\n")
	}

	// Prune prompt.
	if m.hasPrunable && !m.pruneDone {
		msg := StyleWarningBox.Render(
			StyleWarning.Render("Orphaned worktrees detected.") +
				"\nRun " + StyleBranch.Render("git worktree prune") + "? [y/N] ",
		)
		sb.WriteString("\n" + msg + "\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

// RunList renders the worktree table and, if prunable entries exist, asks the
// user whether to prune. Returns true if the user confirmed pruning.
func RunList(items []EnrichedWorktree) (bool, error) {
	m := newListModel(items)
	p := tea.NewProgram(m, tea.WithOutput(lipgloss.DefaultRenderer().Output()))
	final, err := p.Run()
	if err != nil {
		return false, err
	}
	if lm, ok := final.(listModel); ok {
		return lm.PruneYes, nil
	}
	return false, nil
}
