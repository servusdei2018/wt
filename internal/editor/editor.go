// Package editor resolves and launches the user's preferred text editor.
package editor

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/servusdei2018/wt/internal/config"
)

// Resolve returns the editor command to use, in priority order:
//
//  1. cfg.Editor.Command (.wt.toml override)
//  2. $EDITOR environment variable
//  3. $VISUAL environment variable
//  4. "code"  (VS Code, if on PATH)
//  5. "nvim"  (if on PATH)
//  6. "vim"   (unconditional last resort)
func Resolve(cfg *config.Config) string {
	if cfg != nil && cfg.Editor.Command != "" {
		return cfg.Editor.Command
	}
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	if v := os.Getenv("VISUAL"); v != "" {
		return v
	}
	for _, candidate := range []string{"code", "nvim", "vim"} {
		if _, err := exec.LookPath(candidate); err == nil {
			return candidate
		}
	}
	return "vim"
}

// Open launches editorCmd with path as the sole argument, inheriting
// the current process stdio so the editor gets a proper terminal.
func Open(editorCmd, path string) error {
	cmd := exec.Command(editorCmd, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor %q: %w", editorCmd, err)
	}
	return nil
}
