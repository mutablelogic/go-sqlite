package indexer

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	// Modules
	sqobj "github.com/djthorpe/go-sqlite/pkg/sqobj"
	"github.com/hashicorp/go-multierror"

	// Import namepaces
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Manager struct {
	SQObjects

	// Channels for operations
	add     chan *Indexer
	reindex chan *Indexer
	render  chan *Indexer

	// Statements
	file            SQClass
	keymark         SQKey
	keyunmark       SQKey
	keyrenderselect SQKey
	keyrenderlock   SQKey

	// Period by which new files are indexed
	delta time.Duration
}

type RenderFunc func(context.Context, IndexerEvent) error

/*
type Doc struct {
	File        int64     `sqlite:"file,primary"` // references File.RowID
	Title       string    `sqlite:"title,not null"`
	Description string    `sqlite:"description"`
	Shortform   string    `sqlite:"shortform"`
	IndexTime   time.Time `sqlite:"idxtime"`
}

type Tag struct {
	Doc   int64  `sqlite:"doc,index:doc"` // references Doc.RowID
	Value string `sqlite:"doc,not null,index:tag"`
}

type Meta struct {
	Doc   int64  `sqlite:"doc,index:doc,unique:docmeta"` // references Doc.RowID
	Name  string `sqlite:"name,not null,index:meta,unique:docmeta"`
	Value string `sqlite:"value"`
}

type Query struct {
	Index []string `json:"index"`
}
*/

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	tableNameFile   = "file"
	pathSeparator   = string(os.PathSeparator)
	defaultCapacity = 100
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new index manager object with a sqlite connection.
// Add SQLITE_FLAG_DELETEIFEXISTS to clear out existing tables
func NewManager(conn SQConnection, schema string, flags SQFlag) (*Manager, error) {
	this := new(Manager)

	// Check arguments
	if conn == nil || schema == "" {
		return nil, ErrBadParameter.With("Invalid connnection or schema")
	} else if conn, err := sqobj.With(conn, schema); err != nil {
		return nil, err
	} else {
		this.SQObjects = conn
	}

	// Register and create files table
	if class, err := this.Register(tableNameFile, &File{}); err != nil {
		return nil, err
	} else if err := this.Create(class, flags); err != nil {
		return nil, err
	} else {
		this.file = class
		this.keymark = class.Set(class.Update("mark").
			Where(Q(N("index"), "=", P)))
		this.keyunmark = class.Set(class.Delete(Q(N("mark"), "=", P), Q(N("index"), "=", P)))
		this.keyrenderselect = class.Set(S(class).To(N("rowid")).Where(Q(N("idxtime"), " IS NOT NULL"), Q(N("mark"), "=", false), Q(N("index"), "=", P)).Order(N("idxtime")).WithLimitOffset(1, 0))
		this.keyrenderlock = class.Set(class.Update("idxtime").
			Where(Q(N("index"), "=", P), Q(N("rowid"), "=", P)))
	}

	// Make indexer channel for adding new indexers
	this.add, this.reindex, this.render = make(chan *Indexer), make(chan *Indexer), make(chan *Indexer)

	// Return success
	return this, nil
}

