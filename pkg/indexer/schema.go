package indexer

import (
	"context"
	"errors"
	"path/filepath"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	filesTableName = "files"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func CreateSchema(ctx context.Context, pool SQPool, schema string) error {
	conn := pool.Get()
	if conn == nil {
		return errors.New("unable to get a connection from pool")
	}
	defer pool.Put(conn)

	// Create table
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
		return nil
	})
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
