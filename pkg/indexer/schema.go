package indexer

import (
	"context"
	"io"
	"path/filepath"
	"reflect"
	"time"

	// Package imports
	sqobj "github.com/mutablelogic/go-sqlite/pkg/sqobj"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
	. "github.com/mutablelogic/go-sqlite/pkg/quote"
)

///////////////////////////////////////////////////////////////////////////////
// Types

type File struct {
	Name     string    `sqlite:"name,primary,index:name,join:name"` // Index name, primary key
	Path     string    `sqlite:"path,primary,index:path,join:path"` // Relative path, primary key
	Parent   string    `sqlite:"parent,index:parent"`               // Parent folder
	Filename string    `sqlite:"filename,notnull,index:filename"`   // Filename
	IsDir    bool      `sqlite:"isdir,notnull"`                     // Is a directory
	Ext      string    `sqlite:"ext,index:ext"`
	ModTime  time.Time `sqlite:"modtime"`
	Size     int64     `sqlite:"size"`
}

type Doc struct {
	Name        string `sqlite:"name,primary,foreign,join:name"` // Index name, primary key
	Path        string `sqlite:"path,primary,foreign,join:path"` // Relative path, primary key
	Title       string `sqlite:"title,notnull"`                  // Title of the document, text
	Description string `sqlite:"description"`                    // Description of the document, text
	Shortform   string `sqlite:"shortform"`                      // Shortform of the document, html
}

// View is used as the content source for the search virtual table
// and is a join between File and Doc
type View struct {
	Name        string `sqlite:"name"`
	Parent      string `sqlite:"parent"`
	Filename    string `sqlite:"filename"`
	Title       string `sqlite:"title"`
	Description string `sqlite:"description"`
	Shortform   string `sqlite:"shortform"`
}

// Search virtual table uses View to get content
type Search struct {
	Name        string `sqlite:"name"`
	Parent      string `sqlite:"parent"`
	Filename    string `sqlite:"filename"`
	Title       string `sqlite:"title"`
	Description string `sqlite:"description"`
	Shortform   string `sqlite:"shortform"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	fileTableName           = "file"
	searchTableName         = "search"
	docTableName            = "doc"
	viewTableName           = "view"
	searchTriggerInsertName = "search_insert"
	searchTriggerDeleteName = "search_delete"
	searchTriggerUpdateName = "search_update"
)

const (
	defaultTokenizer = "porter unicode61"
)

var (
	filesTypeCast = []reflect.Type{
		reflect.TypeOf(""),
		reflect.TypeOf(""),
		reflect.TypeOf(""),
		reflect.TypeOf(""),
		reflect.TypeOf(false),
		reflect.TypeOf(""),
		reflect.TypeOf(time.Time{}),
		reflect.TypeOf(int64(0)),
	}
)

var (
	fileTable   = sqobj.MustRegisterClass(N(fileTableName), File{})
	docTable    = sqobj.MustRegisterClass(N(docTableName), Doc{}).ForeignKey(fileTable)
	viewTable   = sqobj.MustRegisterView(N(viewTableName), View{}, true, fileTable, docTable)
	searchTable = sqobj.MustRegisterVirtual(N(searchTableName), "fts5", Search{}, "content="+Quote(viewTableName))
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
		if err := fileTable.Create(txn, schema); err != nil {
			return err
		}
		if err := docTable.Create(txn, schema); err != nil {
			return err
		}
		if err := viewTable.Create(txn, schema); err != nil {
			return err
		}
		if err := searchTable.Create(txn, schema, "tokenize="+Quote(tokenizer)); err != nil {
			return err
		}
		// triggers to keep the FTS index up to date
		// https://www.sqlite.org/fts5.html
		if _, err := txn.Query(N(searchTriggerInsertName).WithSchema(schema).CreateTrigger(fileTableName,
			Q("INSERT INTO ", searchTableName, " (rowid, name, parent, filename) VALUES (new.rowid, new.name, new.parent, new.filename)"),
		).After().Insert().IfNotExists()); err != nil {
			return err
		}
		if _, err := txn.Query(N(searchTriggerDeleteName).WithSchema(schema).CreateTrigger(fileTableName,
			Q("INSERT INTO ", searchTableName, " (", searchTableName, ", rowid, name, parent, filename) VALUES ('delete', old.rowid, old.name, old.parent, old.filename)"),
		).After().Delete().IfNotExists()); err != nil {
			return err
		}
		if _, err := txn.Query(N(searchTriggerUpdateName).WithSchema(schema).CreateTrigger(fileTableName,
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
		s := Q("SELECT name,COUNT(*) AS count FROM ", N(fileTableName).WithSchema(schema), " GROUP BY name")
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
	return N(fileTableName).WithSchema(schema).Insert(
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
	return N(fileTableName).WithSchema(schema).Delete(Q("name=?"), Q("path=?")),
		[]interface{}{evt.Name, evt.Path}
}

func GetFile(schema string, rowid int64) (SQStatement, []interface{}, []reflect.Type) {
	return S(N(fileTableName).WithSchema(schema)).
		To(N("name"), N("path"), N("parent"), N("filename"), N("isdir"), N("ext"), N("modtime"), N("size")).
		Where(Q("rowid", "=", P)), []interface{}{rowid}, filesTypeCast
}

func UpsertDoc(txn SQTransaction, doc *Doc) (int64, error) {
	if n, err := docTable.UpsertKeys(txn, doc); err != nil {
		return 0, err
	} else {
		return n[0], nil
	}
}

func Query(schema string, snippet bool) SQSelect {
	// Set the query join
	queryJoin := J(
		N(searchTableName).WithSchema(schema),
		N(fileTableName).WithSchema(schema),
	).LeftJoin(Q(N(searchTableName), ".rowid=", N(fileTableName), ".rowid"))
	// Set the snippet expression
	snippetExpr := V("")
	if snippet {
		snippetExpr = Q("SNIPPET(", searchTableName, ",-1, '<em>', '</em>', '...', 64) AS snippet")
	}
	// Return the select
	return S(queryJoin).To(
		N("rowid").WithSchema(searchTableName),
		N("rank").WithSchema(searchTableName),
		snippetExpr,
		N("name").WithSchema(fileTableName),
		N("path").WithSchema(fileTableName),
		N("parent").WithSchema(fileTableName),
		N("filename").WithSchema(fileTableName),
		N("isdir").WithSchema(fileTableName),
		N("ext").WithSchema(fileTableName),
		N("modtime").WithSchema(fileTableName),
		N("size").WithSchema(fileTableName),
	).Where(Q(searchTableName, " MATCH ", P)).Order(N("rank"))
}
