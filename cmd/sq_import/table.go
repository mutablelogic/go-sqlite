/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package main

import (
	"fmt"
	"strconv"

	// Frameworks
	"github.com/araddon/dateparse"
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////

type Columns struct {
	Name       []string
	Candidates []map[string]bool
}

////////////////////////////////////////////////////////////////////////////////

func NewColumns(filename string) *Columns {
	this := new(Columns)
	return this
}

func (this *Columns) SetNames(row []string) error {
	if len(row) == 0 {
		return gopi.ErrBadParameter
	} else {
		this.Name = row
		return nil
	}
}

func (this *Columns) SetTypes(row []string) error {
	if len(row) == 0 {
		return gopi.ErrBadParameter
	} else if this.Name == nil {
		this.Name = make([]string, len(row))
		for i := 0; i < len(row); i++ {
			this.Name[i] = fmt.Sprintf("column%03d", i)
		}
	}
	if len(row) != len(this.Name) {
		return fmt.Errorf("Row size mismatch")
	}
	// Set up candidates
	if this.Candidates == nil {
		supportedTypes := sqlite.SupportedTypes()
		this.Candidates = make([]map[string]bool, len(row))
		for i := range this.Candidates {
			this.Candidates[i] = make(map[string]bool, len(supportedTypes))
			for _, t := range supportedTypes {
				this.Candidates[i][t] = true
			}
		}
	}
	// Check off candidates for this row
	for i, value := range row {
		candidates := this.Candidates[i]
		if ok, exists := candidates["BOOL"]; ok && exists {
			candidates["BOOL"] = isBool(value)
		}
		if ok, exists := candidates["INTEGER"]; ok && exists {
			candidates["INTEGER"] = isInteger(value)
		}
		if ok, exists := candidates["FLOAT"]; ok && exists {
			candidates["FLOAT"] = isFloat(value)
		}
		if ok, exists := candidates["DATETIME"]; ok && exists {
			candidates["DATETIME"] = isDatetime(value)
		}
		if ok, exists := candidates["TIMESTAMP"]; ok && exists {
			candidates["TIMESTAMP"] = isTimestamp(value)
		}
		if ok, exists := candidates["BLOB"]; ok && exists {
			candidates["BLOB"] = isBlob(value)
		}
	}
	return nil
}

func (this *Columns) Types() []string {
	types := make([]string, len(this.Candidates))
	for i, candidates := range this.Candidates {
		if ok, exists := candidates["BOOL"]; exists && ok {
			types[i] = "BOOL"
		} else if ok, exists := candidates["INTEGER"]; exists && ok {
			types[i] = "INTEGER"
		} else if ok, exists := candidates["BLOB"]; exists && ok {
			types[i] = "BLOB"
		} else if ok, exists := candidates["FLOAT"]; exists && ok {
			types[i] = "FLOAT"
		} else {
			// Always fallback to TEXT
			types[i] = "TEXT"
		}
	}
	return types
}

func (this *Columns) String() string {
	return fmt.Sprintf("<Columns>{ names=%v types=%v }", this.Name, this.Types())
}

func isBool(value string) bool {
	if _, err := strconv.ParseBool(value); err != nil {
		return false
	} else {
		return true
	}
}

func isInteger(value string) bool {
	if _, err := strconv.ParseInt(value, 10, 64); err != nil {
		return false
	} else {
		return true
	}
}

func isFloat(value string) bool {
	if _, err := strconv.ParseFloat(value, 64); err != nil {
		return false
	} else {
		return true
	}
}

func isDatetime(value string) bool {
	if _, err := dateparse.ParseAny(value); err != nil {
		fmt.Println(value, "=> false")
		return false
	} else {
		// TODO: Support DD/MM/YYYY
		fmt.Println(value, "=> true")
		return true
	}
}

func isTimestamp(value string) bool {
	// Not supported
	return false
}

func isBlob(value string) bool {
	// Not supported
	return false
}
