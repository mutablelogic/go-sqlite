package indexer

import (
	"context"
	"errors"
	"io"
	"path/filepath"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	filesTableName    = "file"
	nameIndexName     = "file_name"
	parentIndexName   = "file_parent"
	filenameIndexName = "file_filename"
	extIndexName      = "file_filename"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func CreateSchema(ctx context.Context, pool SQPool, schema string) error {
	conn := pool.Get()
	if conn == nil {
		return errors.New("unable to get a connection from pool")
	}
	defer pool.Put(conn)

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
		// Create the indexes
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
			row, err := r.Next()
			if err != nil {
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
	return N(filesTableName).WithSchema(schema).Replace("name", "path", "parent", "filename", "isdir", "ext", "modtime", "size"),
		[]interface{}{
			evt.Name,
			evt.Path,
			filepath.Dir(evt.Path),
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
