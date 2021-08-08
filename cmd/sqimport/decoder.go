package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"mime"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type decoder struct {
	cols   []string
	csvd   *csv.Reader
	reader func() (map[string]interface{}, error)
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewDecoder(r io.Reader, mimetype string) (*decoder, error) {
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

	// Return success
	return this, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *decoder) Read() (map[string]interface{}, error) {
	return this.reader()
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *decoder) csv() (map[string]interface{}, error) {
	if row, err := this.csvd.Read(); err != nil {
		return nil, err
	} else if this.cols == nil {
		// TODO: Add columns to table
		this.cols = row
		return nil, nil
	} else {
		// TODO: Zip row and columns
		fmt.Println(row)
	}
	return nil, nil
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
