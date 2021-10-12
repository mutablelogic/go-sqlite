package sqlite3

import (
	"fmt"
	"sync"
	"sync/atomic"

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
	cap uint32 // Capacity of the cache, defaults to 100 prepared statements
	n   uint32
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultCapacity = 100
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (cache *ConnCache) String() string {
	str := "<cache"
	str += fmt.Sprint(" cap=", cache.cap)
	str += fmt.Sprint(" n=", cache.n)
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cache *ConnCache) SetCap(cap uint32) {
	cache.cap = maxUnt32(0, cap)
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

		// Need a mutex around prepare
		cache.Mutex.Lock()
		defer cache.Mutex.Unlock()
		if st, err = conn.PrepareCached(q, cache.cap > 0); err != nil {
			return nil, err
		}

		// Store in cache
		if st.Cached() {
			cache.store(q, st)
		}
	}
	return NewResults(st), nil
}

// Close all conn cache prepared statements
func (cache *ConnCache) Close() error {
	cache.Mutex.Lock()
	defer cache.Mutex.Unlock()

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

func (cache *ConnCache) store(key string, st *sqlite3.StatementEx) {
	cache.Map.Store(key, st)
	if n := atomic.AddUint32(&cache.n, 1); n > cache.cap {
		fmt.Println("TODO: Cached statements exceeeds cap", n)
	}
}

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
		// Increment "used" counter by one
		st.Inc(1)
		// Report if higher than capacity
		return st
	}
}
