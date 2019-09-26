# sqlite

[![CircleCI](https://circleci.com/gh/djthorpe/sqlite/tree/master.svg?style=svg)](https://circleci.com/gh/djthorpe/sqlite/tree/master)

This repository contains a higher-level interface to SQLite, to provide database persistence. It implements two components for [gopi](http://github.com/djthorpe/gopi) and some example programs in the `cmd` folder. The repository depends on golang version 1.12 and above (in order to support modules).

## Components

The gopi components provided by this repository are:

| Component Path | Description                            | Component Name |
| -------------- | -------------------------------------- |--------------- |
| sys/sqlite     | SQL Database persisence using sqlite   | db/sqlite      |
| sys/sqlite     | SQL Language Builder                   | db/sqlang      |
| sys/sqobj      | Lightweight object layer               | db/sqobj       |

## Building and installing examples

You may not have sqlite3 installed, so on Debian (and Raspian) Linux you
can install using the following commands:

```bash
bash% sudo apt install sqlite3 sqlite3-doc
```

There is a makefile which can be used for testing and installing bindings and examples:

```
bash% git clone git@github.com:djthorpe/sqlite.git
bash% cd sqlite
bash% make all
```

The resulting binaries are as follows. Use the `-help` flag to see the different options for each:

  * `sq_import` Imports data from CSV files into an SQLite database
  * `fs_indexer` Indexe files into a database to implement a full-text search

You can also build these tools separately using `make sq_import` and `make fs_indexer` respectively.

## Using `db/sqlite`

Database persistence is implemented using the `db/sqlite` component. Here's an example of how to use the component:

```go
package main

import (
  // Frameworks
  gopi "github.com/djthorpe/gopi"
  sqlite "github.com/djthorpe/sqlite"

  // Modules
  _ "github.com/djthorpe/gopi/sys/logger"
  _ "github.com/djthorpe/sqlite/sys/sqlite"
)

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
  db := app.ModuleInstance("db/sqlite").(sqlite.Connection)
  if rows, err := db.QueryOnce("SELECT * FROM table"); err != nil {
    return err
  } else {
    for {
      if row := rows.Next(); row == nil {
        break
      } else {
        fmt.Println(sqlite.RowString(row))
      }
    }
  }
  return nil
}

func main() {
  os.Exit(gopi.CommandLineTool2(gopi.NewAppConfig("db/sqlite"), Main))
}
```

### Returning database information

The `Connection` component includes methods for returning database
information (`Version`, `Tables` and `ColumnsForTable`) and executing
actions on the data:

```go
type Connection interface {
  // Return version number of the underlying sqlite library
  Version() string

  // Return attached schemas
  Schemas() []string

  // Return table names for the main schema. The extended version
  // returns table names for a different schema, or can include
  // temporary tables
  Tables() []string
  TablesEx(schema string, include_temporary bool) []string
  
  // Return the columns for a table
  ColumnsForTable(name, schema string) ([]Column, error)
}
```

The table column interface is specified as follows:

```go
type Column interface {
  // Name returns the column name
  Name() string
  
  // DeclType returns the declared type for the column
  DeclType() string
  
  // Nullable returns true if  a column cell can include the NULL value
  Nullable() bool
  
  // PrimaryKey returns true if the column is part of the primary key
  PrimaryKey() bool
}
```

### Database actions

The `Query` and `QueryOnce` methods return a `Rows` object which can
be used to iterate through the set of results, whereas the `Do` and
`DoOnce` methods can be used for inserting, updating and deleting
table rows (and executing other database actions which don't return
a set of results):

```go
type Connection interface {
  Transaction
}

type Transaction interface {
  // Execute statement (without returning the rows)
  Do(Statement, ...interface{}) (Result, error)
  DoOnce(string, ...interface{}) (Result, error)

  // Query to return rows from result
  Query(Statement, ...interface{}) (Rows, error)
  QueryOnce(string, ...interface{}) (Rows, error)
}
```

The difference between `Do` and `DoOnce` are that the statements
are either built through building a `Statement` object (which is
subsequently prepared for repeated use in subsequent actions) and
parsing the string into an action, and discarding when the action
has been performed. 

### Action results

When no error is returned which calling `Do` or `DoOnce`, a `Result`
will be returned with information about the action executed:

```go
type Result struct {
  // LastInsertId returns the rowid on INSERT or REPLACE
  LastInsertId int64
  
  // RowsAffected returns number of affected rows on UPDATE or DELETE
  RowsAffected uint64
}
```

### Queries and Rows

When using `Query` and `QueryOnce` a `Rows` object is returned,
which provides details on each set of results:

```go
type Rows interface {
  // Columns returns the columns for the rows
  Columns() []Column

  // Next returns the next row of the results, or nil otherwise
  Next() []Value
}

type Value interface {
  String() string       // Return value as string
  Int() int64           // Return value as int
  Bool() bool           // Return value as bool
  Float() float64       // Return value as float
  Timestamp() time.Time // Return value as timestamp
  Bytes() []byte        // Return value as blob

  // IsNull returns true if value is NULL
  IsNull() bool
}
```

When returning a timestamp, the timezone is by default the local timezone,
or another timezone when it's specified by the `-sqlite.tz` command-line flag.

### Transactions

Actions can be performed within a transaction, which can either be commited
or rolled back if an error occurs. 

```go
type Connection interface {
	// Perform operations within a transaction, rollback on error
	Txn(func(Transaction) error) error
}
```

A transaction is specified using a function which should return an error if the actions should be rolled back:

```go
func DeleteRows(rows []int) error {
  var db sqlite.Connection
  var sql sqlite.Language
  return db.Txn(func(txn sqlite.Transaction) error {
    // Prepare DELETE FROM
    delete := sql.Delete("table",sql.In("_rowid_",rows))
    // Execute the statement
    if _, err := txn.Do(delete); err != nil {
      // Rollback
      return err
    } else {
      // Commit
      return nil
    }
  })
}
```

TODO:
   * Releasing resources
   * Attach and detach
   * Result Sets
   * Supported types
   * Utility methods

## Using `db/sqlang`

SQL statements can be created from a string using the `NewStatement` function:

```go
type Transaction interface {
  NewStatement(string) Statement
}
```

However, it's also possible to use a `sqlite.Language` object to construct SQL statements programmatically. For example,

```go
// Returns SELECT * FROM table LIMIT <nnn>
func NewQuery(limit uint) sqlite.Statement {
  var sql sqlite.Language
  return sql.Select("table").Limit(limit)
}
```

TODO:
   * Language builder

## Using `db/sqobj`

This component implements a very lite object persisence later.

TODO:
   * Defining tables
   * Registering struct as a table, and setting the table name
   * Inserting into the table and updating
   * Querying the table
   * Deleting from the table

## Example Commands

### sq_import

The `sq_import` command line tool allows you to import and query CSV files. In order to compile
the sq_import tool,

```bash
bash% git clone git@github.com:djthorpe/sqlite.git
bash% cd sqlite
bash% make sq_import
```

The command line arguments are:

```

sq_import <flags> <csv_file>...

Flags:
  -noheader
      Do not use the first row as column names
  -notnull
      Dont use NULL values for empty values
  -sqlite.dsn string
      Database source (default ":memory:")
```

