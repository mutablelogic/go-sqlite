package sqlite3_test

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	// Namespace Imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
	. "github.com/mutablelogic/go-sqlite/pkg/sqlite3"
)

func Test_Pool_001(t *testing.T) {
	errs, cancel := handleErrors(t)
	pool, err := NewPool(":memory:", errs)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(pool)
	}
	defer pool.Close()
	defer cancel()
}

func Test_Pool_002(t *testing.T) {
	errs, cancel := handleErrors(t)
	pool, err := NewPool(":memory:", errs)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(pool)
	}
	pool.SetMax(5000)

	// Get/put connections
	var wg sync.WaitGroup
	for i := 0; i < 5000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-time.After(randomDuration(10 * time.Millisecond))
			conn := pool.Get()
			if conn != nil {
				t.Log("conn [", i, "] => ", conn, " cur=", pool.Cur())
				conn.Do(context.Background(), 0, func(txn SQTransaction) error {
					_, err := txn.Query(Q("SELECT NULL"))
					return err
				})
				// Wait for a random amount of time before we open the next connection
				<-time.After(randomDuration(10 * time.Millisecond))
				// Return connection
				pool.Put(conn)
				t.Log("  returned conn [", i, "] => cur=", pool.Cur())
			}
		}(i)
	}

	// Wait for all Get/Puts to complete
	wg.Wait()

	// Close pool
	if err := pool.Close(); err != nil {
		t.Error(err)
	}

	// Cancel errors
	cancel()
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func randomDuration(max time.Duration) time.Duration {
	return time.Duration(time.Duration(rand.Int63()) % time.Duration(max))
}

func handleErrors(t *testing.T) (chan<- error, context.CancelFunc) {
	var wg sync.WaitGroup
	errs := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		defer wg.Done()
	FOR_LOOP:
		for {
			select {
			case <-ctx.Done():
				break FOR_LOOP
			case err := <-errs:
				if err != nil {
					t.Error(err)
				}
			}
		}
		close(errs)
	}()
	return errs, func() { cancel(); wg.Wait() }
}
