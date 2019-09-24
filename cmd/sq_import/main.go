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
	"unicode/utf8"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	sqlite "github.com/djthorpe/sqlite"
	tablewriter "github.com/olekukonko/tablewriter"

	// Modules
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/sqlite/sys/sqlite"
)

// BoundRow returns a slice of type []interface{} from a slice of type []string
func BoundRow(row []string, notnull bool) []interface{} {
	row_ := []interface{}{}
	for i := range row {
		if row[i] == "" && notnull == false {
			row_ = append(row_, nil)
		} else {
			row_ = append(row_, row[i])
		}
	}
	return row_
}

// CreateTable creates a new table and inserts rows from CSV file
func CreateTable(db sqlite.Connection, lang sqlite.Language, table *Table) (int, error) {
	affectedRows := 0

	// Wrap SQL statements in a transaction
	return affectedRows, db.Txn(func(txn sqlite.Transaction) error {
		if _, err := txn.Do(lang.NewDropTable(table.Name).IfExists()); err != nil {
			return err
		}
		if _, err := txn.Do(lang.NewCreateTable(table.Name, table.Columns...)); err != nil {
			return err
		}
		if insert := lang.NewInsert(table.Name); insert == nil {
			return gopi.ErrBadParameter
		} else {
			for {
				if row, _, err := table.Next(); err == io.EOF {
					break
				} else if err != nil {
					return err
				} else if result, err := db.Do(insert, BoundRow(row, table.NotNull)...); err != nil {
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
func ShowTable(db sqlite.Connection, lang sqlite.Language, table *Table) error {
	if src := lang.NewSource(table.Name); src == nil {
		return gopi.ErrBadParameter
	} else if st := lang.NewSelect(src); st == nil {
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

func Process(app *gopi.AppInstance, name string, fh io.ReadSeeker) error {
	// Components
	db := app.ModuleInstance("db/sqlite").(sqlite.Connection)
	lang := app.ModuleInstance("db/sqlang").(sqlite.Language)

	// Create a table
	table := NewTable(fh, db, name)
	table.NoHeader, _ = app.AppFlags.GetBool("noheader")
	table.NotNull, _ = app.AppFlags.GetBool("notnull")

	// Set the comment
	if comment, exists := app.AppFlags.GetString("comment"); exists {
		table.Comment, _ = utf8.DecodeRuneInString(comment)
	}

	// Infer column headers and types
	if affectedRows, err := table.Scan(); err != nil {
		return err
	} else {
		app.Logger.Info("%v rows scanned", affectedRows)
	}

	app.Logger.Info("Creating table %v with %d columns", strconv.Quote(table.Name), len(table.Columns))

	// Repeat until all rows read
	if affectedRows, err := CreateTable(db, lang, table); err != nil {
		return err
	} else if err := ShowTable(db, lang, table); err != nil {
		return err
	} else {
		app.Logger.Info("%v rows imported", affectedRows)
	}

	// Return success
	return nil
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	if len(app.AppFlags.Args()) == 0 {
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
				if err := Process(app, name, fh); err != nil {
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
	config := gopi.NewAppConfig("db/sqlite", "db/sqlang")

	// Set arguments
	config.AppFlags.FlagBool("noheader", false, "Do not use the first row as column names")
	config.AppFlags.FlagString("comment", "#", "Comment line prefix")
	config.AppFlags.FlagBool("notnull", false, "Don't use NULL values for empty values")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main))
}
