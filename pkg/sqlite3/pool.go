package sqlite3

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	// Modules
	multierror "github.com/hashicorp/go-multierror"
	sqlite3 "github.com/mutablelogic/go-sqlite/sys/sqlite3"

	// Namespace Imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// PoolConfig is the starting configuration for a pool
type PoolConfig struct {
	Max     int32             `yaml:"max"`       // The maximum number of connections in the pool
	Schemas map[string]string `yaml:"databases"` // Schema names mapped onto path for database file
	Create  bool              `yaml:"create"`    // When false, do not allow creation of new file-based databases
	Auth    SQAuth            // Authentication and Authorization interface
	Trace   TraceFunc         // Trace function
	Flags   SQFlag            // Flags for opening connections
}

// Pool is a connection pool object
type Pool struct {
	cfg   PoolConfig   // The configuration for the pool
	pool  sync.Pool    // The pool of connections
	errs  chan<- error // Errors are sent to this channel
	n     int32        // The number of connections in the pool
	drain int32        // Pool is draining (boolean)
}

// TraceFunc is a function that is called when a statement is executed or prepared
type TraceFunc func(c *Conn, q string, delta time.Duration)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reSchemaName      = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_-]+$")
	defaultPoolConfig = PoolConfig{
		Max:    5,
		Create: true,
		Flags:  SQFlag(sqlite3.SQLITE_OPEN_CREATE) | SQFlag(sqlite3.SQLITE_OPEN_SHAREDCACHE) | SQLITE_OPEN_CACHE,
	}
)

////////////////////////////////////////////////////////////////////////////////
// CONFIGURATION OPTIONS

// Create a new default configuraiton for the pool
func NewConfig() PoolConfig {
	cfg := defaultPoolConfig
	cfg.Schemas = map[string]string{DefaultSchema: defaultMemory}
	return cfg
}

// Enable authentication and authorization
func (cfg PoolConfig) WithAuth(auth SQAuth) PoolConfig {
	cfg.Auth = auth
	return cfg
}

// Enable trace of statement execution
func (cfg PoolConfig) WithTrace(fn TraceFunc) PoolConfig {
	cfg.Trace = fn
	return cfg
}

// Enable or disable creation of database files
func (cfg PoolConfig) WithCreate(create bool) PoolConfig {
	cfg.Create = create
	return cfg
}

// Set maxmimum concurrent connections
func (cfg PoolConfig) WithMaxConnections(n int) PoolConfig {
	if n >= 0 {
		cfg.Max = int32(n)
	}
	return cfg
}

