package sqlite

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Queries represents a collection of named SQL statements parsed from a reader.
// SQL statements are separated by comment lines in the format: -- <identifier>
// or -- <identifier>.<identifier>, with any number of dots between them. Each
// statement is stored with its key.
type Queries struct {
	keys    []string
	queries map[string]string
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	// A key is identifiers separated by dots, where an identifier is letters,
	// digits and underscores, which doesn't start with a digit, so a comment
	// which ends in a word and a period, such as "-- form.", isn't a key
	reQuerySeparator = regexp.MustCompile(`^--\s*([A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)*)\s*$`)
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewQueries parses SQL statements from a reader and returns a Queries collection.
// SQL statements must be separated by comment lines with a key: -- <identifier>
// or -- <identifier>.<identifier>, with any number of dots, where an identifier
// consists of letters, digits and underscores, and doesn't start with a digit.
// Other comment lines, such as those which describe a statement, or one which
// ends a sentence with a period, such as "-- form.", and comments at the end of
// a line, are removed from the statement, except within strings, quoted
// identifiers and dollar-quoted bodies, such as those of functions.
// Returns an error if duplicate keys are found or if reading fails.
//
// Example input format:
//
//	-- user.select
//	SELECT * FROM users WHERE id = $1;
//
//	-- user.insert
//	INSERT INTO users (name, email) VALUES ($1, $2);
func NewQueries(r io.Reader) (*Queries, error) {
	var key string
	var sql strings.Builder

	scanner := bufio.NewScanner(r)
	self := &Queries{
		queries: make(map[string]string),
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Check if line matches separator pattern
		if matches := reQuerySeparator.FindStringSubmatch(line); matches != nil {
			// Save previous statement if exists
			if key != "" {
				self.queries[key] = strings.TrimSpace(sql.String())
				self.keys = append(self.keys, key)
			}

			// Get new key
			key = matches[1]

			// Check for duplicate key
			if _, exists := self.queries[key]; exists {
				return nil, ErrBadParameter.Withf("duplicate SQL statement key: %q", key)
			}

			sql.Reset()
			continue
		}

		// Accumulate SQL lines for current statement
		sql.WriteString(line)
		sql.WriteString("\n")
	}

	// Save the last statement
	if key != "" {
		self.queries[key] = strings.TrimSpace(sql.String())
		self.keys = append(self.keys, key)
	}

	// Remove the comments, and any ; at the end of each SQL statement
	for k, v := range self.queries {
		self.queries[k] = strings.TrimSuffix(strings.TrimSpace(stripComments(v)), ";")
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Return success
	return self, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Keys returns a slice of all SQL statement identifiers in the order they were parsed.
func (s *Queries) Keys() []string {
	return s.keys
}

// Query retrieves the SQL statement associated with the given key.
// Returns an empty string if the key does not exist.
func (s *Queries) Query(key string) string {
	if sql, ok := s.queries[key]; ok {
		return sql
	}
	return ""
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// stripComments removes the -- comments from a statement: lines which are
// only a comment are removed, and a comment at the end of a line is removed
// with the space before it. A -- within a string, such as 'a--b', a quoted
// identifier, such as "a--b", or a dollar-quoted body, such as a function's,
// isn't a comment, and a dollar-quoted body is kept as it is.
func stripComments(sql string) string {
	var result strings.Builder
	var line strings.Builder
	var quote string // the quote which is open: ', E', ", or a dollar tag
	comment := false // whether the rest of the line is a comment

	// endLine writes the line, without any space at its end, unless it was
	// only a comment
	endLine := func(eol bool) {
		text := line.String()
		if trimmed := strings.TrimRight(text, " \t\r"); trimmed != "" || !comment {
			if quote == "" {
				text = trimmed
			}
			result.WriteString(text)
			if eol {
				result.WriteByte('\n')
			}
		}
		line.Reset()
		comment = false
	}

	for i := 0; i < len(sql); i++ {
		c := sql[i]
		switch {
		case c == '\n':
			endLine(true)
		case comment:
			// Skip the comment to the end of the line
		case quote == "'" || quote == "E'":
			line.WriteByte(c)
			if c == '\\' && quote == "E'" && i+1 < len(sql) {
				// A backslash escapes the next character in an E'' string
				i++
				line.WriteByte(sql[i])
			} else if c == '\'' {
				if i+1 < len(sql) && sql[i+1] == '\'' {
					// '' is a quote within the string
					i++
					line.WriteByte(sql[i])
				} else {
					quote = ""
				}
			}
		case quote == `"`:
			line.WriteByte(c)
			if c == '"' {
				quote = ""
			}
		case quote != "":
			// A dollar-quoted body ends with its tag
			if strings.HasPrefix(sql[i:], quote) {
				line.WriteString(quote)
				i += len(quote) - 1
				quote = ""
			} else {
				line.WriteByte(c)
			}
		case c == '-' && i+1 < len(sql) && sql[i+1] == '-':
			comment = true
		case c == '\'':
			if n := line.Len(); n > 0 && (line.String()[n-1] == 'E' || line.String()[n-1] == 'e') && (n == 1 || !isIdentChar(line.String()[n-2])) {
				quote = "E'"
			} else {
				quote = "'"
			}
			line.WriteByte(c)
		case c == '"':
			quote = `"`
			line.WriteByte(c)
		case c == '$':
			if tag := dollarTag(sql[i:]); tag != "" && (line.Len() == 0 || !isIdentChar(line.String()[line.Len()-1])) {
				quote = tag
				line.WriteString(tag)
				i += len(tag) - 1
			} else {
				line.WriteByte(c)
			}
		default:
			line.WriteByte(c)
		}
	}
	endLine(false)

	// Return the statement
	return result.String()
}

// dollarTag returns the dollar quote at the start of the string, such as $$
// or $body$, or an empty string if it doesn't start with one, such as $1 or
// ${"schema"}
func dollarTag(s string) string {
	for i := 1; i < len(s); i++ {
		switch c := s[i]; {
		case c == '$':
			return s[:i+1]
		case c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (i > 1 && c >= '0' && c <= '9') || c >= 0x80:
			// Part of the tag
		default:
			return ""
		}
	}
	return ""
}

// isIdentChar reports whether the byte can be part of an identifier
func isIdentChar(c byte) bool {
	return c == '_' || c == '$' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c >= 0x80
}
