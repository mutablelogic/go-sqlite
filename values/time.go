package values

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	// The formats of text which is a time, as SQLite's date and time
	// functions read and write them, and RFC 3339, of which Set writes
	// times, in which a time without a time zone is UTC
	timeFormats = []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02T15:04:05.999999999",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02T15:04Z07:00",
		"2006-01-02 15:04Z07:00",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02",
	}
)

const (
	// The Julian day of the Unix epoch
	unixJulianDay = 2440587.5
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Time returns the value's time, as SQLite's date and time functions read
// them: text, of RFC 3339, or of YYYY-MM-DD HH:MM:SS, with optional fractions
// of a second and time zone, which is otherwise UTC, an integer, of seconds
// since the Unix epoch, or a float, of a Julian day
func Time(v Value) (time.Time, error) {
	switch v.Type() {
	case Text:
		text := strings.TrimSpace(v.Text())
		for _, format := range timeFormats {
			if t, err := time.Parse(format, text); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("%q isn't a time", text)
	case Integer:
		return time.Unix(v.Int64(), 0).UTC(), nil
	case Float:
		days := v.Float() - unixJulianDay
		seconds, fraction := math.Modf(days * 86400)
		return time.Unix(int64(seconds), int64(math.Round(fraction*1e9))).UTC(), nil
	}
	return time.Time{}, fmt.Errorf("a %s isn't a time", v.Type())
}

// Duration returns the value's duration: an integer, of nanoseconds, which
// Set writes, a float, of seconds, which SQLite's date and time arithmetic
// returns, such as (julianday(b) - julianday(a)) * 86400, or text, of Go's
// format, such as 1h30m, or of nanoseconds
func Duration(v Value) (time.Duration, error) {
	switch v.Type() {
	case Integer:
		return time.Duration(v.Int64()), nil
	case Float:
		seconds := v.Float() * float64(time.Second)
		if seconds != seconds || seconds >= math.MaxInt64 || seconds <= math.MinInt64 {
			return 0, fmt.Errorf("%g seconds isn't a duration", v.Float())
		}
		return time.Duration(math.Round(seconds)), nil
	case Text:
		text := strings.TrimSpace(v.Text())
		if d, err := time.ParseDuration(text); err == nil {
			return d, nil
		}
		if n, err := strconv.ParseInt(text, 10, 64); err == nil {
			return time.Duration(n), nil
		}
		return 0, fmt.Errorf("%q isn't a duration", text)
	}
	return 0, fmt.Errorf("a %s isn't a duration", v.Type())
}
