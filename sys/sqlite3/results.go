package sqlite3

import "fmt"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Results struct {
	st   *Statement
	err  error
	cols []interface{}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r *Results) String() string {
	str := "<results"
	if r.st != nil {
		str += " " + r.st.String()
	}
	if r.err != nil && r.err != SQLITE_ROW {
		str += fmt.Sprintf(" err=%q", r.err.Error())
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

// Return a new results object
func results(st *Statement, err error) *Results {
	r := new(Results)
	r.st = st
	r.err = err
	r.cols = make([]interface{}, 0, st.ColumnCount())
	return r
}

// Return next row of values, or nil if there are
// no more rows. If there is an error, this method will currently
// panic.
func (r *Results) Next() []interface{} {
	if r.err == SQLITE_DONE {
		r.st.Finalize()
		r.st = nil
		r.cols = nil
		return nil
	} else if r.err == SQLITE_ROW {
		len := r.st.DataCount()
		r.cols = r.cols[:len]
		for i := 0; i < len; i++ {
			r.cols[i] = r.value(i)
		}
		r.err = r.st.Step()
		return r.cols
	} else {
		// Not yet handling other errors
		panic(r.err)
	}
}

// Return column names for the next row to be fetched
func (r *Results) ColumnNames() []string {
	if r.st == nil {
		return nil
	}
	len := r.st.ColumnCount()
	result := make([]string, len)
	for i := 0; i < len; i++ {
		result[i] = r.st.ColumnName(i)
	}
	return result
}

// Return column types for the next row to be fetched
func (r *Results) ColumnTypes() []Type {
	if r.st == nil {
		return nil
	}
	len := r.st.ColumnCount()
	result := make([]Type, len)
	for i := 0; i < len; i++ {
		result[i] = r.st.ColumnType(i)
	}
	return result
}

// Return column decltypes for the next row to be fetched
func (r *Results) ColumnDeclTypes() []string {
	if r.st == nil {
		return nil
	}
	len := r.st.ColumnCount()
	result := make([]string, len)
	for i := 0; i < len; i++ {
		result[i] = r.st.ColumnDeclType(i)
	}
	return result
}

// Return the source database schema name for the next row to be fetched
func (r *Results) ColumnDatabaseNames() []string {
	if r.st == nil {
		return nil
	}
	len := r.st.ColumnCount()
	result := make([]string, len)
	for i := 0; i < len; i++ {
		result[i] = r.st.ColumnDatabaseName(i)
	}
	return result
}

// Return the source table name for the next row to be fetched
func (r *Results) ColumnTableNames() []string {
	if r.st == nil {
		return nil
	}
	len := r.st.ColumnCount()
	result := make([]string, len)
	for i := 0; i < len; i++ {
		result[i] = r.st.ColumnTableName(i)
	}
	return result
}

// Return the origin for the next row to be fetched
func (r *Results) ColumnOriginNames() []string {
	if r.st == nil {
		return nil
	}
	len := r.st.ColumnCount()
	result := make([]string, len)
	for i := 0; i < len; i++ {
		result[i] = r.st.ColumnOriginName(i)
	}
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (r *Results) value(index int) interface{} {
	switch r.st.ColumnType(index) {
	case SQLITE_INTEGER:
		return r.st.ColumnInt64(index)
	case SQLITE_FLOAT:
		return r.st.ColumnDouble(index)
	case SQLITE_TEXT:
		return r.st.ColumnText(index)
	case SQLITE_BLOB:
		return r.st.ColumnBlob(index)
	case SQLITE_NULL:
		return nil
	default:
		panic("Invalid type" + r.st.ColumnType(index).String())
	}
	return 0
}
