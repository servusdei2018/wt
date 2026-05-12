// Package ui contains terminal UI components built on the Charmbracelet stack.
package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette.
	colorPrimary = lipgloss.Color("#7C3AED") // violet
	colorMuted   = lipgloss.Color("#6B7280") // cool gray
	colorSuccess = lipgloss.Color("#10B981") // emerald
	colorWarning = lipgloss.Color("#F59E0B") // amber
	colorDanger  = lipgloss.Color("#EF4444") // red
	colorAccent  = lipgloss.Color("#38BDF8") // sky blue — current worktree

	// Status indicator glyphs.
	IndicatorDirty    = "✖"
	IndicatorClean    = "✔"
	IndicatorCurrent  = "▶"
	IndicatorPrunable = "!"

	// Base styles — composable building blocks.
	StyleBranch  = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	StyleMuted   = lipgloss.NewStyle().Foreground(colorMuted)
	StyleSuccess = lipgloss.NewStyle().Foreground(colorSuccess)
	StyleWarning = lipgloss.NewStyle().Foreground(colorWarning)
	StyleDanger  = lipgloss.NewStyle().Foreground(colorDanger)
	StyleCurrent = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	StyleHeader  = lipgloss.NewStyle().Bold(true).Underline(true)

	// Box styles for output messages.
	StyleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1)

	StyleSuccessBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSuccess).
			Padding(0, 1)

	StyleWarningBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorWarning).
			Padding(0, 1)

	StyleDangerBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDanger).
			Padding(0, 1)
)

// FormatSize formats a byte count into a human-readable string (KB, MB, GB…).
func FormatSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// FormatAge converts a duration into a compact human string ("2h", "3d", "1w").
func FormatAge(d interface{ Hours() float64 }) string {
	h := d.Hours()
	switch {
	case h < 1:
		return "just now"
	case h < 24:
		return fmt.Sprintf("%.0fh", h)
	case h < 24*7:
		return fmt.Sprintf("%.0fd", h/24)
	case h < 24*30:
		return fmt.Sprintf("%.0fw", h/(24*7))
	default:
		return fmt.Sprintf("%.0fmo", h/(24*30))
	}
}
