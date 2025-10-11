//go:build !windows
// +build !windows

package main

import (
	"os"
	"syscall"
)

func getFileID(fi os.FileInfo) uint64 {
	if stat, ok := fi.Sys().(*syscall.Stat_t); ok {
		return uint64(stat.Ino)
	}
	return 0
}
