package importer

import (
	"fmt"
	"mime"

	// Namespace Imports
	. "github.com/djthorpe/go-sqlite"
	"github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return a new decoder for the given mimetype, or guess the mimetype when
// opening the file
func (this *Importer) Decoder(mimetype string) (SQImportDecoder, error) {
	// Open the data source
	r, guessedmimetype, err := open(this.url)
	if err != nil {
		return nil, err
	} else if mimetype != "" {
		guessedmimetype = mimetype
	}

	// Parse mediatype
	mediatype, params, err := mime.ParseMediaType(guessedmimetype)
	if err != nil {
		if err_ := r.Close(); err != nil {
			err = multierror.Append(err, err_)
		}
		return nil, err
	}

	// Set charset
	cr, err := charsetReader(r, params["charset"])
	if err != nil {
		if err_ := r.Close(); err != nil {
			err = multierror.Append(err, err_)
		}
		return nil, err
	}

	// Set decoder based on mediatype and other possible
	// parameters
	switch {
	case mediatype == "application/vnd.ms-excel":
		return this.NewXLSDecoder(r)
	case mediatype == "application/excel":
		return this.NewXLSDecoder(r)
	case mediatype == "application/x-excel":
		return this.NewXLSDecoder(r)
	case mediatype == "application/x-msexcel":
		return this.NewXLSDecoder(r)
	case mediatype == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return this.NewXLSDecoder(r)
	case mediatype == "text/csv":
		return this.NewCSVDecoder(r, cr, ',')
	case mediatype == "text/tsv":
		return this.NewCSVDecoder(r, cr, '\t')
	case mediatype == "text/plain" && this.c.Ext == ".csv":
		return this.NewCSVDecoder(r, cr, ',')
	case mediatype == "text/plain" && this.c.Ext == ".tsv":
		return this.NewCSVDecoder(r, cr, '\t')
	default:
		return nil, fmt.Errorf("unsupported media type: %q (file extension %q)", mediatype, this.c.Ext)
	}
}
