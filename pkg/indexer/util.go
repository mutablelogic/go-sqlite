package indexer

import (
	"os"
	"path/filepath"
)

const (
	pathSeparator = string(os.PathSeparator)
)

func stringSliceContains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func pathToParent(path string) string {
	parent := filepath.Dir(path)
	if parent == "." {
		return pathSeparator
	} else {
		return parent
	}
}

func boolToInt64(v bool) int64 {
	if v {
		return 1
	}
	return 0
}

func uintMin(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}
