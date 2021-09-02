package sqobj

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	// Modules
	. "github.com/djthorpe/go-sqlite"
	//. "github.com/djthorpe/go-sqlite/pkg/lang"
	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"
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
	if err := this.Do(func(txn SQTransaction) error {

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
		for _, create := range CreateTableAndIndexes(class, true, class.Proto()) {
			if _, err := txn.Exec(create); err != nil {
				return err
			}
		}

		// Return success
		return nil
	}); err != nil {
		return err
	}

	// Prepare SQL statements outside of transaction
	if err := class.(*sqclass).prepare(this); err != nil {
		return err
	}

	// Return success
	return nil
}

// Write objects to database in a single transaction and return the results from writing
// on error rollback occurs
func (this *sqobj) Write(v ...interface{}) ([]SQResult, error) {
	result := make([]SQResult, 0, len(v))
	if err := this.Do(func(txn SQTransaction) error {
		for _, v := range v {
			class, err := this.classFor(v)
			if err != nil {
				return err
			}
			if params, err := class.params(v); err != nil {
				return err
			} else if r, err := txn.Exec(class.statement(classKeyWrite), params...); err != nil {
				return err
			} else {
				result = append(result, r)
			}
		}
		// Return success
		return nil
	}); err != nil {
		return nil, err
	}
	// Return success
	return result, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *sqobj) classFor(v interface{}) (*sqclass, error) {
	if v == nil {
		return nil, ErrBadParameter.With("unexpected nil value")
	}
	if class, exists := this.class[reflect.TypeOf(v)]; !exists {
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
