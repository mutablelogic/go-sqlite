package sqlite3

import (
	"strconv"
	"strings"
	"time"
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

// intMax returns the maximum of two int values
func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// intMin returns the minimum of two int values
func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// maxInt32 returns the maximum of two int32 values
func maxInt32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

// durationMin returns the minimum of two time.Duration values
func durationMin(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// durationMax returns the maximum of two time.Duration values
func durationMax(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
