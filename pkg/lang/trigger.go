package lang

import (
	"strings"

	// Import namespaces
	. "github.com/mutablelogic/go-sqlite"
	"github.com/mutablelogic/go-sqlite/pkg/quote"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type trigger struct {
	source
	temporary   bool
	ifnotexists bool
	table       string
	when        string
	action      string
	statements  []SQStatement
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Insert values into a table with a name and defined column names
func (this *source) CreateTrigger(table string, st ...SQStatement) SQTrigger {
	if len(st) == 0 {
		return nil
	} else {
		return &trigger{source{this.name, this.schema, "", false}, false, false, table, "AFTER", "INSERT", st}
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *trigger) String() string {
	return this.Query()
}

func (this *trigger) Query() string {
	tokens := []string{"CREATE"}

	// Add keywords into the query
	if this.temporary {
		tokens = append(tokens, "TEMPORARY")
	}
	if this.ifnotexists {
		tokens = append(tokens, "TRIGGER IF NOT EXISTS")
	} else {
		tokens = append(tokens, "TRIGGER")
	}

	// Add source and action
	tokens = append(tokens, this.source.Query(), this.when, this.action, "ON", quote.QuoteIdentifier(this.table))

	// Add Begin and End
	tokens = append(tokens, "BEGIN")
	for _, st := range this.statements {
		tokens = append(tokens, st.Query()+";")
	}
	tokens = append(tokens, "END")

	// Return the query
	return strings.Join(tokens, " ")
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *trigger) IfNotExists() SQTrigger {
	copy := *this
	copy.ifnotexists = true
	return &copy
}

func (this *trigger) WithTemporary() SQTrigger {
	copy := *this
	copy.temporary = true
	return &copy
}

func (this *trigger) Before() SQTrigger {
	copy := *this
	copy.when = "BEFORE"
	return &copy
}

func (this *trigger) After() SQTrigger {
	copy := *this
	copy.when = "AFTER"
	return &copy
}

func (this *trigger) InsteadOf() SQTrigger {
	copy := *this
	copy.when = "INSTEAD OF"
	return &copy
}

func (this *trigger) Delete() SQTrigger {
	copy := *this
	copy.action = "DELETE"
	return &copy
}

func (this *trigger) Insert() SQTrigger {
	copy := *this
	copy.action = "INSERT"
	return &copy
}

func (this *trigger) Update(col ...string) SQTrigger {
	copy := *this
	if len(col) == 0 {
		copy.action = "UPDATE"
	} else {
		copy.action = "UPDATE OF (" + quote.QuoteIdentifiers(col...) + ")"
	}
	return &copy
}
