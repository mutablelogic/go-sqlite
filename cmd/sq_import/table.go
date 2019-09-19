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
	// Skip lines which start with a comment (// or #)
	SkipComments bool
	// NotNull excludes NULL values from columns
	NotNull bool
	// Columns is the name of the columns
	Columns []sqlite.Column
	// Candidates for the column type
	Candidates []map[string]bool
	// File handle
	fh io.ReadSeeker
	// CSV Reader
	reader *csv.Reader
	// First row (seek to zero positon)
	first int
}

////////////////////////////////////////////////////////////////////////////////

// Create a new empty table to be imported
func NewTable(fh io.ReadSeeker, name string) *Table {
	this := new(Table)
	this.Name = strings.ToLower(name)
	this.NoHeader = false
	this.NotNull = false
	this.SkipComments = true
	this.fh = fh
	this.reader = csv.NewReader(fh)
	this.first = -1
	return this
}

// Scan the CSV file and set the column name and type
func (this *Table) Scan() (int, error) {
	// Seek to start of file
	if _, err := this.fh.Seek(0, io.SeekStart); err != nil {
		return -1, err
	}
	// Iterate through values to set header names and types
	affectedRows := 0
	for i := 0; true; i++ {
		if row, err := this.reader.Read(); err == io.EOF {
			// EOF
			break
		} else if err != nil {
			return i + 1, err
		} else if len(row) == 0 || (len(row) == 1 && strings.TrimSpace(row[0]) == "") {
			// Skip empty rows
		} else if this.SkipComments && isComment(row) {
			// Skip rows with comments
		} else if i == 0 && this.NoHeader == false {
			// Set column headers
			this.Columns = namedColumns(row)
		} else {
			if this.Columns == nil {
				// Set columns with generic names
				this.Columns = genericColumns(len(row))
			}
			if this.Candidates == nil {
				// Set empty type candidates
				this.Candidates = make([]map[string]bool, len(row))
				for i := range this.Candidates {
					this.Candidates[i] = newCandidates()
				}
			}
			// Infer types
			for i, value := range row {
				if len(this.Candidates[i]) > 1 {
					checkCandidates(this.Candidates[i], sqlite.SupportedTypesForValue(value))
				}
			}
			affectedRows = affectedRows + 1
		}
	}

	// Success
	return affectedRows, nil
}

// Next returns the next row
func (this *Table) Next() ([]string, error) {
	// Seek to start of file
	if this.first == -1 {
		if _, err := this.fh.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
	}
	for {
		this.first = this.first + 1
		if row, err := this.reader.Read(); err == io.EOF {
			this.first = -1
			return nil, err
		} else if len(row) == 0 || (len(row) == 1 && strings.TrimSpace(row[0]) == "") {
			// Skip empty rows
		} else if this.SkipComments && isComment(row) {
			// Skip comments
		} else if this.first == 0 && this.NoHeader == false {
			// Skip header
		} else {
			return row, nil
		}
	}
}

// Remove unsupported types for a column
func (this *Table) TypeForColumn(i int) string {
	supported_types := sqlite.SupportedTypes()
	candidates := this.Candidates[i]
	for j := len(supported_types) - 1; j >= 0; j-- {
		decltype := supported_types[j]
		if _, exists := candidates[decltype]; exists {
			return decltype
		}
	}
	return supported_types[0]
}

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

func (this *Table) CreateTable() string {
	columns := make([]string, len(this.Columns))
	for i := range this.Columns {
		columns[i] = fmt.Sprintf("%v %v", sqlite.QuoteIdentifier(this.Columns[i]), this.TypeForColumn(i))
	}
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %v (%v)", sqlite.QuoteIdentifier(this.Name), strings.Join(columns, ","))
}

func (this *Table) InsertRow() string {
	columns := make([]string, len(this.Columns))
	for i := range this.Columns {
		columns[i] = "?"
	}
	return fmt.Sprintf("INSERT INTO %v VALUES (%v)", sqlite.QuoteIdentifier(this.Name), strings.Join(columns, ","))
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func isComment(row []string) bool {
	if len(row) == 0 {
		return false
	} else if strings.HasPrefix(row[0], "#") {
		return true
	} else if strings.HasPrefix(row[0], "//") {
		return true
	} else {
		return false
	}
}

func genericColumns(size int) []sqlite.Column {
	columns := make([]string, size)
	for i := 0; i < size; i++ {
		columns[i] = fmt.Sprintf("column%03d", i)
	}
	return columns
}

func namedColumns(names []string) []sqlite.Column {
	columns := make([]sqlite.Column, len(names))

}

func newCandidates() map[string]bool {
	supportedTypes := sqlite.SupportedTypes()
	candidates := make(map[string]bool, len(supportedTypes))
	for _, t := range supportedTypes {
		candidates[t] = true
	}
	return candidates
}