// Add schema to the pool
func (cfg PoolConfig) WithSchema(name, path string) PoolConfig {
	cfg.Schemas[name] = path
	return cfg
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewPool returns a new default pool with a shared cache and maxiumum pool
// size of 5 connections. If filename is not empty, this database is opened
// or else memory is used. Pass a channel to receive errors, or nil to ignore
func NewPool(path string, errs chan<- error) (*Pool, error) {
	cfg := NewConfig()
	if path != "" {
		cfg.Schemas = map[string]string{DefaultSchema: path}
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
		config.Max = maxInt32(config.Max, 1)
	}

	// Set default flags if not set
	if config.Flags == 0 {
		config.Flags = defaultPoolConfig.Flags
	}

	// Update create flag
	if config.Create {
		config.Flags |= SQFlag(sqlite3.SQLITE_OPEN_CREATE)
	} else {
		config.Flags &^= SQFlag(sqlite3.SQLITE_OPEN_CREATE)
	}

	// Set up pool
	p.cfg = config
	p.errs = errs
	p.pool = sync.Pool{New: func() interface{} {
		if conn, errs := p.new(); errs != nil {
			p.err(errs)
			return nil
		} else {
			return conn
		}
	}}

	// Create a single connection and put in the pool
	if conn, errs := p.new(); errs != nil {
		return nil, errs
	} else {
		p.Put(conn)
		p.n = 0
	}

	// Return success
	return p, nil
}

// Close waits for all connections to be released and then
// releases resources
func (p *Pool) Close() error {
	// Drain the pool
	atomic.StoreInt32(&p.drain, 1)

	var result error
	for {
		conn := p.pool.Get()
		if conn == nil {
			break
		} else if err := conn.(*Conn).Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *Pool) String() string {
	str := "<pool"
	str += fmt.Sprintf(" ver=%q", Version())
	str += fmt.Sprint(" flags=", sqlite3.OpenFlags(p.cfg.Flags))
	str += fmt.Sprint(" cur=", p.Cur())
	str += fmt.Sprint(" max=", p.Max())
	for schema := range p.cfg.Schemas {
		str += fmt.Sprintf(" <schema %s=%q>", strings.TrimSpace(schema), p.pathForSchema(schema))
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *Pool) Get() SQConnection {
	if conn, ok := p.pool.Get().(SQConnection); ok {
		// Increment counter of open connections
		atomic.AddInt32(&p.n, 1)
		return conn
	} else {
		return nil
	}
}

func (p *Pool) Put(conn SQConnection) {
	if conn != nil {
		// Decrement counter of open connections
		atomic.AddInt32(&p.n, -1)
		p.pool.Put(conn)
	}
}

// Return number of "checked out" (used) connections
func (p *Pool) Cur() int {
	return int(atomic.LoadInt32(&p.n))
}

// Return maximum allowed connections
func (p *Pool) Max() int {
	return int(p.cfg.Max)
}

// Set maximum number of "checked out" connections
func (p *Pool) SetMax(n int) {
	if n == 0 {
		p.cfg.Max = defaultPoolConfig.Max
	} else {
		p.cfg.Max = maxInt32(int32(n), 1)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (p *Pool) new() (SQConnection, error) {
	// If pool is being drained, return nil
	if atomic.LoadInt32(&p.drain) != 0 {
		return nil, nil
	}

	// If cur >= max, then reject
	if p.cfg.Max != 0 && atomic.LoadInt32(&p.n) >= p.cfg.Max {
		return nil, ErrChannelBlocked.Withf("Maximum number of connections reached (%d)", p.cfg.Max)
	}

	// Open connection to main schema, which is required
	defaultPath := p.pathForSchema(DefaultSchema)
	if defaultPath == "" {
		return nil, ErrNotFound.Withf("No default schema %q found", DefaultSchema)
	}

	// Always allow memory databases to be created and read/write
	flags := p.cfg.Flags
	if defaultPath == defaultMemory {
		flags |= SQFlag(sqlite3.SQLITE_OPEN_CREATE | sqlite3.SQLITE_OPEN_READWRITE)
	}

	// Perform the open
	conn, err := OpenPath(defaultPath, flags)
	if err != nil {
		return nil, err
	}

	// Set trace
	if p.cfg.Trace != nil {
		conn.ConnEx.SetTraceHook(func(_ sqlite3.TraceType, a, b unsafe.Pointer) int {
			p.trace(conn, (*sqlite3.Statement)(a), *(*int64)(b))
			return 0
		}, sqlite3.SQLITE_TRACE_PROFILE)
	}

	// Attach additional databases
	var result error
	for schema := range p.cfg.Schemas {
		schema = strings.TrimSpace(schema)
		path := p.pathForSchema(schema)
		if schema == DefaultSchema {
			continue
		}
		if path == "" {
			result = multierror.Append(result, ErrBadParameter.Withf("Schema %q", schema))
		} else if err := conn.Attach(schema, path); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Set auth
	if p.cfg.Auth != nil {
		conn.SetAuthorizerHook(func(action sqlite3.SQAction, args [4]string) sqlite3.SQAuth {
			if err := p.auth(conn.ctx, action, args); err == nil {
				return sqlite3.SQLITE_ALLOW
			} else {
				p.err(err)
				return sqlite3.SQLITE_DENY
			}
		})
	}

	// Check for errors
	if result != nil {
		return nil, result
	}

	// Success
	return conn, nil
}

// err will pass an error to a channel unless channel is blocked
func (p *Pool) err(err error) {
	if p.errs != nil {
		select {
		case p.errs <- err:
			return
		default:
			return
		}
	}
}

// pathForSchema returns the path for the specified schema
// or an empty string if the schema name is not valid
func (p *Pool) pathForSchema(schema string) string {
	if schema == "" {
		return p.pathForSchema(DefaultSchema)
	} else if !reSchemaName.MatchString(schema) {
		return ""
	} else if path, exists := p.cfg.Schemas[schema]; !exists {
		return ""
	} else {
		return path
	}
}

// Trace
func (p *Pool) trace(c *Conn, s *sqlite3.Statement, ns int64) {
	if p.cfg.Trace != nil {
		p.cfg.Trace(c, s.SQL(), time.Duration(ns)*time.Nanosecond)
	}
}
