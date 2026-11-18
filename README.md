# go-sqlite

A SQLite client for Go, of the same shape as [go-pg](https://github.com/mutablelogic/go-pg). Pure Go, without cgo.

- **Drivers:** the root package works over `database/sql`. A driver opens databases and maps their errors. [`ncruces`](ncruces) uses [ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3), which is SQLite compiled to WebAssembly.
- **One writer, many readers:** writes and transactions are on one connection, and reads on several, of `WithMaxReaders`. A file's journal is a write-ahead log, so reads see what's committed while a transaction writes.
- **Queries:** named SQL from a file (`NewQueries`), `${key}` substitutions and `@key` arguments (`Bind`), and objects which read, write and select themselves (`Reader`, `Writer`, `Selector`).

## Open a database

```go
import (
	sqlite "github.com/mutablelogic/go-sqlite"
	ncruces "github.com/mutablelogic/go-sqlite/ncruces"
)

db, err := sqlite.Open(ctx, "index.db", ncruces.New())
if err != nil {
	return err
}
defer db.Close()
```

- **In memory:** `sqlite.Memory`, which is one connection.
- **Pragmas:** foreign keys are on, the journal is WAL, and the busy timeout is 10s. Others are set with `sqlite.WithPragma`.
- **Functions and extensions:** `ncruces.WithInit` calls a function on each connection, such as one which creates a function, or an extension's `Register`.

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

## Drivers

Another driver implements `sqlite.Driver`, and passes `sqlitetest.Run(t, driver)`.
