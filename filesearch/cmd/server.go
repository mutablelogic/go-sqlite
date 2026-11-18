//go:build !client

// Package cmd is the commands, and flags, of a file search manager, which a
// program composes, and which imports the handlers of go-metadata which it
// reads files with.
package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"

	// Packages
	tmdb "github.com/mutablelogic/go-metadata/pkg/tmdb"
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

	// The names of files and folders which aren't indexed
	Exclude []string `name:"exclude" help:"Regular expressions, of the names of files and folders which aren't indexed, such as '\\.bak$'"`
	// The movies and episodes on TMDB which videos are
	TMDBKey string `name:"tmdb-key" env:"TMDB_TOKEN" help:"The API Read Access Token of TMDB, with which videos are looked up on TMDB, of their names, when they're indexed. Set it in the environment, rather than as a flag, to keep it out of the shell's history."`

	DefaultExclude bool `name:"default-exclude" help:"Don't index hidden files and folders, lock files, backups, temporary files, and the files of systems, and dependencies" default:"true" negatable:""`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// pseudoVersion matches a version of a commit, rather than a release
var pseudoVersion = regexp.MustCompile(`-(?:[0-9A-Za-z.]+\.)?\d{14}-[0-9a-f]{12}$`)

const (
	// The modules of filesearch and go-metadata, of whose versions what reads
	// files is
	filesearchModule = "github.com/mutablelogic/go-sqlite/filesearch"
	metadataModule   = "github.com/mutablelogic/go-metadata"

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
	opts = append([]filesearch.Opt{
		filesearch.WithReader(flags.reader()),
		filesearch.WithTracer(ctx.Tracer()),
		filesearch.WithMaxReaders(flags.Readers),
		filesearch.WithExclude(flags.Exclude...),
	}, opts...)
	if !flags.DefaultExclude {
		opts = append(opts, filesearch.WithoutDefaultExclude())
	}
	if flags.TMDBKey != "" {
		client, err := tmdb.New(flags.TMDBKey)
		if err != nil {
			return err
		}
		opts = append(opts, filesearch.WithTMDB(client))
	}
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

// reader returns what reads files, of the versions of filesearch and
// go-metadata, the program's build tags, such as media, of the handlers it
// imports, and whether videos are looked up on TMDB, so a volume which
// another read is indexed again. A commit of a build which isn't released
// isn't part of it, so building it again doesn't index everything again.
func (flags Flags) reader() string {
	parts := []string{"filesearch"}
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Path == filesearchModule {
			parts = append(parts, release(info.Main.Version))
		}
		for _, dep := range info.Deps {
			switch dep.Path {
			case filesearchModule:
				parts = append(parts, release(dep.Version))
			case metadataModule:
				parts = append(parts, "go-metadata "+release(dep.Version))
			}
		}
		for _, setting := range info.Settings {
			if setting.Key == "-tags" && setting.Value != "" {
				parts = append(parts, "tags "+setting.Value)
			}
		}
	}
	if flags.TMDBKey != "" {
		parts = append(parts, "tmdb")
	}
	return strings.Join(parts, " ")
}

// release returns a module's version, or "devel" when it's built from a
// commit which isn't released, such as v0.0.0-20261118123708-ead0a2da1cc5+dirty
func release(version string) string {
	if version == "" || version == "(devel)" || pseudoVersion.MatchString(version) || strings.Contains(version, "+") {
		return "devel"
	}
	return version
}

// path returns the path of the database, which is in the user's cache
// folder, which is created, and only the user can read, when --db isn't set
func (flags Flags) path() (string, error) {
	if flags.DB != "" {
		return flags.DB, nil
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	// Its folder is the user's alone, as the database has what's read of
	// their files
	folder := filepath.Join(cache, "filesearch")
	if err := os.MkdirAll(folder, 0o700); err != nil {
		return "", err
	}
	if err := os.Chmod(folder, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(folder, DefaultName), nil
}
