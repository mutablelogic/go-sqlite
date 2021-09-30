package indexer_test

import (
	"context"
	"sync"
	"testing"
	"time"

	. "github.com/mutablelogic/go-sqlite/pkg/indexer"
)

const (
	TEST_PATH_1 = "../../.."
)

func Test_Indexer_000(t *testing.T) {
	indexer, err := NewIndexer("test", TEST_PATH_1)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(indexer)
	}
}

func Test_Indexer_001(t *testing.T) {
	// Create channel for errors
	errs, cancel := catchErrors(t)
	defer cancel()

	// Create indexer
	indexer, err := NewIndexer("test", TEST_PATH_1)
	if err != nil {
		t.Fatal(err)
	}

	// Create context for running the indexer
	ctx, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	// Run indexer in background and end when context done
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := indexer.Run(ctx, errs); err != nil {
			t.Error(err)
		}
	}()

	// Queue up an indexing operation
	indexer.Exclude("/waveshare")
	if err := indexer.Walk(ctx); err != nil {
		t.Error(err)
	}

	// Wait for end of goroutine
	wg.Wait()
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// catchErrors returns an error channel and a function to cancel catching the errors
func catchErrors(t *testing.T) (chan<- error, context.CancelFunc) {
	var wg sync.WaitGroup

	errs := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		for {
			select {
			case err := <-errs:
				t.Error(err)
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	return errs, func() {
		cancel()
		wg.Wait()
		close(errs)
	}
}
