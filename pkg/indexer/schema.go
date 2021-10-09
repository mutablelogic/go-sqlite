package indexer

import (
	"context"
	"io"
	"path/filepath"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
	. "github.com/mutablelogic/go-sqlite/pkg/quote"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	filesTableName          = "file"
	nameIndexName           = "file_name"
	searchTableName         = "search"
	parentIndexName         = "file_parent"
	filenameIndexName       = "file_filename"
	extIndexName            = "file_filename"
	searchTriggerInsertName = "search_insert"
	searchTriggerDeleteName = "search_delete"
	searchTriggerUpdateName = "search_update"
)

const (
	defaultTokenizer = "porter unicode61"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func CreateSchema(ctx context.Context, conn SQConnection, schema string, tokenizer string) error {
	// Set default tokenizer as porter
	if tokenizer == "" {
		tokenizer = defaultTokenizer
	}

	// Create tables
	return conn.Do(ctx, 0, func(txn SQTransaction) error {
		if _, err := txn.Query(N(filesTableName).WithSchema(schema).CreateTable(
			C("name").WithPrimary(),
			C("path").WithPrimary(),
			C("parent"),
			C("filename").NotNull(),
			C("isdir").WithType("INTEGER").NotNull(),
			C("ext"),
			C("modtime"),
			C("size").WithType("INTEGER"),
		).IfNotExists()); err != nil {
			return err
		}
		// Create the file indexes
		if _, err := txn.Query(N(nameIndexName).WithSchema(schema).CreateIndex(
			filesTableName, "name",
		).IfNotExists()); err != nil {
			return err
		}
		if _, err := txn.Query(N(parentIndexName).WithSchema(schema).CreateIndex(
			filesTableName, "parent",
		).IfNotExists()); err != nil {
			return err
		}
		if _, err := txn.Query(N(filenameIndexName).WithSchema(schema).CreateIndex(
			filesTableName, "filename",
		).IfNotExists()); err != nil {
			return err
		}
		if _, err := txn.Query(N(extIndexName).WithSchema(schema).CreateIndex(
			filesTableName, "ext",
		).IfNotExists()); err != nil {
			return err
		}
		// Create the search table
		if _, err := txn.Query(N(searchTableName).WithSchema(schema).CreateVirtualTable(
			"fts5",
			"name",
			"parent",
			"filename",
			"content="+filesTableName,
			"tokenize="+Quote(tokenizer),
		).IfNotExists()); err != nil {
			return err
		}
		// triggers to keep the FTS index up to date
		// https://www.sqlite.org/fts5.html
		if _, err := txn.Query(N(searchTriggerInsertName).WithSchema(schema).CreateTrigger(filesTableName,
			Q("INSERT INTO ", searchTableName, " (rowid, name, parent, filename) VALUES (new.rowid, new.name, new.parent, new.filename)"),
		).After().Insert().IfNotExists()); err != nil {
			return err
		}
		if _, err := txn.Query(N(searchTriggerDeleteName).WithSchema(schema).CreateTrigger(filesTableName,
			Q("INSERT INTO ", searchTableName, " (", searchTableName, ", rowid, name, parent, filename) VALUES ('delete', old.rowid, old.name, old.parent, old.filename)"),
		).After().Delete().IfNotExists()); err != nil {
			return err
		}
		if _, err := txn.Query(N(searchTriggerUpdateName).WithSchema(schema).CreateTrigger(filesTableName,
			Q("INSERT INTO ", searchTableName, " (", searchTableName, ", rowid, name, parent, filename) VALUES ('delete', old.rowid, old.name, old.parent, old.filename)"),
			Q("INSERT INTO ", searchTableName, " (rowid, name, parent, filename) VALUES (new.rowid, new.name, new.parent, new.filename)"),
		).After().Update().IfNotExists()); err != nil {
			return err
		}
		return nil
	})
}

// Get indexes and count of documents for each index
func ListIndexWithCount(ctx context.Context, conn SQConnection, schema string) (map[string]int64, error) {
	results := make(map[string]int64)
	if err := conn.Do(ctx, 0, func(txn SQTransaction) error {
		s := Q("SELECT name,COUNT(*) AS count FROM ", N(filesTableName).WithSchema(schema), " GROUP BY name")
		r, err := txn.Query(s)
		if err != nil && err != io.EOF {
			return err
		}
		for {
			row := r.Next()
			if row == nil {
				break
			}
			if len(row) == 2 {
				results[row[0].(string)] = row[1].(int64)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Return success
	return results, nil
}

func Replace(schema string, evt *QueueEvent) (SQStatement, []interface{}) {
	return N(filesTableName).WithSchema(schema).Insert(
			"name", "path", "parent", "filename", "isdir", "ext", "modtime", "size",
		).WithConflictUpdate("name", "path"),
		[]interface{}{
			evt.Name,
			evt.Path,
			pathToParent(evt.Path),
			evt.Info.Name(),
			boolToInt64(evt.Info.IsDir()),
			filepath.Ext(evt.Info.Name()),
			evt.Info.ModTime(),
			evt.Info.Size(),
		}
}

func Delete(schema string, evt *QueueEvent) (SQStatement, []interface{}) {
	return N(filesTableName).WithSchema(schema).Delete(Q("name=?"), Q("path=?")),
		[]interface{}{evt.Name, evt.Path}
}

func Query(schema string) SQSelect {
	// Set the query join
	queryJoin := J(
		N(searchTableName).WithSchema(schema),
		N(filesTableName).WithSchema(schema),
	).LeftJoin(Q(N(searchTableName), ".rowid=", N(filesTableName), ".rowid"))

	// Return the select
	return S(queryJoin).To(
		N("rowid").WithSchema(searchTableName),
		N("rank").WithSchema(searchTableName),
		N("name").WithSchema(filesTableName),
		N("path").WithSchema(filesTableName),
		N("parent").WithSchema(filesTableName),
		N("filename").WithSchema(filesTableName),
		N("isdir").WithSchema(filesTableName),
		N("ext").WithSchema(filesTableName),
		N("modtime").WithSchema(filesTableName),
		N("size").WithSchema(filesTableName),
	).Where(Q(searchTableName, " MATCH ", P)).Order(N("rank"))
}
