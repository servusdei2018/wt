package hook

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/servusdei2018/wt/internal/config"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		cfg      *config.Config
		wantHook *Hook
	}{
		{
			name: "custom hook",
			cfg: &config.Config{
				Hooks: config.HooksConfig{
					PostCreate: "echo hello",
				},
			},
			wantHook: &Hook{Name: "custom", Command: []string{"echo hello"}},
		},
		{
			name:     "npm fallback",
			files:    []string{"package.json"},
			wantHook: &Hook{Name: "npm install", Command: []string{"npm", "install"}},
		},
		{
			name:     "pnpm priority",
			files:    []string{"package.json", "pnpm-lock.yaml"},
			wantHook: &Hook{Name: "pnpm install", Command: []string{"pnpm", "install"}},
		},
		{
			name:     "bun priority",
			files:    []string{"package.json", "bun.lockb"},
			wantHook: &Hook{Name: "bun install", Command: []string{"bun", "install"}},
		},
		{
			name:     "bun lock (new format) priority",
			files:    []string{"package.json", "bun.lock"},
			wantHook: &Hook{Name: "bun install", Command: []string{"bun", "install"}},
		},
		{
			name:     "yarn priority",
			files:    []string{"package.json", "yarn.lock"},
			wantHook: &Hook{Name: "yarn install", Command: []string{"yarn", "install"}},
		},
		{
			name:     "go detection",
			files:    []string{"go.mod"},
			wantHook: &Hook{Name: "go mod download", Command: []string{"go", "mod", "download"}},
		},
		{
			name:     "pip requirements",
			files:    []string{"requirements.txt"},
			wantHook: &Hook{Name: "pip install", Command: []string{"pip", "install", "-r", "requirements.txt"}},
		},
		{
			name:     "pip pyproject",
			files:    []string{"pyproject.toml"},
			wantHook: &Hook{Name: "pip install", Command: []string{"pip", "install", "-e", "."}},
		},
		{
			name:     "uv priority",
			files:    []string{"pyproject.toml", "uv.lock"},
			wantHook: &Hook{Name: "uv sync", Command: []string{"uv", "sync"}},
		},
		{
			name:     "deno detection",
			files:    []string{"deno.json"},
			wantHook: &Hook{Name: "deno install", Command: []string{"deno", "install"}},
		},
		{
			name:     "deno jsonc detection",
			files:    []string{"deno.jsonc"},
			wantHook: &Hook{Name: "deno install", Command: []string{"deno", "install"}},
		},
		{
			name:     "no hook",
			files:    []string{"README.md"},
			wantHook: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			for _, f := range tt.files {
				if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("{}"), 0644); err != nil {
					t.Fatal(err)
				}
			}

			got, err := Detect(tmpDir, tt.cfg)
			if err != nil {
				t.Errorf("Detect() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.wantHook) {
				t.Errorf("Detect() got = %v, want %v", got, tt.wantHook)
			}
		})
	}
}
