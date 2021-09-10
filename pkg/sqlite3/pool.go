package sqlite3

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	// Modules
	sqlite3 "github.com/djthorpe/go-sqlite/sys/sqlite3"
	multierror "github.com/hashicorp/go-multierror"

	// Namespace Imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
	. "github.com/djthorpe/go-sqlite/pkg/quote"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// PoolConfig is the starting configuration for a pool
type PoolConfig struct {
	Max     int64             `yaml:"max"`   // The maximum number of connections in the pool
	Schemas map[string]string `yaml:"db"`    // Schema names mapped onto path for database file
	Trace   bool              `yaml:"trace"` // Profiling for statements
	Flags   sqlite3.OpenFlags // Flags for opening connections
}

// Pool is a connection pool object
type Pool struct {
	sync.WaitGroup
	sync.Pool
	PoolConfig
	PoolCache

	errs   chan<- error
	ctx    context.Context
	cancel context.CancelFunc
	n      int64
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reSchemaName      = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_-]+$")
	defaultPoolConfig = PoolConfig{
		Max:     5,
		Trace:   false,
		Schemas: map[string]string{defaultSchema: defaultMemory},
		Flags:   sqlite3.DefaultFlags | sqlite3.SQLITE_OPEN_SHAREDCACHE,
	}
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewPool returns a new default pool with a shared cache and maxiumum pool
// size of 5 connections. If filename is not empty, this database is opened
// or else memory is used. Pass a channel to receive errors, or nil to ignore
func NewPool(path string, errs chan<- error) (*Pool, error) {
	cfg := defaultPoolConfig
	if path != "" {
		cfg.Schemas = map[string]string{defaultSchema: path}
	}
	return OpenPool(cfg, errs)
}

// OpenPool returns a new pool with the specified configuration
func OpenPool(config PoolConfig, errs chan<- error) (*Pool, error) {
	p := new(Pool)

	// Set config.Max to default if zero, or minimum of 1
	// connection
	if config.Max == 0 {
		config.Max = defaultPoolConfig.Max
	} else {
		config.Max = maxInt64(config.Max, 1)
	}

	// Set default flags if not set
	if config.Flags == 0 {
		config.Flags = defaultPoolConfig.Flags
	}

	// Set up pool
	p.PoolConfig = config
	p.Pool = sync.Pool{New: func() interface{} {
		if conn, errs := p.new(); errs != nil {
			p.err(errs)
			return nil
		} else {
			return conn
		}
	}}
	p.errs = errs
	p.ctx, p.cancel = context.WithCancel(context.Background())

	// Create a single connection and put in the pool
	if conn, errs := p.new(); errs != nil {
		return nil, errs
	} else {
		p.Pool.Put(conn)
	}

	// Return success
	return p, nil
}

