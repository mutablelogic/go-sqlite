# sqlite

[![CircleCI](https://circleci.com/gh/djthorpe/sqlite/tree/master.svg?style=svg)](https://circleci.com/gh/djthorpe/sqlite/tree/master)

This repository contains a higher-level interface to SQLite, to provide database persistence. It implements two components for [gopi](http://github.com/djthorpe/gopi) and some example programs in the `cmd` folder. The repository depends on golang version 1.13 and above (in order to support modules and the new `errors` which are used in testing).

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
  * `fs_indexer` Index files into a database to implement a full-text search
  * `fs_indexer_service` A file indexer which is accessed remotely though gRPC calls

You can also build these tools separately using `make sq_import`, `make fs_indexer` and `make fs_indexer_service`
respectively.

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

### Attaching external databases

The following methods can be used to attach external databases and query the schemas defined:

```go
type Connection interface {
	Attach(schema, dsn string) error
	Detach(schema string) error

  // Return schemas defined, including "main" and "temp"
  Schemas() []string
}
```

The `dsn` is either a filename or the token `":memory:"`for a separate database. The `schema` name cannot be set to `main` or `temp` as those schemas are pre-defined.

### Supported types

The following go types are supported, and their associated database declared type:

| Type       | Decltype  | Example              |
| ---------- | --------- | -------------------- |
| string     | TEXT      | "string"             |
| bool       | BOOL      | true                 |
| int64      | INTEGER   | 1234                 |
| float64    | REAL      | 3.1416               |
| timestamp  | DATETIME  | 2019-07-10T12:34:56Z |
| []byte     | BLOB      | 123456789ABCDE       |

TODO:
  * Converting from bound argments...
  * Quote identifiers...
  * Quote strings...


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

This component implements a very light object persisence later, simply in order to reduce the amount of boilerplate code required to read, write and delete data. In order to use the component, you define a __class__ from a prototype object and then read and write objects between your running application and the database:

```go
package main

type File struct {
	Id   int64  `sql:"id,primary"`
	Root string `sql:"root,primary"`
	Path string `sql:"path"`
	sqlite.Object
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
  sqobj := app.ModuleInstance("db/sqobj").(sqlite.Objects)

  // Register the 'File' class of objects with database
  if _,err := sqobj.RegisterStruct(&File{}); err != nil {
    return err
  }

  // Create an empty object
  obj := &File{}

  // Insert an empty object. Will fail if object with same primary
  // key already exists
  if _, err := sqobj.Write(sqlite.INSERT,obj); err != nil {
    return err
  }

  // Update the object. Will fail if object with same primary key
  // does not exist
  if _, err := sqobj.Write(sqlite.UPDATE,obj); err != nil {
    return err
  }

  // Delete the object. Will fail if object with same primary key
  // does not exist
  if _, err := sqobj.Delete(obj); err != nil {
    return err
  }

  // Return success
  return nil
}

func main() {
  // Create the configuration
  config := gopi.NewAppConfig("db/sqobj")

  // Run the command line tool
  os.Exit(gopi.CommandLineTool2(config, Main))
}
```

### Registering `struct` classes

An object __class__ maps directly to the database: currently a class can
be a 1:1 mapping of a `struct` type but in future classes may also be defined
for views, joins and so forth (these class types have not yet been implemented).

A class which is mapped onto a structure requires:

  * The structure to be named and include at least one non-private field;
  * The structure can embed an `sqlite.Object` structure within it, which means
    no other fields need to be primary;
  * Where a structure does not embed an `sqlite.Object` then to execute update or
    delete operations, the structure must contain one or more primary fields.

Additional properties for a `struct` definition uses the `sql` tag optionally. For
example,

```go
// CREATE TABLE A (Id INTEGER)
//   -- Can only insert and select
type struct A {
  Id int
}

// CREATE TABLE A (Id INTEGER)
//   -- Can insert, update, delete and select
type struct A {  
  Id int 
  sqlite.Object
}

// CREATE TABLE A (primary_id INTEGER PRIMARY KEY NOT NULL)
//   -- Can insert, update, delete and select
type struct A {  
  Id int `sql:"primary_id,primary"`
}

// CREATE TABLE A (
//   column_a INTEGER NOT NULL,
//   column_b TEXT,
//   PRIMARY(column_a,column_b)
// )
//   -- Can insert, update, delete and select
type struct A {  
  Id int `sql:"column_a,primary"`
  Key string `sql:"column_b,primary"`
}
```

Within the stuct definitions, The field tags allowed are:

  * The first token is always the column name. If empty, defaults
    to the field name;
  * A `primary` token defines the column as part of the primary key;
  * A `nullable` token allows the column to represent NULL values, although
    setting and getting NULL values is not currently supported.
  * Where a token is a supported 
    type `TEXT`,`BLOB`,`DATETIME`,`TIMESTAMP`,`FLOAT`,`INTEGER` or `BOOL` overrides the type defined in the field.

The structure is registered using the `RegisterStruct` method, which
returns the class:

```go
  // Register the 'File' class of objects with database
  if class,err := sqobj.RegisterStruct(File{}); err != nil {
    return err
  } else {
    // ...
  }
```

This would create the table in the database if it doesn't exist. If it does exist, an
error is returned if the table cannot represent the defined `struct` (for example, it
has missing columns). You can override the name of the table in the database by defining
a static method as follows:

```go
func (File) TableName() string {
  return "file_table"
}
```

### Writing to the database

Writing to the database consists of three operations:

  * Inserting a new object into the database is represented by flag `sqlite.FLAG_INSERT`;
  * Updating an existing object is represented by flag `sqlite.FLAG_UPDATE`;
  * Replacing or inserting objects is represented by flag `sqlite.FLAG_UPDATE|sqlite.FLAG_INSERT`

You can use the `Write` method, which takes both the flag and the objects to be written, assuming the structure has already been registered with the database:

```go
  new_file := &File{}
  if affected_rows, err := db.Write(sqlite.FLAG_INSERT,new_file); err != nil {
    // ...
  } else {
    fmt.Println("rowid=",new_file.RowId)
  }
```

The `Write` method returns the number of objects inserted, updated or replaced. If 
your `struct` embeds a `sqlite.Object` then you can access the `RowId` field. This
will only work on insert if you pass the pointer to object.

### Deleting and Counting

You can delete specific objects from the database using the `Delete` method and also return a count of objects in the database by referring to the
class:

```go
type Objects interface {
	// Delete structs by key or rowid, rollback on error
	// and return number of affected rows
	Delete(objs ...interface{}) (uint64, error)

	// Count number of objects of a particular class
	Count(Class) (uint64, error)
}

```

### Reading from the database

In order to read from the database, you need to provide an array of objects with enough capacity to hold you objects. For example, in the following example a slice with capacity of 100 is created and passed. The second argument to `Read` defaults to a `LIMIT` of 100 when set to 0, but can be set to anything from 1 to the capacity of the slice otherwise:

```go
  objs := make(File,0,100)
  if returned_rows, err := db.Read(&obj,0); err != nil {
    // ...
  } else {
    fmt.Println(objs)
  }
```

### Additional notes

Here's some additional functionality which is still needed:

  * Ordering reads using ORDER BY
  * Document supported types
  * Supporting NULL/nil
  * Supporting OFFSET, filtering
  * Supporting JOINs and VIEWs and grouping
  * Pre-write and pre-delete hooks 

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

