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

	gopi "github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *sqobj) isExistingTable(name string) bool {
	for _, table := range this.conn.Tables() {
		if table == name {
			return true
		}
	}
	return false
}

func (this *sqobj) registeredClass(name, pkgpath string) *sqclass {
	if classes, exists := this.class[pkgpath]; exists == false {
		return nil
	} else if class, exists := classes[name]; exists == false {
		return nil
	} else {
		return class
	}
}

// classesFor maps an array of objects v to their classes
func (this *sqobj) classesFor(v []interface{}) ([]*sqclass, error) {
	classmap := make([]*sqclass, len(v))
	for i, value := range v {
		if name, pkgpath := this.reflectName(value); name == "" {
			return nil, fmt.Errorf("%w: No struct name", gopi.ErrAppError)
		} else if class := this.registeredClass(name, pkgpath); class == nil {
			return nil, fmt.Errorf("%w: No registered class for %v (in path %v)", gopi.ErrNotFound, strconv.Quote(name), strconv.Quote(pkgpath))
		} else {
			classmap[i] = class
		}
	}
	// Success
	return classmap, nil
}
