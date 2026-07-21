package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/servusdei2018/wt/internal/git"
)

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		got := FormatSize(tt.bytes)
		if got != tt.want {
			t.Errorf("FormatSize(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		dur  time.Duration
		want string
	}{
		{30 * time.Minute, "just now"},
		{5 * time.Hour, "5h"},
		{3 * 24 * time.Hour, "3d"},
		{2 * 7 * 24 * time.Hour, "2w"},
	}

	for _, tt := range tests {
		got := FormatAge(tt.dur)
		if got != tt.want {
			t.Errorf("FormatAge(%v) = %q, want %q", tt.dur, got, tt.want)
		}
	}
}

func TestRenderList(t *testing.T) {
	items := []EnrichedWorktree{
		{
			Worktree: git.Worktree{
				Path:   "/repo",
				Branch: "main",
				HEAD:   "1a2b3c4",
				IsMain: true,
			},
			Index: 1,
			Age:   10 * time.Minute,
		},
		{
			Worktree: git.Worktree{
				Path:   "/repo/.worktrees/feature",
				Branch: "feature",
				HEAD:   "9f8e7d6",
			},
			Index: 2,
			Age:   2 * time.Hour,
			Status: git.Status{
				IsDirty: true,
			},
		},
	}

	output := RenderList(items)
	if !strings.Contains(output, "Worktrees") {
		t.Error("RenderList missing Worktrees header")
	}
	if !strings.Contains(output, "main") || !strings.Contains(output, "feature") {
		t.Error("RenderList missing branch names")
	}
	if !strings.Contains(output, "1a2b3c4") || !strings.Contains(output, "9f8e7d6") {
		t.Error("RenderList missing HEAD commit SHAs")
	}
}
