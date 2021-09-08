package sqimport

import (
	// Modules

	"github.com/djthorpe/go-sqlite"
	sqlite3 "github.com/djthorpe/go-sqlite/sys/sqlite3"
	multierror "github.com/hashicorp/go-multierror"

	// Namespace imports
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type sqlwriter struct {
	*sqlite3.ConnEx
	overwrite    bool
	name, schema string
	cols         []string
	insert       *sqlite3.StatementEx
	n            int
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewSQLWriter(c SQImportConfig, db *sqlite3.ConnEx) (sqlite.SQWriter, error) {
	return &sqlwriter{db, c.Overwrite, "", "", nil, nil, 0}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *sqlwriter) Write(name, schema string, col []string, row []interface{}) error {
	if name != this.name || schema != this.schema {
		// End any previous transaction
		if !this.Autocommit() {
			if err := this.Commit(); err != nil {
				return err
			}
		}

		// Start transaction
		if err := this.Begin(sqlite3.SQLITE_TXN_DEFAULT); err != nil {
			return err
		}

		// Drop table if overwrite is enabled
		if this.overwrite {
			if err := this.dropTable(name, schema); err != nil {
				return err
			}
		}

		// Create table if it doesn't exist
		if err := this.createTable(name, schema, col); err != nil {
			return err
		}

		// Set current name/schema
		this.name = name
		this.schema = schema
		this.cols = nil

		// Reset prepared insert statement
		if this.insert != nil {
			if err := this.insert.Close(); err != nil {
				return err
			}
			this.insert = nil
		}
	}

	// Add rows
	if this.cols == nil {
		// Alter table to add columns
		if err := this.addColumns(name, schema, sqlToCols(col)); err != nil {
			return err
		}

		// Set the columns
		this.cols = col
	}

	// Insert columns
	if err := this.insertRow(name, schema, col, row); err != nil {
		return err
	} else {
		this.n = this.n + 1
	}

	// Return success
	return nil
}

func (this *sqlwriter) Close() error {
	var result error

	// End any previous transaction
	if !this.Autocommit() {
		if err := this.Commit(); err != nil {
			return err
		}
	}

	// Release non-connection resources
	this.name = ""
	this.schema = ""
	this.cols = nil

	if this.insert != nil {
		if err := this.insert.Close(); err != nil {
			result = err
		}
	}

	// Return any errors
	return result
}

func (this *sqlwriter) Count() int {
	return this.n
}

func (this *sqlwriter) Reset() {
	this.n = 0
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *sqlwriter) dropTable(name, schema string) error {
	drop := N(name).WithSchema(schema).DropTable().IfExists()
	if err := this.Exec(drop.Query(), nil); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *sqlwriter) createTable(name, schema string, cols []string) error {
	create := N(name).WithSchema(schema).CreateTable(sqlToCols(cols)...).IfNotExists()
	if err := this.Exec(create.Query(), nil); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *sqlwriter) addColumns(name, schema string, cols []SQColumn) error {
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

func (this *sqlwriter) insertRow(name, schema string, cols []string, row []interface{}) error {
	// Prepare insert statement
	if this.insert == nil || len(cols) != len(this.cols) {
		if st, err := this.Prepare(N(name).WithSchema(schema).Insert(cols...).Query()); err != nil {
			return err
		} else {
			this.insert = st
		}
	}

	// Insert row
	if _, err := this.insert.Exec(0, row...); err != nil {
		return err
	}

	// Return success
	return nil
}

func sqlToCols(colnames []string) []SQColumn {
	result := make([]SQColumn, len(colnames))
	for i, colname := range colnames {
		result[i] = C(colname)
	}
	return result
}

func columnExists(v []sqlite.SQColumn, name string) bool {
	for _, col := range v {
		if col.Name() == name {
			return true
		}
	}
	return false
}
