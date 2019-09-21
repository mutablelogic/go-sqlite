/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	// Frameworks
	"github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////

type Table struct {
	// Table is the name of the table
	Name string
	// NoHeader when set to false uses row 1 as header names
	NoHeader bool
	// Comment prefix
	Comment rune
	// NotNull excludes NULL values from columns
	NotNull bool
	// Columns is the name of the columns
	Columns []sqlite.Column
	// Candidates for the column type
	candidates []map[string]bool
	// Database connection
	db sqlite.Connection
	// File handles
	fh  io.ReadSeeker
	csv *csv.Reader
	// First row (seek to zero positon)
	first bool
	row   int
}

////////////////////////////////////////////////////////////////////////////////

// Create a new empty table to be imported
func NewTable(fh io.ReadSeeker, db sqlite.Connection, name string) *Table {
	this := new(Table)
	this.Name = strings.ToLower(name)
	this.NoHeader = false
	this.NotNull = false
	this.candidates = make([]map[string]bool, 0, 10)
	this.Columns = make([]sqlite.Column, 0, 10)
	this.db = db
	this.fh = fh
	this.csv = csv.NewReader(fh)
	this.first = true
	this.row = -1
	return this
}

// NextRow scans the CSV file and returns an io.EOF error on end of the file
func (this *Table) nextRow() ([]string, error) {
	if this.first {
		// Seek to start of file
		if _, err := this.fh.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		// Reset
		this.csv.Comment = this.Comment
		this.first = false
	}
	if row, err := this.csv.Read(); err == io.EOF {
		this.first = true
		return nil, err
	} else {
		return row, err
	}
}

// Scan the whole CSV file and set the column name and types
// and return the number of scanned columns on success
func (this *Table) Scan() (int, error) {
	affectedRows := 0
	maxColumns := 0
	for {
		row, err := this.nextRow()
		is_header := affectedRows == 0 && this.NoHeader == false
		if err == io.EOF {
			// EOF
			break
		} else if is_header {
			// Set column headers
			this.Columns = this.namedColumns(row)
		}

		// Alter the maximum number of columns
		if len(row) > maxColumns {
			maxColumns = len(row)
		}

		// Append any new candidates and columns
		for {
			if len(this.candidates) < maxColumns {
				this.candidates = append(this.candidates, newCandidates())
			} else {
				break
			}
		}
		for {
			if len(this.Columns) < maxColumns {
				this.Columns = append(this.Columns, this.genericColumn(len(this.Columns)))
			} else {
				break
			}
		}

		// Increment affectedRows
		affectedRows++

		// Infer types
		if is_header == false {
			for i, value := range row {
				if len(this.candidates[i]) > 1 {
					checkCandidates(this.candidates[i], sqlite.SupportedTypesForValue(value))
				}
			}
		}
	}

	// Now we can set the declared types for columns
	for i, column := range this.Columns {
		this.Columns[i] = this.db.NewColumn(column.Name(), this.TypeForColumn(i), column.Nullable(), column.PrimaryKey())
	}

	return affectedRows, nil
}

func (this *Table) Next() ([]string, int, error) {
	if this.row == -1 {
		this.row = 0
	}
	for {
		if row, err := this.nextRow(); err == io.EOF {
			return row, this.row, err
		} else if err != nil {
			return row, this.row, err
		} else if this.row == 0 && this.NoHeader == false {
			// Skip header
			this.row = 1
		} else {
			this.row++
			return row, this.row, nil
		}
	}
}

// Remove unsupported types for a column
func (this *Table) TypeForColumn(i int) string {
	supported_types := sqlite.SupportedTypes()
	candidates := this.candidates[i]
	for j := len(supported_types) - 1; j >= 0; j-- {
		decltype := supported_types[j]
		if _, exists := candidates[decltype]; exists {
			return decltype
		}
	}
	return supported_types[0]
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Remove candidates if they are no longer valid
func checkCandidates(candidates map[string]bool, types []string) {
	// Make a map
	types_map := make(map[string]bool, len(types))
	for _, decltype := range types {
		types_map[decltype] = true
	}
	// Remove candidates which don't yet exist
	for decltype, _ := range candidates {
		if _, exists := types_map[decltype]; exists == false {
			delete(candidates, decltype)
		}
	}
}

func isComment(line string) bool {
	if strings.HasPrefix(line, "#") {
		return true
	} else if strings.HasPrefix(line, "//") {
		return true
	} else {
		return false
	}
}

func (this *Table) genericColumn(pos int) sqlite.Column {
	default_type := sqlite.SupportedTypes()[0]
	return this.db.NewColumn(genericNameForColumn(pos), default_type, this.NotNull == false, false)
}

func (this *Table) namedColumns(names []string) []sqlite.Column {
	columns := make([]sqlite.Column, len(names))
	default_type := sqlite.SupportedTypes()[0]
	for i, name := range names {
		if name == "" {
			name = genericNameForColumn(i)
		}
		columns[i] = this.db.NewColumn(name, default_type, this.NotNull == false, false)
	}
	return columns
}

func newCandidates() map[string]bool {
	supportedTypes := sqlite.SupportedTypes()
	candidates := make(map[string]bool, len(supportedTypes))
	for _, t := range supportedTypes {
		candidates[t] = true
	}
	return candidates
}

func genericNameForColumn(pos int) string {
	return fmt.Sprintf("column%03d", pos+1)
}
