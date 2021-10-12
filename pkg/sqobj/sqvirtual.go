package sqobj

import (
	"fmt"
	"reflect"

	// Import Namespaces
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Virtual struct {
	*SQReflect
	SQSource

	// Prepared statements and in-place parameters
	module string
	opts   []string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// MustRegisterVirtual registers a SQObject virtual table class, panics if an error
// occurs.
func MustRegisterVirtual(source SQSource, module string, proto interface{}, options ...string) *Virtual {
	if cls, err := RegisterVirtual(source, module, proto, options...); err != nil {
		panic(err)
	} else {
		return cls
	}
}

// RegisterVirtual registers a SQObject virtual table class, returns the class and
// any errors
func RegisterVirtual(source SQSource, module string, proto interface{}, options ...string) (*Virtual, error) {
	this := new(Virtual)

	// Check name
	if source.Name() == "" {
		return nil, ErrBadParameter.With("source")
	} else {
		this.SQSource = source
	}
	// Check module
	if module == "" {
		return nil, ErrBadParameter.With("module")
	} else {
		this.module = module
		this.opts = options
	}

	// Do reflection
	if r, err := NewReflect(proto); err != nil {
		return nil, err
	} else {
		this.SQReflect = r
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Virtual) String() string {
	str := "<sqvirtual"
	str += fmt.Sprintf(" name=%q", this.Name())
	str += fmt.Sprintf(" module=%q", this.module)
	if len(this.opts) > 0 {
		str += fmt.Sprintf(" options=%q", this.opts)
	}
	if schema := this.Schema(); schema != "" {
		str += fmt.Sprintf(" schema=%q", this.Schema())
	}
	str += " " + fmt.Sprint(this.SQReflect)
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

// Proto returns a prototype of the class
func (this *Virtual) Proto() reflect.Value {
	return reflect.New(this.t)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Create creates a virtual table. If
// the flag SQLITE_OPEN_OVERWRITE is set when creating the connection, then tables
// are dropped and then re-created.
func (this *Virtual) Create(txn SQTransaction, schema string, options ...string) error {
	// If schema then set it
	if schema != "" {
		this.SQSource = this.SQSource.WithSchema(schema)
	}

	if txn.Flags().Is(SQLITE_OPEN_OVERWRITE) && hasElement(txn.Tables(this.Schema()), this.Name()) {
		// Drop table
		if _, err := txn.Query(this.DropTable()); err != nil {
			return err
		}
	}

	// Create tables if they don't exist
	for _, st := range this.Virtual(this.SQSource, this.module, true, append(options, this.opts...)...) {
		if _, err := txn.Query(st); err != nil {
			return err
		}
	}

	// Return success
	return nil
}
