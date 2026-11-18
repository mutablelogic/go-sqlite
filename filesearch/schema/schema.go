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
	Version = 1

	// ListLimit is the most objects which are listed, or searched, at once
	ListLimit = 100
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
		if version < 1 {
			for _, key := range objects.Keys() {
				if err := conn.With(sqlite.TraceSpanNameArg, key).Exec(ctx, objects.Query(key)); err != nil {
					return err
				}
			}
			if err := sqlite.SetVersion(ctx, conn, 1); err != nil {
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
