/*
	SQLite client
	(c) Copyright David Thorpe 2017
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqlite

import (
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	// Frameworks
	"github.com/araddon/dateparse"
)

/////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	reserved_words = `ABORT ACTION ADD AFTER ALL ALTER ANALYZE AND 
	AS ASC ATTACH AUTOINCREMENT
	BEFORE BEGIN BETWEEN BY CASCADE CASE CAST CHECK COLLATE COLUMN 
	COMMIT CONFLICT CONSTRAINT CREATE CROSS CURRENT_DATE CURRENT_TIME 
	CURRENT_TIMESTAMP DATABASE DEFAULT DEFERRABLE DEFERRED DELETE 
	DESC DETACH DISTINCT DROP EACH ELSE END ESCAPE EXCEPT EXCLUSIVE 
	EXISTS EXPLAIN FAIL FOR FOREIGN FROM FULL GLOB GROUP HAVING IF
	IGNORE IMMEDIATE IN INDEX INDEXED INITIALLY INNER INSERT INSTEAD
	INTERSECT INTO IS ISNULL JOIN KEY LEFT LIKE LIMIT MATCH NATURAL
	NO NOT NOTNULL NULL OF OFFSET ON OR ORDER OUTER PLAN PRAGMA PRIMARY
	QUERY RAISE RECURSIVE REFERENCES REGEXP REINDEX RELEASE RENAME
	REPLACE RESTRICT RIGHT ROLLBACK ROW SAVEPOINT SELECT SET TABLE
	TEMP TEMPORARY THEN TO TRANSACTION TRIGGER UNION UNIQUE UPDATE
	USING VACUUM VALUES VIEW VIRTUAL WHEN WHERE WITH WITHOUT`

	reserved_types = `TEXT BLOB DATETIME TIMESTAMP FLOAT INTEGER BOOL`
)

/////////////////////////////////////////////////////////////////////
// GLOBAL VARIABLES

var (
	reservedWords            = make(map[string]bool, 0)
	reservedTypes            = make(map[string]bool, 0)
	regexpBareIdentifier     = regexp.MustCompile("^[A-Za-z_][A-Za-z0-9_]*$")
	regexpDatetimeDDMMYYYY   = regexp.MustCompile("^(\\d{2})(\\d{2})(\\d{4})$")
	regexpDatetimeDD_MM_YYYY = regexp.MustCompile("^(\\d{2})\\/(\\d{2})\\/(\\d{4})$")
	regexpDatetimeDDMMYY     = regexp.MustCompile("^(\\d{2})(\\d{2})(\\d{2})$")
	regexpDatetimeDD_MM_YY   = regexp.MustCompile("^(\\d{2})\\/(\\d{2})\\/(\\d{2})$")
)

var (
	ErrUnsupportedType = errors.New("Unsupported type")
	ErrInvalidDate     = errors.New("Invalid date")
)

/////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// DoubleQuote puts double quotes around a string
// and escapes existing double quotes
func DoubleQuote(value string) string {
	// Change " into ""
	if strings.Contains(value, "\"") {
		value = strings.Replace(value, "\"", "\"\"", -1)
	}
	return "\"" + value + "\""
}

// QuoteIdentifier returns a safe version of an identifier
func QuoteIdentifier(value string) string {
	if isReservedWord(value) {
		// Check for reserved keyword
		return DoubleQuote(value)
	} else if isBareIdentifier(value) {
		return value
	} else {
		return DoubleQuote(value)
	}
}

// QuoteIdentifiers returns a safe version of a list of identifiers,
// separated by commas
func QuoteIdentifiers(values ...string) string {
	if len(values) == 0 {
		return ""
	} else if len(values) == 1 {
		return QuoteIdentifier(values[0])
	} else {
		arr := make([]string, len(values))
		for i, value := range values {
			arr[i] = QuoteIdentifier(value)
		}
		return strings.Join(arr, ",")
	}
}

// IsSupportedType returns true if the value provided is
// a reserved type
func IsSupportedType(value string) bool {
	return isReservedType(value)
}

// SupportedTypes returns all supported types
func SupportedTypes() []string {
	return strings.Fields(reserved_types)
}

// RowString returns a row as a string
func RowString(row []Value) string {
	if row == nil {
		return "<nil>"
	}
	str := make([]string, len(row))
	for i, v := range row {
		decltype := ""
		if v.Column() != nil {
			decltype = v.Column().DeclType()
		}
		switch {
		case v.IsNull():
			str[i] = "<nil>"
		case decltype == "TEXT":
			str[i] = strconv.Quote(v.String())
		case decltype == "INTEGER" || decltype == "FLOAT" || decltype == "BOOL":
			str[i] = v.String()
		case decltype == "TIMESTAMP" || decltype == "DATETIME":
			str[i] = v.Timestamp().Format(time.RFC3339)
		case decltype == "BLOB":
			str[i] = strings.ToUpper(hex.EncodeToString(v.Bytes()))
		default:
			str[i] = fmt.Sprintf("<%v>%v", decltype, strconv.Quote(v.String()))
		}
	}
	return fmt.Sprint("[" + strings.Join(str, ",") + "]")
}

// RowMap returns a row as <name>,<value> pairs
func RowMap(row []Value) map[string]Value {
	if row == nil {
		return nil
	}
	row_ := make(map[string]Value, len(row))
	for _, value := range row {
		if value.Column() != nil {
			row_[value.Column().Name()] = value
		}
	}
	return row_
}

// ToSupportedType converts from a string to a supported type. If the
// string is empty, will return <nil> if notnull is false
func ToSupportedType(value, decltype string, notnull bool, timezone *time.Location) (interface{}, error) {
	value_ := strings.TrimSpace(value)
	switch strings.ToUpper(decltype) {
	case "BOOL":
		if value_ == "" && notnull == false {
			return nil, nil
		} else if bool_, err := strconv.ParseBool(value_); err != nil {
			return nil, err
		} else {
			return bool_, nil
		}
	case "TEXT":
		if value == "" && notnull == false {
			return nil, nil
		} else {
			return value, nil
		}
	case "INTEGER":
		if value_ == "" && notnull == false {
			return nil, nil
		} else if int_, err := strconv.ParseInt(value_, 10, 64); err != nil {
			return nil, err
		} else {
			return int_, nil
		}
	case "FLOAT":
		if value_ == "" && notnull == false {
			return nil, nil
		} else if float_, err := strconv.ParseFloat(value_, 64); err != nil {
			return nil, err
		} else {
			return float_, nil
		}
	case "BLOB":
		if value_ == "" && notnull == false {
			return nil, nil
		} else if bytes_, err := hex.DecodeString(value_); err != nil {
			return nil, err
		} else {
			return bytes_, nil
		}
	case "TIMESTAMP":
		if value_ == "" && notnull == false {
			return nil, nil
		} else if timestamp_, err := time.Parse(time.RFC3339, value_); err == nil {
			return timestamp_, nil
		} else if timestamp_, err := time.Parse(time.RFC3339Nano, value_); err == nil {
			return timestamp_, nil
		} else {
			return nil, err
		}
	case "DATETIME":
		if value_ == "" && notnull == false {
			return nil, nil
		} else if datetime_, err := parseDatetime(value_, timezone); err != nil {
			return nil, err
		} else {
			return datetime_, nil
		}
	default:
		return nil, ErrUnsupportedType
	}
}

// SupportedTypesForValue returns most likely supported types
// for a value, in order. It will always return at least one type,
// with the most likely type as the zero'th element
func SupportedTypesForValue(value string) []string {
	all_types := SupportedTypes()
	supported_types := make([]string, 0, len(all_types))

	// Go in reverse order, assuming not an empty string
	if strings.TrimSpace(value) != "" {
		for i := len(all_types) - 1; i > 0; i-- {
			if _, err := ToSupportedType(value, all_types[i], true, time.UTC); err == nil {
				supported_types = append(supported_types, all_types[i])
			}
		}
	}

	// Always append the default type
	if _, err := ToSupportedType(value, all_types[0], true, time.UTC); err == nil {
		supported_types = append(supported_types, all_types[0])
	}

	// Return supported types in priority order
	return supported_types
}

func SupportedTypeForType(v interface{}) string {
	switch v.(type) {
	case int, int8, int16, int32, int64:
		return "INTEGER"
	case uint, uint8, uint16, uint32, uint64:
		return "INTEGER"
	case string:
		return "TEXT"
	case []byte:
		return "BLOB"
	case time.Time:
		return "TIMESTAMP"
	case float32, float64:
		return "FLOAT"
	case bool:
		return "BOOL"
	default:
		return ""
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func isBareIdentifier(value string) bool {
	return regexpBareIdentifier.MatchString(value)
}

func isReservedWord(value string) bool {
	var once sync.Once
	once.Do(func() {
		for _, word := range strings.Fields(reserved_words) {
			reservedWords[strings.ToUpper(word)] = true
		}
	})
	value = strings.TrimSpace(strings.ToUpper(value))
	_, exists := reservedWords[value]
	return exists
}

func isReservedType(value string) bool {
	var once sync.Once
	once.Do(func() {
		for _, word := range strings.Fields(reserved_types) {
			reservedTypes[strings.ToUpper(word)] = true
		}
	})
	value = strings.TrimSpace(strings.ToUpper(value))
	_, exists := reservedTypes[value]
	return exists
}

func parseDatetime(value string, timezone *time.Location) (time.Time, error) {
	var matches []string
	var yy int

	if matches = regexpDatetimeDDMMYY.FindStringSubmatch(value); len(matches) == 4 {
		// Matched Euro date, add on 2000
		yy = 1900
	} else if matches = regexpDatetimeDD_MM_YY.FindStringSubmatch(value); len(matches) == 4 {
		// Matched Euro date,add on 2000
		yy = 1900
	} else if matches = regexpDatetimeDDMMYYYY.FindStringSubmatch(value); len(matches) == 4 {
		// Matched Euro date
	} else if matches = regexpDatetimeDD_MM_YYYY.FindStringSubmatch(value); len(matches) == 4 {
		// Matched Euro date
	} else if datetime, err := dateparse.ParseIn(value, timezone); err != nil {
		// Parse error
		return time.Time{}, err
	} else {
		return datetime, err
	}

	if day, _ := strconv.ParseInt(matches[1], 10, 32); day < 1 || day > 31 {
		return time.Time{}, ErrInvalidDate
	} else if month, _ := strconv.ParseInt(matches[2], 10, 32); month < 1 || month > 12 {
		return time.Time{}, ErrInvalidDate
	} else if year, _ := strconv.ParseInt(matches[3], 10, 32); year > 9999 {
		return time.Time{}, ErrInvalidDate
	} else if datetime := time.Date(int(year)+yy, time.Month(month), int(day), 0, 0, 0, 0, timezone); datetime.IsZero() {
		return time.Time{}, ErrInvalidDate
	} else {
		return datetime, nil
	}
}
