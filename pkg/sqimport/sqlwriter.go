package sqimport

import (

	// Modules
	"github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type sqlwriter struct {
	sqlite.SQConnection
	overwrite bool
	cols      []string
	insert    sqlite.SQStatement
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewSQLWriter(c sqlite.SQImportConfig, db sqlite.SQConnection) (sqlite.SQWriter, error) {
	return &sqlwriter{db, c.Overwrite, nil, nil}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *sqlwriter) Write(name, schema string, col []string, row []interface{}) error {
	// Create table if it doesn't exist
	if this.cols == nil {
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
	}

	// Return success
	return nil
}

func (this *sqlwriter) Close() error {
	// Release non-connection resources
	this.cols = nil
	this.insert = nil

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *sqlwriter) dropTable(name, schema string) error {
	drop := N(name).WithSchema(schema).DropTable().IfExists()
	if _, err := this.Exec(drop); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *sqlwriter) createTable(name, schema string, cols []string) error {
	create := N(name).WithSchema(schema).CreateTable(sqlToCols(cols)...).IfNotExists()
	if _, err := this.Exec(create); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *sqlwriter) addColumns(name, schema string, cols []sqlite.SQColumn) error {
	zip := zipCols(this.ColumnsEx(name, schema))
	for _, col := range cols {
		if _, exists := zip[col.Name()]; !exists {
			if _, err := this.Exec(N(name).WithSchema(schema).AlterTable().AddColumn(col)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *sqlwriter) insertRow(name, schema string, cols []string, row []interface{}) error {
	// Prepare insert statement
	if this.insert == nil || len(cols) != len(this.cols) {
		if q, err := this.Prepare(N(name).WithSchema(schema).Insert(cols...)); err != nil {
			return err
		} else {
			this.insert = q
		}
	}
	// Insert data
	if _, err := this.Exec(this.insert, row...); err != nil {
		return err
	} else {
		return nil
	}
}

func sqlToCols(colnames []string) []sqlite.SQColumn {
	result := make([]sqlite.SQColumn, len(colnames))
	for i, colname := range colnames {
		result[i] = C(colname)
	}
	return result
}

func zipCols(v []sqlite.SQColumn) map[string]sqlite.SQColumn {
	result := make(map[string]sqlite.SQColumn, len(v))
	for _, v := range v {
		result[v.Name()] = v
	}
	return result
}
