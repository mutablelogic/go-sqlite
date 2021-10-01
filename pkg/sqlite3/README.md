# sqlite3 package

This package provides a high-level interface for [sqlite3](http://sqlite.org/)
including connection pooling, transaction and execution management.

This package is part of a wider project, `github.com/mutablelogic/go-sqlite`.
Please see the [module documentation](https://github.com/mutablelogic/go-sqlite/blob/master/README.md)
for more information.

## Building

This module does not include a full
copy of __sqlite__ as part of the build process, but expect a `pkgconfig`
file called `sqlite3.pc` to be present (and an existing set of header
files and libraries to be available to link against, of course).

In order to locate the correct installation of `sqlite3` use two environment variables:

  * `PKG_CONFIG_PATH` is used for locating `sqlite3.pc`
  * `DYLD_LIBRARY_PATH` is used for locating the dynamic library when testing and/or running

On Macintosh with homebrew, for example:

```bash
[bash] brew install sqlite3
[bash] git clone git@github.com:djthorpe/go-sqlite.git
[bash] cd go-sqlite
[bash] go mod tidy
[bash] SQLITE_LIB="/usr/local/opt/sqlite/lib"
[bash] PKG_CONFIG_PATH="${SQLITE_LIB}/pkgconfig" DYLD_LIBRARY_PATH="${SQLITE_LIB}" go test -v ./pkg/sqlite3
```

On Debian Linux you shouldn't need to locate the correct path to the sqlite3 library:

```bash
[bash] sudo apt install libsqlite3-dev
[bash] git clone git@github.com:djthorpe/go-sqlite.git
[bash] cd go-sqlite
[bash] go mod tidy
[bash] go test -v ./pkg/sqlite3
```

There are some examples in the `cmd` folder of the main repository on how to use
the package, and various pseudo examples in this document.

## Contributing & Distribution

Please do file feature requests and bugs [here](https://github.com/mutablelogic/go-sqlite/issues).
The license is Apache 2 so feel free to redistribute. Redistributions in either source
code or binary form must reproduce the copyright notice, and please link back to this
repository for more information:

> Copyright (c) 2021, David Thorpe, All rights reserved.

## Overview

The package includes:

  * A Connection __Pool__ for managing connections to sqlite3 databases;
  * A __Connection__ for executing queries;
  * An __Auth__ interface for managing authentication and authorization;
  * A __Cache__ for managing prepared statements.

It's possible to create custom functions (both in a scalar and aggregate context)
and use perform streaming read and write operations on large binary (BLOB) objects.

In order to create a connection pool, you can create a default pool using the `NewPool`
method:

```go
package main

import (
  sqlite3 "github.com/mutablelogic/go-sqlite/pkg/sqlite3"
)

func main() {
  pool, err := sqlite3.NewPool(path, nil)
  if err != nil {
    panic(err)
  }
  defer pool.Close()

  // Onbtain a connection from pool, put back when done
  conn := pool.Get()
  defer pool.Put(conn)

  // Enumerate the tables in the database
  tables := conn.Tables()

  // ...
}
```

In this example, a database is opened and the `Get` method obtains a connection
to the databaseand `Put` will return it to the pool. The `Tables` method enumerates 
the tables in the database. The following sections outline how to interact with the
`sqlite3` package in more detail.

## Connection Pool

A __Pool__ is a common  pattern for managing connections to a database, where there
is a limited number of connections available with concurrent accesses to the database.
The pool can be created in two ways:

  1. `sqlite3.NewPool(path string, errs chan<- error) (*Pool, error)` creates a standard pool
     for a single database (referred to by file path or as `:memory:` for an in-memory database);
  2. `sqlite3.OpenPool(config sqlite3.PoolConfig, errs chan<- error) (*Pool, error) ` creates
     a pool with a configuration; more information on the configuration can be found below.

In either case, the second argument can be `nil` or a channel for receiving errors. Errors are
received in this way so that the pool method `Get` can return `nil` if an error occurs, but
errors can still be reported for debugging.

### Pool Configuration

By default the pool will create a maximum of 5 simultaneous connections to the database.
However, you can use the `NewConfig()` method with options to alter the default configuration.
For example,


```go
package main

import (
  sqlite3 "github.com/mutablelogic/go-sqlite/pkg/sqlite3"
)

func main() {
  cfg := sqlite3.NewConfig().SetMaxConnections(100)
  pool, err := sqlite3.OpenPool(cfg, nil)
  if err != nil {
    panic(err)
  }
  defer pool.Close()

  // ...
}
```

The different options to modify the default configuration are:

  * `func (PoolConfig) WithAuth(SQAuth)` sets the authentication and authorization
    interface for any executed statements. More information about this interface can 
    be found in the section below.
  * `func (PoolConfig) WithTrace(TraceFunc)` sets a trace function for the pool, so that
    you can monitor the activity executing statements. More information about this
    can be found in the section below.
  * `func (PoolConfig) WithMaxConnections(int)` sets the maximum number of connections
    to the database. Setting a value of `0` will use the default number of connections.
  * `func (PoolConfig) WithSchema(name, path string)` adds a database schema to the
    connection pool. One schema should always be named `main`. Setting the path argument
    to `:memory:` will set the schema to an in-memory database, otherwise the schema will
    be read from disk.

### Getting a Connection

Once you have created a pool, you can obtain a connection from the pool using the `Get` method,
which should always be paired with a `Put` call:

```go
package main

import (
  sqlite3 "github.com/mutablelogic/go-sqlite/pkg/sqlite3"
)

func main() {
  pool, err := sqlite3.NewPool(":memory:", nil)
  if err != nil {
    panic(err)
  }
  defer pool.Close()

  // Get connection
  conn := pool.Get()
  if conn == nil {
    panic("No connection")
  }
  defer pool.Put(conn)

  // ...
}
```

The `Get` method may return nil if the maximum number of connections has been reached.
Once a connection has been `Put` back into the pool, it should no longer be used (there
is nothing presently to prevent use of a connection after it has been `Put` back, but
it could be added in later).

### Example code for reporting errors

In general you should pass a channel for receiving errors. Here is some sample code
you can use for doing this:

```go
func createErrorChannel() (chan<- error, context.CancelFunc) {
  var wg sync.WaitGroup
  errs := make(chan error)
  ctx, cancel := context.WithCancel(context.Background())
  wg.Add(1)
  go func() {
    defer wg.Done()
      for {
        select {
        case <-ctx.Done():
          close(errs)
          return
        case err := <-errs:
          if err != nil {
            // Replace the following line
            log.Print(err)
          }
        }
      }
  }()

  return errs, func() { cancel(); wg.Wait() }
}
```

The function will return the error channel and a cancel function. The cancel function
should only be called after the pool has been closed to ensure that the pool does not
try and report errors to a closed error channel.

## Transactions and Queries

There are two connection methods for executing queries:

### Execution outside a transaction

The function `func (SQConnection) Exec(SQStatement, SQExecFunc) error` executes a
callback function with the result of the query. For example,

```go
package main

import (
  sqlite3 "github.com/mutablelogic/go-sqlite/pkg/sqlite3"
  . "github.com/mutablelogic/go-sqlite/pkg/sqlang"
)

func main() {
  // ...
  conn.Exec(Q("PRAGMA module_list"), func(row, cols []string) bool {
    fmt.Println(row)
    return false
  })
}
```

Use this method to run statements which don't need committing or rolling back
on errors, or which only need text information returned.

### Execution in a transaction

On the whole you will want to operate the database inside tansactions. In order
to do this, call the function `func (SQConnection) Do(context.Context, SQFlag, SQTxnFunc) error` with a callback function of type `SQTxnFunc` as an argument. The signature of the transaction
function is `func (SQTransaction) error` and if it returns any error, the transaction will be rolled back, otherwise any modifications within the transaction will be committed.

You can pass zero (`0`) for the `SQFlag` argument if you don't need to use any flags, or else pass any combination of the following flags:

  * `SQLITE_TXN_DEFAULT` Default (deferred) transaction flag (can be omitted if not needed)
  * `SQLITE_TXN_IMMEDIATE` Immediate transaction
  * `SQLITE_TXN_EXCLUSIVE` Exclusive transaction
  * `SQLITE_TXN_NO_FOREIGNKEY_CONSTRAINTS` Drop foreign key constraints within the transaction

More information about different types of transactions is documented [here](https://www.sqlite.org/lang_transaction.html).

For example,

```go
package main

import (
  sqlite3 "github.com/mutablelogic/go-sqlite/pkg/sqlite3"
  . "github.com/mutablelogic/go-sqlite/pkg/sqlang"
)

func main() {
  // ...
  conn.Do(context.Background(),0,func(txn SQTransaction) error {
    if _, err := txn.Query(N("test").Insert().WithDefaultValues()); err != nil {
      return err
    }
    // Commit transaction
    return nil
  })
}
```

## Custom Types

TODO

## Custom Functions

TODO

## Authentication and Authorization

TODO

## Pool Status

There are two methods which can be used for getting and setting pool status:

  * `func (SQPool) Cur() int` returns the current number of connections in the pool;
  * `func (SQPool) SetMax(n int)` sets the maximum number of connections allowed in the pool.
    This will not affect the number of connections currently in the pool, however.

## Reading and Writing Large Objects

TODO

## Backup

TODO
