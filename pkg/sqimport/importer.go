package sqimport

import (
	"fmt"
	"io"
	"mime"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	// Modules
	sqlite "github.com/djthorpe/go-sqlite"
	"github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type importer struct {
	sync.Mutex
	c        sqlite.SQImportConfig
	r        io.ReadCloser
	w        sqlite.SQWriter
	dec      sqlite.SQImportDecoder
	url      *url.URL
	mimetype string
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	DefaultConfig = sqlite.SQImportConfig{
		Header:     true,
		TrimSpace:  true,
		LazyQuotes: true,
	}
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create an importer with default configuation
func DefaultImporter(url string, writer sqlite.SQWriter) (sqlite.SQImporter, error) {
	return NewImporter(DefaultConfig, url, writer)
}

// Create a new importer with a database writer
func NewImporter(c sqlite.SQImportConfig, u string, writer sqlite.SQWriter) (sqlite.SQImporter, error) {
	this := &importer{
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
		return nil, sqlite.ErrBadParameter.With("Writer")
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

func (this *importer) URL() *url.URL {
	return this.url
}

// Read a row from the source data and potentially insert into the table. On end
// of data, returns io.EOF.
func (this *importer) Read() error {
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
	if err := this.dec.Read(); err != nil {
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

// Return a new decoder for the given mimetype
func (this *importer) Decoder(mimetype string) (sqlite.SQImportDecoder, error) {
	// Parse mediatype
	mediatype, params, err := mime.ParseMediaType(mimetype)
	if err != nil {
		return nil, err
	}

	// Set charset
	r, err := charsetReader(this.r, params["charset"])
	if err != nil {
		return nil, err
	}

	// Set decoder based on mediatype and other possible
	// parameters
	switch {
	case mediatype == "text/csv":
		return this.NewCSVDecoder(r, this.w, ',')
	case mediatype == "text/tsv":
		return this.NewCSVDecoder(r, this.w, '\t')
	case mediatype == "text/plain" && this.c.Ext == ".csv":
		return this.NewCSVDecoder(r, this.w, ',')
	case mediatype == "text/plain" && this.c.Ext == ".tsv":
		return this.NewCSVDecoder(r, this.w, '\t')
	default:
		return nil, fmt.Errorf("unsupported media type: %q (file extension %q)", mediatype, this.c.Ext)
	}
}
