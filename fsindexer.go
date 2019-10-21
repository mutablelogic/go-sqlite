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
// INTERFACES

// FSIndexer defines a file system indexer
type FSIndexer interface {
	gopi.Driver

	// Inititate indexing a filesystem
	Index(string) error
}
