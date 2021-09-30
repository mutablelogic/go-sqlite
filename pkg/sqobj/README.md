
# sqlite3 objects (SQObjects)

This package provides a method forwriting, reading and deleting go objects (structures)
to and from SQLite databases. Not exactly a full "Object Relational Database" but a way
to reduce the amount of boilerplate code and error handling need to keep a database
in syncronization with objects.

This package is part of a wider project, `github.com/mutablelogic/go-sqlite`.
Please see the [module documentation](https://github.com/mutablelogic/go-sqlite/blob/master/README.md)
for more information.

The general method is:

  1. Use [tags](https://golang.org/ref/spec#Struct_types) on your `struct` definition to
    define the database table, columns, indexes and foreign keys;
  2. Create a database connection and register the `struct`s you want to use to
    syncronize with the database. You also need to register foreign key relationships;
  3. Create the tables and indexes in the database;
  4. Read, write and delete objects using the `Read`, `Write` and `Delete` methods. You
     may need to use some __hook__ functions to handle foreign key relationships.

For example, the following definition:

```go
type Doc struct {
	A int    `sqlite:"a,autoincrement"`
	B int    `sqlite:"b,unique"`
	C string `sqlite:"c"`
}
```

Will create a table with the following statement:

```sql
CREATE TABLE IF NOT EXISTS main.doc (
  a INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  b INTEGER,
  c TEXT,
  UNIQUE (b)
)
```

A fuller explanation of the tags and supported types is provided below. __SQObjects__ is currently in development.

## Introduction

TODO

## Registering a definition

TODO

## Supported scalar types

TODO

## Writing objects (inserting and updating)

TODO

## Reading objects (selecting)

TODO

## Deleting objects

TODO

