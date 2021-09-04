package sqobj

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	// Modules
	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"

	// Import namespaces
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type sqobj struct {
	sync.Mutex
	SQConnection
	schema string
	class  map[reflect.Type]*sqclass
	names  []string
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultSchema = "main"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func Open(path string, location *time.Location) (SQObjects, error) {
	if conn, err := sqlite.Open(path, location); err != nil {
		return nil, err
	} else {
		return With(conn, defaultSchema)
	}
}

func New(location ...*time.Location) (SQObjects, error) {
	if conn, err := sqlite.New(location...); err != nil {
		return nil, err
	} else {
		return With(conn, defaultSchema)
	}
}

func With(conn SQConnection, schema string) (SQObjects, error) {
	this := new(sqobj)
	this.class = make(map[reflect.Type]*sqclass)

	// Set connection
	if conn == nil {
		return nil, ErrBadParameter.With("SQConnection")
	} else {
		this.SQConnection = conn
	}

	// Set schema
	if schema == "" {
		schema = defaultSchema
	}
	if containsString(this.Schemas(), schema) == false {
		return nil, ErrNotFound.With("schema: ", strconv.Quote(schema))
	} else {
		this.schema = schema
	}

	// Set foreign key support
	if err := this.SetForeignKeyConstraints(true); err != nil {
		return nil, err
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *sqobj) String() string {
	str := "<sqobj"
	str += fmt.Sprintf(" schema=%q", this.schema)
	str += fmt.Sprint(" ", this.SQConnection)
	for _, class := range this.class {
		str += fmt.Sprint(" ", class)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *sqobj) Register(name string, proto interface{}) (SQClass, error) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Create the class, check constraints
	class := NewClass(name, this.schema, proto)
	if class == nil {
		return nil, ErrBadParameter.With("Invalid prototype")
	} else if containsString(this.names, class.Name()) {
		return nil, ErrDuplicateEntry.With(class.Name())
	} else if _, exists := this.class[class.Type()]; exists {
		return nil, ErrDuplicateEntry.With(class.Name())
	} else {
		this.names = append(this.names, class.Name())
		this.class[class.Type()] = class
	}

	// Return success
	return class, nil
}

func (this *sqobj) Create(class SQClass, flags SQFlag) error {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Suspend foreign key constraints whilst dropping and creating
	if err := this.SetForeignKeyConstraints(false); err != nil {
		return err
	}
	defer this.SetForeignKeyConstraints(true)

	// Drop and create within a transaction
	return this.Do(func(txn SQTransaction) error {
		// Check for existence of table
		if containsString(this.TablesEx(this.schema, false), class.Name()) && flags&SQLITE_FLAG_DELETEIFEXISTS != 0 {
			// Drop indexes
			for _, index := range this.IndexesEx(class.Name(), this.schema) {
				if !index.Auto() {
					if _, err := txn.Exec(index.DropIndex()); err != nil {
						return err
					}
				}
			}
			// Drop table
			if _, err := txn.Exec(class.DropTable()); err != nil {
				return err
			}
		}

		// Create table and indexes
		for _, create := range class.Get(SQKeyCreate) {
			if _, err := txn.Exec(create); err != nil {
				return err
			}
		}

		// Return success
		return nil
	})
}

// Write objects to database in a single transaction and return the results from writing
func (this *sqobj) Write(v ...interface{}) ([]SQResult, error) {
	return this.WriteWithHook(nil, v...)
}

// WriteHook inserts or updates objects to database in a single transaction and
// return the results from writing. On error rollback occurs. Call hook after each
// write with next object, or nil if this was the last write
func (this *sqobj) WriteWithHook(fn SQWriteHook, v ...interface{}) ([]SQResult, error) {
	result := make([]SQResult, 0, len(v))
	if err := this.Do(func(txn SQTransaction) error {
		for i, v_ := range v {
			class, err := this.classFor(v_)
			if err != nil {
				return err
			}
			params, err := class.params(v_)
			if err != nil {
				return err
			}

			// Insert or update, return number of affected rows
			r, err := txn.Exec(class.Get(SQKeyWrite)[0], params...)
			if err != nil {
				return err
			}

			// Blank out the rowid if no update was made
			if r.RowsAffected == 0 {
				r.LastInsertId = 0
			}

			// Re-fetch the row to obtain the rowid if primary key is used
			if len(class.PrimaryColumnNames()) > 0 {
				params, err := class.primaryvalues(v_)
				if err != nil {
					return err
				}
				rs, err := txn.Query(class.Get(SQKeyGetRowId)[0], params...)
				if err != nil {
					return err
				}
				defer rs.Close()
				if rowid := rs.NextArray(); rowid != nil {
					r.LastInsertId = rowid[0].(int64)
				}
			}

			// Call hook on the next object
			if fn != nil {
				if i+1 < len(v) {
					if err := fn(r, v[i+1]); err != nil {
						return err
					}
				} else {
					if err := fn(r, nil); err != nil {
						return err
					}
				}
			}

			// Append r
			result = append(result, r)
		}
		// Return success
		return nil
	}); err != nil {
		return nil, err
	}
	// Return success
	return result, nil
}

// Delete objects from database in a single transaction and return the results from deletion
// on error rollback occurs
func (this *sqobj) Delete(v ...interface{}) ([]SQResult, error) {
	result := make([]SQResult, 0, len(v))
	if err := this.Do(func(txn SQTransaction) error {
		for _, v := range v {
			class, err := this.classFor(v)
			if err != nil {
				return err
			}
			params, err := class.primaryvalues(v)
			if err != nil {
				return err
			}
			if len(params) == 0 {
				return ErrBadParameter.Withf("No primary key for class %q", class.Name())
			}

			// Fetch the rowid for this row
			rs, err := txn.Query(class.Get(SQKeyGetRowId)[0], params...)
			if err != nil {
				return err
			}
			defer rs.Close()
			var r SQResult
			if row := rs.NextArray(); row != nil {
				r.LastInsertId = row[0].(int64)
			}
			// If we have a rowid then delete the row
			if r.LastInsertId > 0 {
				r_, err := txn.Exec(class.Get(SQKeyDelete)[0], params...)
				if err != nil {
					return err
				} else {
					r.RowsAffected = r_.RowsAffected
				}
			}
			// Append r
			result = append(result, r)
		}
		// Return success
		return nil
	}); err != nil {
		return nil, err
	}

	// Return success
	return result, nil
}

func (this *sqobj) Read(class SQClass) (SQIterator, error) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	if class, ok := class.(*sqclass); !ok {
		return nil, ErrBadParameter.Withf("Invalid class %q", class.Name())
	} else if rs, err := this.Query(class.Get(SQKeyRead)[0]); err != nil {
		return nil, err
	} else {
		return class.NewIterator(rs), nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *sqobj) classFor(v interface{}) (*sqclass, error) {
	if v == nil {
		return nil, ErrBadParameter.With("unexpected nil value")
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if class, exists := this.class[rv.Type()]; !exists {
		return nil, ErrNotFound.With("class: ", reflect.TypeOf(v))
	} else {
		return class, nil
	}
}

func containsString(arr []string, v string) bool {
	for _, elem := range arr {
		if elem == v {
			return true
		}
	}
	return false
}
