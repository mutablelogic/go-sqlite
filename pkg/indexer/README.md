
# sqlite3 file indexer

This package provides a file indexer, which indexes files in one or more folders (and their
child folders) in an sqlite3 database, and can update when files and folders are added,
changed and deleted.

This package is part of a wider project, `github.com/mutablelogic/go-sqlite`.
Please see the [module documentation](https://github.com/mutablelogic/go-sqlite/blob/master/README.md)
for more information.

## Introduction

An indexing process consists of:

  1. In `Indexer` object, which runs the indexing process;
  2. A `Queue` object, which receives change events from the indexer and passes
     them onto the storage;
  3. A `Store` object, which consumes change events and maintains the database.

To create an indexer, use the `NewIndexer` function. Pass the unique name for the index,
the path to the folder to index, and a `Queue` object. For example, 

```go
package main

import (
	// Packages
	"github.com/mutablelogic/go-sqlite/pkg/indexer"
)

func main() {
    name, path := // ....

    // Create a queue and indexer
    queue := indexer.NewQueue()
	indexer, err := indexer.NewIndexer(name, path, queue)
	if err != nil {
        panic(err)
	}

    // Perform indexing in background process...
    var wg sync.WaitGroup
    go Index(&wg, indexer)
    go Walk(&wg, indexer)
    go Process(&wg, queue)
   
	// Wait for all goroutines to finish
	wg.Wait()

    // Do any cleanup here
}
```

The three background goroutines are:

  1. `Index`: Watches for changes to the folder and indexes the files;
  2. `Walk`: Performs a recursive walk of the folder and indexes the files;
  3. `Process`: Consumes change events from the queue and updates the database.

## Consuming change events

TODO

## File and path inclusions and exclusions

TODO

## Example Applications

There is an example application [here](https://github.com/mutablelogic/go-sqlite/tree/master/cmd) 
which will index a folder into an sqlite database, and a plugin for 
[go-server](https://github.com/mutablelogic/go-server)
[here](https://github.com/mutablelogic/go-sqlite/tree/master/plugin) 
which provides a REST API to the indexing process. You may want to build your own 
application to index your files, in which case read on.
