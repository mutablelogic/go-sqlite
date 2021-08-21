package sqlite

import (
	"fmt"
	"strings"

	// Modules
	sqlite "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

func (this *connection) Schemas() []string {
	// Perform the query
	rs, err := this.Query(Q("PRAGMA database_list"))
	if err != nil {
		return nil
	}
	defer rs.Close()

	// Collate the results
	schemas := make([]string, 0, 1)
	for {
		row := rs.NextMap()
		if row == nil {
			break
		}
		schemas = append(schemas, row["name"].(string))
	}

	// Return success
	return schemas
}

func (this *connection) Tables() []string {
	return this.TablesEx("", false)
}

func (this *connection) TablesEx(schema string, temp bool) []string {
	// Create the query
	query := ""
	if temp {
		query = `
			SELECT name FROM 
   				(SELECT name,type FROM %ssqlite_master UNION ALL SELECT name,type FROM %ssqlite_temp_master)
			WHERE type=? AND name NOT LIKE 'sqlite_%%'
			ORDER BY name ASC
		`
	} else {
		query = `
			SELECT name FROM 
				%ssqlite_master 
			WHERE type=? AND name NOT LIKE 'sqlite_%%'
			ORDER BY name ASC -- %s
		`
	}

	// Append the schema
	if schema != "" {
		query = fmt.Sprintf(query, sqlite.QuoteIdentifier(schema)+".", sqlite.QuoteIdentifier(schema)+".")
	} else {
		query = fmt.Sprintf(query, "", "")
	}

	// Perform the query
	rows, err := this.Query(Q(query), "table")
	if err != nil {
		return nil
	}
	defer rows.Close()

	// Collate the results
	names := make([]string, 0, 10)
	for {
		values := rows.NextArray()
		if values == nil {
			break
		} else if len(values) != 1 {
			return nil
		} else {
			names = append(names, fmt.Sprint(values[0]))
		}
	}

	// Return success
	return names
}

func (this *connection) Columns(name string) []sqlite.SQColumn {
	return this.ColumnsEx(name, "")
}

func (this *connection) ColumnsEx(name, schema string) []sqlite.SQColumn {
	// Perform query
	query := "table_info(" + sqlite.QuoteIdentifier(name) + ")"
	if schema != "" {
		query = "PRAGMA " + sqlite.QuoteIdentifier(schema) + "." + query
	} else {
		query = "PRAGMA " + query
	}
	rs, err := this.Query(Q(query))
	if err != nil {
		return nil
	}
	defer rs.Close()

	// Collate results, estimate up to 10 columns
	columns := make([]sqlite.SQColumn, 0, 10)
	for {
		row := rs.NextMap()
		if row == nil {
			break
		}
		col := N(row["name"].(string)).WithType(row["type"].(string))
		if row["notnull"].(int64) != 0 {
			col = col.NotNull()
		}
		columns = append(columns, col)
	}
	return columns
}

func (this *connection) Modules(prefix string) []string {
	// Perform query
	rs, err := this.Query(Q("PRAGMA module_list"))
	if err != nil {
		return nil
	}
	defer rs.Close()

	// Collate results, estimate up to 10 rows
	result := make([]string, 0, 10)
	for {
		row := rs.NextArray()
		if row == nil || len(row) < 1 {
			break
		}
		module := row[0].(string)
		if prefix == "" || strings.HasPrefix(module, prefix) {
			result = append(result, row[0].(string))
		}
	}

	// Return nil if no matching results
	if len(result) == 0 {
		return nil
	} else {
		return result
	}
}
