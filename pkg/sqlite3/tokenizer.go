package sqlite3

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite/pkg/quote"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Tokenizer struct {
	*bufio.Scanner
}

type (
	KeywordToken    string
	TypeToken       string
	NameToken       string
	ValueToken      string
	PuncuationToken string
	WhitespaceToken string
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

func NewTokenizer(v string) *Tokenizer {
	t := &Tokenizer{bufio.NewScanner(strings.NewReader(v))}
	t.Scanner.Split(sqlSplit)
	return t
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

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
