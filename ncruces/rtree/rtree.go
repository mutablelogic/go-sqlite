// Package rtree adds SQLite's spatial indexes, of the rtree and geopoly virtual tables, to each
// connection, which package ncruces opens, when it's imported:
//
//	import _ "github.com/mutablelogic/go-sqlite/ncruces/rtree"
//
// https://sqlite.org/rtree.html
package rtree

import (
	// Packages
	sqlite3 "github.com/ncruces/go-sqlite3"
	ext "github.com/ncruces/go-sqlite3/ext/rtree"
)

func init() {
	sqlite3.AutoExtension(ext.Register)
}
