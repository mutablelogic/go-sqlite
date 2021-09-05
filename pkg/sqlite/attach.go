package sqlite

import (
	// Import namespaces
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Attach database as schema
func (this *connection) Attach(schema, dsn string) error {
	if dsn == "" {
		return this.Attach(schema, sqLiteMemory)
	}
	query := Q("ATTACH DATABASE ", DoubleQuote(dsn), " AS ", QuoteIdentifier(schema))
	if _, err := this.Exec(query); err != nil {
		return err
	}
	// Success
	return nil
}

// Detach database as schema
func (this *connection) Detach(schema string) error {
	query := Q("DETACH DATABASE ", QuoteIdentifier(schema))
	if _, err := this.Exec(query); err != nil {
		return err
	}
	// Success
	return nil
}
