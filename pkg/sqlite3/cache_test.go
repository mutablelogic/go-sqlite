package sqlite3_test

import (
	"context"
	"math/rand"
	"sync"
	"testing"

	// Module imports
	sqlite3 "github.com/mutablelogic/go-sqlite/sys/sqlite3"

	// Namespace Imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
	. "github.com/mutablelogic/go-sqlite/pkg/sqlite3"
)

func Test_Cache_001(t *testing.T) {
	// Create a connection and enable the connection cache (which caches prepared statements)
	conn, err := OpenPath(":memory:", SQFlag(sqlite3.DefaultFlags)|SQLITE_OPEN_CACHE)
	if err != nil {
		t.Fatal(err)
	}

	// Perform caching in a transaction
	conn.Do(context.Background(), 0, func(txn SQTransaction) error {
		// SELECT n between 0-9 over 1000 executions in parallel should
		// return the same result, with a perfect cache hit rate of 9 in 10?
		var wg sync.WaitGroup
		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func() {
				txn.Lock()
				defer txn.Unlock()
				defer wg.Done()
				n := rand.Uint32() % 10
				r, err := txn.Query(Q("SELECT ", n))
				if err != nil {
					t.Error("Query Error: ", err)
					return
				}
				defer r.Close()
				for {
					row := r.Next()
					if row == nil {
						break
					} else if len(row) != 1 {
						t.Error("Unexpected row length: ", row, " expected [", n, "]", r.Columns())
					} else if int64(n) != row[0] {
						t.Error("Unexpected row value: ", row, " expected [", n, "]", r.Columns())
					}
				}
			}()
		}

		// Wait for all goroutines to complete
		wg.Wait()

		// Return success
		return nil
	})

	// Close cache, release resources
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
}
