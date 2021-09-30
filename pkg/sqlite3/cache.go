package sqlite3

import (
	"sync"

	// Packages
	multierror "github.com/hashicorp/go-multierror"
	sqlite3 "github.com/mutablelogic/go-sqlite/sys/sqlite3"

	// Namespace Imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ConnCache struct {
	sync.Mutex
	sync.Map
	cap int // Capacity of the cache, defaults to 100 prepared statements
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultCapacity = 100
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cache *ConnCache) SetCap(cap int) {
	cache.cap = intMax(0, cap)
}

// Return a prepared statement from the cache, or prepare a new statement
// and put it in the cache before returning
func (cache *ConnCache) Prepare(conn *sqlite3.ConnEx, q string) (*Results, error) {
	if conn == nil {
		return nil, ErrInternalAppError
	}
	st := cache.load(q)
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
	}
	return NewResults(st), nil
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

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (cache *ConnCache) load(key string) *sqlite3.StatementEx {
	// If cache is switched off, then return nil
	if cache.cap == 0 {
		return nil
	}
	// Load statement from cache and return
	st, _ := cache.Map.Load(key)
	if st, ok := st.(*sqlite3.StatementEx); !ok {
		return nil
	} else {
		// Increment counter by one
		st.Inc(1)
		return st
	}
}
