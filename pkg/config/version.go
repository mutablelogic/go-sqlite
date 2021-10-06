package config

import (
	"fmt"
	"io"
	"runtime"

	// Packages
	sqlite3 "github.com/mutablelogic/go-sqlite/sys/sqlite3/"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	GitSource   string
	GitTag      string
	GitBranch   string
	GitHash     string
	GoBuildTime string
)

func PrintVersion(w io.Writer) {
	if GitSource != "" {
		fmt.Fprintf(w, "  URL: https://%v\n", GitSource)
	}
	if GitTag != "" || GitBranch != "" {
		fmt.Fprintf(w, "  Version: %v (branch: %q hash:%q)\n", GitTag, GitBranch, GitHash)
	}
	if GoBuildTime != "" {
		fmt.Fprintf(w, "  Build Time: %v\n", GoBuildTime)
	}
	fmt.Fprintf(w, "  go: %v (%v/%v)\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(w, "  sqlite3: %v\n", sqlite3.Version())
}
