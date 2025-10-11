//go:build windows
// +build windows

package main

import (
	"os"
	"syscall"
)

func getFileID(fi os.FileInfo) uint64 {
	if stat, ok := fi.Sys().(*syscall.Win32FileAttributeData); ok {
		// FileIndex is not available directly, so fallback to file size and mod time
		return uint64(stat.FileSizeHigh)<<32 | uint64(stat.FileSizeLow) ^ uint64(stat.LastWriteTime.Nanoseconds())
	}
	return 0
}
