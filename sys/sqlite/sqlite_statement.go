/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqlite

import (
	"fmt"
	"strconv"
)

////////////////////////////////////////////////////////////////////////////////
// STATEMENT IMPLEMENTATION

func (this *statement) Query() string {
	if this.statement == nil {
		return ""
	} else {
		return this.query
	}
}

func (this *statement) String() string {
	if this.statement != nil {
		return fmt.Sprintf("<sqlite.Statement>{ %v num_input=%v }", strconv.Quote(this.query), this.statement.NumInput())
	} else {
		return fmt.Sprintf("<sqlite.Statement>{ nil }")
	}
}
