package sqlite

import "net/url"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type SQImportConfig struct {
	// Schema defines the table schema to import into. Optional.
	Schema string `sqlite:"schema"`

	// Name defines the table name to import into, if empty will be inferred
	// from the import source URL
	Name string `sqlite:"name"`

	// Ext defines the extension to infer the mimetype from. Optional.
	Ext string `sqlite:"ext"`

	// Header when true indicates the first line of a CSV file is a header
	Header bool `sqlite:"header"`

	// TrimSpace when true indicates the CSV file should be trimmed of whitespace
	// for each field
	TrimSpace bool `sqlite:"trimspace"`

	// Comment defines the character which indicates a line is a comment. Optional.
	Comment rune `sqlite:"comment"`

	// Delimiter defines the character which indicates a field delimiter. Optional.
	Delimiter rune `sqlite:"delimiter"`

	// LazyQuotes when true indicates the CSV file should allow non-standard quotes.
	LazyQuotes bool `sqlite:"lazyquotes"`

	// Overwrite existing table (will append data otherwise)
	Overwrite bool `sqlite:"overwrite"`
}

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

type SQImporter interface {
	// Read from the source. Returns io.EOF when no more data is available.
	Read() error

	// Return the URL of the source
	URL() *url.URL

	// Return the Table name for the destination
	Name() string

	// Return a decoder for a mimetype
	Decoder(string) (SQImportDecoder, error)
}

type SQWriter interface {
	// Write is called to add a row to the table with the named columns
	Write(name, schema string, cols []string, row []interface{}) error

	// Reset the counter
	Reset()

	// Count returns the number of rows written
	Count() int

	// Close completes the writing, flushing any records
	Close() error
}

type SQImportDecoder interface {
	// Read from the source, and write rows. Returns io.EOF when no more data is available.
	Read(SQWriter) error
}
