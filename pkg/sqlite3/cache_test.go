package sqlite3_test

import (
	"context"
	"errors"
	"io"
	"math/rand"
	"sync"
	"testing"

	// Module imports
	sqlite3 "github.com/djthorpe/go-sqlite/sys/sqlite3"

	// Namespace Imports
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
	. "github.com/djthorpe/go-sqlite/pkg/sqlite3"
)

func Test_Cache_001(t *testing.T) {
	// Create a connection and enable the connection cache (which caches prepared statements)
	conn, err := OpenPath(":memory:", sqlite3.DefaultFlags|sqlite3.SQLITE_OPEN_CONNCACHE)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Perform caching in a transaction
	conn.Do(context.Background(), 0, func(txn SQTransaction) error {
		// SELECT n between 0-9 over 100 executions in parallel should
		// return the same result, with a perfect cache hit rate of 9 in 10?
		var wg sync.WaitGroup
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				n := rand.Uint32() % 10
				r, err := txn.Query(Q("SELECT ", n))
				if err != nil {
					t.Error("Query Error: ", err)
					return
				}
				for {
					row, err := r.Next()
					if errors.Is(err, io.EOF) {
						break
					} else if err != nil {
						t.Error("Next Error: ", err)
					} else if len(row) != 1 {
						t.Error("Unexpected row length: ", row)
					} else if int64(n) != row[0] {
						t.Error("Unexpected row value: ", row, " expected [", n, "]")
					} else {
						t.Log(n, "=>", row[0])
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
