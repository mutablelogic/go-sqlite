# sqlite3 bindings

This package provides bindings for [sqlite3](http://sqlite.org/) which
I am sure is very similar to other bindings! In my defence :-) learning more
about the internals of sqlite is a good exercise in itself.

The bindings do not add a lot of functionality beyond replicating the API
in a more golang pattern. They are bindings afterall. It is assumed that
a separate package would be used to provide a more useful API, including
connection pooling, transaction and execution management, and so forth.

This package is part of a wider project, `github.com/djthorpe/go-sqlite`.
Please see the [module documentation](https://github.com/djthorpe/go-sqlite/blob/master/README.md)
for more information.

## Building

Unlike some of the other bindings I have seen, these do not include a full
copy of __sqlite__ as part of the build process, but expect a `pkgconfig`
file called `sqlite.pc` to be present (and an existing set of header
files and libraries to be available to link against, of course).

In order to locate the __pkgconfig__ file in a non-standard location, use
the `PKG_CONFIG_PATH` environment variable. For example, I have installed
sqlite using `brew install sqlite` and this is how I run the tests:

```bash
[bash] git clone git@github.com:djthorpe/go-sqlite.git
[bash] cd go-sqlite
[bash] go mod tidy
[bash] PKG_CONFIG_PATH="/usr/local/opt/sqlite/lib/pkgconfig" go test -v ./sys/sqlite3
```

There are some examples in the `cmd` folder of the main repository on how to use
the bindings, and various pseudo examples in this document.

## Contributing & Distribution

Please do file feature requests and bugs [here](https://github.com/djthorpe/go-sqlite/issues).
The license is Apache 2 so feel free to redistribute. Redistributions in either source
code or binary form must reproduce the copyright notice, and please link back to this
repository for more information:

> Copyright (c) 2021, David Thorpe, All rights reserved.

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

Five methods will execute a query:

  * `func (*ConnEx) Exec(string, func (row, cols []string) bool) error` will execute
    one or more SQL queries (separated by a semi-colon) without bound parameters, 
    and invoke a function callback with the results. Return `true` from this 
    callback to abort any subsequent results being returned;
  * `func (*ConnEx) ExecEx(string, func (row, cols []string) bool,...interface{}) error` will execute
    one or more SQL queries (separated by a semi-colon) with bound parameters, 
    and invoke a function callback with the results. Return `true` from this 
    callback to abort any subsequent results being returned;
  * `func (*ConnEx) Begin(SQTransaction) error` will start a transaction. Include
    an argument `sqlite3.SQLITE_TXN_DEFAULT`, `sqlite3.SQLITE_TXN_IMMEDIATE` or
    `sqlite3.SQLITE_TXN_EXCLUSIVE` to set the transaction type;
  * `func (*ConnEx) Commit() error` will commit a transaction;
  * `func (*ConnEx) Rollback() error` will rollback a transaction.

The following methods return and set information about the connection. These can be
used for both `*Conn` and `*ConnEx` types:  

  * `func (*Conn) Filename(string) string` returns the filename for an attached
    database;
  * `func (*Conn) Readonly(string) bool` returns the readonly status for an attached
    database;
  * `func (*Conn) Autocommit() bool ` returns false if the connection is in a transaction;
  * `func (*Conn) LastInsertId() int64` returns the `RowId` of the last row inserted;
  * `func (*Conn) Changes() int64` returns the number of rows affected by the last query;

Finally,

  * `func (*Conn) Interrupt()` interrupts any running queries for the connection.

When errors are returned from any methods, their error message is 
[documented here](https://www.sqlite.org/rescode.html). The result codes can
be printed or cast to an integer or other numeric type as necessary.

## Statements & Bindings

In order to execute a query or set of queries, they first need to be prepared.
The method `func (*ConnEx) Prepare(q string) (*StatementEx, error)` returns
a prepared statement. It is your responsibility to call `func (*ConnEx) Close() error`
on the statement when you are finished with it. For example,

```go
package main

import (
    "github.com/djthorpe/go-sqlite/sys/sqlite3"
)

func main() {
    path := "..."
    db, err := sqlite3.OpenPathEx(path, sqlite3.SQLITE_OPEN_CREATE, "")
    // ...
    stmt, err := db.Prepare("SELECT * FROM table")
    if err != nil {
        // ...
    }
    defer stmt.Close()
    // ...
}
```

You can then either:

  * Set bound parameters using `func (*StatementEx) Bind(...interface{}) error`
    or `func (*StatementEx) BindNamed(...interface{}) error` to bind
    parameters to the statement, and then call `func (*StatementEx) Exec() (*Results, error)`
    with no arguments to execute the statement;
  * Or, call `func (*StatementEx) Exec(...interface{}) (*Results, error)` with bound parameters
    directly.

Any parameters which are not bound are assumed to be NULL. If your prepared statement has
multiple queries, then you can call `Exec` repeatedly until no more results are returned.
For example,

```go
package main

import (
    "github.com/djthorpe/go-sqlite/sys/sqlite3"
)

func main() {
    path := "..."
    db, err := sqlite3.OpenPathEx(path, sqlite3.SQLITE_OPEN_CREATE, "")
    // ...
    stmt, err := db.Prepare("SELECT * FROM table")
    if err != nil {
        // ...
    }
    defer stmt.Close()
    for {
        r, err := stmt.Exec()
        if err != nil {
            // Handle error
        } else if r == nil {
            // No more result queries to execute
            break
        } else {
            // Read results from query
        }
    }
}
```

### Binding Values To Prepared Statements

[Bound values](https://www.sqlite.org/c3ref/bind_blob.html) are arguments
in calls to the following methods:

  * `func (*StatementEx) Bind(...interface{}) error` to bind parameters in numerical order;
  * `func (*StatementEx) BindNamed(...interface{}) error` to bind parameters with name, value 
    pairs;
  * `func (*StatementEx) Exec(...interface{}) (*Results, error)` to bind parameters in numerical 
    order and execute the statement. If no argumet is given, previously bound parameters are used;
  * `func (*ConnEx) ExecEx(string, func (row, cols []string) bool,...interface{}) error` to 
    execute a query directly with parameters in numerical order.

Each value is translated into an sqlite type as per the following table, where N can be
8 or 16 (in the case of integers) or 32 or 64 (in the case of integers and floats):

| go             | sqlite                |
| -------------- | ----------------------| 
| `nil`          | NULL                  |
| `int`,`intN`   | INTEGER               |
| `uint`,`uintN` | INTEGER               |
| `floatN`       | FLOAT                 |
| `string`       | TEXT                  |
| `bool`         | INTEGER               |
| `[]byte`       | BLOB                  |

> It might be extended to time.Time and custom types (using marshalling) later.

In the SQL statement text input literals may be replaced by a parameter that matches one of `?`, `?N`, `:V`, `@V` or `$V`
where N is an integer and V is an alpha-numeric string. For example,

```go
package main

import (
    "github.com/djthorpe/go-sqlite/sys/sqlite3"
)

func main() {
    path := "..."
    db, err := sqlite3.OpenPathEx(path, sqlite3.SQLITE_OPEN_CREATE, "")
    // ...
    stmt, err := db.Prepare("SELECT * FROM table WHERE a=:A AND b=:B")
    if err != nil {
        // ...
    }
    defer stmt.Close()

    for {
        if err := stmt.BindNamed(":A", 100, ":B", 200); err != nil {
            // Handle error
        }
        r, err := stmt.Exec()
        if err != nil {
            // Handle error
        } else if r == nil {
            // No more result queries to execute
        } else if err := ReadResults(r); err != nil {
            // Handle error
        }
    }
}
```

## Results

Results are returned from the `Exec` method after a statement is executed. If there are no results,
then a call to `func (*Results) Next() ([]interface{},error)` will return `nil` in place of an
array of values. You should repeatedly call the `Next` method until this occurs. For example,

```go
func ReadResults(r *Results) error {
    for {
        row, err := r.Next()
        if err != nil {
            return err
        } else if row == nil {
            return nil
        }
        // Handle row
        // ...
    }
}
```

When `Next` is invoked without arguments, the values returned are interpreted as the above table
but in reverse. For example, a `NULL` value is returned as `nil`. `INTEGER` values are returned
as `int64` and `FLOAT` values are returned as `float64`. If you invoke `Next` with a slice of
`reflect.Type` then the values returned are converted to the types specified in the slice. For
example,


```go
func ReadResults(r *Results) error {
    cast := []reflect.Type{ reflect.TypeOf(bool), reflect.TypeOf(uint) }
    for {
        row, err := r.Next(cast...)
        if err != nil {
            return err
        } else if row == nil {
            return nil
        }
        // Handle row which has bool as first element and uint as second element
        // ...
    }
}
```

If a value cannot be cast by a call to `Next`, then an error is returned.

> Will be extended to time.Time and custom types (using unmarshalling) later.

Reflection on the results can be used through the following method calls:

  * `func (*Results) ColumnNames() []string` returns column names for the results
  * `func (*Results) ColumnCount() int` returns column count
  * `func (*Results) ColumnTypes() []Type` returns column types for the results
  * `func (*Results) ColumnDeclTypes() []string` returns column decltypes for the results
  * `func (*Results) ColumnDatabaseNames() []string` returns the source database schema name for the results
  * `func (*Results) ColumnTableNames() []string` returns the source table name for the results
  * `func (*Results) ColumnOriginNames() []string` returns the origin for the results

These allocate new arrays on each call so you should use them sparingly.

## User-Defined Functions

You can [define scalar and aggregate user-defined functions](https://www.sqlite.org/appfunc.html)
(and override existing ones) for use in statement execution:

  * A __scalar function__ takes zero or more argument values and returns a single value or an error;
  * An __aggregate function__ is called for every result within the grouping and then returns a single value or an error.

The types for the function calls in go are:

  * Scalar function `type StepFunc func(*Context, []*Value)`
  * Aggregate function to collate each result `type StepFunc func(*Context, []*Value)`
  * Aggregate function to finalize `type FinalFunc func(*Context)`

To register a user-defined function use the following methods:

  * `func (*ConnEx) CreateScalarFunction(string,int,bool,StepFunc) error` where the first argument is
    the name of the function, the second is the number of arguments accepted
    (or -1 for variable number of arguments), the third flag indicates that the function
    returns the same value for the same input arguments, and the fourth argument is the callback.
  * `func (*ConnEx) CreateAggregateFunction(string,int,bool,StepFunc,FinalFunc) error` has the same
    arguments as above, but the fourth and fifth arguments are the step and final callbacks.

You can register multiple calls for the same function name. See the [documentation](https://www.sqlite.org/appfunc.html)
for more information.

### Values

Values are passed to the step function callbacks and include arguments to the function. See the
[documentation](https://www.sqlite.org/c3ref/value.html) for more information. In addition
the method `func (*Value) Interface() interface{}` can be used to convert the value to a go type.

### Context

The `*Context` is passed to all the user-defined function callbacks. The context is used to store
the return value and errors. See the [documentation](https://www.sqlite.org/c3ref/context.html)
for more information. In addition, the method `func (*Context) ResultInterface(v interface{}) error`
can be called to set a go value, and returns an error if the conversion could not be
perfomed.

## Commit, Update and Rollback Hooks

TODO

## Authentication and Authorization Hook

TODO

## Binary Object (Blob IO) Interface

In addition to the standard interface which inserts, updates and deletes binary objects atomically,
it's possible to read and write data to binary objects incrementally.
The [documentation is here](https://www.sqlite.org/c3ref/blob.html).

In order to create a blob, use the SQL method `INSERT INTO table VALUES ZEROBLOB(?)` for example
with a size parameter. Then use the last inserted rowid to read and write to the blob.

  * Use `func (*Conn) OpenBlob(schema, table, column string, rowid int64, flags OpenFlags) (*Blob, error)`
    to return a handle to a blob;
  * Use `func (*Conn) OpenBlobEx(schema, table, column string, rowid int64, flags OpenFlags) (*BlobEx, error)`
    to return a handle to a blob which provides an `io.Reader` and `io.Writer` interface;
  * Use `func (*Blob) Close() error` to close the blob on either a `*Blob` or `*BlobEx` handle;
  * The method `func (*Blob) Bytes() int` returns the size of the blob;
  * The method `func (*Blob) Reopen(int64) error` opens a new row with the existing blob handle.

See the documentation for the [`io.Reader`](https://golang.org/pkg/io/#Reader) and
[`io.Writer`](https://golang.org/pkg/io/#Writer)
interfaces for more information on `Read`, `Write`, `Seek`, `ReadAt` and `WriteAt` methods.

## Backup Interface

TODO 

## Status and Limits

TODO

## Miscellaneous

Status Counters
Limits
Shared Cache Mode
