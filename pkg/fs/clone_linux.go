//go:build linux

package fs

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// cloneFileOS performs a Linux FICLONE ioctl call to clone src to dst.
func cloneFileOS(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src: %w", err)
	}
	defer func() {
		_ = srcFile.Close()
	}()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("stat src: %w", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode().Perm())
	if err != nil {
		return fmt.Errorf("open dst: %w", err)
	}
	defer func() {
		_ = dstFile.Close()
	}()

	if err := unix.IoctlFileClone(int(dstFile.Fd()), int(srcFile.Fd())); err != nil {
		return err
	}
	return nil
}
