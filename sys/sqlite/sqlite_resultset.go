/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package sqlite

import (
	"fmt"
	"strconv"

	// Frameworks
	sq "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// COLUMN IMPLEMENTATION

func (this *column) Pos() int {
	return this.pos
}

func (this *column) DeclType() string {
	return this.decltype
}

func (this *column) Name() string {
	return this.name
}

func (this *column) String() string {
	return fmt.Sprintf("<sqlite.Column>{ name=%v decltype=%v pos=%v }", strconv.Quote(this.name), strconv.Quote(this.decltype), this.pos)
}

////////////////////////////////////////////////////////////////////////////////
// RESULTSET IMPLEMENTATION

// Return column names
func (this *resultset) Columns() []sq.Column {
	return this.columns
}

// Return next row or nil
func (this *resultset) Next() []sq.Value {
	if this.rows == nil {
		return nil
	} else if err := this.rows.Next(this.values); err != nil {
		this.rows.Close()
		this.rows = nil
		return nil
	} else {
		values := make([]sq.Value, len(this.values))
		for i, v := range this.values {
			values[i] = &value{v, this.columns[i].(*column)}
		}
		return values
	}
}

// Destroy resultset
func (this *resultset) Destroy() error {
	var err error
	if this.rows != nil {
		err = this.rows.Close()
		this.rows = nil
	}
	return err
}

// Stringify
func (this *resultset) String() string {
	if this.rows != nil {
		return fmt.Sprintf("<sqlite.Resultset>{ columns=%v }", this.columns)
	} else {
		return fmt.Sprintf("<sqlite.Resultset>{ nil }")
	}
}
