package sqobj

import (
	"fmt"
	"sync"

	// Modules

	// Import Namespaces
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type sqclass struct {
	sync.RWMutex
	*SQReflect
	SQSource

	// Prepared statements
	s map[SQKey]SQStatement
}

type sqpreparefunc func(*sqclass) SQStatement

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	statements = map[SQKey]sqpreparefunc{
		SQKeyRead: sqSelect,
	}
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewClass(name, schema string, proto interface{}) (*sqclass, error) {
	this := new(sqclass)
	this.s = make(map[SQKey]SQStatement)

	// Check name
	if name == "" {
		return nil, ErrBadParameter.Withf("name")
	}

	// Do reflection
	if r, err := NewReflect(proto); err != nil {
		return nil, err
	} else {
		this.SQReflect = r
	}

	// Set source
	this.SQSource = N(name).WithSchema(schema)

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *sqclass) String() string {
	str := "<sqclass"
	str += fmt.Sprintf(" name=%q", this.Name())
	if schema := this.Schema(); schema != "" {
		str += fmt.Sprintf(" schema=%q", this.Schema())
	}
	str += " " + fmt.Sprint(this.SQReflect)
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *sqclass) Create(conn SQConnection, flags SQFlag) error {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()

	// Suspend foreign key constraints whilst dropping and creating
	if err := conn.SetForeignKeyConstraints(false); err != nil {
		return err
	}
	defer conn.SetForeignKeyConstraints(true)

	return conn.Do(func(txn SQTransaction) error {
		// Drop table if it exists
		if hasElement(conn.TablesEx(this.Schema(), false), this.Name()) && flags&SQLITE_FLAG_DELETEIFEXISTS != 0 {
			// Drop indexes
			for _, index := range conn.IndexesEx(this.Name(), this.Schema()) {
				if !index.Auto() {
					if _, err := txn.Exec(index.DropIndex()); err != nil {
						return err
					}
				}
			}
			// Drop table
			if _, err := txn.Exec(this.DropTable()); err != nil {
				return err
			}
		}

		// Create tables if they don't exist
		for _, st := range this.Table(this.SQSource, true) {
			if _, err := txn.Exec(st); err != nil {
				return err
			}
		}

		// Prepare statements
		for key, fn := range statements {
			st := fn(this)
			if st == nil {
				return ErrBadParameter.Withf("Prepare: %q", key)
			}
			if st == nil {
				continue
			}
			if st, err := conn.Prepare(st); err != nil {
				return ErrUnexpectedResponse.Withf("Prepare: %v: %v", key, err)
			} else {
				this.s[key] = st
			}
			fmt.Println(st)
		}

		// Return success
		return nil
	})
}

// Read from table and return an iterator. It is expected that Read would
// accept a query, including: order, limit, offset, distinct and a
// list of expressions
func (this *sqclass) Read(conn SQConnection) (SQIterator, error) {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()

	// Read from database
	st, exists := this.s[SQKeyRead]
	if !exists {
		return nil, ErrOutOfOrder.Withf("Read: %q", this.Name())
	}
	rs, err := conn.Query(st)
	if err != nil {
		return nil, err
	} else {
		return NewIterator(this, rs), nil
	}
}
