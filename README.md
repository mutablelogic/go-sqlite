# sqlite

[![CircleCI](https://circleci.com/gh/djthorpe/sqlite/tree/master.svg?style=svg)](https://circleci.com/gh/djthorpe/sqlite/tree/master)

This repository contains a higher-level interface to SQLite, to provide database persistence.
It implements two components for [gopi](http://github.com/djthorpe/gopi) and some example 
programs in the `cmd` folder. The repository depends on golang version 1.12 and 
above (in order to support modules).

## Components

The gopi components provided by this repository are:

| Component Path | Description                            | Component Name |
| -------------- | -------------------------------------- |--------------- |
| sys/sqlite     | SQL Database persisence using sqlite   | db/sqlite      |
| sys/sqlite     | SQL Language Builder                   | db/sqlang      |
| sys/sqobj      | General Purpose Hardware Input/Output  | db/sqobj       |

## Building and installing examples

There is a makefile which can be used for testing and installing bindings and examples:

```
bash% git clone git@github.com:djthorpe/sqlite.git
bash% cd sqlite
bash% make all
```

The resulting binaries are as follows. Use the `-help` flag to see the different options for each:

  * `sq_import` Imports data from CSV files into an SQLite database

## Using `db/sqlite`

Database persistence is implemented using the `db/sqlite` component. Here's an example of how
to use the component:

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

  // Return table names for a particulr schema. Temporary tables
  // are not included by default
  Tables(schema string, include_temporary bool) []string
  
  // Return the columns for a table
  ColumnsForTable(name, schema string) ([]Column, error)
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
  gopi.Driver

  // Execute statement (without returning the rows)
  Do(Statement, ...interface{}) (Result, error)
  DoOnce(string, ...interface{}) (Result, error)

  // Query to return sets of results
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
  // The last primary key (rowid) on INSERT or REPLACE
  LastInsertId int64
  
  // The number of rows affected by the action if UPDATE or DELETE
  RowsAffected uint64
}
```

### Queries and Rows

When using `Query` and `QueryOnce` a `Rows` object is returned,
which provides details on each set of results:

TODO

### Transactions

A number of update, delete and insert actions can be performed
within a transaction, which can either be commited to the database
or rolled back if an error occurs:

```go
type Connection interface {
	// Perform operations within a transaction, rollback on error
	Tx(func(Transaction) error) error
}
```

Here's an example of a transaction:

```go
func DeleteRows(rows []int) error {
  var db sqlite.Connection
  var sql sqlite.Language
  return db.Tx(func(txn sqlite.Transaction) error {
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
   * Result Sets
   * Supported types
   * Utility methods
   * Language builder

## Using `db/sqlang`

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