// Close waits for all connections to be released and then
// releases resources
func (p *Pool) Close() error {
	// Set max to 0 to prevent new connections, send cancel signal to all workers
	// and wait for them to exit
	p.SetMax(0)
	p.cancel()
	p.Wait()

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *Pool) String() string {
	str := "<pool"
	str += fmt.Sprint(" cur=", p.Cur())
	str += fmt.Sprint(" max=", p.Max())
	str += fmt.Sprint(" flags=", p.Flags)
	for schema := range p.Schemas {
		str += fmt.Sprintf(" <schema %s=%q>", strings.TrimSpace(schema), p.pathForSchema(schema))
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Max returns the maximum number of connections allowed
func (p *Pool) Max() int64 {
	return atomic.LoadInt64(&p.PoolConfig.Max)
}

// SetMax allowed connections released from pool. Note this does not change
// the maximum instantly, it will settle to this value over time. Set as value
// zero to disable opening new connections
func (p *Pool) SetMax(n int64) {
	atomic.StoreInt64(&p.PoolConfig.Max, maxInt64(n, 0))
}

// Cur returns the current number of used connections
func (p *Pool) Cur() int64 {
	return atomic.LoadInt64(&p.n)
}

// Get a connection from the pool, and return it to the pool when the context
// is cancelled or it is put back using the Put method. If there are no
// connections available, nil is returned.
func (p *Pool) Get(ctx context.Context) *Conn {
	// Return error if maximum number of connections has been reached
	if p.Cur() >= p.Max() {
		p.err(ErrChannelBlocked.Withf("Maximum number of connections (%d) reached", p.Cur()))
		return nil
	}

	// Get a connection from the pool, add one to counter
	conn := p.Pool.Get().(*Conn)
	if conn == nil {
		return nil
	} else if conn.c != nil {
		panic("Expected conn.c to be nil")
	} else {
		conn.c = make(chan struct{})
		atomic.AddInt64(&p.n, 1)
	}

	// Release the connection in the background
	p.WaitGroup.Add(1)
	go func() {
		defer p.WaitGroup.Done()
		select {
		case <-ctx.Done():
			p.put(conn)
		case <-conn.c:
			p.put(conn)
		case <-p.ctx.Done():
			p.put(conn)
		}
	}()

	// Return the connection
	return conn
}

// Return connection to the pool
func (p *Pool) Put(conn *Conn) {
	if conn != nil {
		conn.c <- struct{}{}
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Create a new connection and attach databases, returns error if unable to
// complete operation
func (p *Pool) new() (*Conn, error) {
	// Open connection to main schema, which is required
	defaultPath := p.pathForSchema(defaultSchema)
	if defaultPath == "" {
		return nil, ErrNotFound.Withf("No default schema %q found", defaultSchema)
	}
	conn, err := OpenPath(defaultPath, p.Flags)
	if err != nil {
		return nil, err
	}

	// Attach additional databases
	var result error
	for schema := range p.Schemas {
		schema = strings.TrimSpace(schema)
		path := p.pathForSchema(schema)
		if path == defaultPath {
			continue
		}
		if path == "" {
			result = multierror.Append(result, ErrBadParameter.Withf("Schema %q", schema))
		} else if err := p.attach(conn, schema, path); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Set trace
	if p.PoolConfig.Trace {
		conn.SetTraceHook(func(_ sqlite3.TraceType, a, b unsafe.Pointer) int {
			p.trace(conn, (*sqlite3.Statement)(a), *(*int64)(b))
			return 0
		}, sqlite3.SQLITE_TRACE_PROFILE)
	}

	// Check for errors
	if result != nil {
		return nil, result
	}

	// Success
	return conn, nil
}

func (p *Pool) put(conn *Conn) {
	// Close channel
	if conn.c == nil {
		panic("Expected conn.c to be non-nil")
	} else {
		close(conn.c)
		conn.c = nil
	}
	// Choose to put back into pool or close connection
	n := atomic.AddInt64(&p.n, -1)
	if n >= p.Max() {
		p.Pool.Put(conn)
	} else if err := conn.Close(); err != nil {
		p.err(err)
	}
}

// pathForSchema returns the path for the specified schema
// or an empty string if the schema name is not valid
func (p *Pool) pathForSchema(schema string) string {
	if schema == "" {
		return p.pathForSchema(defaultSchema)
	} else if !reSchemaName.MatchString(schema) {
		return ""
	} else if path, exists := p.Schemas[schema]; !exists {
		return ""
	} else {
		return path
	}
}

// err will pass an error to a channel unless channel is blocked
func (p *Pool) err(err error) {
	select {
	case p.errs <- err:
		return
	default:
		return
	}
}

// maxInt64 returns the maximum of two values
func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// Attach database as schema. If path is empty then a new in-memory database
// is attached.
func (p *Pool) attach(conn *Conn, schema, path string) error {
	if schema == "" {
		return ErrBadParameter.Withf("%q", schema)
	}
	if path == "" {
		return p.attach(conn, schema, defaultMemory)
	}
	return conn.Exec(Q("ATTACH DATABASE ", DoubleQuote(path), " AS ", QuoteIdentifier(schema)), nil)
}

// Detach named database as schema
func (p *Pool) detach(conn *Conn, schema string) error {
	return conn.Exec(Q("DETACH DATABASE ", QuoteIdentifier(schema)), nil)
}

// Trace
func (p *Pool) trace(c *Conn, s *sqlite3.Statement, ns int64) {
	fmt.Printf("TRACE %q => %v\n", s, time.Duration(ns)*time.Nanosecond)
}
