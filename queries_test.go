package sqlite

import (
	"strings"
	"testing"

	// Packages
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

func Test_reQuerySeparator(t *testing.T) {
	for line, key := range map[string]string{
		// Keys
		"-- user.select":            "user.select",
		"-- feed.table.url":         "feed.table.url",
		"-- pgqueue.ticker_next":    "pgqueue.ticker_next",
		"-- _private.v2":            "_private.v2",
		"--heartbeat.list_for_user": "heartbeat.list_for_user",
		"--   spaced.key   ":        "spaced.key",

		// Not keys, such as a comment of one word
		"-- select":               "",
		"-- change":               "",
		"-- form.":                "",
		"-- creates.":             "",
		"-- user.":                "",
		"-- .user":                "",
		"-- user..select":         "",
		"-- 123":                  "",
		"-- user.1select":         "",
		"-- user-select":          "",
		"-- The words to search":  "",
		"-- user.select;":         "",
		"--":                      "",
		"SELECT 1 -- user.select": "",
	} {
		if matches := reQuerySeparator.FindStringSubmatch(line); key == "" {
			assert.Nil(t, matches, line)
		} else if assert.NotNil(t, matches, line) {
			assert.Equal(t, key, matches[1], line)
		}
	}
}

func Test_NewQueries(t *testing.T) {
	queries, err := NewQueries(strings.NewReader(`
-- user.table
-- The users, whose comment ends a sentence on its own line, which isn't a
-- key, and so is part of the statement.
CREATE TABLE users (id INT);

-- user.select
SELECT * FROM users -- the users
WHERE id = @id; -- by their identifier
`))
	require.NoError(t, err)
	assert.Equal(t, []string{"user.table", "user.select"}, queries.Keys())
	assert.Equal(t, "CREATE TABLE users (id INT)", queries.Query("user.table"))
	assert.Equal(t, "SELECT * FROM users\nWHERE id = @id", queries.Query("user.select"))

	// A line which looks like a key, in a string, or a block comment, isn't
	queries, err = NewQueries(strings.NewReader("-- query.sample\nSELECT 'first\n-- embedded.key\nlast';\n-- query.other\nSELECT /*\n-- inside.key\n*/ 2;\n"))
	require.NoError(t, err)
	assert.Equal(t, []string{"query.sample", "query.other"}, queries.Keys())
	assert.Equal(t, "SELECT 'first\n-- embedded.key\nlast'", queries.Query("query.sample"))
	assert.Equal(t, "SELECT /*\n-- inside.key\n*/ 2", queries.Query("query.other"))

	// A comment line of one word isn't a key, so it's removed from the
	// statement, which isn't empty
	queries, err = NewQueries(strings.NewReader("-- volume.put\n-- which are modified when they\n-- change\nINSERT INTO volume VALUES (1);\n"))
	require.NoError(t, err)
	assert.Equal(t, []string{"volume.put"}, queries.Keys())
	assert.Equal(t, "INSERT INTO volume VALUES (1)", queries.Query("volume.put"))

	// A key without a statement, such as before another key, or at the end,
	// is an error
	_, err = NewQueries(strings.NewReader("-- user.select\n-- a comment\n-- user.insert\nINSERT INTO users VALUES (1);\n"))
	assert.ErrorIs(t, err, ErrBadParameter)
	_, err = NewQueries(strings.NewReader("-- user.select\nSELECT 1;\n-- user.insert\n;\n"))
	assert.ErrorIs(t, err, ErrBadParameter)

	// Keys can't be duplicated
	_, err = NewQueries(strings.NewReader("-- user.select\nSELECT 1;\n-- user.select\nSELECT 2;\n"))
	assert.ErrorIs(t, err, ErrBadParameter)
}

func Test_stripComments(t *testing.T) {
	for _, test := range []struct {
		sql, want string
	}{
		// Comments on their own lines, and at the end of lines
		{"-- comment\nSELECT 1", "SELECT 1"},
		{"SELECT 1 -- comment", "SELECT 1"},
		{"SELECT 1, -- one\n  2 -- two\n", "SELECT 1,\n  2\n"},
		{"  -- indented comment\nSELECT 1", "SELECT 1"},
		{"SELECT 1--comment", "SELECT 1"},
		{"SELECT 1; -- comment", "SELECT 1;"},

		// Blank lines which aren't comments are kept
		{"SELECT 1\n\nSELECT 2", "SELECT 1\n\nSELECT 2"},

		// A -- within a string or a quoted identifier isn't a comment
		{"SELECT 'a--b' -- comment", "SELECT 'a--b'"},
		{"SELECT 'it''s -- not' AS x", "SELECT 'it''s -- not' AS x"},
		{`SELECT "a--b" FROM t -- comment`, `SELECT "a--b" FROM t`},
		{"SELECT 'line one\n-- line two' -- comment", "SELECT 'line one\n-- line two'"},

		{`SELECT "a""--b" FROM t -- comment`, `SELECT "a""--b" FROM t`},

		// SQLite's other quoted identifiers, [a] and `a`, aren't comments
		{"SELECT 1 AS [a--b] -- comment", "SELECT 1 AS [a--b]"},
		{"SELECT 1 AS `a--b` -- comment", "SELECT 1 AS `a--b`"},
		{"SELECT 1 AS `a``--b` -- comment", "SELECT 1 AS `a``--b`"},

		// A block comment is kept as it is, with -- in it, on its lines
		{"SELECT /* -- note */ 1 -- comment", "SELECT /* -- note */ 1"},
		{"SELECT /* one\n-- two\n*/ 1 -- comment", "SELECT /* one\n-- two\n*/ 1"},

		// Parameters and placeholders are kept
		{"SELECT @id -- param", "SELECT @id"},
		{`SELECT * FROM ${"schema"}.t -- table`, `SELECT * FROM ${"schema"}.t`},
	} {
		assert.Equal(t, test.want, stripComments(test.sql), test.sql)
	}
}

func Test_Statement(t *testing.T) {
	for _, test := range []struct {
		sql, want string
	}{
		{"SELECT 1", "SELECT 1"},
		{"SELECT 1;", "SELECT 1"},
		{"SELECT 1 ; ;\n", "SELECT 1"},
		{"SELECT 1; -- the one", "SELECT 1"},
		{"SELECT ';' /* ; */ FROM t;", "SELECT ';' /* ; */ FROM t"},
		{"SELECT 1; /* explanation */", "SELECT 1"},
		{"SELECT 1 /* a */ ; /* b */ ;", "SELECT 1"},
		{"SELECT '; /* x */' AS [a;]", "SELECT '; /* x */' AS [a;]"},
		{"SELECT `a;` ;", "SELECT `a;`"},
	} {
		assert.Equal(t, test.want, Statement(test.sql), test.sql)
	}
}
