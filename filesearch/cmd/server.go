//go:build !client

// Package cmd is the commands, and flags, of a file search manager, which a
// program composes, and which imports the handlers of go-metadata which it
// reads files with.
package cmd

import (
	"errors"
	"os"
	"path/filepath"

	// Packages
	server "github.com/mutablelogic/go-server"
	servercmd "github.com/mutablelogic/go-server/pkg/cmd"
	filesearch "github.com/mutablelogic/go-sqlite/filesearch"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Flags are the flags of a file search manager
type Flags struct {
	DB      string `name:"db" help:"The database, which is created when it doesn't exist, or :memory: (default: filesearch.db, in the user's cache folder)"`
	Readers int    `name:"readers" help:"The most connections which read the database at once, such as to search it" default:"4"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// DefaultName is the name of the database, in the user's cache folder,
	// when it isn't set
	DefaultName = "filesearch.db"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// WithManager creates a file search manager, of the database of the flags,
// with the server's tracer, and then the options, calls the function with
// it, and closes it, once the function returns, returning any error. The
// server is where its HTTP handlers are registered, when there is one, once
// there are some.
func (flags Flags) WithManager(ctx server.Cmd, srv *servercmd.RunServer, fn func(*filesearch.Manager) error, opts ...filesearch.Opt) (err error) {
	path, err := flags.path()
	if err != nil {
		return err
	}
	opts = append([]filesearch.Opt{filesearch.WithTracer(ctx.Tracer()), filesearch.WithMaxReaders(flags.Readers)}, opts...)
	manager, err := filesearch.New(ctx.Context(), path, opts...)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, manager.Close()) }()
	ctx.Logger().DebugContext(ctx.Context(), "file search", "path", path)

	// Call the function with the manager
	return fn(manager)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// path returns the path of the database, which is in the user's cache
// folder, which is created, when --db isn't set
func (flags Flags) path() (string, error) {
	if flags.DB != "" {
		return flags.DB, nil
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	folder := filepath.Join(cache, "filesearch")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(folder, DefaultName), nil
}
