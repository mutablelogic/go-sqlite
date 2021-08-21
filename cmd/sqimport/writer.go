package main

import (
	"fmt"

	// Modules
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type writer struct {
	SQConnection

	// The writer's table name
	Name string

	// When Overwrite is true, drop table before creating it
	Overwrite bool
}

type row struct {
	SQStatement
	args []string
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewWriter(db SQConnection) *writer {
	return &writer{db, "", false}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *writer) String() string {
	str := "<writer"
	str += fmt.Sprint(" db=", this.SQConnection)
	if this.Name != "" {
		str += fmt.Sprintf(" name=%q", this.Name)
	}
	str += fmt.Sprint(" overwrite=", this.Overwrite)
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *writer) CreateTable(cols []string) []SQStatement {
	result := []SQStatement{}

	// Set all as text columns
	c := []SQColumn{}
	for _, name := range cols {
		c = append(c, C(name))
	}

	// Drop table
	if this.Overwrite {
		result = append(result, N(this.Name).DropTable().IfExists())
	}

	// Create table
	result = append(result, N(this.Name).CreateTable(c...))

	// Return success
	return result
}

func (this *writer) Insert(cols, rows []string) []SQStatement {
	return []SQStatement{
		&row{N(this.Name).Insert(cols...), rows},
	}
}

func (this *writer) Select() (SQRows, error) {
	return this.SQConnection.Query(S(N(this.Name)))
}

func (this *writer) Do(st []SQStatement) error {
	return this.SQConnection.Do(func(txn SQTransaction) error {
		for _, st := range st {
			fmt.Println(st)
			if st_, ok := st.(*row); ok {
				if _, err := txn.Exec(st, to_interface(st_.args)...); err != nil {
					return err
				}
			} else if _, err := txn.Exec(st); err != nil {
				return err
			}
		}
		return nil
	})
}

func to_interface(args []string) []interface{} {
	var result []interface{}
	for _, arg := range args {
		result = append(result, arg)
	}
	return result
}
