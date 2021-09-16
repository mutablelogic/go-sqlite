package importer

import (
	"fmt"
	"io"

	// Namespace Imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite"

	// Package imports
	excelize "github.com/xuri/excelize/v2"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type xlsdecoder struct {
	f      *excelize.File
	r      *excelize.Rows
	sheet  string
	header bool
	cols   []string
	values []interface{}
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewXLSDecoder returns a XLS decoder setting options
func (this *Importer) NewXLSDecoder(r io.Reader) (SQImportDecoder, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, err
	}

	// Make decoder, set sheet to import
	decoder := &xlsdecoder{f, nil, "", this.c.Header, nil, nil}
	if sheet := f.GetActiveSheetIndex(); sheet > 0 {
		decoder.sheet = f.GetSheetName(sheet)
	} else if f.SheetCount >= 1 {
		decoder.sheet = f.GetSheetName(sheet)
	}
	if decoder.sheet == "" {
		return nil, ErrBadParameter.With("No active sheet")
	}

	// Make iterator
	if rows, err := f.Rows(decoder.sheet); err != nil {
		return nil, err
	} else {
		decoder.r = rows
	}

	// Return success
	return decoder, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (dec *xlsdecoder) String() string {
	return fmt.Sprintf("<application/vnd.ms-excel>")
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

// Read reads a CSV record, and returns io.EOF on end of reading.
// May return nil for values to skip a write.
func (this *xlsdecoder) Read() ([]string, []interface{}, error) {
	// Read a row
	noteof := this.r.Next()
	if !noteof {
		return nil, nil, io.EOF
	}
	row, err := this.r.Columns()
	if err != nil {
		return nil, nil, err
	} else if len(row) == 0 {
		return nil, nil, nil
	}

	// Initialize the reader
	if this.cols == nil {
		this.cols = make([]string, len(row))
		for i, col := range row {
			if this.header && col != "" {
				this.cols[i] = col
			} else {
				this.cols[i] = this.makeCol(i)
			}
		}
		if this.header {
			return nil, nil, nil
		}
	}

	// Add new column headings as necessary
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
func (this *xlsdecoder) makeCol(i int) string {
	return fmt.Sprintf("col_%02d", i)
}
