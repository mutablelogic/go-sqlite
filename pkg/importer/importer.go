package sqimport

import (
	"net/url"
	"path/filepath"
	"strings"

	// Namespace Imports
	. "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Importer struct {
	c        SQImportConfig
	url      *url.URL
	mimetype string
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
func DefaultImporter(url string) (*Importer, error) {
	return NewImporter(DefaultConfig, url)
}

// Create a new importer with a database writer
func NewImporter(c SQImportConfig, u string) (*Importer, error) {
	this := &Importer{
		c: c,
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

func (this *Importer) URL() *url.URL {
	return this.url
}

func (this *Importer) Name() string {
	return this.c.Name
}

// Read a row from the source data and potentially insert into the table. On end
// of data, returns io.EOF.
func (this *Importer) ReadWrite(dec SQImportDecoder, fn SQImportWriterFunc) error {
	row, err := dec.Read()
	if err != nil {
		return err
	} else if row != nil {
		return fn(row)
	} else {
		return nil
	}
}
