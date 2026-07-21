//go:build windows

package fs

import (
	"os"
)

func getFileStat(info os.FileInfo) (uint64, int64) {
	return 0, info.ModTime().UnixNano()
}
