package sqlite

func (this *connection) Attach(schema, dsn string) error {
	query := this.Q("ATTACH DATABASE " + DoubleQuote(dsn) + " AS " + QuoteIdentifier(schema))
	if _, err := this.Exec(query); err != nil {
		return err
	}
	// Success
	return nil
}

func (this *connection) Detach(schema string) error {
	query := this.Q("DETACH DATABASE " + QuoteIdentifier(schema))
	if _, err := this.Exec(query); err != nil {
		return err
	}
	// Success
	return nil
}
