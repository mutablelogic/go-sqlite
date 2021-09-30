package importer

import (
	"errors"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	// Package imports
	multierror "github.com/hashicorp/go-multierror"

	// Namespace Imports
	. "github.com/mutablelogic/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Importer struct {
	c   SQImportConfig
	w   SQImportWriter
	fn  SQImportWriterFunc
	url *url.URL
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	DefaultConfig = SQImportConfig{
		Header:     true,
		TrimSpace:  true,
		LazyQuotes: true,
	}
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create an importer with default configuation
func DefaultImporter(url string, w SQImportWriter) (*Importer, error) {
	return NewImporter(DefaultConfig, url, w)
}

// Create a new importer with a database writer
func NewImporter(c SQImportConfig, u string, w SQImportWriter) (*Importer, error) {
	this := &Importer{
		c: c,
		w: w,
	}

	// Set the URL source
	if url, err := url.Parse(u); err != nil {
		return nil, err
	} else {
		this.url = url
	}

	// Set the table name and extension if not already set
	if this.c.Name == "" {
		this.c.Name = filepath.Base(this.url.String())
		if ext := filepath.Ext(this.c.Name); ext != "" {
			this.c.Name = strings.TrimSuffix(this.c.Name, ext)
			this.c.Ext = ext
		}
	} else {
		this.c.Name = c.Name
	}
	if this.c.Ext == "" {
		if ext := filepath.Ext(filepath.Base(this.url.String())); ext != "" {
			this.c.Ext = ext
		}
	}
	this.c.Ext = strings.ToLower(this.c.Ext)

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (i *Importer) URL() *url.URL {
	return i.url
}

func (i *Importer) Name() string {
	return i.c.Name
}

// Read a row from the source data and potentially insert into the table. On end
// of data, returns io.EOF.
func (i *Importer) ReadWrite(dec SQImportDecoder) error {
	var result error

	// Read next row, end transaction if at EOF or other error
	cols, values, err := dec.Read()
	if err != nil {
		result = multierror.Append(result, err)
		if err := i.w.End(errors.Is(err, io.EOF)); err != nil {
			result = multierror.Append(result, err)
		}
	} else if cols == nil || values == nil {
		return nil
	}

	// Begin transaction, get function
	if result == nil {
		if i.fn == nil {
			if fn, err := i.w.Begin(i.c.Name, i.c.Schema, cols); err != nil {
				result = multierror.Append(result, err)
			} else {
				i.fn = fn
			}
		}
	}

	// Write row
	if result == nil && i.fn != nil {
		if err := i.fn(values); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Return any errors
	return result
}
