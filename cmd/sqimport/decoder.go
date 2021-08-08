package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"mime"
	"strings"

	// Modules
	sqlite "github.com/djthorpe/go-sqlite"
	charmap "golang.org/x/text/encoding/charmap"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type decoder struct {
	cols   []string
	csvd   *csv.Reader
	reader func() ([]sqlite.SQStatement, error)
	writer *writer
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewDecoder(r io.Reader, w *writer, mimetype string) (*decoder, error) {
	this := new(decoder)

	// Parse mediatype
	mediatype, params, err := mime.ParseMediaType(mimetype)
	if err != nil {
		return nil, err
	}

	// Set charset
	r, err = charsetReader(r, params["charset"])
	if err != nil {
		return nil, err
	}

	// Set decoder
	switch mediatype {
	case "text/csv", "text/plain":
		this.csvd = csv.NewReader(r)
		this.reader = this.csv
	default:
		return nil, fmt.Errorf("unsupported media type: %q", mediatype)
	}

	// Set writer
	this.writer = w

	// Return success
	return this, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *decoder) Read() error {
	statements, err := this.reader()
	if err != nil {
		return err
	}
	return this.writer.Do(statements)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *decoder) csv() ([]sqlite.SQStatement, error) {
	if row, err := this.csvd.Read(); err != nil {
		return nil, err
	} else if this.cols == nil {
		this.cols = row
		return this.writer.CreateTable(row), nil
	} else {
		return this.writer.Insert(this.cols, row), nil
	}
}

func charsetReader(r io.Reader, charset string) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "utf8", "utf-8", "":
		// Default
		return r, nil
	case "windows-1252":
		return charmap.Windows1252.NewDecoder().Reader(r), nil
	case "iso-8859-1":
		return charmap.ISO8859_1.NewDecoder().Reader(r), nil
	default:
		return nil, fmt.Errorf("unsupported charset: %q", charset)
	}
}
