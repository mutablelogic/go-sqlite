package importer

import (
	"encoding/csv"
	"fmt"
	"io"

	// Namespace Imports
	. "github.com/mutablelogic/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type csvdecoder struct {
	c      io.Closer
	r      *csv.Reader
	header bool
	cols   []string
	values []interface{}
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewCSVDecoder returns a CSV decoder setting options
func (this *Importer) NewCSVDecoder(c io.Closer, r io.Reader, delimiter rune) (SQImportDecoder, error) {
	decoder := &csvdecoder{c, csv.NewReader(r), this.c.Header, nil, nil}

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

func (dec *csvdecoder) Close() error {
	return dec.c.Close()
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (dec *csvdecoder) String() string {
	return fmt.Sprintf("<text/csv delimiter=%q>", dec.r.Comma)
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

// Read reads a CSV record, and returns io.EOF on end of reading.
// May return nil for values to skip a write.
func (this *csvdecoder) Read() ([]string, []interface{}, error) {
	// Read a row
	row, err := this.r.Read()
	if err != nil {
		return nil, nil, err
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
			return nil, nil, nil
		}
	}

	// Add new column headings as necessary, populate values
	for len(row) > len(this.cols) {
		this.cols = append(this.cols, this.makeCol(len(this.cols)))
	}
	if len(this.values) != len(row) {
		this.values = make([]interface{}, len(this.cols))
	}
	for i, v := range row {
		this.values[i] = v
	}

	// Return
	return this.cols, this.values, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Return a column heading for the given index
func (this *csvdecoder) makeCol(i int) string {
	return fmt.Sprintf("col_%02d", i)
}
