# go-sqlite

SQLite for Go, whose operations map directly onto API calls. You write the SQL; the library binds it and scans the results into your types. Pure Go, without cgo, with the same semantics as [go-pg](https://github.com/mutablelogic/go-pg) for PostgreSQL.

Documentation: <https://pkg.go.dev/github.com/mutablelogic/go-sqlite>

## Operations map onto API calls

Each operation is an API call. A request type selects the rows, and a response type reads them:

| Operation | API call | Interfaces | Connection |
| --- | --- | --- | --- |
| **Insert** a row, and return it | _POST_ or _PUT_ | `Writer`, and a `Reader` of what's returned | writer |
| **Update** rows, and optionally return them | _PATCH_ | `Selector` and `Writer`, and a `Reader` | writer |
| **Delete** rows, and optionally return them | _DELETE_ | `Selector`, and a `Reader` | writer |
| **Get** a row | _GET_ | `Selector` and `Reader` | a reader |
| **List** rows, and count them | _GET_ | `Selector`, and a `Reader` or `ListReader` | a reader |

So a handler is one call, and `sqlite.HTTPError` turns its errors into HTTP errors, such as not found or conflict:

```go
// GET /objects/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request, id int64) error {
  var obj MyObject
  if err := h.db.Get(r.Context(), &obj, MyObject{Id: id}); err != nil {
    return httpresponse.Error(w, sqlite.HTTPError(err))
  }
  return httpresponse.JSON(w, http.StatusOK, 2, obj)
}

// POST /objects
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) error {
  var req MyObject
  if err := httprequest.Read(r, &req); err != nil {
    return httpresponse.Error(w, err)
  }
  var obj MyObject
  if err := h.db.Insert(r.Context(), &obj, req); err != nil {
    return httpresponse.Error(w, sqlite.HTTPError(err))
  }
  return httpresponse.JSON(w, http.StatusCreated, 2, obj)
}
```

A handler for go-pg, against PostgreSQL, has the same shape, with `pg.HTTPError`.

## SQL first, not an ORM

There's no query builder, no generated SQL, and no mapping of struct tags to tables. Your types return the SQL for each operation, and bind its arguments:

```go
func (obj MyObject) Select(bind *sqlite.Bind, op sqlite.Op) (string, error) {
  bind.Set("id", obj.Id)
  switch op {
  case sqlite.Get:
    return `SELECT id, name FROM mytable WHERE id = @id`, nil
  case sqlite.Delete:
    return `DELETE FROM mytable WHERE id = @id RETURNING id, name`, nil
  }
  return "", sqlite.ErrNotImplemented
}
```

* **What runs is what you read.** Queries can be reviewed, and tested in the `sqlite3` shell, as they are.
* **All of SQLite:** full-text search, JSON, window functions, common table expressions, `RETURNING` and `UPSERT`, with nothing to wait for or work around.
* **Queries in files:** keep them in `.sql` files, by name, as [saved queries](#saved-queries), with `${key}` substitutions for names, such as a schema.
* **Thin types:** a `Selector` binds and returns SQL, a `Reader` scans a row, and a `Writer` binds values. Nothing else needs to know about the database.

## Features

* **One writer and many readers,** with WAL, so reads don't wait for writes, and writes don't fail with "database is locked".
* **Transactions,** nested as savepoints, and bulk operations, which reuse their statements.
* **Values:** Go values to SQLite's, and back, including JSON, times, durations and UUIDs.
* **Extensions,** such as full-text search, and functions written in Go, which SQL calls.
* **Tracing,** with OpenTelemetry.
* **Changes:** rows which are written, published to a broadcaster once they're committed.

For a working example, see [filesearch](filesearch): a command-line tool which indexes and searches files, with transactions, saved queries and full-text search.

## Interfaces

Each operation:

1. Calls the `Writer` and the `Selector`, which bind their parameters, and return the SQL.
2. Waits for a connection: the writer, which writes one statement or transaction at a time, or one of the readers, which read at once.
3. Replaces the SQL's `${key}` substitutions, and binds its `@key` arguments.
4. Executes it, and calls the `Reader` with each row it returns.

### Selector

```go
type Selector interface {
  // Bind row selection variables, returning the SQL statement required for the operation.
  // The operation is Get, Update, Delete or List.
  Select(*Bind, Op) (string, error)
}
```

A `Selector` selects rows from a table, for Get, List, Update and Delete operations.

### Reader

```go
type Reader interface {
  // Scan a row into the receiver
  Scan(Row) error
}

type ListReader interface {
  Reader

  // Scan a count of all the rows into the receiver
  ScanCount(Row) error
}
```

A `Reader` translates a row into a Go type. When there are several rows, `Scan` is called with each. Get calls it with the first row only. A `ListReader` is also given the count of all the rows a List selects, without its offset and limit.

### Writer

```go
type Writer interface {
  // Bind insert parameters, returning the SQL statement required for the insert
  Insert(*Bind) (string, error)

  // Bind update parameters
  Update(*Bind) error
}
```

A `Writer` binds a Go type's fields to SQL parameters, for Insert and Update operations.

## Opening a Database

The [`ncruces`](ncruces) package opens databases with [ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3), which is SQLite compiled to WebAssembly:

```go
import (
  sqlite "github.com/mutablelogic/go-sqlite"
  ncruces "github.com/mutablelogic/go-sqlite/ncruces"
)

func main() {
  db, err := ncruces.Open(ctx, "index.db", sqlite.WithMaxReaders(4))
  if err != nil {
    panic(err)
  }
  defer db.Close()

  // ...
}
```

* **Files** are created when they don't exist. Their journal is a write-ahead log, so reads see what's committed while a transaction writes.
* **In memory:** `sqlite.Memory`, which is one connection, which writes and reads.
* **Pragmas:** foreign keys are on, `synchronous` is `NORMAL`, and the busy timeout is 10s.

The options are:

* `sqlite.WithMaxReaders(int)` - The most connections which read at once. The default is 4.
* `sqlite.WithPragma(string, string)` - Set a pragma on each connection, such as `cache_size`, of `-4000`.
* `sqlite.WithFunction(string, bool, sqlite.Fn)` - Add a user-defined function, called from SQL. See [Functions](#functions).
* `sqlite.WithBind(string, any)` - Set a bind variable for every query.
* `sqlite.WithTracer(trace.Tracer)` - Set an OpenTelemetry tracer. See [Tracing](#tracing).

## Executing Statements

To execute a statement, use `Exec`:

```go
if err := db.Exec(ctx, `CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)`); err != nil {
  panic(err)
}
```

Values are bound with `With`, which returns a connection with them. See [Bindings](#bindings).

```go
if err := db.With("table", "test", "name", name).Exec(ctx, `INSERT INTO ${"table"} (name) VALUES (@name)`); err != nil {
  panic(err)
}
```

A query is one statement, or several, which are executed in turn.

## Bindings

A `Bind` is a set of named values, which a statement uses in two ways:

* **Arguments:** `@key`, `:key` and `$key` are bound by SQLite, so a value is never SQL. Use these for values.
* **Substitutions:** `${key}` is replaced with the value, as it is, before the statement is prepared. Use these for SQL, such as a table's name, or a clause:
  * `${key}` - The value, as it is, such as `ORDER BY name`.
  * `${"key"}` - The value, as an identifier, such as `"my table"`, with its `"` escaped.
  * `${'key'}` - The value, as a string, such as `'it''s'`. A `[]string` is a list of them, `'a','b'`, such as for `IN (${'list'})`.

Binds are layered:

* **Every query:** `sqlite.WithBind(key, value)`, when the database is opened.
* **A connection:** `conn.With(key, value, ...)` returns a connection with them, as well as its own. Transactions within it have them too.
* **An operation:** a `Selector` or `Writer` sets them on the `Bind` it's given, which is a copy, so operations don't change each other's.

Within a `Selector` or `Writer`:

* `bind.Set(key, value)` - Set a value, and return its argument, `@key`.
* `bind.Get(key)`, `bind.Has(key)` and `bind.Del(key)` - Get, check and delete a value.
* `bind.Append(key, value)` and `bind.Join(key, sep)` - Build a list, such as of a `WHERE` clause's conditions, and join it.
* `bind.Query(key)` - The SQL of a saved query. See [Saved Queries](#saved-queries).

For example, a `Selector` which filters rows by what's set:

```go
func (req MyListRequest) Select(bind *sqlite.Bind, op sqlite.Op) (string, error) {
  if req.Name != "" {
    bind.Append("where", "name = "+bind.Set("name", req.Name))
  }
  if len(req.Tags) > 0 {
    bind.Append("where", "tag IN (SELECT value FROM json_each("+bind.Set("tags", req.Tags)+"))")
  }
  if bind.Has("where") {
    bind.Set("where", "WHERE "+bind.Join("where", " AND "))
  } else {
    bind.Set("where", "")
  }
  req.OffsetLimit.Bind(bind, 100)
  return `SELECT name FROM mytable ${where} ORDER BY name`, nil
}
```

The conditions are SQL, which is substituted, and their values are arguments, which are bound. `req.Tags`, a `[]string`, is bound as JSON, which `json_each` reads.

## Implementing Get

To get a row, implement a `Selector` and a `Reader`:

```go
type MyObject struct {
  Id   int64
  Name string
}

// Reader - scan the row into the object
func (obj *MyObject) Scan(row sqlite.Row) error {
  return row.Scan(&obj.Id, &obj.Name)
}

// Selector - select the row
func (obj MyObject) Select(bind *sqlite.Bind, op sqlite.Op) (string, error) {
  switch op {
  case sqlite.Get:
    bind.Set("id", obj.Id)
    return `SELECT id, name FROM mytable WHERE id = @id`, nil
  }
  return "", sqlite.ErrNotImplemented
}

func main() {
  // ...
  var obj MyObject
  if err := db.Get(ctx, &obj, MyObject{Id: 1}); errors.Is(err, sqlite.ErrNotFound) {
    // No row
  } else if err != nil {
    panic(err)
  }
}
```

## Implementing List

List rows, a page at a time, with an offset and limit:

```go
type MyListRequest struct {
  sqlite.OffsetLimit
  Name string
  Tags []string
}

type MyList struct {
  Count uint64
  Names []string
}

// Reader - called with each row; note it's a pointer receiver
func (list *MyList) Scan(row sqlite.Row) error {
  var name string
  if err := row.Scan(&name); err != nil {
    return err
  }
  list.Names = append(list.Names, name)
  return nil
}

// ListReader - optionally, the count of all the rows
func (list *MyList) ScanCount(row sqlite.Row) error {
  return row.Scan(&list.Count)
}

// Selector - select the rows, and bind the offset and limit, of at most 100
func (req MyListRequest) Select(bind *sqlite.Bind, op sqlite.Op) (string, error) {
  switch op {
  case sqlite.List:
    req.OffsetLimit.Bind(bind, 100)
    return `SELECT name FROM mytable ORDER BY name`, nil
  }
  return "", sqlite.ErrNotImplemented
}
```

* **The offset and limit** are `${offsetlimit}`, which List adds after the query, so don't include them in it.
* **The count** is of the query, without them, and is read in the same transaction as the rows.
* **No rows** isn't an error.

## Implementing Insert

To insert a row, implement a `Writer`:

```go
// Writer - bind insert parameters
func (obj MyObject) Insert(bind *sqlite.Bind) (string, error) {
  bind.Set("name", obj.Name)
  return `INSERT INTO mytable (name) VALUES (@name) RETURNING id, name`, nil
}

// Writer - bind update parameters
func (obj MyObject) Update(bind *sqlite.Bind) error {
  bind.Set("name", obj.Name)
  return nil
}

func main() {
  // ...
  obj := MyObject{Name: "hello"}
  if err := db.Insert(ctx, &obj, obj); err != nil {
    panic(err)
  }
  fmt.Println("Inserted with ID:", obj.Id)
}
```

The `RETURNING` clause returns the inserted row, including generated values, such as its `INTEGER PRIMARY KEY`.

## Implementing Patch

To update rows, implement a `Selector`, of `Update`, and a `Writer`:

```go
func (obj MyObject) Select(bind *sqlite.Bind, op sqlite.Op) (string, error) {
  bind.Set("id", obj.Id)
  switch op {
  case sqlite.Get:
    return `SELECT id, name FROM mytable WHERE id = @id`, nil
  case sqlite.Update:
    return `UPDATE mytable SET name = @name WHERE id = @id RETURNING id, name`, nil
  case sqlite.Delete:
    return `DELETE FROM mytable WHERE id = @id RETURNING id, name`, nil
  }
  return "", sqlite.ErrNotImplemented
}

func main() {
  // ...
  obj := MyObject{Id: 1, Name: "updated"}
  if err := db.Update(ctx, &obj, obj, obj); err != nil {
    panic(err)
  }
}
```

The `Writer`'s `Update` is called before the `Selector`, so the `Selector` can choose its SQL from what's updated.

## Implementing Delete

To delete rows, implement a `Selector`, of `Delete`, as above:

```go
var deleted MyObject
if err := db.Delete(ctx, &deleted, MyObject{Id: 1}); err != nil {
  panic(err)
}
fmt.Println("Deleted:", deleted.Name)
```

The `RETURNING` clause is optional. Without it, pass a nil `Reader`.

## Saved Queries

Queries can be kept in a file of SQL, apart from the Go which uses them, with a key before each:

```sql
-- mytable.create
-- The table of names, which is created once
CREATE TABLE IF NOT EXISTS mytable (
  id   INTEGER PRIMARY KEY,
  name TEXT NOT NULL -- not empty
);

-- mytable.get
SELECT id, name FROM ${"table"} WHERE id = @id;

-- mytable.insert
INSERT INTO ${"table"} (name) VALUES (@name) RETURNING id, name;
```

* **Keys** are names, separated by dots, such as `mytable.get`, on a line of their own, after `--`.
* **Comments** are removed: lines which are only a comment, such as a description, and comments at the end of lines. `--` within strings, quoted identifiers and block comments is kept.
* **The `;`** which ends each statement is removed, so it can be within another, such as `List`'s count.
* **Keys are unique** in a file. A duplicate is an error.

The queries are read with `NewQueries`, and bound to a connection with `WithQueries`, so its operations have them:

```go
//go:embed queries.sql
var sql string

queries, err := sqlite.NewQueries(strings.NewReader(sql))
if err != nil {
  panic(err)
}
conn := db.WithQueries(queries).With("table", "mytable")

// Create the table
if err := conn.Exec(ctx, queries.Query("mytable.create")); err != nil {
  panic(err)
}
```

A `Selector` or `Writer` returns a query by its key, with `bind.Query`, which also names its span, when it's [traced](#tracing):

```go
func (obj MyObject) Select(bind *sqlite.Bind, op sqlite.Op) (string, error) {
  switch op {
  case sqlite.Get:
    bind.Set("id", obj.Id)
    return bind.Query("mytable.get"), nil
  }
  return "", sqlite.ErrNotImplemented
}
```

Its `${key}` substitutions are replaced once, as it's executed, so a value with a `$` in it, such as a path, is kept as it is.

## Transactions

Transactions are executed within a function, with `Tx`:

```go
if err := db.Tx(ctx, func(conn sqlite.Conn) error {
  if err := conn.Exec(ctx, `INSERT INTO test (name) VALUES ('hello')`); err != nil {
    return err
  }
  return conn.Insert(ctx, &obj, obj)
}); err != nil {
  panic(err)
}
```

* **Committed** when the function returns nil, and **rolled back** when it returns an error, panics, or its context is done.
* **Begun immediately**, on the writer, so another transaction waits for it.
* **Nested:** a transaction within one is a savepoint, which is rolled back on its own.
* **Use its `conn`:** within a transaction, the pool's writer waits for it to end. Reads on the pool don't wait, and see what's committed before it.

`Bulk` is a transaction, as `Tx` is, which prepares each statement once, and reuses it, such as for many inserts. It's about twice as fast:

```go
err := db.Bulk(ctx, func(conn sqlite.Conn) error {
  for _, obj := range objs {
    if err := conn.Insert(ctx, nil, obj); err != nil {
      return err
    }
  }
  return nil
})
```

## Values

The [`values`](values) package converts Go values to SQLite's, and back:

| Go | SQLite |
| --- | --- |
| `nil`, and nil pointers | `NULL` |
| `bool` | `INTEGER`, 0 or 1 |
| integers, and floats | `INTEGER` and `REAL` |
| `string` | `TEXT` |
| `[]byte`, and arrays of bytes | `BLOB` |
| `time.Time` | `TEXT`, of RFC 3339 |
| `time.Duration` | `INTEGER`, of nanoseconds |
| `driver.Valuer` | its value |
| `encoding.TextMarshaler`, such as `uuid.UUID` and `netip.Addr` | `TEXT` |
| `encoding.BinaryMarshaler` | `BLOB` |
| `json.Marshaler`, slices, maps and structs | `TEXT`, of JSON |

* **Named types**, such as `type Status string`, are of their kind, and pointers are of their values.
* **Scanning** reverses it, into those types, and `sql.Scanner` and `any`. A value which doesn't fit, `NULL` into a value which isn't a pointer, slice or map, and text which isn't a number or time, are errors.
* **Times** are read as SQLite's date and time functions read them: RFC 3339 or `YYYY-MM-DD HH:MM:SS` text, Unix seconds, or Julian days.
* **JSON** is text, so SQLite's json functions read it, such as `json_each`.

## Extensions

Importing a subfolder of `ncruces` adds its extension to each connection:

```go
import (
  _ "github.com/mutablelogic/go-sqlite/ncruces/fts5"
)
```

* `ncruces/fts5` - Full-text search, of `fts5` virtual tables.
* `ncruces/rtree` - Spatial indexes, of `rtree` and `geopoly` virtual tables.
* `ncruces/vec1` - Vector search, of `vec1` virtual tables.

For example, with `fts5`:

```go
db.Exec(ctx, `CREATE VIRTUAL TABLE docs USING fts5(title, body)`)
db.Get(ctx, &doc, Search{Query: "rain"}) // SELECT rowid, title FROM docs WHERE docs MATCH @query ORDER BY rank
```

## Functions

`sqlite.WithFunction` adds user-defined functions, called from SQL:

```go
db, err := ncruces.Open(ctx, path,
  sqlite.WithFunction("double", true, sqlite.Func1(func(n int64) (int64, error) {
    return n * 2, nil
  })),
  sqlite.WithFunction("elapsed", true, sqlite.Func2(func(from, to time.Time) (time.Duration, error) {
    return to.Sub(from), nil
  })),
)
```

```sql
SELECT double(21)                                          -- 42
SELECT elapsed('2026-11-18 11:00', '2026-11-18 12:30')     -- 5400000000000, 90 minutes
```

* **Arity:** `Func0`, `Func1`, `Func2` and `Func3` are of no, one, two and three arguments. `FuncN` is of any number of one type, such as `FuncN(func(args ...any) (any, error))`.
* **Arguments** are scanned, as rows are, so they're of any type a row is scanned into, such as `time.Time`, a slice, of JSON, or a pointer, which is nil when it's NULL. An argument which isn't of its type is an error, which says which.
* **Results** are converted as [Values](#values) are.
* **Errors** are the query's.
* **Deterministic** functions always return the same result for the same arguments, so they can be used in indexes.

## Schema Versions

`sqlite.Version` and `sqlite.SetVersion` read and set SQLite's `user_version`, so a schema is created, and then migrated, as it changes:

```go
err := db.Tx(ctx, func(conn sqlite.Conn) error {
  version, err := sqlite.Version(ctx, conn)
  if err != nil {
    return err
  }
  if version < 1 {
    // ...create the schema...
    if err := sqlite.SetVersion(ctx, conn, 1); err != nil {
      return err
    }
  }
  if version < 2 {
    // ...migrate it from version 1...
    if err := sqlite.SetVersion(ctx, conn, 2); err != nil {
      return err
    }
  }
  return nil
})
```

Each step sets its own version, so a database which is already newer isn't changed, and one which fails part way is rolled back to the version it was.

## Errors

Errors are of these kinds, with `errors.Is`:

* `sqlite.ErrNotFound` - Get found no row, or Insert, Update or Delete, with a `Reader`, returned none.
* `sqlite.ErrConflict` and `sqlite.ErrUniqueViolation` - A unique or primary key constraint.
* `sqlite.ErrBadParameter` - A foreign key, not null or check constraint, with `ErrForeignKeyViolation`, `ErrNotNullViolation` or `ErrCheckViolation`, or a value which can't be bound.
* `sqlite.ErrNotAvailable` - The database is busy, locked, or closed.
* `sqlite.ErrDatabase` - Any of SQLite's errors.

When a query is interrupted, or stops waiting for a lock, as its context is done, the error is also the context's, such as `context.DeadlineExceeded`.

For example, `sqlite.HTTPError` converts an error to an HTTP error, of its status, which a handler returns, with go-server's `httpresponse`:

```go
func (h *Handler) Get(w http.ResponseWriter, r *http.Request, id int64) error {
  var obj MyObject
  if err := h.db.Get(r.Context(), &obj, MyObject{Id: id}); err != nil {
    return httpresponse.Error(w, sqlite.HTTPError(err))
  }
  return httpresponse.JSON(w, http.StatusOK, 2, obj)
}
```

| Error | Status |
| --- | --- |
| `sqlite.ErrNotFound` | 404 Not Found |
| `sqlite.ErrBadParameter`, such as a foreign key, not null or check constraint | 400 Bad Request |
| `sqlite.ErrConflict`, such as a unique constraint | 409 Conflict |
| `sqlite.ErrNotImplemented` | 501 Not Implemented |
| `sqlite.ErrNotAvailable`, such as a busy or closed database | 503 Service Unavailable |
| Others | 500 Internal Server Error |

## Tracing

`sqlite.WithTracer` sets an OpenTelemetry tracer, of a span for each query:

```go
import (
  otel "go.opentelemetry.io/otel"
)

db, err := ncruces.Open(ctx, path,
  sqlite.WithTracer(otel.Tracer("myapp")),
)
```

A span's parent is the operation's context, so a query is within the span of the request which made it.


* **Named** by the query's name, of `bind.Query`, such as `mytable.get`, or `sqlite.query`.
* **Transactions:** `sqlite.transaction`, `sqlite.commit` and `sqlite.rollback`, of savepoints too. Opening the database is `sqlite.open`.
* **Attributes:** `db.system`, `db.statement`, and the arguments it bound, as `arg.` and their names.
* **Errors** are the span's, except not found.

For example, a `Get` of a saved query, `mytable.get`, which binds `@id`, is a span:

```text
mytable.get
  db.system     sqlite
  db.statement  SELECT id, name FROM "mytable" WHERE id = @id
  arg.id        1
```

## Changes

Consumers often need to know when data changes, without polling for it. For example, a web page can update as soon as a row is added. Or a server can stream changes to its clients, or refresh a cache.

`sqlite.WithBroadcaster` publishes each row which is inserted, updated or deleted, once it's committed. It works with go-server's [broadcaster](https://github.com/mutablelogic/go-server/tree/main/broadcaster), or anything with a non-blocking `Publish(schema.Change) int`:

```go
import (
  broadcaster "github.com/mutablelogic/go-server/broadcaster"
)

changes := broadcaster.New()
db, err := ncruces.Open(ctx, path,
  sqlite.WithBroadcaster(changes, "mytable"),
)
```

* **A change** is its action (`create`, `update` or `delete`), its database (`main`), its table, and its rowid as `Row`.
* **Tables:** the tables named, or all of them. The shadow tables of virtual tables, such as fts5's, are left out.
* **Committed:** changes are published after the transaction commits, so a watcher reads what changed. Rolled-back transactions and savepoints publish nothing.
* **Once per row:** a row written several times in a transaction is one change. Created then deleted is none.
* **Many rows:** over `sqlite.MaxChanges` rows of a table in one transaction are one change of the table, without a `Row`.
* **Every row:** rows deleted by a foreign key's cascade, by a `DELETE` without a `WHERE`, or replaced by `INSERT OR REPLACE`, are published. A row whose rowid changes is a delete and a create.
* **Publish** is called while the writer is held, so it mustn't block or query the database. A watcher reads changes on its own goroutine, as go-server's broadcaster's do.
* **Failed statements:** in a transaction, rows undone by a failed statement aren't published, and rows kept by `INSERT OR FAIL` are.
* **Not seen:** writes by other processes, and of tables `WITHOUT ROWID`. Savepoints written by hand in SQL aren't tracked, so their rolled-back rows are still published.

## License & Distribution

This software is licensed under the **Apache License 2.0**.

> **go-sqlite**  
> [https://github.com/mutablelogic/go-sqlite/](https://github.com/mutablelogic/go-sqlite/)  
> Copyright (c) 2017-2026 The go-sqlite authors, All rights reserved.

Under the Apache License 2.0, you are free to use, modify, and distribute this software for any purpose, including commercial applications. **Attribution requirement:** when redistributing this software or derivative works, in source or binary form, include the copyright notice above, and a copy of the [LICENSE](LICENSE) and [NOTICE](NOTICE) files. If you modify the code, you must state your changes. See the [Apache License 2.0](http://www.apache.org/licenses/LICENSE-2.0) for complete terms.

When distributing binaries which include this software, also comply with the licenses of the libraries which are compiled into them. Their copyright notices are in the [NOTICE](NOTICE) file.

| Library | License |
| --- | --- |
| [SQLite](https://sqlite.org/copyright.html), including its fts5, rtree and vec1 extensions | Public domain |
| [ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3) | MIT |
| [ncruces/go-sqlite3-wasm](https://github.com/ncruces/go-sqlite3-wasm), SQLite translated to Go | MIT-0, with the original licenses of what's translated |
| [ncruces/julianday](https://github.com/ncruces/julianday) | MIT |
| [OpenTelemetry for Go](https://github.com/open-telemetry/opentelemetry-go) | Apache 2.0 |
| [go-server](https://github.com/mutablelogic/go-server) | Apache 2.0 |
| [xxhash](https://github.com/cespare/xxhash) | MIT |
| [golang.org/x/sys](https://pkg.go.dev/golang.org/x/sys) | BSD 3-clause |

The [`test`](test) package, which is for tests, also uses [testify](https://github.com/stretchr/testify) (MIT), the OpenTelemetry SDK and [logr](https://github.com/go-logr/logr) (Apache 2.0), [google/uuid](https://github.com/google/uuid) (BSD 3-clause) and [yaml](https://github.com/yaml/go-yaml) (MIT and Apache 2.0).
