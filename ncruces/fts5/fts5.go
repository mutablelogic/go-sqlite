// Package fts5 adds SQLite's full-text search, of the fts5 virtual table, to each
// connection, which package ncruces opens, when it's imported:
//
//	import _ "github.com/mutablelogic/go-sqlite/ncruces/fts5"
//
// https://sqlite.org/fts5.html
package fts5

import (
	// Packages
	sqlite3 "github.com/ncruces/go-sqlite3"
	ext "github.com/ncruces/go-sqlite3/ext/fts5"
)

func init() {
	sqlite3.AutoExtension(ext.Register)
}
