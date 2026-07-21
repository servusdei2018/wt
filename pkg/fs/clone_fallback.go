//go:build !linux && !darwin

package fs

import "errors"

// ErrReflinkUnsupported is returned when CoW reflink cloning is not supported by the OS.
var ErrReflinkUnsupported = errors.New("reflink is not supported on this operating system")

func cloneFileOS(src, dst string) error {
	return ErrReflinkUnsupported
}
