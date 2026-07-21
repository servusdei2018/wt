package editor

import (
	"context"
	"testing"

	"github.com/servusdei2018/wt/internal/config"
)

func TestResolveConfigOverride(t *testing.T) {
	cfg := &config.Config{
		Editor: config.EditorConfig{Command: "custom-editor"},
	}
	got := Resolve(cfg)
	if got != "custom-editor" {
		t.Errorf("Resolve() = %q, want %q", got, "custom-editor")
	}
}

func TestResolveEnvVar(t *testing.T) {
	t.Setenv("EDITOR", "nano")
	t.Setenv("VISUAL", "gvim")

	cfg := &config.Config{}
	got := Resolve(cfg)
	if got != "nano" {
		t.Errorf("Resolve() = %q, want %q", got, "nano")
	}
}

func TestResolveVisualEnvVar(t *testing.T) {
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "gvim")

	cfg := &config.Config{}
	got := Resolve(cfg)
	if got != "gvim" {
		t.Errorf("Resolve() = %q, want %q", got, "gvim")
	}
}

func TestOpenWithFlags(t *testing.T) {
	ctx := context.Background()
	// Test if Open splits arguments correctly. We can use a simple echo command.
	// Since echo exits 0, this should succeed.
	err := Open(ctx, "echo -n", "test-path")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
