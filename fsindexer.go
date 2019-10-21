/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqlite

import (
	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type FSStatus uint

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

// FSIndexer defines a file system indexer
type FSIndexer interface {
	gopi.Driver

	// Index inititates indexing a filesystem at
	// a particular path. The 'watch' argument when
	// true will watch for updates to the index as they
	// occur
	Index(string, bool) (FSIndex, error)

	// Indexes returns all registered indexes and statistics
	Indexes() []FSIndex

	// Delete will cancel indexing and remove an index
	Delete(FSIndex) error
}

// FSIndex represents an index of files or documents
type FSIndex interface {
	Id() int64        // Id of the index
	Name() string     // Name of the index
	Count() uint64    // Count of the documents or files indexed
	Status() FSStatus // Status of the index
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
