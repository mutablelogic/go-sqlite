package lang

import (
	"strings"

	sqlite "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type drop struct {
	source
	class    string
	ifexists bool
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Drop a table
func (this *source) DropTable() sqlite.SQDrop {
	return &drop{source{this.name, this.schema, ""}, "TABLE", false}
}

// Drop a index
func (this *source) DropIndex() sqlite.SQDrop {
	return &drop{source{this.name, this.schema, ""}, "INDEX", false}
}

// Drop a trigger
func (this *source) DropTrigger() sqlite.SQDrop {
	return &drop{source{this.name, this.schema, ""}, "TRIGGER", false}
}

// Drop a view
func (this *source) DropView() sqlite.SQDrop {
	return &drop{source{this.name, this.schema, ""}, "VIEW", false}
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *drop) IfExists() sqlite.SQDrop {
	return &drop{this.source, this.class, true}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *drop) Query() string {
	tokens := []string{"DROP", this.class}
	if this.ifexists {
		tokens = append(tokens, "IF EXISTS")
	}
	tokens = append(tokens, this.source.String())

	// Return the query
	return strings.Join(tokens, " ")
}

func (this *drop) String() string {
	return this.Query()
}
