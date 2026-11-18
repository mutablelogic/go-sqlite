package sqlite

import (
	"context"
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// userVersion is the version of the database's schema, which SQLite keeps,
// as user_version, for its user, and doesn't use itself
type userVersion struct {
	Version int64
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Version returns the version of the database's schema, which is 0 until
// it's set, so a schema can be created, and then migrated, as it changes.
func Version(ctx context.Context, conn Conn) (int64, error) {
	var version userVersion
	if err := conn.With(TraceSpanNameArg, "sqlite.version").Get(ctx, &version, version); err != nil {
		return 0, err
	}
	return version.Version, nil
}

// SetVersion sets the version of the database's schema, such as within the
// transaction which migrates it.
func SetVersion(ctx context.Context, conn Conn, version int64) error {
	if version < 0 || version > 1<<31-1 {
		return ErrBadParameter.Withf("version: %d", version)
	}
	return conn.With(TraceSpanNameArg, "sqlite.set_version").Exec(ctx, fmt.Sprintf("PRAGMA user_version = %d", version))
}

////////////////////////////////////////////////////////////////////////////////
// SELECTOR AND READER

func (userVersion) Select(*Bind, Op) (string, error) {
	return "PRAGMA user_version", nil
}

func (v *userVersion) Scan(row Row) error {
	return row.Scan(&v.Version)
}
