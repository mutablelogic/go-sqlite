/*
	SQLite client
	(c) Copyright David Thorpe 2017
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqlite

/////////////////////////////////////////////////////////////////////
// TYPES

type q_CreateTable struct {
	Name        string
	Schema      string
	IfNotExists bool
	Columns     []Column
}

///////////////////////////////////////////////////////////////////////////////
// CONSTRUCT QUERIES

func CreateTable(name string) Query {
	return &q_CreateTable{
		Name: name,
	}
}