func (this *Manager) Run(ctx context.Context, fn RenderFunc) error {
	var result error
	var wg sync.WaitGroup

	// Channel for accepting events
	c := make(chan IndexerEvent, defaultCapacity)

FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case indexer := <-this.reindex:
			wg.Add(1)
			go func() {
				defer wg.Done()

				// Bomb out if not in correct state
				if indexer.State() != IndexerStateRunning {
					return
				}

				// Set state and reset it when finished
				indexer.setState(IndexerStateReindexing)
				defer indexer.setState(IndexerStateRunning)

				// Mark and walk
				if err := this.mark(indexer, true); err != nil {
					result = multierror.Append(result, err)
				} else if err := indexer.walk(ctx, c); err != nil {
					result = multierror.Append(result, err)
				}

				// Wait for all events to be processed, then unmark
				<-time.After(time.Second)
				if err := this.mark(indexer, false); err != nil {
					result = multierror.Append(result, err)
				}
			}()
		case indexer := <-this.render:
			if rowid, err := this.renderselect(indexer); err != nil {
				result = multierror.Append(result, err)
			} else if rowid > 0 {
				if err := this.renderexec(ctx, indexer, rowid, fn); err != nil {
					result = multierror.Append(result, err)
				}
			}
		case indexer := <-this.add:
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := indexer.run(ctx, c, this.render); err != nil {
					result = multierror.Append(result, err)
				}
			}()
		case evt := <-c:
			if err := this.process(evt); err != nil {
				fmt.Println(err)
			}
		}
	}

	// Wait for all indexers to finish
	wg.Wait()

	// Close message channels
	close(c)
	close(this.add)
	close(this.reindex)
	close(this.render)

	// Return any errors
	return result
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Manager) String() string {
	str := "<indexer.manager"
	str += fmt.Sprint(" ", this.SQObjects)
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// NewIndexer adds a file indexer and initiates walking the file tree
func (this *Manager) NewIndexer(name, path string) (*Indexer, error) {
	// Create an indexer object, send to process
	indexer, err := NewIndexer(name, path)
	if err != nil {
		return nil, err
	} else {
		this.add <- indexer
	}

	// Return success
	return indexer, nil
}

// Exclude adds a path or file extension exclusion to an index, removing any existing
// files which match these exclusions. If the exclusion begins with a '.' then a
// file extension exlusion is added, if it begins with a '/' then a path prefix
// exclusion is added. Path prefix exclusions are case-sensitive,
// file extension exclusions are not.
func (this *Manager) Exclude(idx *Indexer, exclusion string) error {
	if idx == nil || exclusion == "" {
		return ErrBadParameter.With("Exclude")
	}
	return idx.exclude(exclusion)
}

// Reindex starts a reindex process to pick up new documents and delete old
// ones which are not in the index. A reindexing can only start when the
// indexer is in running state
func (this *Manager) Reindex(indexer *Indexer) error {
	if indexer == nil {
		return ErrBadParameter.With("Reindex")
	} else if state := indexer.State(); state != IndexerStateRunning {
		return ErrOutOfOrder.With("Reindex in state: ", state)
	} else {
		this.reindex <- indexer
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// retrieve a document for rendering
func (this *Manager) renderselect(indexer *Indexer) (int64, error) {
	rs, err := this.Query(this.file.Get(this.keyrenderselect)[0], indexer.Name())
	if err != nil {
		return 0, err
	}
	defer rs.Close()
	if row := rs.NextArray(); row != nil {
		return row[0].(int64), nil
	} else {
		return 0, nil
	}
}

// render a document
func (this *Manager) renderexec(ctx context.Context, indexer *Indexer, rowid int64, fn RenderFunc) error {
	return this.Do(func(txn SQTransaction) error {
		r, err := txn.Exec(this.file.Get(this.keyrenderlock)[0], nil, indexer.Name(), rowid)
		if err != nil {
			return err
		}
		if r.RowsAffected == 0 {
			fmt.Println("no rows affected", rowid)
			return nil
		}
		fmt.Println("locked", rowid)

		// Execute the rendering function, if error rollback occurs
		return fn(ctx, nil)
	})
}

// process events
func (this *Manager) process(evt IndexerEvent) error {
	switch evt.Type() {
	case EVENT_TYPE_ADDED, EVENT_TYPE_CHANGED:
		if rowid, file, err := this.processadd(evt); err != nil {
			return err
		} else if rowid != 0 {
			fmt.Println("Add:", rowid, file)
		}
	case EVENT_TYPE_REMOVED:
		if rowid, file, err := this.processremove(evt); err != nil {
			return err
		} else if rowid != 0 {
			fmt.Println("Del:", rowid, file)
		}
	case EVENT_TYPE_RENAMED:
		if evt.FileInfo() != nil {
			if rowid, file, err := this.processadd(evt); err != nil {
				return err
			} else if rowid != 0 {
				fmt.Println("Add:", rowid, file)
			}
		} else {
			if rowid, file, err := this.processremove(evt); err != nil {
				return err
			} else if rowid != 0 {
				fmt.Println("Del:", rowid, file)
			}
		}
	}

	// Return success
	return nil
}

// mark database as dirty or clean (and delete dirty)
func (this *Manager) mark(indexer *Indexer, value bool) error {
	if indexer == nil {
		return ErrBadParameter.With("Mark")
	}
	return this.Do(func(txn SQTransaction) error {
		if !value {
			_, err := txn.Exec(this.file.Get(this.keyunmark)[0], !value, indexer.Name())
			if err != nil {
				return err
			}
		} else {
			_, err := txn.Exec(this.file.Get(this.keymark)[0], value, indexer.Name())
			if err != nil {
				return err
			}
		}
		// Return success
		return nil
	})
}

func (this *Manager) processadd(evt IndexerEvent) (int64, *File, error) {
	file := NewFile(evt, true)
	r, err := this.Write(file)
	if err != nil {
		return 0, nil, err
	}
	// If a new record was created, report
	if r[0].RowsAffected != 0 {
		return r[0].LastInsertId, file, nil
	} else {
		return 0, nil, nil
	}
}

func (this *Manager) processremove(evt IndexerEvent) (int64, *File, error) {
	file := NewFile(evt, true)
	r, err := this.Delete(file)
	if err != nil {
		return 0, nil, err
	}
	if r[0].RowsAffected != 0 {
		return r[0].LastInsertId, file, nil
	} else {
		return 0, nil, nil
	}
}
