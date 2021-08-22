package sqimport

import (
	"encoding/csv"
	"fmt"
	"io"

	// Modules
	sqlite "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type csvdecoder struct {
	r            *csv.Reader
	header       bool
	name, schema string
	cols         []string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewCSVDecoder returns a CSV decoder setting options
func (this *importer) NewCSVDecoder(r io.Reader, delimiter rune) (sqlite.SQImportDecoder, error) {
	decoder := &csvdecoder{csv.NewReader(r), this.c.Header, this.c.Name, this.c.Schema, nil}

	// Set delimiter
	if this.c.Delimiter != 0 {
		decoder.r.Comma = this.c.Delimiter
	} else if delimiter != 0 {
		decoder.r.Comma = delimiter
	}

	// Set other options
	if this.c.Comment != 0 {
		decoder.r.Comment = this.c.Comment
	}
	decoder.r.TrimLeadingSpace = this.c.TrimSpace
	decoder.r.LazyQuotes = this.c.LazyQuotes
	decoder.r.ReuseRecord = true

	// Return success
	return decoder, nil
}

// Read reads a CSV record, and returns io.EOF on end of reading
func (this *csvdecoder) Read(w sqlite.SQWriter) error {
	// Check arguments
	if w == nil {
		return sqlite.ErrBadParameter.With("SQLWriter")
	}
	// Read a row
	row, err := this.r.Read()
	if err != nil {
		return err
	}

	// Initialize the reader
	if this.cols == nil {
		this.cols = make([]string, len(row))
		for i, col := range row {
			if this.header {
				this.cols[i] = col
			} else {
				this.cols[i] = this.makeCol(i)
			}
		}
		if this.header {
			return nil
		}
	}

	// Add new column headings as necessary
	for len(row) > len(this.cols) {
		this.cols = append(this.cols, this.makeCol(len(this.cols)))
	}

	// Write the row
	return w.Write(this.name, this.schema, this.cols[:len(row)], csvRow(row))
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Return a column heading for the given index
func (this *csvdecoder) makeCol(i int) string {
	return fmt.Sprintf("col_%02d", i)
}

// Return row as []interface{} from []string
func csvRow(v []string) []interface{} {
	result := make([]interface{}, len(v))
	for i, s := range v {
		result[i] = s
	}
	return result
}
