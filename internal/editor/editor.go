// Package editor resolves and launches the user's preferred text editor.
package editor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

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
func Open(ctx context.Context, editorCmd, path string) error {
	parts := strings.Fields(editorCmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty editor command")
	}
	args := append(parts[1:], path)
	cmd := exec.CommandContext(ctx, parts[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor %q: %w", editorCmd, err)
	}
	return nil
}
