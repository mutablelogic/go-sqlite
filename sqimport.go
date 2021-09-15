package sqlite

import (
	"io"
	"net/url"
)

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
	// ReadWrite will read from the source, and write to destination. This function
	// should be called multiple times until io.EOF is returned, indicating that
	// no more data is available.
	ReadWrite(SQImportDecoder, SQImportWriterFunc) error

	// Return the URL of the source
	URL() *url.URL

	// Return the Table name for the destination
	Name() string

	// Return a decoder for a reader, mimetype or file extension (when starts with a .)
	// Will return nil if no decoder is available. The mimetype can include
	// the character set (e.g. text/csv; charset=utf-8)
	Decoder(io.Reader, string) (SQImportDecoder, error)
}

// SQWriterFunc callback invoked for each row
type SQImportWriterFunc func(map[string]interface{}) error

// SQImportWriter is an interface for writing decoded rows to a destination
type SQImportWriter interface {
	// Begin the writer process for a destination and return a writer callback
	Begin(name, schema string, cols []string) (SQImportWriterFunc, error)

	// End the transaction with success (true) or failure (false). On failure, rollback
	End(bool) error
}

type SQImportDecoder interface {
	io.Closer

	// Read from the source, return column names and values. May
	// return nil to skip a write. Returns io.EOF when no more data is available.
	Read() (map[string]interface{}, error)
}
