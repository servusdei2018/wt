//go:build !windows

package fs

import (
	"os"
	"syscall"
)

func getFileStat(info os.FileInfo) (uint64, int64) {
	mtime := info.ModTime().UnixNano()
	var inode uint64
	if sys, ok := info.Sys().(*syscall.Stat_t); ok {
		inode = sys.Ino
	}
	return inode, mtime
}
