# go-sqlite

This module provides an alternative interface for sqlite, including:

  * Opening in-memory databases and databases by file path;
  * Transactions (committing changes and rolling back on errors);
  * Reading results into a struct, map or slice.
  * Reflection on databases (schemas, tables, columns, indexes, etc);
  * Attaching and detaching databases by schema name;
  * Executing arbitrary statements or building statements programmatically;

Presently the module is in development and the API is subject to change.

## Opening and creating databases

You can create a new in-memory database using the `New` method, or you can open an existing file-based database using the `Open` method.
Both methods take an optional `*time.Location` argument which allows interpretation of time values with time zone. For example,

```go
package main

import (
  "github.com/djthorpe/go-sqlite/pkg/sqlite"
)

func main() {
  db, err := sqlite.New() // Open in-memory database with local time zone
  if err != nil {
    // ...
  }
  defer db.Close()

  path := // ...
  db, err := sqlite.Open(path,time.UTC) // Open file database with UTC time zone
  if err != nil {
    // ...
  }
  defer db.Close()
}
```

Use `db.Close()` to release database resources.

## Executing queries and transactions

Statements are executed using the `db.Exec` method on a query. In order to create
a query from a string, use the `Q()` method (see below for information on building
SQL statements). You can include bound parameters after the query:

For example,

```go
package main

import (
  "github.com/djthorpe/go-sqlite/pkg/sqlite"
  . "github.com/djthorpe/go-sqlite/pkg/lang"
)

func main() {
  db, err := sqlite.New() // Open in-memory database with local time zone
  if err != nil {
    // ...
  }
  defer db.Close()

  if result,err := db.Exec(Q("CREATE TABLE test (a TEXT,b TEXT)")); err != nil {
    // ...
  } else {
    fmt.Println(result)
  }
}
```

The result object returned (of type SQResult) contains fields `LastInsertId` and `RowsAffected`
which may or may not be set depending on the query executed. To return data, a result set is
returned which should allows you to iterate across the results as a map of values, a slice
of values:


```go
package main

import (
  "github.com/djthorpe/go-sqlite/pkg/sqlite"
  . "github.com/djthorpe/go-sqlite/pkg/lang"
)

func main() {
  db, err := sqlite.New() // Open in-memory database with local time zone
  if err != nil {
    // ...
  }
  defer db.Close()

  rs,err := db.Select(Q("SELECT a,b FROM test WHERE a=?"),"foo")
  if err != nil {
    // ...
  }
  defer rs.Close()
  for {
    row := rs.Next()
    if row == nil {
      break
    }
    // ...
  }
}
```

You can also create a block of code, which when returning any error will rollback
any database snapshot, or else commit the snapshot if no error occurred:

```go
package main

import (
  "github.com/djthorpe/go-sqlite/pkg/sqlite"
  . "github.com/djthorpe/go-sqlite/pkg/lang"
)

func main() {
  db, err := sqlite.New() // Open in-memory database with local time zone
  if err != nil {
    // ...
  }
  defer db.Close()

  db.Do(func (txn sqlite.SQTransaction) error {
    _, err := txn.Exec(Q("..."))
    if err != nil {
      // Rollback any database changes
      return err
    }
    
    // Perform further operatins here...

    // Return success, commit transaction
    return nil
  })
}
```

## Supported Column Types

The following types are supported, and expect the declared column types:

| Scalar Type | Column Type           |
| ------------| ----------------------| 
| int64       | INTEGER               |
| float64     | FLOAT                 |
| string      | TEXT                  |
| bool        | BOOL                  |
| time.Time   | TIMESTAMP or DATETIME |
| []byte      | BLOB                  |

If you pass other integer and unsigned integer types into the `Exec` and `Query` functions then they are converted to one of the above types. You can also define methods `MarshalSQ` and `UnmarshalSQ` in order to convert your custom types into
scalar types. For example, the following methods convert between supported scalar types:

```go
type CustomParam struct {
	A, B string
}

func (c CustomParam) MarshalSQ() (interface{}, error) {
	if data, err := json.Marshal(c); err != nil {
		return nil, err
	} else {
		return string(data), err
	}
}

func (c *CustomParam) UnmarshalSQ(v interface{}) error {
	if data, ok := v.(string); ok {
		return json.Unmarshal([]byte(data), c)
	} else {
		return fmt.Errorf("Invalid type: %T", v)
	}
}
```

## Attaching databases by schema name

You can load additional databases to a database by schema name. Use `Attach` and `Detach` to attach and detach databases. For example,



## Database reflection

## Building statements programmatically

A statement builder can be used for generating SQL statements programmatially. It is intended you use
the following primitves to build your statements:

  * `P` is a placeholder for a value, which binds to the corresponding placeholder in `Query` or `Exec` methods;
  * `V()` is the value function;
  * `N()` is the name function, which corresponds to a table or column name;
  * `Q()` is the quote function, which allows insertation or execution or arbitary queries;
  * `S()` is the select function, which builds up a SELECT statement;

In order to use these primitives within your code, it is suggested you import the laguage namespace directly into
your code. For example:

```go
package main

import (
  . "github.com/djthorpe/go-sqlite/pkg/lang"
)

func main() {
  s := S(N("a"),N("b").Distinct().Where(N("a").Is(P))
  fmt.Println(s) // Prints SELECT DISTINCT * FROM a,b WHERE a=?
}
```

If the symbols P,V,N,Q or S clash with any symbols in your code namespace, you can import the package
without the dot prefix and refer to the sumbols prefixed with `lang.` instead.

## Reading results into a struct, map or slice


## Importing data
