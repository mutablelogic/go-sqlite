package sqlite3

import (
	// Import namespaces
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *Conn) ForeignKeyConstraints() (bool, error) {
	var enable bool
	if err := this.Exec(Q("PRAGMA foreign_keys"), func(row, _ []string) bool {
		enable = stringToBool(row[0])
		return false
	}); err != nil {
		return false, err
	}
	// Return success
	return enable, nil
}

func (this *Conn) SetForeignKeyConstraints(enable bool) error {
	if v, err := this.ForeignKeyConstraints(); err != nil {
		return err
	} else if v == enable {
		return nil
	}
	return this.Exec(Q("PRAGMA foreign_keys=", V(enable)), nil)
}
