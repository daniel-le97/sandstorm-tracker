//go:build windows
// +build windows

package tail

import (
	"os"
)

// getFileID returns 0 on Windows to avoid false positives; only use size shrinkage for rotation detection
func getFileID(fi os.FileInfo) uint64 {
	return 0
}
