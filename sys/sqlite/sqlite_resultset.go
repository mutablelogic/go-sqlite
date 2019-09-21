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
	"strings"

	// Frameworks
	sq "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// COLUMN IMPLEMENTATION

func (this *column) DeclType() string {
	return this.decltype
}

func (this *column) Name() string {
	return this.name
}

func (this *column) Nullable() bool {
	return this.nullable
}

func (this *column) PrimaryKey() bool {
	return this.primary
}

func (this *column) Index() int {
	return this.index
}

func (this *column) Query() string {
	if this.nullable {
		return fmt.Sprintf("%v %v", sq.QuoteIdentifier(this.name), this.decltype)
	} else {
		return fmt.Sprintf("%v %v NOT NULL", sq.QuoteIdentifier(this.name), this.decltype)
	}
}

func (this *column) String() string {
	tokens := []string{}
	if this.primary {
		tokens = append(tokens, "primary")
	}
	if this.nullable {
		tokens = append(tokens, "nullable")
	}
	if this.index >= 0 {
		tokens = append(tokens, "index="+fmt.Sprint(this.index))
	}
	return fmt.Sprintf("<sqlite.Column>{ name=%v decltype=%v %v }", strconv.Quote(this.name), strconv.Quote(this.decltype), strings.Join(tokens, " "))
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
