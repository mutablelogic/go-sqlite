package sqobj

import (
	"context"
	"fmt"
	"reflect"

	// Import Namespaces
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-sqlite"

	// Packages
	sqlite3 "github.com/mutablelogic/go-sqlite/pkg/sqlite3"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Objects struct {
	schema string
	m      map[reflect.Type]*Class
	p      SQPool
	c      SQConnection
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// WithPool creates an SQObjects with a connection pool and named schema
func WithPool(pool SQPool, schema string, classes ...SQClass) (*Objects, error) {
	objects := new(Objects)

	// Check parameters
	if pool == nil || len(classes) == 0 {
		return nil, ErrBadParameter.With("WithPool")
	} else {
		objects.p = pool
	}

	return objects.with(schema, classes...)
}

// With creates an SQObjects with a database connection and named schema
func With(conn SQConnection, schema string, classes ...SQClass) (*Objects, error) {
	objects := new(Objects)

	// Check parameters
	if conn == nil || len(classes) == 0 {
		return nil, ErrBadParameter.With("With")
	} else {
		objects.c = conn
	}

	return objects.with(schema, classes...)
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (obj *Objects) String() string {
	str := "<sqobjects"
	str += fmt.Sprintf(" schema=%q", obj.schema)
	for _, c := range obj.m {
		str += " " + c.String()
	}
	str += fmt.Sprint(" ", obj.SQConnection)
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Write objects (insert or update) to the database
func (obj *Objects) Write(ctx context.Context, v ...interface{}) error {
	return obj.c.Do(ctx, SQLITE_NONE, func(txn SQTransaction) error {
		for _, v := range v {
			rv := ValueOf(v)
			class, exists := obj.m[rv.Type()]
			if !exists {
				return ErrBadParameter.Withf("Write: %v", v)
			}
			if r, err := class.UpsertKeys(txn, v); err != nil {
				return err
			} else {
				// TODO: Pass rowid and primary keys to next object
				fmt.Println(r[0])
			}
		}
		return nil
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (objects *Objects) conn(ctx context.Context) SQConnection {

}

func (objects *Objects) with(schema string, classes ...SQClass) (*Objects, error) {
	if schema == "" {
		schema = sqlite3.DefaultSchema
	}

	// Set connection, classes
	objects.m = make(map[reflect.Type]*Class, len(classes))
	objects.schema = schema

	if schema == "" {
		schema = sqlite3.DefaultSchema
	}

	// Set connection, classes
	objects.c = conn
	objects.m = make(map[reflect.Type]*Class, len(classes))
	objects.schema = schema

	// Check schema
	if !hasElement(conn.Schemas(), schema) {
		return nil, ErrNotFound.Withf("schema %q", schema)
	}

	// Error if foreign keys not supported
	if !conn.Flags().Is(SQLITE_OPEN_FOREIGNKEYS) {
		return nil, ErrBadParameter.With("SQLITE_OPEN_FOREIGNKEYS")
	}

	// Register classes
	for _, class := range classes {
		if class, ok := class.(*Class); !ok {
			return nil, ErrBadParameter.With(class.Name())
		} else {
			objects.m[class.t] = class
		}
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
