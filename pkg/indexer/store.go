package indexer

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	// Packages
	"github.com/hashicorp/go-multierror"

	// Import namepaces
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Store struct {
	pool    SQPool
	queue   *Queue
	workers uint
	schema  string
}

type operation struct {
	q    SQStatement
	args []interface{}
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	defaultWorkers = uint(runtime.NumCPU() * 2)
	defaultFlush   = 500 * time.Millisecond
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new store object
func NewStore(pool SQPool, schema string, queue *Queue, workers uint) *Store {
	s := new(Store)
	s.pool = pool
	s.queue = queue

	// Create workers - use cpus
	if workers == 0 {
		s.workers = defaultWorkers
	} else {
		s.workers = workers
	}

	// Increase pool size if necessary for the number of workers
	if s.workers > uint(pool.Max()) && pool.Max() != 0 {
		pool.SetMax(int(s.workers))
	}

	// Get a database connection
	conn := pool.Get()
	if conn == nil {
		return nil
	}
	defer pool.Put(conn)

	// Check schema exists
	if !stringSliceContains(conn.Schemas(), schema) {
		return nil
	}

	// Return success
	return s
}

func (s *Store) Run(ctx context.Context, errs chan<- error) error {
	var wg sync.WaitGroup
	var result error

	// Create the schema
	if err := CreateSchema(ctx, s.pool, s.schema); err != nil {
		return err
	}

	// Create workers
	for i := uint(0); i < s.workers; i++ {
		wg.Add(1)
		go func(i uint) {
			defer wg.Done()
			if err := s.worker(ctx, i, errs); err != nil {
				result = multierror.Append(result, err)
			}
		}(i)
	}

	// Wait for end of goroutine
	<-ctx.Done()

	// Wait for end of workers
	wg.Wait()

	// Return any errors
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (s *Store) Schema() string {
	return s.schema
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (s *Store) worker(ctx context.Context, id uint, errs chan<- error) error {
	// Get database connection
	conn := s.pool.Get()
	if conn == nil {
		return ErrInternalAppError.Withf("Worker %d could not obtain database connection", id)
	}
	defer s.pool.Put(conn)

	// Create a timer for flushing
	timer := time.NewTicker(defaultFlush)
	defer timer.Stop()

	// Create an array of operations which is preiodically flushed
	ops := make([]operation, 0)

	// Loop until context cancelled
	for {
		select {
		case <-ctx.Done():
			if n, err := s.flush(context.Background(), conn, ops); err != nil {
				errs <- err
			} else if n > 0 {
				errs <- fmt.Errorf("flush: rows affected %d", n)
			}
			return nil
		case <-timer.C:
			if n, err := s.flush(ctx, conn, ops); err != nil {
				errs <- err
			} else if n > 0 {
				errs <- fmt.Errorf("flush: rows affected %d", n)
			}
			// Flush ops array
			ops = ops[:0]
		default:
			if evt := s.queue.Next(); evt != nil {
				ops = append(ops, s.process(evt))
			}
		}
	}
}

func (s *Store) process(evt *QueueEvent) operation {
	switch evt.Event {
	case EventAdd:
		if replace, args := Replace(s.schema, evt); replace != nil {
			return operation{replace, args}
		}
	case EventRemove:
		if replace, args := Delete(s.schema, evt); replace != nil {
			return operation{replace, args}
		}
	case EventReindexStarted:
		fmt.Println("TODO: INDEX START: ", evt.Path)
	case EventReindexCompleted:
		fmt.Println("TODO: INDEX STOP: ", evt.Path)
	}

	// By default, return empty operation
	return operation{}
}

func (s *Store) flush(ctx context.Context, conn SQConnection, ops []operation) (int, error) {
	n := 0
	if len(ops) == 0 {
		return 0, nil
	}
	err := conn.Do(ctx, 0, func(txn SQTransaction) error {
		for _, op := range ops {
			if op.q != nil {
				if r, err := txn.Query(op.q, op.args...); err != nil {
					return err
				} else {
					n += r.RowsAffected()
				}
			}
		}

		// Return succcess
		return nil
	})
	if err != nil {
		return 0, err
	} else {
		return n, nil
	}
}
