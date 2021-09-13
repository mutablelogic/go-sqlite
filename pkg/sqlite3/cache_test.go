package sqlite3_test

import (
	"context"
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
	// Create a connection
	conn, err := OpenPath(":memory:", sqlite3.DefaultFlags)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Perform caching in a transaction
	conn.Do(context.Background(), 0, func(txn SQTransaction) error {
		// Read values from the cache
		var wg sync.WaitGroup
		for i := 0; i < 99; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				n := rand.Uint32() % 5
				if r, err := txn.Query(Q("SELECT ", n)); err != nil {
					t.Error(err)
				} else {
					t.Log(r)
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
