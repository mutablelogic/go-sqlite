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
		"-- select":                 "select",
		"-- user.select":            "user.select",
		"-- feed.table.url":         "feed.table.url",
		"-- pgqueue.ticker_next":    "pgqueue.ticker_next",
		"-- _private.v2":            "_private.v2",
		"--heartbeat.list_for_user": "heartbeat.list_for_user",
		"--   spaced.key   ":        "spaced.key",

		// Not keys
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
		{`SELECT E'a\'--b' -- comment`, `SELECT E'a\'--b'`},
		{"SELECT 'line one\n-- line two' -- comment", "SELECT 'line one\n-- line two'"},

		// A dollar-quoted body is kept as it is, with its comments
		{"CREATE FUNCTION f() RETURNS INT AS $$\n  -- the answer\n  SELECT 42; -- forty-two\n$$ LANGUAGE SQL -- comment", "CREATE FUNCTION f() RETURNS INT AS $$\n  -- the answer\n  SELECT 42; -- forty-two\n$$ LANGUAGE SQL"},
		{"DO $body$ BEGIN -- start\nEND $body$; -- done", "DO $body$ BEGIN -- start\nEND $body$;"},

		// Positional parameters and placeholders aren't dollar quotes
		{"SELECT $1, $2 -- params", "SELECT $1, $2"},
		{`SELECT * FROM ${"schema"}.t -- table`, `SELECT * FROM ${"schema"}.t`},
		{"SELECT a$b$ -- not a tag", "SELECT a$b$"},
	} {
		assert.Equal(t, test.want, stripComments(test.sql), test.sql)
	}
}
