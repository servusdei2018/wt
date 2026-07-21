// Package seed handles copying and interpolating local configuration & secret files
// into new worktrees upon creation.
package seed

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/servusdei2018/wt/internal/config"
)

var slugRegexp = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// TemplateData contains variables available for interpolation in seed files.
type TemplateData struct {
	Branch       string
	BranchSlug   string
	Port         int
	WorktreePath string
	WorktreeName string
	RepoRoot     string
}

// SeedReport tracks the result of seeding operations.
type SeedReport struct {
	FilesCopied []string
}

// Count returns the number of files copied.
func (r *SeedReport) Count() int {
	return len(r.FilesCopied)
}

// ComputePort returns a deterministic port number in the range 3000-8999
// based on the FNV-1a hash of the branch name.
func ComputePort(branch string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(branch))
	return 3000 + int(h.Sum32()%6000)
}

// Slugify sanitizes a branch name into a web/hostname friendly slug.
func Slugify(branch string) string {
	slug := slugRegexp.ReplaceAllString(branch, "-")
	return strings.Trim(slug, "-")
}

// SeedFiles copies uncommitted local config files listed in cfg.Include from repoRoot
// to targetWorktree, interpolating template variables for text files.
func SeedFiles(repoRoot, targetWorktree, branch string, cfg config.SeedConfig) (*SeedReport, error) {
	report := &SeedReport{}
	if len(cfg.Include) == 0 {
		return report, nil
	}

	data := TemplateData{
		Branch:       branch,
		BranchSlug:   Slugify(branch),
		Port:         ComputePort(branch),
		WorktreePath: targetWorktree,
		WorktreeName: filepath.Base(targetWorktree),
		RepoRoot:     repoRoot,
	}

	for _, relPath := range cfg.Include {
		if !filepath.IsLocal(relPath) {
			return nil, fmt.Errorf("seed file path outside repository: %q", relPath)
		}
		cleanRel := filepath.Clean(relPath)

		srcPath := filepath.Join(repoRoot, cleanRel)
		srcInfo, err := os.Stat(srcPath)
		if errors.Is(err, os.ErrNotExist) {
			// Skip missing optional seed files
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("stat seed file %q: %w", srcPath, err)
		}
		if srcInfo.IsDir() {
			continue
		}

		destPath := filepath.Join(targetWorktree, cleanRel)
		if _, err := os.Stat(destPath); err == nil {
			// Skip if already present in worktree (e.g. tracked file)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return nil, fmt.Errorf("mkdir seed dir %q: %w", filepath.Dir(destPath), err)
		}

		content, err := os.ReadFile(srcPath)
		if err != nil {
			return nil, fmt.Errorf("read seed file %q: %w", srcPath, err)
		}

		var finalBytes []byte
		if isBinary(content) {
			finalBytes = content
		} else {
			var err error
			finalBytes, err = interpolate(cleanRel, content, data)
			if err != nil {
				return nil, fmt.Errorf("seed file %q interpolation error: %w", cleanRel, err)
			}
		}

		if err := os.WriteFile(destPath, finalBytes, srcInfo.Mode().Perm()); err != nil {
			return nil, fmt.Errorf("write seed file %q: %w", destPath, err)
		}

		report.FilesCopied = append(report.FilesCopied, cleanRel)
	}

	return report, nil
}

func isBinary(data []byte) bool {
	checkLen := min(len(data), 512)
	return bytes.IndexByte(data[:checkLen], 0) != -1
}

func interpolate(name string, content []byte, data TemplateData) ([]byte, error) {
	tmpl, err := template.New(name).Option("missingkey=error").Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute: %w", err)
	}

	return buf.Bytes(), nil
}
