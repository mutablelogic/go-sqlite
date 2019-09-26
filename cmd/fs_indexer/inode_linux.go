package main

import (
	"os"
	"syscall"
)

func inodeForInfo(info os.FileInfo) int64 {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return int64(stat.Ino)
	} else {
		return 0
	}
}
