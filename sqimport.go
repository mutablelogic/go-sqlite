package sqlite

import "net/url"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type SQImportConfig struct {
	// Schema defines the table schema to import into. Optional.
	Schema string

	// Name defines the table name to import into, if empty will be inferred
	// from the import source URL
	Name string

	// Ext defines the extension to infer the mimetype from. Optional.
	Ext string

	// Header when true indicates the first line of a CSV file is a header
	Header bool

	// TrimSpace when true indicates the CSV file should be trimmed of whitespace
	// for each field
	TrimSpace bool

	// Comment defines the character which indicates a line is a comment. Optional.
	Comment rune

	// Delimiter defines the character which indicates a field delimiter. Optional.
	Delimiter rune

	// LazyQuotes when true indicates the CSV file should allow non-standard quotes.
	LazyQuotes bool

	// Overwrite existing table (will append data otherwise)
	Overwrite bool
}

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

type SQImporter interface {
	// Read from the source. Returns io.EOF when no more data is available.
	Read() error

	// Return the URL of the source
	URL() *url.URL

	// Return a decoder for a mimetype
	Decoder(string) (SQImportDecoder, error)
}

type SQWriter interface {
	// Write is called to add a row to the table with the named columns
	Write(name, schema string, cols []string, row []interface{}) error

	// Close completes the writing, flushing any records
	Close() error
}

type SQImportDecoder interface {
	// Read from the source. Returns io.EOF when no more data is available.
	Read() error
}
