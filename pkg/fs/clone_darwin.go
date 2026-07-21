//go:build darwin

package fs

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// cloneFileOS performs a macOS APFS clonefile call to clone src to dst.
func cloneFileOS(src, dst string) error {
	if err := unix.Clonefile(src, dst, 0); err != nil {
		return fmt.Errorf("clonefile: %w", err)
	}
	return nil
}
