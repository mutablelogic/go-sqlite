# sqlite3 bindings

This package provides bindings for [sqlite3](http://sqlite.org/) which
I am sure is very similar to other bindings! In my defence :-) learning more
about the internals of sqlite is a good exercise in itself.

The bindings do not add a lot of functionality beyond replicating the API
in a more golang pattern. They are bindings afterall. It is assumed that
a separate package would be used to provide a more useful API.

## Building

Unlike some of the other bindings I have seen, these do not include a full
copy of __sqlite__ as part of the build process, but expect a `pkgconfig`
file called `sqlite.pc` to be present (and an existing set of header
files and libraries to be available to link against, of course).

In order to locate the __pkgconfig__ file in a non-standard location, use
the `PKG_CONFIG_PATH` environment variable. For example, to run the tests:

```bash
[bash] git clone git@github.com:djthorpe/go-sqlite.git
[bash] cd go-sqlite
[bash] go mod tidy
[bash] PKG_CONFIG_PATH="/usr/local/opt/sqlite/lib/pkgconfig" go test -v ./sys/sqlite3
```

There are some examples in the `cmd` folder of the main repository.

## Connection

The `Conn` type is a wrapper around the `sqlite3` C API, and the `ConnEx` type
also implements various callback hooks. I recommend using the `ConnEx` type
for full functionality. See 
the [associated C API docmentation](https://www.sqlite.org/cintro.html)
for more information about each method.

To open a connection to a database:

```go
package main

import (
    "github.com/djthorpe/go-sqlite/sys/sqlite3"
)

func main() {
    path := "..."
    db, err := sqlite3.OpenPathEx(path, sqlite3.SQLITE_OPEN_CREATE, "")
    if err != nil {
        t.Error(err)
    }
    defer db.Close()
    // ...
}
```

The `OpenUrlEx` version is also available which treats the first parameter as
a URL rather than a path, and 
[includes various options](https://www.sqlite.org/c3ref/open.html).

A default busy timeout for acquiring locks is set to five seconds. Change the
busy timeout or set a custom busy handler using the `SetBusyTimeout` and
`SetBusyHandler` methods. In addition, `SetProgressHandler` can be used 
to set a callback for progress during long running queries, which allows
for cancellation mid-query.

Four methods will execute a query:

  * `func (*ConnEx) Exec(string,func (row,cols []string) bool) error` will execute
    one or more SQL queries (separated by semi-colon) without bound parameters, 
    and call a function with the results. Return `true` from this method to abort
    any subsequent results being returned;
  * `func (*ConnEx) Begin(SQTransaction) error` will start a transaction. Include
    an argument `sqlite3.SQLITE_TXN_DEFAULT`, `sqlite3.SQLITE_TXN_IMMEDIATE` or
    `sqlite3.SQLITE_TXN_EXCLUSIVE` to set the transaction type;
  * `func (*ConnEx) Commit() error` will commit a transaction;
  * `func (*ConnEx) Rollback() error` will rollback a transaction.

The following methods return and set information about the connection:

  * `func (*Conn) Filename(string) string` returns the filename for an attached
    database;
  * `func (*Conn) Readonly(string) bool` returns the readonly status for an attached
    database;
  * `func (*Conn) Autocommit() bool ` returns false if the connection is in a transaction;
  * `func (*Conn) LastInsertId() int64` returns the `RowId` of the last row inserted;
  * `func (*Conn) Changes() int64` returns the number of rows affected by the last query;
  * `func (*Conn) Interrupt()` interrupts any running queries.

## Statements

TODO

## Results

TODO

## User-Defined Functions

TODO

## Commit, Update and Rollback Hooks


## Status and Limits

## Miscellaneous

