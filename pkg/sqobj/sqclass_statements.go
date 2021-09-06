package sqobj

import (
	// Modules

	// Import Namespaces
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type sqpreparefunc func(*sqclass, SQFlag) SQStatement

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	statements = map[SQKey]sqpreparefunc{
		SQKeySelect: sqSelect,
		SQKeyInsert: sqInsert,
		SQKeyDelete: sqDelete,
	}
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - STATEMENTS

func sqSelect(class *sqclass, _ SQFlag) SQStatement {
	cols := make([]SQSource, len(class.col)+1)
	// first row is the rowid
	cols[0] = N("rowid")
	for i, col := range class.col {
		cols[i+1] = col.Col.WithAlias("")
	}
	return S(class.SQSource).To(cols...)
}

func sqInsert(class *sqclass, _ SQFlag) SQStatement {
	cols := make([]string, len(class.col))
	for i, col := range class.col {
		cols[i] = col.Col.Name()
	}
	return class.SQSource.Insert(cols...)
}

func sqDelete(class *sqclass, _ SQFlag) SQStatement {
	cols := make([]interface{}, 0, len(class.col))
	for _, col := range class.col {
		if col.Primary {
			cols = append(cols, Q(N(col.Name), "=", P))
		}
	}
	if len(cols) > 0 {
		return class.SQSource.Delete(cols...)
	} else {
		return nil
	}
}

/*

func (this *sqclass) addStatements(flags SQFlag) error {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()

	// Create statments for create,insert and delete
	this.addStatement(SQKeyCreate, this.sqCreate())
	this.addStatement(SQKeyWrite, this.sqInsert(flags))
	this.addStatement(SQKeyRead, this.sqSelect())

	// If we have primary keys, other operations are possible
	if len(this.PrimaryColumnNames()) > 0 {
		this.addStatement(SQKeyDelete, this.sqDelete())
		this.addStatement(SQKeyGetRowId, this.sqGetRowId())
	}

	// Create index statements
	for _, index := range this.indexes {
		this.addStatement(SQKeyCreate, this.sqIndex(index))
	}

	// Return success
	return nil
}

func (this *sqclass) addStatement(key SQKey, st SQStatement) {
	this.s[key] = append(this.s[key], st)
}

func (this *sqclass) sqCreate() SQTable {
	st := this.CreateTable(this.Columns()...).IfNotExists()
	for _, column := range this.columns {
		if column.Unique {
			st = st.WithUnique(column.Field.Name)
		} else if column.Index {
			st = st.WithIndex(column.Field.Name)
		}
	}
	return st
}

func (this *sqclass) sqIndex(index *sqindex) SQStatement {
	st := N(this.Name()+"_"+index.name).
		WithSchema(this.Schema()).
		CreateIndex(this.Name(), index.cols...).IfNotExists()
	if index.unique {
		st = st.WithUnique()
	}
	return st
}

func (this *sqclass) sqInsert(flags SQFlag) SQStatement {
	st := this.Insert(this.ColumnNames()...)

	// Add conflict resolution for any primary key field
	st = st.WithConflictUpdate(this.PrimaryColumnNames()...)

	// Add conflict resolution for any unique fields
	for _, column := range this.columns {
		if column.Unique && flags&SQLITE_FLAG_UPDATEONINSERT != 0 {
			st = st.WithConflictUpdate(column.Field.Name)
		}
	}

	// Add conflict for any unique indexes
	for _, index := range this.indexes {
		if index.unique && flags&SQLITE_FLAG_UPDATEONINSERT != 0 {
			st = st.WithConflictUpdate(index.cols...)
		}
	}

	// Return success
	return st
}

func (this *sqclass) sqDelete() SQStatement {
	expr := []interface{}{}
	for _, name := range this.PrimaryColumnNames() {
		expr = append(expr, Q(N(name), "=", P))
	}
	return this.Delete(expr...)
}

func (this *sqclass) sqGetRowId() SQStatement {
	expr := []interface{}{}
	for _, name := range this.PrimaryColumnNames() {
		expr = append(expr, Q(N(name), "=", P))
	}
	return S(this.SQSource).To(N("rowid")).Where(expr...)
}

func (this *sqclass) sqSelect() SQStatement {
	return S(this.SQSource).To(this.ColumnSources()...)
}
*/
