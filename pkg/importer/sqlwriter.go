package importer

import (
	// Modules

	sqlite3 "github.com/djthorpe/go-sqlite/sys/sqlite3"
	multierror "github.com/hashicorp/go-multierror"

	// Namespace imports
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type SQLWriter struct {
	*sqlite3.ConnEx
	overwrite bool
	n         int
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewSQLWriter(c SQImportConfig, db *sqlite3.ConnEx) (*SQLWriter, error) {
	return &SQLWriter{db, c.Overwrite, 0}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Begin a transaction, passing a writing function back to the caller
func (w *SQLWriter) Begin(name, schema string, cols []string) (SQImportWriterFunc, error) {
	// Start transaction
	if err := w.ConnEx.Begin(sqlite3.SQLITE_TXN_DEFAULT); err != nil {
		return nil, err
	}

	// Set schema
	if schema == "" {
		schema = sqlite3.DefaultSchema
	}

	// Drop table if overwrite is enabled
	if w.overwrite {
		if err := w.dropTable(name, schema); err != nil {
			w.ConnEx.Rollback()
			return nil, err
		}
	}

	// Create table if it doesn't exist
	if err := w.createTable(name, schema, cols); err != nil {
		w.ConnEx.Rollback()
		return nil, err
	}

	// Add columns
	if err := w.addColumns(name, schema, cols); err != nil {
		w.ConnEx.Rollback()
		return nil, err
	}

	// Make function to write rows
	fn := func(row []interface{}) error {
		changes, err := w.writer(name, schema, cols, row)
		w.n += changes
		return err
	}

	// Reset counter
	w.n = 0

	// Return function
	return fn, nil
}

func (w *SQLWriter) End(success bool) error {
	if success {
		return w.ConnEx.Commit()
	} else {
		return w.ConnEx.Rollback()
	}
}

func (w *SQLWriter) Count() int {
	return w.n
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (w *SQLWriter) writer(name, schema string, cols []string, values []interface{}) (int, error) {
	if err := w.ConnEx.ExecEx(N(name).WithSchema(schema).Insert(cols...).Query(), nil, values...); err != nil {
		return int(w.ConnEx.LastInsertId()), err
	} else {
		return int(w.ConnEx.Changes()), nil
	}
}

func (this *SQLWriter) dropTable(name, schema string) error {
	drop := N(name).WithSchema(schema).DropTable().IfExists()
	if err := this.Exec(drop.Query(), nil); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *SQLWriter) createTable(name, schema string, cols []string) error {
	create := N(name).WithSchema(schema).CreateTable(sqlToCols(cols)...).IfNotExists()
	if err := this.Exec(create.Query(), nil); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *SQLWriter) addColumns(name, schema string, cols []string) error {
	var result error
	if err := this.Exec(Q("PRAGMA ", N(schema), ".table_info(", N(name), ")").Query(), func(row, col []string) bool {
		if !columnExists(cols, row[1]) {
			if err := this.Exec(N(name).WithSchema(schema).AlterTable().AddColumn(C(row[1])).Query(), nil); err != nil {
				result = multierror.Append(result, err)
				return true
			}
		}
		return false
	}); err != nil {
		result = multierror.Append(result, err)
	}

	// Return any errors
	return result
}

func sqlToCols(colnames []string) []SQColumn {
	result := make([]SQColumn, len(colnames))
	for i, colname := range colnames {
		result[i] = C(colname)
	}
	return result
}

func columnExists(v []string, name string) bool {
	for _, col := range v {
		if col == name {
			return true
		}
	}
	return false
}
