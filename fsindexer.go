/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqlite

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type FSStatus uint

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

// FSIndexer defines a file system indexer
type FSIndexer interface {
	gopi.Driver

	// Indexes returns all registered indexes and statistics
	Indexes() []FSIndex

	// Returns an existing index
	IndexById(int64) FSIndex

	// Index inititates indexing a filesystem at a particular path. The 'watch'
	// argument when true will watch for updates to the index as they
	// occur
	AddIndex(string, bool) (int64, error)

	// DeleteById will cancel indexing and remove an index
	// by unique id
	DeleteIndexById(int64) error

	// ReindexById initiates a reindex of an existing index
	// by unique id
	ReindexById(int64) error
}

// FSIndex represents an index of files or documents
type FSIndex interface {
	Id() int64        // Id of the index
	Name() string     // Name of the index
	Count() uint64    // Count of the documents or files indexed
	Status() FSStatus // Status of the index
}

////////////////////////////////////////////////////////////////////////////////
// RPC SERVICE CLIENTS

type FSIndexerIndexClient interface {
	gopi.RPCClient

	Ping() error                            // Ping remote serviice
	List() ([]FSIndex, error)               // List returns a list of indexes
	AddIndex(string, bool) (FSIndex, error) // AddIndex folder and optionally start watching
	DeleteIndex(int64) error                // DeleteIndex removes an index by identifier
}

type FSIndexerQueryClient interface {
	gopi.RPCClient

	Ping() error // Ping remote serviice
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	// FS_STATUS represents the status of an FSIndex objects
	FS_STATUS_NONE FSStatus = iota
	FS_STATUS_INDEXING
	FS_STATUS_IDLE
	FS_STATUS_WATCHING
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s FSStatus) String() string {
	switch s {
	case FS_STATUS_NONE:
		return "FS_STATUS_NONE"
	case FS_STATUS_INDEXING:
		return "FS_STATUS_INDEXING"
	case FS_STATUS_IDLE:
		return "FS_STATUS_IDLE"
	case FS_STATUS_WATCHING:
		return "FS_STATUS_WATCHING"
	default:
		return "[?? Invalid FSStatus value]"
	}
}
