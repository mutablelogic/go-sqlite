package quote

import (
	"regexp"
	"strings"
)

/////////////////////////////////////////////////////////////////////
// GLOBAL VARIABLES

var (
	regexpBareIdentifier = regexp.MustCompile("^[A-Za-z_][A-Za-z0-9_]*$")
)

/////////////////////////////////////////////////////////////////////
// FUNCTIONS

// Quote puts single quotes around a string and escapes existing single quotes
func Quote(value string) string {
	return "'" + strings.Replace(value, "'", "''", -1) + "'"
}

// DoubleQuote puts double quotes around a string and escapes existing double quotes
func DoubleQuote(value string) string {
	return "\"" + strings.Replace(value, "\"", "\"\"", -1) + "\""
}

// QuoteIdentifier returns a safe version of an identifier
func QuoteIdentifier(v string) string {
	if IsReservedWord(v) {
		return DoubleQuote(v)
	} else if isBareIdentifier(v) {
		return v
	} else {
		return DoubleQuote(v)
	}
}

// QuoteIdentifiers returns a safe version of a list of identifiers,
// separated by commas
func QuoteIdentifiers(v ...string) string {
	if len(v) == 0 {
		return ""
	}
	if len(v) == 1 {
		return QuoteIdentifier(v[0])
	}
	result := make([]string, len(v))
	for i, v_ := range v {
		result[i] = QuoteIdentifier(v_)
	}
	return strings.Join(result, ",")
}

// QuoteDeclType returns a supported type or quotes type
// TEXT => TEXT
// TIMESTAMP => TIMESTAMP
// some other type => "some other type"
func QuoteDeclType(v string) string {
	if IsType(v) {
		return v
	}
	if isBareIdentifier(v) {
		return v
	}
	return DoubleQuote(v)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func isBareIdentifier(value string) bool {
	return regexpBareIdentifier.MatchString(value)
}
