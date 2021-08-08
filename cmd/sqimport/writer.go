package main

import (
	"github.com/djthorpe/go-sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type writer struct {
	sqlite.SQConnection

	// The writer's table name
	Name string

	// When Overwrite is true, drop table before creating it
	Overwrite bool
}

type row struct {
	sqlite.SQStatement
	args []string
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewWriter(db sqlite.SQConnection) *writer {
	return &writer{db, "", false}
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

func (this *writer) CreateTable(cols []string) []sqlite.SQStatement {
	result := []sqlite.SQStatement{}

	// Set all as text columns
	c := []sqlite.SQColumn{}
	for _, name := range cols {
		c = append(c, this.Column(name, "TEXT"))
	}

	// Drop table
	if this.Overwrite {
		result = append(result, this.DropTable(this.Name).IfExists())
	}

	// Create table
	result = append(result, this.SQConnection.CreateTable(this.Name, c...))

	// Return success
	return result
}

func (this *writer) Insert(cols, rows []string) []sqlite.SQStatement {
	return []sqlite.SQStatement{
		&row{this.SQConnection.Insert(this.Name, cols...), rows},
	}
}

func (this *writer) Do(st []sqlite.SQStatement) error {
	return this.SQConnection.Do(func(txn sqlite.SQTransaction) error {
		for _, st := range st {
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
