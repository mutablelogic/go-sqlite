package sqimport

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
	insert    *sqlite3.StatementEx
}

type WriterFunc func(row []interface{}) error

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewSQLWriter(c SQImportConfig, db *sqlite3.ConnEx) (*SQLWriter, error) {
	return &SQLWriter{db, c.Overwrite, nil}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Begin a transaction, passing a writing function back to the caller
func (this *SQLWriter) Begin(name, schema string, cols []string) (WriterFunc, error) {
	// Start transaction
	if err := this.ConnEx.Begin(sqlite3.SQLITE_TXN_DEFAULT); err != nil {
		return nil, err
	}

	// Drop table if overwrite is enabled
	if this.overwrite {
		if err := this.dropTable(name, schema); err != nil {
			this.ConnEx.Rollback()
			return nil, err
		}
	}

	// Create table if it doesn't exist
	if err := this.createTable(name, schema, cols); err != nil {
		this.ConnEx.Rollback()
		return nil, err
	}

	// TODO: Add columns onto the table if necessary

	// Make prepared statement for insert
	if st, err := this.ConnEx.Prepare(N(name).WithSchema(schema).Insert(cols...).Query()); err != nil {
		this.ConnEx.Rollback()
		return nil, err
	} else {
		this.insert = st
	}

	// Pass back the function to write a row of data
	return this.writer, nil
}

func (this *SQLWriter) End(success bool) error {
	var result error

	// Close prepared statement
	if err := this.insert.Close(); err != nil {
		result = multierror.Append(result, err)
	}

	// Return success or failure
	if success {
		result = multierror.Append(result, this.ConnEx.Commit())
	} else {
		result = multierror.Append(result, this.ConnEx.Rollback())
	}

	// Return any errors
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *SQLWriter) writer(row []interface{}) error {
	if _, err := this.insert.Exec(0, row...); err != nil {
		return err
	} else {
		return nil
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

func (this *SQLWriter) addColumns(name, schema string, cols []SQColumn) error {
	var result error
	if err := this.Exec(Q("PRAGMA table_info(", N(name).WithSchema(schema), ")").Query(), func(row, col []string) bool {
		if columnExists(cols, row[1]) == false {
			if err := this.Exec(N(name).WithSchema(schema).AlterTable().AddColumn(C(row[1])).Query(), nil); err != nil {
				// Abort on error
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

func columnExists(v []SQColumn, name string) bool {
	for _, col := range v {
		if col.Name() == name {
			return true
		}
	}
	return false
}
