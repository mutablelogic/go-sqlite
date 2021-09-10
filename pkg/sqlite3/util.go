package sqlite3

import (
	"strconv"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	trueValues = []string{"1", "y", "yes", "true", "ok", "on"}
)

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// inList returns true if a case-insensitive name
// either matches or is a prefix in the list of values provided
func inList(values []string, name string, prefix bool) bool {
	name = strings.ToUpper(name)
	for _, p := range values {
		p = strings.ToUpper(p)
		if name == p {
			return true
		} else if prefix && strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

// stringToBool returns true is a case-insensitive value provided
// is Y, TRUE, YES, OK, ON or any positive integer else it returns
// false
func stringToBool(v string) bool {
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	} else if inList(trueValues, v, false) {
		return true
	} else if n, err := strconv.ParseUint(v, 0, 32); err == nil {
		return n != 0
	} else {
		return false
	}
}
