package sqlite3

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type StatementEx struct {
	sync.Mutex
	st []*Statement
	n  uint32
	ts int64
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Prepare query string and return prepared statements
func (c *ConnEx) Prepare(q string) (*StatementEx, error) {
	s := new(StatementEx)
	for {
		if q == "" {
			break
		}
		st, extra, err := c.Conn.Prepare(q)
		if err != nil {
			return nil, err
		}
		s.st = append(s.st, st)
		q = strings.TrimSpace(extra)
	}

	// Set "last used" timestamp
	s.ts = time.Now().UnixNano()

	// Report on missing close
	_, file, line, _ := runtime.Caller(1)
	runtime.SetFinalizer(s, func(s *StatementEx) {
		if s.st != nil {
			panic(fmt.Sprintf("%s:%d: Prepare() missing call to Close()", file, line))
		}
	})

	// Return statement
	return s, nil
}

// Release resources for statements
func (s *StatementEx) Close() error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	// Finalize all statements
	var result error
	for _, st := range s.st {
		if err := st.Finalize(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Release resources
	s.st = nil
	s.ts = 0

	// Return any errors
	return result
}

// Execute prepared statement n, when called with arguments, this
// calls Bind() first
func (s *StatementEx) Exec(n uint, v ...interface{}) (*Results, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	// Return nil result and SQLITE_DONE if no more statements to execute
	if n >= uint(len(s.st)) {
		return nil, SQLITE_DONE
	}

	// Step to next statement
	st := s.st[int(n)]
	if err := st.Reset(); err != nil {
		return nil, err
	}

	// Bind parameters
	if len(v) > 0 {
		if err := st.Bind(v...); err != nil {
			return nil, err
		}
	}

	// Perform the step
	if err := st.Step(); errors.Is(err, SQLITE_DONE) || errors.Is(err, SQLITE_ROW) {
		return results(st, err), nil
	} else {
		return nil, err
	}
}

// Increment adds n to the statement counter and updates the timestamp
func (s *StatementEx) Inc(n uint32) {
	atomic.StoreInt64(&s.ts, time.Now().UnixNano())
	atomic.AddUint32(&s.n, n)
}

// Returns current count. Used to count the frequency of calls for caching purposes.
func (s *StatementEx) Count() uint32 {
	return atomic.LoadUint32(&s.n)
}

// Returns last accessed timestamp for caching purposes as an int64
func (s *StatementEx) Timestamp() int64 {
	return atomic.LoadInt64(&s.ts)
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s *StatementEx) String() string {
	str := "[statements"
	for _, st := range s.st {
		str += fmt.Sprint(" " + st.String())
	}
	return str + "]"
}
