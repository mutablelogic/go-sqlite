// Package sqlitetest is the tests which a Driver passes, so a driver can be
// replaced by another, which behaves the same.
package sqlitetest

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// user is a row of the users table
type user struct {
	ID    int64
	Name  string
	Email *string
}

// userList is a list of users, and their count
type userList struct {
	Count uint64
	Users []user
}

// userSelector selects a user by their identifier, or lists them
type userSelector struct {
	ID int64
}

// userMeta is what's inserted, or updated, of a user
type userMeta struct {
	Name  string
	Email *string
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var queries = `
-- users.create
CREATE TABLE IF NOT EXISTS users (
	id    INTEGER PRIMARY KEY,
	name  TEXT NOT NULL CHECK (name <> ''), -- not empty
	email TEXT UNIQUE
);

-- posts.create
CREATE TABLE IF NOT EXISTS posts (
	id   INTEGER PRIMARY KEY,
	user INTEGER NOT NULL REFERENCES users (id)
);

-- users.insert
INSERT INTO ${"table"} (name, email) VALUES (@name, @email) RETURNING id, name, email;

-- users.get
SELECT id, name, email FROM ${"table"} WHERE id = @id;

-- users.update
UPDATE ${"table"} SET name = @name WHERE id = @id RETURNING id, name, email;

-- users.delete
DELETE FROM ${"table"} WHERE id = @id RETURNING id, name, email;

-- users.list
SELECT id, name, email FROM ${"table"} ORDER BY id
`

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Run runs the tests of the driver, on databases in memory, and in files,
// in a temporary folder.
func Run(t *testing.T, driver sqlite.Driver) {
	t.Run("Open", func(t *testing.T) { testOpen(t, driver) })
	for _, kind := range []string{"Memory", "File"} {
		t.Run(kind, func(t *testing.T) {
			t.Run("CRUD", func(t *testing.T) { testCRUD(t, open(t, driver, kind)) })
			t.Run("List", func(t *testing.T) { testList(t, open(t, driver, kind)) })
			t.Run("Errors", func(t *testing.T) { testErrors(t, open(t, driver, kind)) })
			t.Run("Tx", func(t *testing.T) { testTx(t, open(t, driver, kind)) })
			t.Run("Version", func(t *testing.T) { testVersion(t, open(t, driver, kind)) })
			t.Run("Concurrent", func(t *testing.T) { testConcurrent(t, open(t, driver, kind)) })
		})
	}
	t.Run("Readers", func(t *testing.T) { testReaders(t, driver) })
}

////////////////////////////////////////////////////////////////////////////////
// TESTS

func testOpen(t *testing.T, driver sqlite.Driver) {
	ctx := t.Context()

	// A driver and a path are needed, and options are checked
	_, err := sqlite.Open(ctx, sqlite.Memory, nil)
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
	_, err = sqlite.Open(ctx, " ", driver)
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
	_, err = sqlite.Open(ctx, sqlite.Memory, driver, sqlite.WithMaxReaders(0))
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
	_, err = sqlite.Open(ctx, sqlite.Memory, driver, sqlite.WithPragma("a; DROP", "1"))
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)

	// A folder which doesn't exist can't be opened
	_, err = sqlite.Open(ctx, filepath.Join(t.TempDir(), "missing", "test.db"), driver)
	assert.Error(t, err)

	// A path with characters which are a URI's, such as ? and #, is its name
	path := filepath.Join(t.TempDir(), "a b?c#d%20.db")
	db, err := sqlite.Open(ctx, path, driver, sqlite.WithPragma("cache_size", "-4000"))
	require.NoError(t, err)
	assert.NoError(t, db.Ping(ctx))
	assert.NoError(t, db.Close())
	assert.FileExists(t, path)
}

