package sqlite

import (
	sqlite "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *connection) Attach(schema, dsn string) error {
	query := Q("ATTACH DATABASE " + sqlite.DoubleQuote(dsn) + " AS " + sqlite.QuoteIdentifier(schema))
	if _, err := this.Exec(query); err != nil {
		return err
	}
	// Success
	return nil
}

func (this *connection) Detach(schema string) error {
	query := Q("DETACH DATABASE " + sqlite.QuoteIdentifier(schema))
	if _, err := this.Exec(query); err != nil {
		return err
	}
	// Success
	return nil
}
