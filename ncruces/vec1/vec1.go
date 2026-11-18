// Package vec1 adds SQLite's vector search, of the vec1 virtual table, to each
// connection, which package ncruces opens, when it's imported:
//
//	import _ "github.com/mutablelogic/go-sqlite/ncruces/vec1"
//
// https://sqlite.org/vec1
package vec1

import (
	// Packages
	sqlite3 "github.com/ncruces/go-sqlite3"
	ext "github.com/ncruces/go-sqlite3/ext/vec1"
)

func init() {
	sqlite3.AutoExtension(ext.Register)
}
