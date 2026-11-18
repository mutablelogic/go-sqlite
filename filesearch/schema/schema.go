// Package schema is the schema of the file indexer: its tables, its queries,
// and the types which are read from, and written to, them.
//
// The schema needs SQLite's full-text search, of fts5, such as:
//
//	import _ "github.com/mutablelogic/go-sqlite/ncruces/fts5"
package schema

import (
	"context"
	_ "embed"
	"strings"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

//go:embed objects.sql
var Objects string

//go:embed queries.sql
var Queries string

const (
	// Version is the version of the schema, of the database's user_version
	Version = 2

	// MinPathLength is the fewest characters of a part of a path, a folder
	// or a name, which is searched
	MinPathLength = 3

	// ListLimit is the most objects which are listed, or searched, at once
	ListLimit = 100

	// DefaultCutoff is the lowest rank of a search, between 0 and 1, unless
	// it's set, so the objects which match it least aren't returned
	DefaultCutoff = 0.15
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Create creates the schema, in a transaction, when the database hasn't
// one, and sets its version. A database whose schema is newer is an error.
func Create(ctx context.Context, conn sqlite.Conn) error {
	objects, err := sqlite.NewQueries(strings.NewReader(Objects))
	if err != nil {
		return err
	}
	return conn.Tx(ctx, func(conn sqlite.Conn) error {
		version, err := sqlite.Version(ctx, conn)
		if err != nil {
			return err
		}
		if version > Version {
			return sqlite.ErrNotImplemented.Withf("schema version %d is newer than %d", version, Version)
		}

		// Version 1's objects are deleted, so they're indexed again, when their
		// volumes are, and then its text, which hasn't the parts of paths, is
		// dropped, as a virtual table's columns can't be added, and its text
		// isn't kept, so it's created again
		if version == 1 {
			for _, sql := range []string{`DELETE FROM object`, `DROP TABLE IF EXISTS object_text`} {
				if err := conn.With(sqlite.TraceSpanNameArg, "filesearch.migrate_2").Exec(ctx, sql); err != nil {
					return err
				}
			}
		}

		// The schema, of its version, which creates what isn't there
		if version < Version {
			for _, key := range objects.Keys() {
				if err := conn.With(sqlite.TraceSpanNameArg, key).Exec(ctx, objects.Query(key)); err != nil {
					return err
				}
			}
			if err := sqlite.SetVersion(ctx, conn, Version); err != nil {
				return err
			}
		}
		return nil
	})
}

// NewQueries returns the queries, which a connection is bound to, with
// WithQueries, so the types' selectors and writers use them
func NewQueries() (*sqlite.Queries, error) {
	return sqlite.NewQueries(strings.NewReader(Queries))
}
