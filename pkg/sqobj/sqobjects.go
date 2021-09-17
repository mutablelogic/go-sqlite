package sqobj

import (

	// Import Namespaces
	"context"
	"fmt"
	"reflect"

	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite"

	// Packages
	"github.com/djthorpe/go-sqlite/pkg/sqlite3"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Objects struct {
	*sqlite3.Conn

	schema string
	m      map[reflect.Type]*Class
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// With creates an SQObjects with an existing database connection and named schema
func With(conn *sqlite3.Conn, schema string, classes ...SQClass) (*Objects, error) {
	objects := new(Objects)

	// Check parameters
	if conn == nil || len(classes) == 0 {
		return nil, ErrBadParameter.With("With")
	}
	if schema == "" {
		schema = sqlite3.DefaultSchema
	}

	// Set connection, classes
	objects.Conn = conn
	objects.m = make(map[reflect.Type]*Class, len(classes))
	objects.schema = schema

	// Check schema
	if !hasElement(conn.Schemas(), schema) {
		return nil, ErrNotFound.Withf("schema %q", schema)
	}

	// Register classes
	for _, class := range classes {
		if class, ok := class.(*Class); !ok {
			return nil, ErrBadParameter.With(class.Name())
		} else {
			objects.m[class.t] = class
		}
	}

	// Set foreign keys on
	if err := conn.SetForeignKeyConstraints(true); err != nil {
		return nil, err
	}

	// Create schema - tables, indexes
	if err := conn.Do(context.Background(), SQLITE_TXN_NO_FOREIGNKEY_CONSTRAINTS, func(txn SQTransaction) error {
		// Create all classes
		for _, class := range classes {
			if err := class.Create(txn, schema); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Return success
	return objects, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (obj *Objects) String() string {
	str := "<sqobjects"
	str += fmt.Sprintf(" schema=%q", obj.schema)
	for _, c := range obj.m {
		str += " " + c.String()
	}
	str += " " + obj.Conn.String()
	return str + ">"
}