func testCRUD(t *testing.T, db sqlite.Conn) {
	ctx := t.Context()
	email := "a@example.com"

	// Insert, which returns the row
	var a user
	require.NoError(t, db.Insert(ctx, &a, userMeta{Name: "a", Email: &email}))
	assert.NotZero(t, a.ID)
	assert.Equal(t, "a", a.Name)
	assert.Equal(t, &email, a.Email)

	// Get it, and one which isn't there
	var got user
	require.NoError(t, db.Get(ctx, &got, userSelector{ID: a.ID}))
	assert.Equal(t, a, got)
	assert.ErrorIs(t, db.Get(ctx, &got, userSelector{ID: a.ID + 1}), sqlite.ErrNotFound)

	// Update it
	var updated user
	require.NoError(t, db.Update(ctx, &updated, userSelector{ID: a.ID}, userMeta{Name: "b"}))
	assert.Equal(t, "b", updated.Name)
	assert.ErrorIs(t, db.Update(ctx, &updated, userSelector{ID: a.ID + 1}, userMeta{Name: "b"}), sqlite.ErrNotFound)

	// Delete it
	var deleted user
	require.NoError(t, db.Delete(ctx, &deleted, userSelector{ID: a.ID}))
	assert.Equal(t, updated, deleted)
	assert.ErrorIs(t, db.Get(ctx, &got, userSelector{ID: a.ID}), sqlite.ErrNotFound)
	assert.ErrorIs(t, db.Delete(ctx, &deleted, userSelector{ID: a.ID}), sqlite.ErrNotFound)

	// A value with a $ in it, and a quote, is kept as it is
	var b user
	require.NoError(t, db.Insert(ctx, &b, userMeta{Name: "it's a$b"}))
	assert.Equal(t, "it's a$b", b.Name)
	assert.Nil(t, b.Email)
}

func testList(t *testing.T, db sqlite.Conn) {
	ctx := t.Context()

	// No rows isn't an error
	var list userList
	require.NoError(t, db.List(ctx, &list, userSelector{}))
	assert.Zero(t, list.Count)
	assert.Empty(t, list.Users)

	// The count is of all of them, and the rows of the offset and limit
	for _, name := range []string{"a", "b", "c", "d", "e"} {
		require.NoError(t, db.Insert(ctx, &user{}, userMeta{Name: name}))
	}
	list = userList{}
	limit := uint64(2)
	require.NoError(t, db.List(ctx, &list, offsetLimit{sqlite.OffsetLimit{Offset: 1, Limit: &limit}}))
	assert.Equal(t, uint64(5), list.Count)
	if assert.Len(t, list.Users, 2) {
		assert.Equal(t, "b", list.Users[0].Name)
		assert.Equal(t, "c", list.Users[1].Name)
	}
}

func testErrors(t *testing.T, db sqlite.Conn) {
	ctx := t.Context()
	email := "a@example.com"
	var a user
	require.NoError(t, db.Insert(ctx, &a, userMeta{Name: "a", Email: &email}))

	// A unique column is a conflict
	err := db.Insert(ctx, &user{}, userMeta{Name: "b", Email: &email})
	assert.ErrorIs(t, err, sqlite.ErrDatabase)
	assert.ErrorIs(t, err, sqlite.ErrConflict)
	assert.ErrorIs(t, err, sqlite.ErrUniqueViolation)

	// A check, and a foreign key, are bad parameters
	err = db.Insert(ctx, &user{}, userMeta{Name: ""})
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
	assert.ErrorIs(t, err, sqlite.ErrCheckViolation)
	err = db.With("user", a.ID+1).Exec(ctx, `INSERT INTO posts (user) VALUES (@user)`)
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
	assert.ErrorIs(t, err, sqlite.ErrForeignKeyViolation)
	err = db.Exec(ctx, `INSERT INTO posts (user) VALUES (NULL)`)
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
	assert.ErrorIs(t, err, sqlite.ErrNotNullViolation)
	assert.NoError(t, db.With("user", a.ID).Exec(ctx, `INSERT INTO posts (user) VALUES (@user)`))

	// SQL which isn't is a database error
	assert.ErrorIs(t, db.Exec(ctx, `NOT SQL`), sqlite.ErrDatabase)

	// A selector's error is returned as it is
	assert.ErrorIs(t, db.Get(ctx, &user{}, userSelector{ID: -1}), sqlite.ErrBadParameter)
}

