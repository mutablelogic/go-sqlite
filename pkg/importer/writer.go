package importer

///////////////////////////////////////////////////////////////////////////////
// TYPES

// SQWriterFunc callback invoked for each row
type SQImportWriterFunc func([]interface{}) error

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// SQImportWriter is an interface for writing decoded rows to a destination
type SQImportWriter interface {
	// Begin the writer process for a destination and return a writer callback
	Begin(name, schema string, cols []string) (SQImportWriterFunc, error)

	// End the transaction with success (true) or failure (false). On failure, rollback
	End(bool) error
}
