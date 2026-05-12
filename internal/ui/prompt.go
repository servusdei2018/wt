package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

// DirtyAction represents the user's chosen handling for a dirty worktree.
type DirtyAction string

const (
	ActionStash       DirtyAction = "stash"
	ActionPush        DirtyAction = "push"
	ActionForceDelete DirtyAction = "force-delete"
	ActionAbort       DirtyAction = "abort"
)

// AskDirtyAction presents an interactive select prompt asking the user how to
// handle uncommitted changes before tearing down a worktree.
func AskDirtyAction(branch string) (DirtyAction, error) {
	var action DirtyAction
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[DirtyAction]().
				Title(fmt.Sprintf("Branch %q has uncommitted changes.", branch)).
				Description("Choose how to proceed before deletion:").
				Options(
					huh.NewOption("Stash changes and continue", ActionStash),
					huh.NewOption("Push branch to remote first", ActionPush),
					huh.NewOption("Force delete (discard all changes)", ActionForceDelete),
					huh.NewOption("Abort", ActionAbort),
				).
				Value(&action),
		),
	).Run()
	return action, err
}

// Confirm presents a yes/no confirmation prompt.
// Returns true if the user confirmed.
func Confirm(title, description string) (bool, error) {
	var ok bool
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Description(description).
				Affirmative("Yes").
				Negative("No").
				Value(&ok),
		),
	).Run()
	return ok, err
}
