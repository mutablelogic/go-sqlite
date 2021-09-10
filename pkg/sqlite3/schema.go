package sqlite3

import (
	"fmt"
	"strconv"
	"strings"

	// Namespace imports
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Schemas returns a list of schemas
func (c *Conn) Schemas() []string {
	var schemas []string
	if err := c.Exec(Q("PRAGMA database_list"), func(row, col []string) bool {
		schemas = append(schemas, row[1])
		return false
	}); err != nil {
		return nil
	}
	return schemas
}

// Filename returns the filename for a schema
func (c *Conn) Filename(schema string) string {
	if schema == "" {
		return c.Filename(defaultSchema)
	}
	return c.ConnEx.Filename(schema)
}

// Tables returns a list of table names in a schema
func (c *Conn) Tables(schema string) []string {
	if schema == "" {
		return c.Tables(defaultSchema)
	}
	return c.objectsInSchema(schema, "table")
}

// ColumnsForTable returns the columns in a table
func (c *Conn) ColumnsForTable(schema, table string) []SQColumn {
	if schema == "" {
		return c.ColumnsForTable(defaultSchema, table)
	}
	var result []SQColumn
	if err := c.Exec(Q("PRAGMA ", N(schema), ".table_info(", N(table), ")"), func(row, k []string) bool {
		// k is "cid" "name" "type" "notnull" "dflt_value" "pk"
		col := C(row[1]).WithType(row[2])
		if stringToBool(row[3]) {
			col = col.NotNull()
		}
		if stringToBool(row[5]) {
			col = col.WithPrimary()
		}
		// TODO: Add default value, auto increment
		result = append(result, col)
		return false
	}); err != nil {
		fmt.Println(err)
		return nil
	}
	return result
}

// ColumnsForIndex returns the indexes associated with a table
func (c *Conn) ColumnsForIndex(schema, index string) []string {
	if schema == "" {
		return c.ColumnsForIndex(defaultSchema, index)
	}

	var result []string
	if err := c.Exec(Q("PRAGMA ", N(schema), ".index_info(", N(index), ")"), func(row, c []string) bool {
		fmt.Printf("%q %q => %q\n", index, row, c)
		return false
	}); err != nil {
		return nil
	}
	return result
}

// IndexesForTable returns the indexes associated with a table
func (c *Conn) IndexesForTable(schema, table string) []SQIndexView {
	if table == "" {
		return nil
	} else if schema == "" {
		return c.IndexesForTable(defaultSchema, table)
	}
	var result []SQIndexView
	if err := c.Exec(Q("PRAGMA ", N(schema), ".index_list(", N(table), ")"), func(row, _ []string) bool {
		// columns are is "seq" "name" "unique" "origin" "partial"

		// Get index column names, abort if error
		names := c.ColumnsForIndex(schema, row[1])
		if names == nil {
			return true
		}

		// Construct index statement
		index := N(row[1]).WithSchema(schema).CreateIndex(table, names...)
		if schema == tempSchema {
			index = index.WithTemporary()
		}
		if stringToBool(row[2]) || row[3] == "u" || row[3] == "pk" {
			index = index.WithUnique()
		}
		if row[3] != "c" {
			index = index.WithAuto()
		}
		result = append(result, index)
		return false
	}); err != nil {
		return nil
	}
	return result
}

// Views returns a list of view names in a schema
func (c *Conn) Views(schema string) []string {
	if schema == "" {
		return c.Views(defaultSchema)
	}
	return c.objectsInSchema(schema, "view")
}

// Modules returns a list of modules in a schema. If an argument is
// provided, then only modules with those name prefixes are returned.
func (c *Conn) Modules(prefix ...string) []string {
	// Get the names, return
	var result []string
	if err := c.Exec(Q("PRAGMA module_list"), func(row, _ []string) bool {
		if module := row[0]; len(prefix) == 0 || inPrefixList(prefix, module) {
			result = append(result, module)
		}
		return false
	}); err != nil {
		return nil
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (c *Conn) objectsInSchema(schema, t string) []string {
	// Set the schema
	tableName := N("sqlite_master").WithSchema(schema)
	if schema == tempSchema {
		tableName = N("sqlite_temp_master").WithSchema(schema)
	}

	// Get the names, return
	var result []string
	if err := c.Exec(Q("SELECT name FROM ", tableName, " WHERE type=", V(t), " AND name NOT LIKE 'sqlite_%%'"), func(row, _ []string) bool {
		result = append(result, row[0])
		return false
	}); err != nil {
		return nil
	}
	return result
}

func (c *Conn) indexesInSchema(schema, t string) []string {
	// Set the schema
	tableName := N("sqlite_master").WithSchema(schema)
	if schema == tempSchema {
		tableName = N("sqlite_temp_master").WithSchema(schema)
	}

	// Get the names, return
	var result []string
	if err := c.Exec(Q("SELECT name FROM ", tableName, " WHERE type=", V(t), " AND name NOT LIKE 'sqlite_%%'"), func(row, _ []string) bool {
		result = append(result, row[0])
		return false
	}); err != nil {
		return nil
	}
	return result
}

func inPrefixList(prefix []string, name string) bool {
	name = strings.ToUpper(name)
	for _, p := range prefix {
		p = strings.ToUpper(p)
		if name == p || strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

func stringToBool(v string) bool {
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	} else if n, err := strconv.ParseUint(v, 0, 32); err == nil {
		return n != 0
	} else {
		return false
	}
}
