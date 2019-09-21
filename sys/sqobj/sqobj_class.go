/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqobj

import (
	"fmt"
	"strconv"

	sq "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *sqobj) NewClass(name string) *sqclass {
	class := &sqclass{name, nil, this.conn, this.log}
	if class.insert = this.conn.NewInsert(name); class.insert == nil {
		return nil
	} else {
		return class
	}
}

func (this *sqclass) Name() string {
	return this.name
}

func (this *sqclass) Insert(v interface{}) (int64, error) {
	var rowid int64
	err := this.conn.Tx(func(txn sq.Connection) error {
		if result, err := txn.Do(this.insert, 0, 0); err != nil {
			return err
		} else {
			rowid = result.LastInsertId
			return nil
		}
	})
	return rowid, err
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *sqclass) String() string {
	return fmt.Sprintf("<sqobj.Class>{ name=%v }", strconv.Quote(this.name))
}
