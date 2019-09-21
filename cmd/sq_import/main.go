/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	sqlite "github.com/djthorpe/sqlite"
	tablewriter "github.com/olekukonko/tablewriter"

	// Modules
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/sqlite/sys/sqlite"
)

// BoundRow returns a slice of type []interface{} from a slice of type []string
func BoundRow(row []string) []interface{} {
	row_ := []interface{}{}
	for i := range row {
		row_ = append(row_, row[i])
	}
	return row_
}

// CreateTable creates a new table and inserts rows from CSV file
func CreateTable(db sqlite.Connection, table *Table) (int, error) {
	affectedRows := 0

	// Wrap SQL statements in a transaction
	return affectedRows, db.Tx(func(db sqlite.Connection) error {
		if _, err := db.Do(db.NewDropTable(table.Name).IfExists()); err != nil {
			return err
		}
		if _, err := db.Do(db.NewCreateTable(table.Name, table.Columns...)); err != nil {
			return err
		}
		if insert := db.NewInsert(table.Name); insert == nil {
			return gopi.ErrBadParameter
		} else {
			for {
				if row, err := table.Next(); err == io.EOF {
					break
				} else if err != nil {
					return err
				} else if result, err := db.Do(insert, BoundRow(row)...); err != nil {
					return err
				} else {
					affectedRows = affectedRows + int(result.RowsAffected)
				}
			}
		}
		return nil
	})
}

// ShowTable outputs an SQL table to the screen
func ShowTable(db sqlite.Connection, table *Table) error {
	if src := db.NewSource(table.Name); src == nil {
		return gopi.ErrBadParameter
	} else if st := db.NewSelect(src); st == nil {
		return gopi.ErrBadParameter
	} else if rows, err := db.Query(st); err != nil {
		return err
	} else {
		tablewriter := tablewriter.NewWriter(os.Stdout)
		header := make([]string, len(rows.Columns()))
		for i, column := range rows.Columns() {
			header[i] = column.Name()
		}
		tablewriter.SetHeader(header)
		tablewriter.SetAutoFormatHeaders(false)

		for {
			if values := rows.Next(); values == nil {
				break
			} else {
				row := make([]string, len(values))
				for i, value := range values {
					if value.IsNull() {
						row[i] = "<null>"
					} else {
						row[i] = value.String()
					}
				}
				tablewriter.Append(row)

			}
		}

		tablewriter.Render()
	}

	// Success
	return nil
}

func Process(app *gopi.AppInstance, db sqlite.Connection, name string, fh io.ReadSeeker) error {
	// Create a table
	table := NewTable(fh, db, name)
	table.NoHeader, _ = app.AppFlags.GetBool("noheader")
	table.SkipComments, _ = app.AppFlags.GetBool("skipcomments")
	table.NotNull, _ = app.AppFlags.GetBool("notnull")

	// Scan rows for column names and types
	if affectedRows, err := table.Scan(); err != nil {
		return fmt.Errorf("%v (line %v)", err, affectedRows)
	}

	// Create the table if it doesn't exist
	app.Logger.Info("Creating table %v with %d columns", strconv.Quote(table.Name), len(table.Columns))

	// Repeat until all rows read
	if affectedRows, err := CreateTable(db, table); err != nil {
		return err
	} else if err := ShowTable(db, table); err != nil {
		return err
	} else {
		app.Logger.Info("%v rows imported", affectedRows)
	}

	// Return success
	return nil
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {

	if db := app.ModuleInstance("db/sqlite").(sqlite.Connection); db == nil {
		return gopi.ErrAppError
	} else if len(app.AppFlags.Args()) == 0 {
		return gopi.ErrHelp
	} else {
		for _, filename := range app.AppFlags.Args() {
			name := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
			if s, err := os.Stat(filename); err != nil {
				return fmt.Errorf("%v: %v", name, err)
			} else if s.Mode().IsRegular() == false {
				return fmt.Errorf("%v: Not a regular file", name)
			} else if fh, err := os.Open(filename); err != nil {
				return err
			} else {
				defer fh.Close()
				if err := Process(app, db, name, fh); err != nil {
					return fmt.Errorf("%v: %v", name, err)
				}
			}
		}
	}

	// Success
	done <- gopi.DONE
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("db/sqlite")

	// Set arguments
	config.AppFlags.FlagBool("noheader", false, "Do not use the first row as column names")
	config.AppFlags.FlagBool("skipcomments", true, "Skip comment lines")
	config.AppFlags.FlagBool("notnull", false, "Don't use NULL values for empty values")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main))
}
