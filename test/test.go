// Package test is the tests which an implementation, such as package
// ncruces, passes, so it can be replaced by another, which behaves the same.
package test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/netip"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
	"uuid"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
	codes "go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	tracetest "go.opentelemetry.io/otel/sdk/trace/tracetest"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// OpenFunc opens a database, of a path, or sqlite.Memory, such as
// ncruces.Open
type OpenFunc func(ctx context.Context, path string, opts ...sqlite.Opt) (sqlite.PoolConn, error)

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

-- types.create
CREATE TABLE IF NOT EXISTS types (
	id INTEGER PRIMARY KEY, i INTEGER, f REAL, t TEXT, b BLOB, ts TEXT, ok INTEGER, n TEXT
);
`

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Run runs the tests of the implementation, which opens databases, in
// memory, and in files, in a temporary folder.
func Run(t *testing.T, fn OpenFunc) {
	t.Run("Open", func(t *testing.T) { testOpen(t, fn) })
	t.Run("Close", func(t *testing.T) { testClose(t, fn) })
	t.Run("Functions", func(t *testing.T) { testFunctions(t, fn) })
	t.Run("Tracer", func(t *testing.T) { testTracer(t, fn) })
	for _, kind := range []string{"Memory", "File"} {
		t.Run(kind, func(t *testing.T) {
			t.Run("CRUD", func(t *testing.T) { testCRUD(t, open(t, fn, kind)) })
			t.Run("List", func(t *testing.T) { testList(t, open(t, fn, kind)) })
			t.Run("Types", func(t *testing.T) { testTypes(t, open(t, fn, kind)) })
			t.Run("Errors", func(t *testing.T) { testErrors(t, open(t, fn, kind)) })
			t.Run("Statements", func(t *testing.T) { testStatements(t, open(t, fn, kind)) })
			t.Run("Tx", func(t *testing.T) { testTx(t, open(t, fn, kind)) })
			t.Run("Bulk", func(t *testing.T) { testBulk(t, open(t, fn, kind)) })
			t.Run("Version", func(t *testing.T) { testVersion(t, open(t, fn, kind)) })
			t.Run("Concurrent", func(t *testing.T) { testConcurrent(t, open(t, fn, kind)) })
		})
	}
	t.Run("Readers", func(t *testing.T) { testReaders(t, fn) })
}

////////////////////////////////////////////////////////////////////////////////
// TESTS

func testOpen(t *testing.T, fn OpenFunc) {
	ctx := t.Context()

	// A path is needed, and options are checked
	_, err := fn(ctx, " ")
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
	_, err = fn(ctx, sqlite.Memory, sqlite.WithMaxReaders(0))
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
	_, err = fn(ctx, sqlite.Memory, sqlite.WithPragma("a; DROP", "1"))
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
	_, err = fn(ctx, sqlite.Memory, sqlite.WithPragma("cache_size", "1; DROP TABLE users"))
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
	_, err = fn(ctx, sqlite.Memory, sqlite.WithFunction("f", true, sqlite.Fn{}))
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
	_, err = fn(ctx, sqlite.Memory, sqlite.WithFunction("f", true, sqlite.Func1[int64, int64](nil)))
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
	_, err = fn(ctx, sqlite.Memory, sqlite.WithFunction("not a name", true, sqlite.Func0(func() (int, error) { return 1, nil })))
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)

	// A context which is done doesn't open, or create, the database
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	path := filepath.Join(t.TempDir(), "cancelled.db")
	_, err = fn(cancelled, path)
	assert.ErrorIs(t, err, context.Canceled)
	assert.NoFileExists(t, path)

	// A folder which doesn't exist can't be opened
	_, err = fn(ctx, filepath.Join(t.TempDir(), "missing", "test.db"))
	assert.Error(t, err)

	// A path with characters which are a URI's, such as ? and #, is its name
	path = filepath.Join(t.TempDir(), "a b?c#d%20.db")
	db, err := fn(ctx, path, sqlite.WithPragma("cache_size", "-4000"))
	require.NoError(t, err)
	assert.NoError(t, db.Ping(ctx))
	assert.NoError(t, db.Close())
	assert.FileExists(t, path)
}

func testClose(t *testing.T, fn OpenFunc) {
	ctx := t.Context()
	db, err := fn(ctx, filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)

	// Once it's closed, it's not available, and it's closed once
	require.NoError(t, db.Close())
	assert.NoError(t, db.Close())
	assert.ErrorIs(t, db.Exec(ctx, "SELECT 1"), sqlite.ErrNotAvailable)
	assert.ErrorIs(t, db.Get(ctx, &value{}, value{query: "SELECT 1"}), sqlite.ErrNotAvailable)
	assert.ErrorIs(t, db.Ping(ctx), sqlite.ErrNotAvailable)
}

func testFunctions(t *testing.T, fn OpenFunc) {
	ctx := t.Context()
	failed := errors.New("failed")

	// Functions, of each connection, the readers' too, whose arguments are
	// scanned into their types, and whose results are set, of theirs
	db, err := fn(ctx, filepath.Join(t.TempDir(), "test.db"),
		sqlite.WithFunction("answer", true, sqlite.Func0(func() (int64, error) {
			return 42, nil
		})),
		sqlite.WithFunction("double", true, sqlite.Func1(func(n int64) (int64, error) {
			return n * 2, nil
		})),
		sqlite.WithFunction("elapsed", true, sqlite.Func2(func(from, to time.Time) (time.Duration, error) {
			return to.Sub(from), nil
		})),
		sqlite.WithFunction("clamp", true, sqlite.Func3(func(n, lo, hi float64) (float64, error) {
			return min(max(n, lo), hi), nil
		})),
		sqlite.WithFunction("nullable", true, sqlite.Func1(func(n *int64) (string, error) {
			if n == nil {
				return "null", nil
			}
			return fmt.Sprint(*n), nil
		})),
		sqlite.WithFunction("tags", true, sqlite.Func1(func(tags []string) (string, error) {
			return strings.Join(tags, "+"), nil
		})),
		sqlite.WithFunction("fails", true, sqlite.Func1(func(string) (any, error) {
			return nil, failed
		})),
		sqlite.WithFunction("result", true, sqlite.Func1(func(kind string) (any, error) {
			text := "text"
			switch kind {
			case "bool":
				return true, nil
			case "uint":
				return uint8(7), nil
			case "float":
				return float32(2.5), nil
			case "bytes":
				return []byte{1}, nil
			case "ptr":
				return &text, nil
			case "nilptr":
				return (*string)(nil), nil
			case "valuer":
				return sql.NullInt64{Int64: 3, Valid: true}, nil
			case "time":
				return time.Date(2026, 11, 18, 12, 30, 0, 0, time.UTC), nil
			case "large":
				return uint64(1 << 63), nil
			case "slice":
				return []string{"a"}, nil
			case "chan":
				return make(chan int), nil
			}
			return nil, nil
		})),
		sqlite.WithFunction("kind", true, sqlite.FuncN(func(args ...any) (string, error) {
			kinds := make([]string, 0, len(args))
			for _, arg := range args {
				kinds = append(kinds, fmt.Sprintf("%T", arg))
			}
			return strings.Join(kinds, ","), nil
		})),
		sqlite.WithFunction("sum", true, sqlite.FuncN(func(args ...int64) (int64, error) {
			var sum int64
			for _, n := range args {
				sum += n
			}
			return sum, nil
		})),
	)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	for _, conn := range []sqlite.Conn{db, txConn(t, db)} {
		v := &value{query: "SELECT double(21)"}
		require.NoError(t, conn.Get(ctx, v, v))
		assert.Equal(t, int64(42), v.value)
	}

	// Of each arity, and type
	for _, test := range []struct {
		query string
		want  any
	}{
		{"SELECT answer()", int64(42)},
		{"SELECT double('21')", int64(42)},
		{"SELECT elapsed('2026-11-18 11:00', '2026-11-18 12:30')", int64(90 * time.Minute)},
		{"SELECT clamp(5, 0, 1.5)", 1.5},
		{"SELECT nullable(NULL)", "null"},
		{"SELECT nullable(7)", "7"},
		{`SELECT tags('["a","b"]')`, "a+b"},
		{"SELECT kind(1, 1.5, 'a', x'00', NULL)", "int64,float64,string,[]uint8,<nil>"},
		{"SELECT sum()", int64(0)},
		{"SELECT sum(1, 2, 3)", int64(6)},
	} {
		var got any
		require.NoError(t, db.Get(ctx, &column{&got}, value{query: test.query}), test.query)
		assert.Equal(t, test.want, got, test.query)
	}

	// Results are of their kinds, and pointers, which are NULL when they're
	// nil, and driver.Valuers
	for _, test := range []struct {
		kind string
		want any
	}{
		{"bool", int64(1)}, {"uint", int64(7)}, {"float", 2.5}, {"bytes", []byte{1}},
		{"ptr", "text"}, {"nilptr", nil}, {"nil", nil}, {"valuer", int64(3)},
		{"time", "2026-11-18T12:30:00Z"}, {"slice", `["a"]`},
	} {
		var got any
		require.NoError(t, db.With("kind", test.kind).Get(ctx, &column{&got}, value{query: "SELECT result(@kind)"}), test.kind)
		assert.Equal(t, test.want, got, test.kind)
	}
	for _, kind := range []string{"large", "chan"} {
		assert.ErrorIs(t, db.With("kind", kind).Get(ctx, &column{new(any)}, value{query: "SELECT result(@kind)"}), sqlite.ErrDatabase, kind)
	}

	// An argument which isn't of its type, NULL, which isn't a pointer, the
	// wrong number of arguments, and the function's error, are the query's,
	// which say which argument, or what failed
	for query, want := range map[string]string{
		"SELECT double('a')":                "argument 1",
		"SELECT double(NULL)":               "argument 1",
		"SELECT clamp(1, 'x', 2)":           "argument 2",
		"SELECT sum(1, 'x')":                "argument 2",
		"SELECT double(1, 2)":               "double",
		"SELECT fails('a')":                 "failed",
		`SELECT tags('not json')`:           "argument 1",
		"SELECT elapsed('x', '2026-11-18')": "argument 1",
	} {
		err := db.Get(ctx, &column{new(any)}, value{query: query})
		assert.ErrorIs(t, err, sqlite.ErrDatabase, query)
		assert.ErrorContains(t, err, want, query)
	}
}

func testTypes(t *testing.T, db sqlite.Conn) {
	ctx := t.Context()
	type name string
	when := time.Date(2026, 11, 18, 12, 30, 0, 123, time.UTC)

	// Values are bound, of their kinds, and pointers, which are NULL when
	// they're nil, and driver.Valuers
	var email *string
	conn := db.With("i", int8(-7), "f", float32(1.5), "t", name("text"), "b", []byte{0, 1}, "ts", when, "ok", true, "n", email, "s", sql.NullString{})
	require.NoError(t, conn.Exec(ctx, `INSERT INTO types (id, i, f, t, b, ts, ok, n) VALUES (1, @i, @f, @t, @b, @ts, @ok, @n)`))
	require.NoError(t, conn.Exec(ctx, `INSERT INTO types (id, i, f, t, b, ts, ok, n) VALUES (2, 300, 2.5, '42', NULL, 'not a time', 0, @s)`))

	// And scanned into theirs
	var (
		i   int8
		u   uint16
		f   float32
		s   name
		b   []byte
		ts  time.Time
		ok  bool
		n   *string
		any any
		ns  sql.NullString
	)
	require.NoError(t, db.Get(ctx, &column{&i, &u, &f, &s, &b, &ts, &ok, &n, &any, &ns}, value{query: `SELECT i, -i, f, t, b, ts, ok, n, f, n FROM types WHERE id = 1`}))
	assert.Equal(t, int8(-7), i)
	assert.Equal(t, uint16(7), u)
	assert.Equal(t, float32(1.5), f)
	assert.Equal(t, name("text"), s)
	assert.Equal(t, []byte{0, 1}, b)
	assert.True(t, when.Equal(ts), ts)
	assert.True(t, ok)
	assert.Nil(t, n)
	assert.Equal(t, 1.5, any)
	assert.False(t, ns.Valid)

	// Text which is a number is an integer, and NULL is a nil []byte
	var count int64
	require.NoError(t, db.Get(ctx, &column{&count, &b}, value{query: `SELECT t, b FROM types WHERE id = 2`}))
	assert.Equal(t, int64(42), count)
	assert.Nil(t, b)

	// A value which doesn't fit, NULL, which isn't a pointer, text which
	// isn't a time, and the wrong number of values, are errors
	assert.Error(t, db.Get(ctx, &column{&i}, value{query: `SELECT i FROM types WHERE id = 2`}))
	assert.Error(t, db.Get(ctx, &column{&u}, value{query: `SELECT -1`}))
	assert.Error(t, db.Get(ctx, &column{&s}, value{query: `SELECT NULL`}))
	assert.Error(t, db.Get(ctx, &column{&ts}, value{query: `SELECT ts FROM types WHERE id = 2`}))
	assert.Error(t, db.Get(ctx, &column{&i, &u}, value{query: `SELECT 1`}))
	assert.Error(t, db.Get(ctx, &column{i}, value{query: `SELECT 1`}))

	// Composites are JSON, which SQLite's json functions read, and are
	// scanned back
	var tags []string
	require.NoError(t, db.With("v", []string{"a", "b"}).Get(ctx, &column{&count}, value{query: `SELECT COUNT(*) FROM json_each(@v)`}))
	assert.Equal(t, int64(2), count)
	require.NoError(t, db.With("v", []string{"a", "b"}).Get(ctx, &column{&tags}, value{query: `SELECT @v`}))
	assert.Equal(t, []string{"a", "b"}, tags)
	meta := map[string]interface{}{"a": float64(1)}
	var gotMeta map[string]interface{}
	require.NoError(t, db.With("v", meta).Get(ctx, &column{&gotMeta}, value{query: `SELECT json(@v)`}))
	assert.Equal(t, meta, gotMeta)

	// Text marshalers, such as netip.Addr, are text, and durations are
	// nanoseconds, or seconds, of SQLite's arithmetic
	var addr netip.Addr
	var kind string
	require.NoError(t, db.With("v", netip.MustParseAddr("10.0.0.1")).Get(ctx, &column{&addr, &kind}, value{query: `SELECT @v, typeof(@v)`}))
	assert.Equal(t, netip.MustParseAddr("10.0.0.1"), addr)
	assert.Equal(t, "text", kind)
	var d time.Duration
	require.NoError(t, db.With("v", 90*time.Minute).Get(ctx, &column{&d}, value{query: `SELECT @v`}))
	assert.Equal(t, 90*time.Minute, d)
	require.NoError(t, db.Get(ctx, &column{&d}, value{query: `SELECT (julianday('2026-11-18 12:30') - julianday('2026-11-18 11:00')) * 86400`}))
	assert.InDelta(t, float64(90*time.Minute), float64(d), float64(time.Millisecond))

	// The standard library's UUIDs are text, and read from text, or blobs
	id := uuid.New()
	var gotID uuid.UUID
	require.NoError(t, db.With("v", id).Get(ctx, &column{&gotID, &kind}, value{query: `SELECT @v, typeof(@v)`}))
	assert.Equal(t, id, gotID)
	assert.Equal(t, "text", kind)
	gotID = uuid.UUID{}
	require.NoError(t, db.With("v", id[:]).Get(ctx, &column{&gotID}, value{query: `SELECT @v`}))
	assert.Equal(t, id, gotID)

	// A value which can't be bound is a bad parameter
	assert.ErrorIs(t, db.With("v", make(chan int)).Exec(ctx, `SELECT @v`), sqlite.ErrBadParameter)
	assert.ErrorIs(t, db.With("v", uint64(1<<63)).Exec(ctx, `SELECT @v`), sqlite.ErrBadParameter)
}

func testTracer(t *testing.T, fn OpenFunc) {
	ctx := t.Context()
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	db := openPath(t, fn, filepath.Join(t.TempDir(), "test.db"), sqlite.WithTracer(provider.Tracer("test")))

	// Spans of opening it, a named query, with the arguments it bound, a
	// transaction, which commits, and one which rolls back, a get which finds
	// nothing, which isn't an error, and SQL which isn't, which is
	email := "a@example.com"
	require.NoError(t, db.Insert(ctx, &user{}, userMeta{Name: "a", Email: &email}))
	require.NoError(t, db.Tx(ctx, func(conn sqlite.Conn) error { return nil }))
	assert.Error(t, db.Tx(ctx, func(conn sqlite.Conn) error { return errors.New("failed") }))
	assert.ErrorIs(t, db.Get(ctx, &user{}, userSelector{ID: 99}), sqlite.ErrNotFound)
	assert.Error(t, db.Exec(ctx, "NOT SQL"))

	spans := make(map[string]sdktrace.ReadOnlySpan)
	for _, span := range recorder.Ended() {
		spans[span.Name()] = span
	}
	attrs := func(span sdktrace.ReadOnlySpan) map[string]string {
		result := make(map[string]string)
		for _, attr := range span.Attributes() {
			result[string(attr.Key)] = attr.Value.Emit()
		}
		return result
	}
	for _, name := range []string{"sqlite.open", "users.insert", "users.get", "sqlite.transaction", "sqlite.commit", "sqlite.rollback", "sqlite.query"} {
		if assert.Contains(t, spans, name) {
			assert.Equal(t, "sqlite", attrs(spans[name])["db.system"], name)
		}
	}
	if span, ok := spans["users.insert"]; ok {
		assert.Equal(t, "a", attrs(span)["arg.name"])
		assert.Equal(t, email, attrs(span)["arg.email"])
		assert.NotContains(t, attrs(span), "arg.table", "a substitution, which isn't an argument")
		assert.Contains(t, attrs(span)["db.statement"], `INSERT INTO "users"`)
		assert.Equal(t, codes.Unset, span.Status().Code)
	}
	if span, ok := spans["users.get"]; ok {
		assert.Equal(t, "99", attrs(span)["arg.id"])
		assert.Equal(t, codes.Unset, span.Status().Code, "not found isn't an error")
	}
	if span, ok := spans["sqlite.query"]; ok {
		assert.Equal(t, "NOT SQL", attrs(span)["db.statement"])
		assert.Equal(t, codes.Error, span.Status().Code)
	}
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

	// A get is of the first row, of several
	require.NoError(t, db.Insert(ctx, &user{}, userMeta{Name: "c"}))
	require.NoError(t, db.Insert(ctx, &user{}, userMeta{Name: "d"}))
	require.NoError(t, db.Get(ctx, &got, query{"SELECT id, name, email FROM users ORDER BY name"}))
	assert.Equal(t, "c", got.Name)

	// An @ in a string, an identifier or a comment isn't an argument
	var at string
	require.NoError(t, db.With("v", "x").Get(ctx, &column{&at}, query{"SELECT 'a@example.com' AS \"@b\" /* @c */ -- @d\n"}))
	assert.Equal(t, "a@example.com", at)

	// A $key, :key or @key is an argument, which is bound, rather than
	// substituted, so a value which is SQL is kept as it is
	sql := "x'); DROP TABLE users; --"
	for _, param := range []string{"$name", ":name", "@name"} {
		var u user
		require.NoError(t, db.With("name", sql).Update(ctx, &u, query{"INSERT INTO users (name) VALUES (" + param + ") RETURNING id, name, email"}, nil), param)
		assert.Equal(t, sql, u.Name, param)
	}

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

	// A query which ends with a ;, or a comment, is listed
	for _, sql := range []string{"SELECT id, name, email FROM users;", "SELECT id, name, email FROM users; -- all of them\n", "SELECT id, name, email FROM users; /* all of them */"} {
		require.NoError(t, db.List(ctx, &list, query{sql}), sql)
	}

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

func testStatements(t *testing.T, db sqlite.Conn) {
	ctx := t.Context()

	// A query of several statements executes them all
	require.NoError(t, db.Exec(ctx, `INSERT INTO users (name) VALUES ('a'); INSERT INTO users (name) VALUES ('b')`))
	assert.Equal(t, uint64(2), count(t, db))

	// Or none, when one fails, or its reader panics
	err := db.Exec(ctx, `INSERT INTO users (name) VALUES ('c'); INSERT INTO users (name) VALUES (NULL)`)
	assert.ErrorIs(t, err, sqlite.ErrNotNullViolation)
	assert.Equal(t, uint64(2), count(t, db))
	panics := readerFunc(func(sqlite.Row) error { panic("scan") })
	assert.Panics(t, func() {
		_ = db.Update(ctx, panics, query{`INSERT INTO users (name) VALUES ('d') RETURNING id, name, email; SELECT 1`}, nil)
	})
	assert.Equal(t, uint64(2), count(t, db))

	// Or none, when they can't be released, such as of a deferred foreign
	// key, and the connection is used again
	require.NoError(t, db.Exec(ctx, `CREATE TABLE IF NOT EXISTS deferred (user INTEGER REFERENCES users (id) DEFERRABLE INITIALLY DEFERRED)`))
	err = db.Exec(ctx, `INSERT INTO users (name) VALUES ('d'); INSERT INTO deferred (user) VALUES (999)`)
	assert.ErrorIs(t, err, sqlite.ErrForeignKeyViolation)
	assert.Equal(t, uint64(2), count(t, db))
	require.NoError(t, db.Exec(ctx, `INSERT INTO users (name) VALUES ('d')`))
	require.NoError(t, db.Exec(ctx, `DELETE FROM users WHERE name = 'd'`))
	assert.Equal(t, uint64(2), count(t, db))

	// A query which controls its own transaction isn't in a savepoint, and
	// one which controls it after its first statement is a bad parameter, of
	// which nothing is committed
	require.NoError(t, db.Exec(ctx, `BEGIN; INSERT INTO users (name) VALUES ('x'); ROLLBACK`))
	assert.Equal(t, uint64(2), count(t, db))
	err = db.Exec(ctx, `INSERT INTO users (name) VALUES ('y'); COMMIT; INSERT INTO users (name) VALUES ('z')`)
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
	assert.Equal(t, uint64(2), count(t, db))

	// A statement with comments, or ;s, after it is one statement, which
	// isn't in a savepoint, such as VACUUM, which can't be
	require.NoError(t, db.Exec(ctx, "VACUUM; -- cleanup\n;; /* done */"))

	// Within a transaction, the query's statements are rolled back, and the
	// transaction goes on
	require.NoError(t, db.Tx(ctx, func(conn sqlite.Conn) error {
		require.NoError(t, conn.Exec(ctx, `INSERT INTO users (name) VALUES ('e')`))
		assert.ErrorIs(t, conn.Exec(ctx, `INSERT INTO users (name) VALUES ('f'); INSERT INTO users (name) VALUES (NULL)`), sqlite.ErrNotNullViolation)
		return conn.Exec(ctx, `INSERT INTO users (name) VALUES ('g'); INSERT INTO users (name) VALUES ('h')`)
	}))
	var list userList
	require.NoError(t, db.List(ctx, &list, userSelector{}))
	names := make([]string, 0, len(list.Users))
	for _, user := range list.Users {
		names = append(names, user.Name)
	}
	assert.Equal(t, []string{"a", "b", "e", "g", "h"}, names)
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

	// As is a transaction whose function cancels its context, and returns
	// nil, and a savepoint's
	cancelled, cancel = context.WithCancel(ctx)
	assert.ErrorIs(t, db.Tx(cancelled, func(conn sqlite.Conn) error {
		require.NoError(t, conn.Insert(cancelled, &user{}, userMeta{Name: "k"}))
		cancel()
		return nil
	}), context.Canceled)
	assert.Equal(t, uint64(4), count(t, db))
	require.NoError(t, db.Tx(ctx, func(conn sqlite.Conn) error {
		cancelled, cancel := context.WithCancel(ctx)
		assert.ErrorIs(t, conn.Tx(cancelled, func(conn sqlite.Conn) error {
			require.NoError(t, conn.Insert(cancelled, &user{}, userMeta{Name: "k"}))
			cancel()
			return nil
		}), context.Canceled)
		return nil
	}))
	assert.Equal(t, uint64(4), count(t, db))

	// As is a query whose context is cancelled, and the database is kept
	assert.Error(t, db.Insert(cancelled, &user{}, userMeta{Name: "m"}))
	assert.Error(t, db.Get(cancelled, &user{}, userSelector{ID: 1}))
	assert.Equal(t, uint64(4), count(t, db))
}

func testBulk(t *testing.T, db sqlite.Conn) {
	ctx := t.Context()
	failed := errors.New("failed")

	// Many inserts, of the same query, with their own arguments, and reads,
	// of the first row, and lists, within it, of what's written
	var ids []int64
	require.NoError(t, db.Bulk(ctx, func(conn sqlite.Conn) error {
		for i := range 500 {
			var u user
			if err := conn.Insert(ctx, &u, userMeta{Name: fmt.Sprint("user", i)}); err != nil {
				return err
			}
			ids = append(ids, u.ID)
		}
		for i, id := range ids[:50] {
			var got user
			if err := conn.Get(ctx, &got, userSelector{ID: id}); err != nil {
				return err
			}
			assert.Equal(t, fmt.Sprint("user", i), got.Name)
			if err := conn.Get(ctx, &got, query{"SELECT id, name, email FROM users ORDER BY id DESC"}); err != nil {
				return err
			}
			assert.Equal(t, "user499", got.Name)
		}
		assert.Equal(t, uint64(500), count(t, conn))
		return nil
	}))
	assert.Equal(t, uint64(500), count(t, db))

	// Queries which differ by a substitution, more than are reused
	require.NoError(t, db.Bulk(ctx, func(conn sqlite.Conn) error {
		for i := range 100 {
			if err := conn.With("name", fmt.Sprint("sub", i)).Exec(ctx, `INSERT INTO users (name) VALUES (${'name'})`); err != nil {
				return err
			}
		}
		return nil
	}))
	assert.Equal(t, uint64(600), count(t, db))

	// An error rolls all of it back, and a savepoint, within it, rolls back
	// its own
	assert.ErrorIs(t, db.Bulk(ctx, func(conn sqlite.Conn) error {
		for range 10 {
			if err := conn.Insert(ctx, &user{}, userMeta{Name: "x"}); err != nil {
				return err
			}
		}
		return failed
	}), failed)
	assert.Equal(t, uint64(600), count(t, db))
	require.NoError(t, db.Bulk(ctx, func(conn sqlite.Conn) error {
		require.NoError(t, conn.Insert(ctx, &user{}, userMeta{Name: "kept"}))
		assert.ErrorIs(t, conn.Bulk(ctx, func(conn sqlite.Conn) error {
			require.NoError(t, conn.Insert(ctx, &user{}, userMeta{Name: "lost"}))
			return failed
		}), failed)
		return conn.Insert(ctx, &user{}, userMeta{Name: "kept"})
	}))
	assert.Equal(t, uint64(602), count(t, db))

	// A reader which panics, within it, is recovered from, and its statement
	// is reused, and outside it, the database is used again
	panics := readerFunc(func(sqlite.Row) error { panic("scan") })
	require.NoError(t, db.Bulk(ctx, func(conn sqlite.Conn) error {
		for range 2 {
			assert.Panics(t, func() { _ = conn.Get(ctx, panics, userSelector{ID: ids[0]}) })
		}
		var got user
		require.NoError(t, conn.Get(ctx, &got, userSelector{ID: ids[0]}))
		assert.Equal(t, "user0", got.Name)
		return nil
	}))
	assert.Panics(t, func() { _ = db.Get(ctx, panics, userSelector{ID: ids[0]}) })
	assert.Equal(t, uint64(602), count(t, db))

	// A constraint error, of a reused statement, is returned, and it's
	// reused again
	email := "a@example.com"
	require.NoError(t, db.Bulk(ctx, func(conn sqlite.Conn) error {
		require.NoError(t, conn.Insert(ctx, &user{}, userMeta{Name: "a", Email: &email}))
		assert.ErrorIs(t, conn.Insert(ctx, &user{}, userMeta{Name: "b", Email: &email}), sqlite.ErrUniqueViolation)
		return conn.Insert(ctx, &user{}, userMeta{Name: "c"})
	}))
	assert.Equal(t, uint64(604), count(t, db))
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

func testReaders(t *testing.T, fn OpenFunc) {
	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "test.db")
	db := openPath(t, fn, path)

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

	// A transaction which waits for another database's, of the same file,
	// to write, stops when its context is done, and says so
	other, err := fn(ctx, path)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, other.Close()) })
	require.NoError(t, db.Tx(ctx, func(conn sqlite.Conn) error {
		waiting, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()
		err := other.Tx(waiting, func(sqlite.Conn) error { return nil })
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.ErrorIs(t, err, sqlite.ErrNotAvailable)
		return nil
	}))

	// Readers only read
	err = db.Get(ctx, &user{}, writeSelector{})
	assert.ErrorIs(t, err, sqlite.ErrDatabase)
	assert.Equal(t, uint64(2), count(t, db))
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// open returns a database, in memory, or in a file, with the tables, and
// the queries bound, which is closed when the test ends
func open(t *testing.T, fn OpenFunc, kind string) sqlite.Conn {
	t.Helper()
	path := sqlite.Memory
	if kind == "File" {
		path = filepath.Join(t.TempDir(), "test.db")
	}
	return openPath(t, fn, path)
}

// openPath returns the database of the path, as open does
func openPath(t *testing.T, fn OpenFunc, path string, opts ...sqlite.Opt) sqlite.Conn {
	t.Helper()
	q, err := sqlite.NewQueries(strings.NewReader(queries))
	require.NoError(t, err)
	db, err := fn(t.Context(), path, append([]sqlite.Opt{sqlite.WithBind("table", "users")}, opts...)...)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	conn := db.WithQueries(q)
	for _, key := range []string{"users.create", "posts.create", "types.create"} {
		require.NoError(t, conn.Exec(t.Context(), q.Query(key)))
	}
	return conn
}

// txConn returns a transaction's connection, which is rolled back when the
// test ends
func txConn(t *testing.T, db sqlite.Conn) sqlite.Conn {
	t.Helper()
	conn := make(chan sqlite.Conn)
	end := make(chan struct{})
	done := make(chan error)
	go func() {
		done <- db.Tx(context.Background(), func(tx sqlite.Conn) error {
			conn <- tx
			<-end
			return errors.New("rolled back")
		})
	}()
	select {
	case tx := <-conn:
		t.Cleanup(func() {
			close(end)
			<-done
		})
		return tx
	case err := <-done:
		t.Fatalf("transaction: %v", err)
		return nil
	}
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

// value is a query's one integer
type value struct {
	query string
	value int64
}

func (v value) Select(*sqlite.Bind, sqlite.Op) (string, error) { return v.query, nil }
func (v *value) Scan(row sqlite.Row) error                     { return row.Scan(&v.value) }

// column scans a row into its values
type column []any

func (c *column) Scan(row sqlite.Row) error { return row.Scan((*c)...) }

// query is a selector of a query
type query struct {
	sql string
}

func (q query) Select(*sqlite.Bind, sqlite.Op) (string, error) { return q.sql, nil }

// readerFunc is a function which scans a row
type readerFunc func(sqlite.Row) error

func (fn readerFunc) Scan(row sqlite.Row) error { return fn(row) }
