package lang

import (
	"strings"

	sqlite "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type insert struct {
	source
	class         string
	defaultvalues bool
	columns       []string
	conflicts     []conflict
}

type conflict struct {
	action string
	target []string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Insert values into a table with a name and defined column names
func (this *source) Insert(columns ...string) sqlite.SQInsert {
	return &insert{source{this.name, this.schema, "", false}, "INSERT", false, columns, nil}
}

// Replace values into a table with a name and defined column names
func (this *source) Replace(columns ...string) sqlite.SQInsert {
	return &insert{source{this.name, this.schema, "", false}, "REPLACE", false, columns, nil}
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *insert) DefaultValues() sqlite.SQInsert {
	return &insert{this.source, this.class, true, this.columns, nil}
}

// WithConflictUpdate sets the conflict resolution to do nothing (that is,
// silently fail)
func (this *insert) WithConflictDoNothing(target ...string) sqlite.SQInsert {
	return &insert{this.source, this.class, this.defaultvalues, this.columns, append(this.conflicts, conflict{"NOTHING", target})}
}

// WithConflictUpdate sets the conflict resolution to update the row only
// when named columns are changed
func (this *insert) WithConflictUpdate(target ...string) sqlite.SQInsert {
	return &insert{this.source, this.class, this.defaultvalues, this.columns, append(this.conflicts, conflict{"UPDATE SET", target})}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *insert) String() string {
	return this.Query()
}

func (this *insert) Query() string {
	tokens := []string{this.class, "INTO"}

	// Add table name
	tokens = append(tokens, this.source.String())

	// Add column names
	if len(this.columns) > 0 {
		tokens = append(tokens, "("+sqlite.QuoteIdentifiers(this.columns...)+")")
	}

	// If default values
	if this.defaultvalues || (len(this.columns) == 0) {
		tokens = append(tokens, "DEFAULT VALUES")
	} else if len(this.columns) > 0 {
		tokens = append(tokens, "VALUES", this.argsN(len(this.columns)))
	} else {
		// No columns, return empty query
		return ""
	}

	// If this is an upsert statement add on conflict resolution
	if len(this.columns) > 0 && len(this.conflicts) > 0 {
		for _, conflict := range this.conflicts {
			tokens = append(tokens, conflict.Query(this.columns))
		}
	}

	// Return the query
	return strings.Join(tokens, " ")
}

func (c conflict) Query(columns []string) string {
	tokens := []string{"ON CONFLICT"}
	if len(c.target) > 0 {
		tokens = append(tokens, "("+sqlite.QuoteIdentifiers(c.target...)+")")
	}
	tokens = append(tokens, "DO", c.action)
	if c.action != "NOTHING" {
		set, where := make([]string, 0, len(columns)), make([]string, 0, len(columns))
		for _, column := range columns {
			set = append(set, sqlite.QuoteIdentifier(column)+"=excluded."+sqlite.QuoteIdentifier(column))
			where = append(where, sqlite.QuoteIdentifier(column)+"<>excluded."+sqlite.QuoteIdentifier(column))
		}
		tokens = append(tokens, strings.Join(set, ","), "WHERE", strings.Join(where, " OR "))
	}
	// When update and number of columns
	// SET a=? WHERE a != excluded.a
	return strings.Join(tokens, " ")
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *insert) argsN(n int) string {
	if n < 1 {
		return ""
	} else {
		return "(" + strings.Repeat("?,", n-1) + "?)"
	}
}
