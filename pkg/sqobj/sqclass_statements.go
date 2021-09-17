package sqobj

import (
	// Import Namespaces

	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type sqpreparefunc func(*Class, SQTransaction) SQStatement
type stkey uint

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	SQKeyNone stkey = iota
	SQKeySelect
	SQKeyInsert
	SQKeyDeleteRows
	SQKeyDeleteKeys
	SQKeyUpdateKeys
	SQKeyUpsertKeys
	SQKeyMax = SQKeyUpdateKeys
)

var (
	statements = map[stkey]sqpreparefunc{
		SQKeySelect:     sqSelect,
		SQKeyInsert:     sqInsert,
		SQKeyDeleteRows: sqDeleteRows,
		SQKeyDeleteKeys: sqDeleteKeys,
		SQKeyUpdateKeys: sqUpdateKeys,
		SQKeyUpsertKeys: sqUpsertKeys,
	}
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - STATEMENTS

func sqSelect(class *Class, _ SQTransaction) SQStatement {
	cols := make([]SQSource, len(class.col)+1)
	// first row is the rowid
	cols[0] = N("rowid")
	for i, col := range class.col {
		cols[i+1] = col.Col.WithAlias("")
	}
	return S(class.SQSource).To(cols...)
}

func sqInsert(class *Class, _ SQTransaction) SQStatement {
	cols := make([]string, len(class.col))
	for i, col := range class.col {
		cols[i] = col.Col.Name()
	}
	return class.SQSource.Insert(cols...)
}

func sqDeleteRows(class *Class, _ SQTransaction) SQStatement {
	return class.SQSource.Delete("rowid=?")
}

func sqDeleteKeys(class *Class, _ SQTransaction) SQStatement {
	cols := make([]interface{}, 0, len(class.col))
	for _, c := range class.col {
		if c.Primary {
			cols = append(cols, Q(N(c.Col.Name()), "=", P))
		}
	}
	return class.SQSource.Delete(cols...)
}

func sqUpdateKeys(class *Class, _ SQTransaction) SQStatement {
	values := make([]string, 0, len(class.col))
	keys := make([]interface{}, 0, len(class.col))
	for _, c := range class.col {
		if !c.Primary {
			values = append(values, c.Col.Name())
		} else {
			keys = append(keys, Q(N(c.Col.Name()), "=", P))
		}
	}
	return class.SQSource.Update(values...).Where(keys...)
}

func sqUpsertKeys(class *Class, _ SQTransaction) SQStatement {
	cols := make([]string, len(class.col))
	keys := make([]string, 0, len(class.col))
	for i, col := range class.col {
		cols[i] = col.Col.Name()
		if col.Primary {
			keys = append(keys, col.Col.Name())
		}
	}
	return class.SQSource.Insert(cols...).WithConflictUpdate(keys...)
}
