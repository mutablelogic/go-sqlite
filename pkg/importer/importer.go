package sqimport

import (
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	// Namespace Imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite"

	// Modules
	multierror "github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Importer struct {
	sync.Mutex
	c        SQImportConfig
	r        io.ReadCloser
	w        SQWriter
	dec      SQImportDecoder
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
func DefaultImporter(url string, writer SQWriter) (*Importer, error) {
	return NewImporter(DefaultConfig, url, writer)
}

// Create a new importer with a database writer
func NewImporter(c SQImportConfig, u string, writer SQWriter) (*Importer, error) {
	this := &Importer{
		c: c,
	}

	// Set the URL source
	if url, err := url.Parse(u); err != nil {
		return nil, err
	} else {
		this.url = url
	}

	// Set the writer
	if writer == nil {
		return nil, ErrBadParameter.With("Writer")
	} else {
		this.w = writer
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
func (this *Importer) ReadWrite() error {
	var result error

	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Open the data source
	if this.r == nil {
		if r, mimetype, err := open(this.url); err != nil {
			result = err
			if err := this.w.Close(); err != nil {
				result = multierror.Append(result, err)
			}
			return result
		} else {
			this.r = r
			this.mimetype = mimetype
		}
		// Skip row
		return nil
	}

	// Set the decoder
	if this.dec == nil {
		if dec, err := this.Decoder(this.mimetype); err != nil {
			result = err
			if err := this.w.Close(); err != nil {
				result = multierror.Append(result, err)
			}
			return result
		} else {
			this.dec = dec
		}
		// Skip row
		return nil
	}

	// Read the row
	if err := this.dec.Read(this.w); err != nil {
		// Release resources
		result = err
		if err := this.r.Close(); err != nil {
			result = multierror.Append(result, err)
		}
		if err := this.w.Close(); err != nil {
			result = multierror.Append(result, err)
		}
		this.dec = nil
		this.r = nil
		this.w = nil
		return result
	}

	// Return sucess
	return nil
}