func testTx(t *testing.T, db sqlite.Conn) {
	ctx := t.Context()
	failed := errors.New("failed")

	// A transaction is committed
	require.NoError(t, db.Tx(ctx, func(conn sqlite.Conn) error {
		return conn.Insert(ctx, &user{}, userMeta{Name: "a"})
	}))
	assert.Equal(t, uint64(1), count(t, db))

	// Or rolled back, when it returns an error, which is returned
	assert.ErrorIs(t, db.Tx(ctx, func(conn sqlite.Conn) error {
		require.NoError(t, conn.Insert(ctx, &user{}, userMeta{Name: "b"}))
		assert.Equal(t, uint64(2), count(t, conn), "within it")
		return failed
	}), failed)
	assert.Equal(t, uint64(1), count(t, db))

	// Or panics, after which the writer is free
	assert.Panics(t, func() {
		_ = db.Tx(ctx, func(conn sqlite.Conn) error {
			require.NoError(t, conn.Insert(ctx, &user{}, userMeta{Name: "c"}))
			panic("panic")
		})
	})
	assert.Equal(t, uint64(1), count(t, db))
	require.NoError(t, db.Insert(ctx, &user{}, userMeta{Name: "d"}))
	assert.Equal(t, uint64(2), count(t, db))

	// A transaction within one is a savepoint, which is rolled back on its
	// own, and the outer one is committed
	require.NoError(t, db.Tx(ctx, func(conn sqlite.Conn) error {
		require.NoError(t, conn.Insert(ctx, &user{}, userMeta{Name: "e"}))
		assert.ErrorIs(t, conn.Tx(ctx, func(conn sqlite.Conn) error {
			require.NoError(t, conn.Insert(ctx, &user{}, userMeta{Name: "f"}))
			return conn.Tx(ctx, func(conn sqlite.Conn) error {
				require.NoError(t, conn.Insert(ctx, &user{}, userMeta{Name: "g"}))
				return failed
			})
		}), failed)
		return conn.Tx(ctx, func(conn sqlite.Conn) error {
			return conn.Insert(ctx, &user{}, userMeta{Name: "h"})
		})
	}))
	var list userList
	require.NoError(t, db.List(ctx, &list, userSelector{}))
	names := make([]string, 0, len(list.Users))
	for _, user := range list.Users {
		names = append(names, user.Name)
	}
	assert.Equal(t, []string{"a", "d", "e", "h"}, names)

	// A bulk operation is a transaction
	assert.ErrorIs(t, db.Bulk(ctx, func(conn sqlite.Conn) error {
		for _, name := range []string{"i", "j"} {
			if err := conn.Insert(ctx, &user{}, userMeta{Name: name}); err != nil {
				return err
			}
		}
		return failed
	}), failed)
	assert.Equal(t, uint64(4), count(t, db))

	// A transaction whose context is cancelled is rolled back
	cancelled, cancel := context.WithCancel(ctx)
	assert.Error(t, db.Tx(cancelled, func(conn sqlite.Conn) error {
		require.NoError(t, conn.Insert(cancelled, &user{}, userMeta{Name: "k"}))
		cancel()
		return conn.Insert(cancelled, &user{}, userMeta{Name: "l"})
	}))
	assert.Equal(t, uint64(4), count(t, db))

	// As is a query whose context is cancelled, and the database is kept
	assert.Error(t, db.Insert(cancelled, &user{}, userMeta{Name: "m"}))
	assert.Error(t, db.Get(cancelled, &user{}, userSelector{ID: 1}))
	assert.Equal(t, uint64(4), count(t, db))
}

func testVersion(t *testing.T, db sqlite.Conn) {
	ctx := t.Context()

	// The version is 0, until it's set, in a transaction, or not
	version, err := sqlite.Version(ctx, db)
	require.NoError(t, err)
	assert.Zero(t, version)
	require.NoError(t, db.Tx(ctx, func(conn sqlite.Conn) error {
		return sqlite.SetVersion(ctx, conn, 3)
	}))
	version, err = sqlite.Version(ctx, db)
	require.NoError(t, err)
	assert.Equal(t, int64(3), version)
	assert.ErrorIs(t, sqlite.SetVersion(ctx, db, -1), sqlite.ErrBadParameter)
}

func testConcurrent(t *testing.T, db sqlite.Conn) {
	ctx := t.Context()

	// Writes, transactions and reads, at once, wait for each other, rather
	// than failing
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Go(func() {
			for range 10 {
				var u user
				if i%2 == 0 {
					assert.NoError(t, db.Insert(ctx, &u, userMeta{Name: "a"}))
				} else {
					assert.NoError(t, db.Tx(ctx, func(conn sqlite.Conn) error {
						return conn.Insert(ctx, &u, userMeta{Name: "b"})
					}))
				}
				assert.NoError(t, db.Get(ctx, &user{}, userSelector{ID: u.ID}))
			}
		})
	}
	wg.Wait()
	assert.Equal(t, uint64(80), count(t, db))
}

