package sqlite3

import (
	"sync"

	// Namespace Imports
	. "github.com/djthorpe/go-errors"
	sqlite3 "github.com/djthorpe/go-sqlite/sys/sqlite3"
	multierror "github.com/hashicorp/go-multierror"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// PoolCache caches prepared statements and profiling information for
// statements so it's possible to see slow queries, etc.
type PoolCache struct{}

type ConnCache struct {
	sync.Mutex
	sync.Map
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultCapacity = 100
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return a prepared statement from the cache, or prepare a new statement
// and put it in the cache before returning
func (cache *ConnCache) Prepare(conn *sqlite3.ConnEx, q string) (*Results, error) {
	if conn == nil {
		return nil, ErrInternalAppError
	}
	st, _ := cache.Map.Load(q)
	if st == nil {
		// Prepare a statement and store in cache
		var err error
		cache.Mutex.Lock()
		defer cache.Mutex.Unlock()
		if st, err = conn.Prepare(q); err != nil {
			return nil, err
		} else {
			cache.Map.Store(q, st)
		}
	} else {
		// Increment counter by one
		st.(*sqlite3.StatementEx).Inc(1)
	}
	return NewResults(st.(*sqlite3.StatementEx)), nil
}

// Close all conn cache prepared statements
func (cache *ConnCache) Close() error {
	var result error
	cache.Map.Range(func(key, value interface{}) bool {
		if err := value.(*sqlite3.StatementEx).Close(); err != nil {
			result = multierror.Append(result, err)
		}
		return true
	})

	// Return any errors
	return result
}
