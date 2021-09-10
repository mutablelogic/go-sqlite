package sqlite3_test

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	// Namespace Imports
	. "github.com/djthorpe/go-sqlite/pkg/sqlite3"
)

func Test_Pool_001(t *testing.T) {
	var wg, wg2 sync.WaitGroup

	errs := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	// Create goroutine to receive errors
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		for {
			select {
			case err := <-errs:
				if err != nil {
					t.Log(err)
				}
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	// Create the pool
	pool, err := NewPool("", errs)
	if err != nil {
		t.Error(err)
	}

	// Set maximum number of connections
	pool.SetMax(100)

	// Add some blocking connection
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		// Create 10 connections which block
		for i := 0; i < 10; i++ {
			if conn := pool.Get(context.Background()); conn != nil {
				t.Log("blocking conn [", i, "] => ", conn)
			}
			// Wait for a random amount of time before we open the next connection
			<-time.After(randomDuration(10 * time.Millisecond))
		}
	}()

	for p := 0; p < 50; p++ {
		wg2.Add(1)
		go func(n int) {
			defer wg2.Done()
			// Create N connections, release at random times, expect to only allow 5 connections
			for i := 0; i < n; i++ {
				ctx, cancel := context.WithTimeout(context.Background(), randomDuration(10*time.Second))
				defer cancel()
				if conn := pool.Get(ctx); conn != nil {
					t.Log("conn [", n, ",", i, "] => ", conn)
					go func() {
						<-time.After(randomDuration(10 * time.Second))
						pool.Put(conn)
					}()
				}
				// Wait for a random amount of time before we open the next connection
				<-time.After(randomDuration(100 * time.Millisecond))
			}
		}(p)
	}

	// Wait for gooutines to complete
	wg2.Wait()

	// Wait for pool to drain then release resources
	if err := pool.Close(); err != nil {
		t.Error(err)
	}

	// Cancel the error reporting goroutine
	cancel()
	wg.Wait()
	close(errs)
}

func randomDuration(max time.Duration) time.Duration {
	return time.Duration(time.Duration(rand.Int63()) % time.Duration(max))
}
