package ncruces

import (
	"strconv"
	"strings"
	"sync"

	// Packages
	schema "github.com/mutablelogic/go-server/broadcaster/schema"
	sqlite "github.com/mutablelogic/go-sqlite"
	sqlite3 "github.com/ncruces/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// changes are the rows which the writer writes, of a transaction, which are
// published to the broadcaster once it's committed, and the writer is
// released, so a watcher which reads them reads what's committed. SQLite's
// hooks say when a row is written, and a transaction is committed, or rolled
// back. They don't say when a savepoint is rolled back to, so savepoint and
// rollbackTo are called around each.
type changes struct {
	sync.Mutex
	publisher sqlite.Publisher
	tables    map[string]bool // The tables which are published, or all of them
	kinds     map[string]bool // Whether a table is published, of its type: not a shadow table

	pending    []written       // Written, in the transaction
	committing []written       // Committed, unless the commit is rolled back
	committed  []schema.Change // Committed, and published when the writer is released
}

// written is a row which is written: how, and its database, table and rowid
type written struct {
	action schema.ChangeAction
	db     string
	table  string
	rowid  int64
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// newChanges returns the changes, of the broadcaster and tables of the
// options, or nil, without a broadcaster
func newChanges(o *sqlite.Options) *changes {
	if o.Broadcaster == nil {
		return nil
	}
	c := &changes{publisher: o.Broadcaster, kinds: make(map[string]bool)}
	if len(o.Tables) > 0 {
		c.tables = make(map[string]bool, len(o.Tables))
		for _, table := range o.Tables {
			c.tables[strings.ToLower(table)] = true
		}
	}
	return c
}

// hook sets the hooks of the writer, which record its changes. The
// pre-update hook, rather than the update hook, is called of each row which
// is deleted, of a DELETE without a WHERE, and of a row which an INSERT OR
// REPLACE replaces, which the update hook isn't.
func (c *changes) hook(db *sqlite3.Conn) {
	db.PreUpdateHook(c.preupdate)
	db.CommitHook(c.commit)
	db.RollbackHook(c.rollback)
}

////////////////////////////////////////////////////////////////////////////////
// HOOKS

// preupdate records a row which is written, of a table which is published.
// A row whose rowid is updated is deleted, of its old rowid, and created, of
// its new one.
func (c *changes) preupdate(data sqlite3.PreUpdateData) {
	if c.tables != nil && !c.tables[strings.ToLower(data.Table)] {
		return
	}
	var rows []written
	switch data.Op {
	case sqlite3.AUTH_INSERT:
		rows = append(rows, written{action: schema.ChangeCreate, rowid: data.NewRowID})
	case sqlite3.AUTH_UPDATE:
		if data.OldRowID == data.NewRowID {
			rows = append(rows, written{action: schema.ChangeUpdate, rowid: data.NewRowID})
		} else {
			rows = append(rows, written{action: schema.ChangeDelete, rowid: data.OldRowID}, written{action: schema.ChangeCreate, rowid: data.NewRowID})
		}
	case sqlite3.AUTH_DELETE:
		rows = append(rows, written{action: schema.ChangeDelete, rowid: data.OldRowID})
	default:
		return
	}
	c.Lock()
	defer c.Unlock()

	// A write after a commit is of another transaction, so the commit
	// succeeded
	c.promote()
	for _, row := range rows {
		row.db, row.table = data.Schema, data.Table
		c.pending = append(c.pending, row)
	}
}

// commit moves the transaction's rows to those which are committing, as
// the commit can still fail, and be rolled back
func (c *changes) commit() bool {
	c.Lock()
	defer c.Unlock()
	c.promote()
	c.committing, c.pending = merge(c.pending), nil
	return true
}

// rollback forgets the transaction's rows, and those of a commit which
// failed
func (c *changes) rollback() {
	c.Lock()
	defer c.Unlock()
	c.pending, c.committing = nil, nil
}

////////////////////////////////////////////////////////////////////////////////
// SAVEPOINTS

// savepoint returns the mark of a savepoint, which is rolled back to
func (c *changes) savepoint() int {
	if c == nil {
		return 0
	}
	c.Lock()
	defer c.Unlock()
	return len(c.pending)
}

// rollbackTo forgets the rows which were written after the savepoint's mark
func (c *changes) rollbackTo(mark int) {
	if c == nil {
		return
	}
	c.Lock()
	defer c.Unlock()
	if mark < len(c.pending) {
		c.pending = c.pending[:mark]
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLISH

// flush publishes what's committed, other than the shadow tables of virtual
// tables, of the writer, which is held, and isn't in a transaction
func (c *changes) flush(db *sqlite3.Conn) {
	if c == nil {
		return
	}
	c.Lock()
	c.promote()
	committed := c.committed
	c.committed = nil
	c.Unlock()
	for _, change := range committed {
		if c.published(db, change.Schema, change.Table) {
			c.publisher.Publish(change)
		}
	}
}

// published returns whether the changes of a table are published, which
// they are unless it's a shadow table, such as fts5's, or the table can't
// be found
func (c *changes) published(db *sqlite3.Conn, name, table string) bool {
	key := strings.ToLower(name + "." + table)
	if published, exists := c.kinds[key]; exists {
		return published
	}
	stmt, _, err := db.Prepare("SELECT type FROM pragma_table_list WHERE schema = ? AND name = ?")
	if err != nil {
		return true
	}
	defer stmt.Close()
	if stmt.BindText(1, name) != nil || stmt.BindText(2, table) != nil {
		return true
	}
	published := true
	if stmt.Step() {
		published = stmt.ColumnText(0) != "shadow"
	}
	if stmt.Err() != nil {
		return true
	}
	c.kinds[key] = published
	return published
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// promote moves the rows which are committing to those which are committed
func (c *changes) promote() {
	for _, change := range collapse(c.committing) {
		c.committed = append(c.committed, change)
	}
	c.committing = nil
}

// merge returns one row of each which is written more than once, in the
// order of their first writes: created, when it's created first, unless
// it's deleted last, when it's neither, or deleted, when it's deleted last,
// or else updated
func merge(rows []written) []written {
	type key struct {
		db, table string
		rowid     int64
	}
	type writes struct {
		first, last schema.ChangeAction
	}
	order := make([]key, 0, len(rows))
	seen := make(map[key]*writes, len(rows))
	for _, r := range rows {
		k := key{r.db, r.table, r.rowid}
		if w, exists := seen[k]; exists {
			w.last = r.action
		} else {
			seen[k] = &writes{first: r.action, last: r.action}
			order = append(order, k)
		}
	}
	result := make([]written, 0, len(order))
	for _, k := range order {
		w := seen[k]
		action := schema.ChangeUpdate
		switch {
		case w.first == schema.ChangeCreate && w.last == schema.ChangeDelete:
			continue
		case w.first == schema.ChangeCreate:
			action = schema.ChangeCreate
		case w.last == schema.ChangeDelete:
			action = schema.ChangeDelete
		}
		result = append(result, written{action: action, db: k.db, table: k.table, rowid: k.rowid})
	}
	return result
}

// collapse returns the changes of the rows, or, of a table of more than
// MaxChanges rows, one change of the table, without a row, whose action is
// theirs, when they're all the same, or else update
func collapse(rows []written) []schema.Change {
	type key struct{ db, table string }
	counts := make(map[key]int)
	actions := make(map[key]schema.ChangeAction)
	for _, r := range rows {
		k := key{r.db, r.table}
		counts[k]++
		if action, exists := actions[k]; !exists {
			actions[k] = r.action
		} else if action != r.action {
			actions[k] = schema.ChangeUpdate
		}
	}
	result := make([]schema.Change, 0, len(rows))
	tables := make(map[key]bool)
	for _, r := range rows {
		k := key{r.db, r.table}
		if counts[k] <= sqlite.MaxChanges {
			result = append(result, schema.Change{Action: r.action, Schema: r.db, Table: r.table, Row: strconv.FormatInt(r.rowid, 10)})
		} else if !tables[k] {
			tables[k] = true
			result = append(result, schema.Change{Action: actions[k], Schema: r.db, Table: r.table})
		}
	}
	return result
}
