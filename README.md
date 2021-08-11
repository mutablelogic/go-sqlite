# go-sqlite

This module provides an alternative interface for sqlite, including:

  * Opening in-memory databases and databases by file path;
  * Attaching and detaching databases by schema name;
  * Reflection on databases (schemas, tables, columns, indexes, etc);
  * Transactions (committing changes and rolling back on errors);
  * Executing arbitrary statements or building statements programmatically;
  * Reading results into a struct, map or slice.

Presently the module is in development and the API is subject to change.

## Opening and creating databases

## Attaching databases by schema name

## Database reflection

## Executing queries and transactions

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
