package sqimport

import (
	"fmt"
	"mime"

	// Namespace Imports
	. "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return a new decoder for the given mimetype
func (this *Importer) Decoder(mimetype string) (SQImportDecoder, error) {
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
		return this.NewCSVDecoder(r, ',')
	case mediatype == "text/tsv":
		return this.NewCSVDecoder(r, '\t')
	case mediatype == "text/plain" && this.c.Ext == ".csv":
		return this.NewCSVDecoder(r, ',')
	case mediatype == "text/plain" && this.c.Ext == ".tsv":
		return this.NewCSVDecoder(r, '\t')
	default:
		return nil, fmt.Errorf("unsupported media type: %q (file extension %q)", mediatype, this.c.Ext)
	}
}
