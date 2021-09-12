package tokenizer

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	// Package imports
	sqlite3 "github.com/djthorpe/go-sqlite/sys/sqlite3"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite/pkg/quote"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// A tokenizer that scans the input SQL statement
type Tokenizer struct {
	*bufio.Scanner
}

type (
	KeywordToken    string // An SQL reserved keyword
	TypeToken       string // An SQL data type
	NameToken       string // A table or column identifier
	ValueToken      string // A value literal
	PuncuationToken string // A punctuation character
	WhitespaceToken string // Whitespace token
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reWhitespace = regexp.MustCompile(`^\s*$`)
	reName       = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	reNumber     = regexp.MustCompile(`^[+-]?([0-9]+([.][0-9]*)?|[.][0-9]+)$`)
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewTokenizer returns a new Tokenizer that scans the input SQL statement
func NewTokenizer(v string) *Tokenizer {
	t := &Tokenizer{bufio.NewScanner(strings.NewReader(v))}
	t.Scanner.Split(sqlSplit)
	return t
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Next returns the next token in the input stream, or returns io.EOF error and
// nil if there are no more tokens to comsume.
func (t *Tokenizer) Next() (interface{}, error) {
	if t.Scanner.Scan() {
		txt := t.Scanner.Text()
		return toToken(txt), nil
	}
	if t.Scanner.Err() != nil {
		return nil, t.Scanner.Err()
	} else {
		return nil, io.EOF
	}
}

// IsComplete returns true if the input string appears to be a complete SQL statement
func IsComplete(v string) bool {
	return sqlite3.IsComplete(v)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func toToken(v string) interface{} {
	if reWhitespace.MatchString(v) {
		return WhitespaceToken(v)
	} else if IsReservedWord(v) {
		return KeywordToken(v)
	} else if IsType(v) {
		return TypeToken(v)
	} else if reName.MatchString(v) {
		return NameToken(v)
	} else if reNumber.MatchString(v) {
		return ValueToken(v)
	} else {
		return PuncuationToken(v)
	}
}

func sqlSplit(data []byte, atEOF bool) (int, []byte, error) {
	advance, token, err := bufio.ScanWords(data, atEOF)
	if err != nil {
		return advance, token, err
	}

	// Check first letter for non-letter or non-digit
	r, width := utf8.DecodeRune(data)
	if width == 0 {
		return 0, token, ErrBadParameter.With("Invalid string")
	}
	if !(unicode.IsDigit(r) || unicode.IsLetter(r) || r == '_') {
		return width, []byte(string(r)), nil
	}

	// Count until non-letter or non-digit
	for i := width; i < len(data); i += width {
		r, width = utf8.DecodeRune(data[i:])
		if width == 0 {
			return 0, token, ErrBadParameter.With("Invalid string")
		}
		if !(unicode.IsDigit(r) || unicode.IsLetter(r) || r == '_') {
			return i, data[:i], nil
		}
	}

	// Return a word
	return advance, token, nil
}
