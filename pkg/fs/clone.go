package fs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// CloneStrategy represents the method used to clone a file.
type CloneStrategy string

const (
	StrategyReflink CloneStrategy = "reflink"
	StrategyCopy    CloneStrategy = "copy"
)

// ErrReflinkFailed indicates that Copy-On-Write reflink is unsupported on the underlying filesystem.
var ErrReflinkFailed = errors.New("reflink cloning unsupported or failed on underlying filesystem")

// CloneOptions specifies settings for file and directory cloning.
type CloneOptions struct {
	RequireReflink bool // If true, error out if CoW reflink is unsupported
	PreserveAttrs  bool // Preserve permissions and modification times
}

// CloneStats summarizes disk cloning operations.
type CloneStats struct {
	FilesTotal     int64         `json:"files_total"`
	FilesReflinked int64         `json:"files_reflinked"`
	FilesCopied    int64         `json:"files_copied"`
	BytesTotal     int64         `json:"bytes_total"`
	BytesReflinked int64         `json:"bytes_reflinked"`
	BytesCopied    int64         `json:"bytes_copied"`
	Duration       time.Duration `json:"duration"`
}

// CloneFile clones src to dst using CoW reflink, falling back to standard file copy
// unless RequireReflink is set in opts.
func CloneFile(src, dst string, opts CloneOptions) (CloneStrategy, error) {
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return "", fmt.Errorf("stat src file: %w", err)
	}

	if srcInfo.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return "", fmt.Errorf("readlink: %w", err)
		}
		_ = os.Remove(dst)
		if err := os.Symlink(target, dst); err != nil {
			return "", fmt.Errorf("symlink: %w", err)
		}
		return StrategyCopy, nil
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", fmt.Errorf("mkdir dst dir: %w", err)
	}

	// Remove destination file if present to support macOS APFS clonefile (which fails if dst exists)
	_ = os.Remove(dst)

	// Attempt Copy-On-Write reflink
	err = cloneFileOS(src, dst)
	if err == nil {
		if opts.PreserveAttrs {
			_ = preserveAttributes(srcInfo, dst)
		}
		return StrategyReflink, nil
	}

	// Clean up dst if created during failed reflink attempt
	_ = os.Remove(dst)

	if opts.RequireReflink {
		return "", fmt.Errorf("%w: %v", ErrReflinkFailed, err)
	}

	// Fallback to standard copy
	if err := copyFile(src, dst, srcInfo, opts.PreserveAttrs); err != nil {
		return "", fmt.Errorf("copy fallback failed: %w", err)
	}

	return StrategyCopy, nil
}

// CloneDir recursively clones a directory tree from srcDir to dstDir.
func CloneDir(srcDir, dstDir string, opts CloneOptions) (*CloneStats, error) {
	start := time.Now()
	stats := &CloneStats{}

	err := filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dstDir, relPath)

		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return os.MkdirAll(targetPath, 0o755)
			}
			return os.MkdirAll(targetPath, info.Mode().Perm())
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		fileSize := info.Size()

		strat, err := CloneFile(path, targetPath, opts)
		if err != nil {
			return fmt.Errorf("failed cloning %s: %w", relPath, err)
		}

		stats.FilesTotal++
		stats.BytesTotal += fileSize

		if strat == StrategyReflink {
			stats.FilesReflinked++
			stats.BytesReflinked += fileSize
		} else {
			stats.FilesCopied++
			stats.BytesCopied += fileSize
		}

		return nil
	})

	stats.Duration = time.Since(start)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func copyFile(src, dst string, srcInfo os.FileInfo, preserveAttrs bool) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = in.Close()
	}()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode().Perm())
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(dst)
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(dst)
		return err
	}

	if preserveAttrs {
		_ = preserveAttributes(srcInfo, dst)
	}

	return nil
}

func preserveAttributes(info os.FileInfo, dst string) error {
	_ = os.Chmod(dst, info.Mode().Perm())
	_ = os.Chtimes(dst, time.Now(), info.ModTime())
	return nil
}
