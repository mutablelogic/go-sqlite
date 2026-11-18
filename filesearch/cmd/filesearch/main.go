package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	// Packages
	servercmd "github.com/mutablelogic/go-server/pkg/cmd"
	version "github.com/mutablelogic/go-server/pkg/version"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	description = "filesearch indexes folders, of their files' metadata, content and artwork, and searches them"

	// exitInterrupted is the status of a command which is interrupted, such
	// as with CTRL+C, as a shell's is
	exitInterrupted = 130
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func main() {
	if err := servercmd.Main(CLI{}, description, version.Version(), servercmd.WithoutHTTP()); errors.Is(err, context.Canceled) {
		// Interrupted, such as with CTRL+C, which the command said
		os.Exit(exitInterrupted)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
