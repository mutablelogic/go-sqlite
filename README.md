# go-sqlite

A SQLite client for Go, of the same shape as [go-pg](https://github.com/mutablelogic/go-pg). Pure Go, without cgo.

- **Implementations:** the root package is the interfaces, queries and errors, without a driver. [`ncruces`](ncruces) implements them with [ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3)'s own API, which is SQLite compiled to WebAssembly.
- **One writer, many readers:** writes and transactions are on one connection, and reads on several, of `WithMaxReaders`. A file's journal is a write-ahead log, so reads see what's committed while a transaction writes.
- **Values:** [`values`](values) converts between Go and SQLite:
  - bools, integers, floats, strings and `[]byte`, of named types too, and pointers, which are NULL when nil;
  - `time.Time` as RFC 3339 text, and `time.Duration` as nanoseconds;
  - `driver.Valuer` and `sql.Scanner`, text and binary marshalers, such as `uuid.UUID` and `netip.Addr`, as text, and `json.Marshaler`;
  - slices, arrays, maps and structs as JSON text, which SQLite's `json_each` reads, and byte arrays as blobs.
- **Queries:** named SQL from a file (`NewQueries`), `${key}` substitutions and `@key` arguments (`Bind`), and objects which read, write and select themselves (`Reader`, `Writer`, `Selector`).

## Open a database

```go
import (
	sqlite "github.com/mutablelogic/go-sqlite"
	ncruces "github.com/mutablelogic/go-sqlite/ncruces"
)

db, err := ncruces.Open(ctx, "index.db")
if err != nil {
	return err
}
defer db.Close()
```

- **In memory:** `sqlite.Memory`, which is one connection.
- **Pragmas:** foreign keys are on, the journal is WAL, and the busy timeout is 10s. Others are set with `sqlite.WithPragma`.
- **Functions:** `sqlite.WithFunction` creates a Go function, which SQL calls, on each connection.
- **Extensions:** `ncruces.WithInit` calls a function with each ncruces connection, such as an extension's `Register`, or one which creates a virtual table.

## Transactions

```go
err := db.Tx(ctx, func(conn sqlite.Conn) error {
	return conn.Insert(ctx, &object, meta)
})
```

- **Rolled back** when the function returns an error, or panics.
- **Nested:** a transaction within one is a savepoint.
- **Use its `conn`:** the pool's writer waits for the transaction to end. Reads on the pool don't wait.

## Schema versions

`sqlite.Version` and `sqlite.SetVersion` read and set SQLite's `user_version`, so a schema is migrated as it changes.

## Implementations

Another implementation opens a `sqlite.PoolConn`, and passes `test.Run(t, open)`, as `ncruces.Open` does. It converts values with `values.Set` and `values.Scan`, by adapting where SQLite gives and takes them.
