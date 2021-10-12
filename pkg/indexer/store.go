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
	. "github.com/mutablelogic/go-server"
	. "github.com/mutablelogic/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Store struct {
	pool     SQPool
	queue    *Queue
	renderer RenderFunc
	workers  uint
	schema   string
}

type operation struct {
	q    SQStatement
	args []interface{}
}

type RenderFunc func(context.Context, string, string) (Document, error)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	defaultWorkers = uint(runtime.NumCPU() * 2)
	defaultFlush   = 500 * time.Millisecond
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new store object
func NewStore(pool SQPool, schema string, queue *Queue, r RenderFunc, workers uint) *Store {
	s := new(Store)
	s.pool = pool
	s.queue = queue
	s.renderer = r

	// Create workers - use double number of cores by default
	if workers == 0 {
		s.workers = defaultWorkers
	} else {
		s.workers = workers
	}

	// Increase pool size if necessary for the number of workers
	if s.workers > uint(pool.Max()) && pool.Max() != 0 {
		pool.SetMax(int(s.workers))
	}

	// Reduce number of workers if the pool max does not match
	if pool.Max() > 0 {
		s.workers = uintMin(s.workers, uint(pool.Max()))
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
	if err := s.createschema(ctx); err != nil {
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
// STRINGIFY

func (s *Store) String() string {
	str := "<store"
	str += fmt.Sprint(" workers=", s.workers)
	if s.schema != "" {
		str += fmt.Sprintf(" schema=%q", s.schema)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (s *Store) Schema() string {
	return s.schema
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (s *Store) createschema(ctx context.Context) error {
	// Get database connection
	conn := s.pool.Get()
	if conn == nil {
		return ErrChannelBlocked.With("Could not obtain database connection")
	}
	defer s.pool.Put(conn)

	// Create the schema
	if err := CreateSchema(ctx, conn, s.schema, "porter"); err != nil {
		return err
	}

	// Return success
	return nil
}

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
	ops := make([]operation, 0, defaultCapacity)

	// Loop until context cancelled
	for {
		select {
		case <-ctx.Done():
			if err := s.flushrender(context.Background(), conn, ops); err != nil {
				errs <- fmt.Errorf("[conn %d] %w", conn.Counter(), err)
			}
			return nil
		case <-timer.C:
			if err := s.flushrender(ctx, conn, ops); err != nil {
				errs <- err
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

func (s *Store) flushrender(ctx context.Context, conn SQConnection, ops []operation) error {
	n, err := s.flush(context.Background(), conn, ops)
	if err != nil {
		return err
	}
	if len(n) == 0 {
		return nil
	}
	if s.renderer == nil {
		return nil
	} else {
		return s.render(ctx, conn, n)
	}
}

func (s *Store) process(evt *QueueEvent) operation {
	switch evt.EventType {
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

// Flush the operations array and return the rowid's for any rows which were affected
func (s *Store) flush(ctx context.Context, conn SQConnection, ops []operation) ([]int64, error) {
	if len(ops) == 0 {
		return nil, nil
	}
	result := make([]int64, 0, len(ops))
	err := conn.Do(ctx, 0, func(txn SQTransaction) error {
		// Create file and search records
		for _, op := range ops {
			if op.q != nil {
				if r, err := txn.Query(op.q, op.args...); err != nil {
					return err
				} else if r.RowsAffected() > 0 {
					result = append(result, r.LastInsertId())
				}
			}
		}

		// Return succcess
		return nil
	})
	if err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

// Render rowid's into documents and insert those into the database
func (s *Store) render(ctx context.Context, conn SQConnection, rowid []int64) error {
	var result error
	conn.Do(ctx, 0, func(txn SQTransaction) error {
		for _, rowid := range rowid {
			if rowid == 0 {
				continue
			}
			q, p, t := GetFile(s.schema, rowid)
			r, err := txn.Query(q, p...)
			if err != nil {
				return err
			}
			row := r.Next(t...)
			if row == nil {
				return ErrInternalAppError.Withf("Could not find row %d", rowid)
			}
			doc, err := s.renderer(ctx, row[0].(string), row[1].(string))
			if err != nil {
				result = multierror.Append(result, err)
			} else if err := s.insert(ctx, txn, row[0].(string), row[1].(string), doc); err != nil {
				result = multierror.Append(result, err)
			}
		}
		// We collect errors but we don't rollback because of them
		return nil
	})
	return result
}

// Insert a document into the database within a transaction
func (s *Store) insert(ctx context.Context, txn SQTransaction, name, path string, doc Document) error {
	if n, err := UpsertDoc(txn, &Doc{
		Name:        name,
		Path:        path,
		Title:       doc.Title(),
		Description: doc.Description(),
		Shortform:   string(doc.Shortform()), // TODO: html2text
	}); err != nil {
		return err
	} else {
		fmt.Println("inserted doc", n)
	}

	// Return success
	return nil
}
