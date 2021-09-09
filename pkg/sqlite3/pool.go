package sqlite3

import (
	"context"
	"sync"

	// Modules
	sqlite3 "github.com/djthorpe/go-sqlite/sys/sqlite3"
	multierror "github.com/hashicorp/go-multierror"

	// Namespace Imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// PoolConfig is the starting configuration for a pool
type PoolConfig struct {
	Min     uint              `yaml:"min"` // The minimum number of connections in the pool
	Max     uint              `yaml:"max"` // The maximum number of connections in the pool
	Schemas map[string]string `yaml:"db"`  // Schema names mapped onto path for database file
	Flags   sqlite3.OpenFlags // Flags for opening connections
}

// Pool is a connection pool object
type Pool struct {
	sync.WaitGroup
	PoolConfig

	workers []*Conn
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	defaultPoolConfig = PoolConfig{
		Min:     1,
		Max:     5,
		Schemas: map[string]string{defaultSchema: defaultMemory},
		Flags:   sqlite3.DefaultFlags | sqlite3.SQLITE_OPEN_SHAREDCACHE,
	}
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewPool returns a new default pool with a shared cache and maxiumum pool
// size of 5 connections
func NewPool() (*Pool, error) {
	return OpenPool(defaultPoolConfig)
}

// OpenPool returns a new pool with the specified configuration
func OpenPool(config PoolConfig) (*Pool, error) {
	this := new(Pool)
	this.PoolConfig = config

	// Set capacity for number of workers
	this.workers = make([]*Conn, 0, this.Max)

	// Return success
	return this, nil
}

// Close waits for all connections to be released and then
// releases resources
func (p *Pool) Close() error {
	var result error

	// Wait for all workers to be released before closing
	p.Wait()

	// Release worker resources
	for _, worker := range p.workers {
		if worker != nil {
			if err := worker.Close(); err != nil {
				result = multierror.Append(result, err)
			}
		}
	}

	// Release pool resources
	p.workers = nil

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get obtains a connection from the pool, will block until one is available
// or context is cancelled
func (p *Pool) Get(ctx context.Context) (*Conn, error) {
	return nil, ErrInternalAppError
}

// Release a worker back to the pool
func (p *Pool) Release(c *Conn) error {

}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Create a new connection and attach databases, returns error if unable to
// complete operation
func (p *Pool) newconn() (*Conn, error) {
	// Open connection
	conn, err := sqlite3.Open(this.Schemas[defaultSchema], this.Flags)
	if err != nil {
		return nil, err
	}

	// Return success
	return NewConn(conn), nil
}
