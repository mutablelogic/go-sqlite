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
        fmt.Println(sqlite.RowSting(row))
      }
    }
  }
  return nil
}

func main() {
	os.Exit(gopi.CommandLineTool2(gopi.NewAppConfig("db/sqlite"), Main))
}
```

## Using `db/sqobj`

This component implements a very lite object persisence later.

## sq_import

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
  -skipcomments
    	Skip comment lines (default true) which start with # or //
  -sqlite.dsn string
    	Database source (default ":memory:")
  -verbose
    	Verbose logging
  -version
    	Print version information and exit
  -debug
    	Set debugging mode
```

