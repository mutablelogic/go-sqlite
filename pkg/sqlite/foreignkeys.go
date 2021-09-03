package sqlite

import (
	"fmt"

	// Import namespaces
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *connection) ForeignKeyConstraints() (bool, error) {
	rows, err := this.Query(Q("PRAGMA foreign_keys"))
	if err != nil {
		return false, err
	}
	defer rows.Close()
	value := rows.NextArray()
	if len(value) != 1 {
		return false, ErrNotImplemented.With("ForeignKeyConstraints")
	}
	switch fmt.Sprint(value[0]) {
	case "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return false, ErrUnexpectedResponse.With(value[0])
	}
}

func (this *connection) SetForeignKeyConstraints(enable bool) error {
	if v, err := this.ForeignKeyConstraints(); err != nil {
		return err
	} else if v != enable {
		if _, err := this.Exec(Q("PRAGMA foreign_keys=" + toSwitch(enable))); err != nil {
			return err
		}
	}
	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func toSwitch(v bool) string {
	if v {
		return "ON"
	} else {
		return "OFF"
	}
}