func testReaders(t *testing.T, driver sqlite.Driver) {
	ctx := t.Context()
	db := open(t, driver, "File")

	// A read, while a transaction writes, doesn't wait for it, and sees what
	// was committed before it
	require.NoError(t, db.Insert(ctx, &user{}, userMeta{Name: "a"}))
	require.NoError(t, db.Tx(ctx, func(conn sqlite.Conn) error {
		require.NoError(t, conn.Insert(ctx, &user{}, userMeta{Name: "b"}))
		assert.Equal(t, uint64(1), count(t, db), "the pool's")
		assert.Equal(t, uint64(2), count(t, conn), "the transaction's")
		return nil
	}))
	assert.Equal(t, uint64(2), count(t, db))

	// Readers only read
	err := db.Get(ctx, &user{}, writeSelector{})
	assert.ErrorIs(t, err, sqlite.ErrDatabase)
	assert.Equal(t, uint64(2), count(t, db))
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// open returns a database, in memory, or in a file, with the tables, and
// the queries bound, which is closed when the test ends
func open(t *testing.T, driver sqlite.Driver, kind string) sqlite.Conn {
	t.Helper()
	path := sqlite.Memory
	if kind == "File" {
		path = filepath.Join(t.TempDir(), "test.db")
	}
	q, err := sqlite.NewQueries(strings.NewReader(queries))
	require.NoError(t, err)
	db, err := sqlite.Open(t.Context(), path, driver, sqlite.WithBind("table", "users"))
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	conn := db.WithQueries(q)
	for _, key := range []string{"users.create", "posts.create"} {
		require.NoError(t, conn.Exec(t.Context(), q.Query(key)))
	}
	return conn
}

// count returns the number of users
func count(t *testing.T, conn sqlite.Conn) uint64 {
	t.Helper()
	var list userList
	require.NoError(t, conn.List(t.Context(), &list, userSelector{}))
	return list.Count
}

////////////////////////////////////////////////////////////////////////////////
// READERS, WRITERS AND SELECTORS

func (u *user) Scan(row sqlite.Row) error {
	return row.Scan(&u.ID, &u.Name, &u.Email)
}

func (l *userList) Scan(row sqlite.Row) error {
	var u user
	if err := u.Scan(row); err != nil {
		return err
	}
	l.Users = append(l.Users, u)
	return nil
}

func (l *userList) ScanCount(row sqlite.Row) error {
	return row.Scan(&l.Count)
}

func (m userMeta) Insert(bind *sqlite.Bind) (string, error) {
	bind.Set("name", m.Name)
	bind.Set("email", m.Email)
	return bind.Query("users.insert"), nil
}

func (m userMeta) Update(bind *sqlite.Bind) error {
	bind.Set("name", m.Name)
	return nil
}

func (s userSelector) Select(bind *sqlite.Bind, op sqlite.Op) (string, error) {
	if s.ID < 0 {
		return "", sqlite.ErrBadParameter.Withf("id: %d", s.ID)
	}
	bind.Set("id", s.ID)
	switch op {
	case sqlite.Get:
		return bind.Query("users.get"), nil
	case sqlite.Update:
		return bind.Query("users.update"), nil
	case sqlite.Delete:
		return bind.Query("users.delete"), nil
	case sqlite.List:
		return bind.Query("users.list"), nil
	}
	return "", sqlite.ErrNotImplemented.With(op)
}

// offsetLimit lists users, of an offset and limit
type offsetLimit struct {
	sqlite.OffsetLimit
}

func (s offsetLimit) Select(bind *sqlite.Bind, op sqlite.Op) (string, error) {
	s.OffsetLimit.Bind(bind, 100)
	return bind.Query("users.list"), nil
}

// writeSelector writes, which a reader can't
type writeSelector struct{}

func (writeSelector) Select(*sqlite.Bind, sqlite.Op) (string, error) {
	return `INSERT INTO users (name) VALUES ('x') RETURNING id, name, email`, nil
}
