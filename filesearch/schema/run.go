package schema

import (
	"time"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// RunMeta is a run of the indexer, of a root, which is started
type RunMeta struct {
	Root    string
	Started time.Time
}

// RunFinish finishes a run, so it's complete, and its root can be swept
type RunFinish struct {
	Finished time.Time
}

// Run is a run of the indexer, of a root, which is complete once it's
// finished
type Run struct {
	ID       int64      `json:"id"`
	Root     string     `json:"root"`
	Started  time.Time  `json:"started"`
	Finished *time.Time `json:"finished,omitempty"`
}

// RunID selects a run, by its id, to get or finish it
type RunID int64

////////////////////////////////////////////////////////////////////////////////
// WRITER

// Insert binds a run, which is started
func (r RunMeta) Insert(bind *sqlite.Bind) (string, error) {
	root, err := absPath(r.Root)
	if err != nil {
		return "", err
	}
	bind.Set("root", root)
	bind.Set("started", r.Started.UTC())
	return bind.Query("filesearch.run_start"), nil
}

// Update isn't implemented: a run is finished, with RunFinish
func (r RunMeta) Update(*sqlite.Bind) error {
	return sqlite.ErrNotImplemented.With("a run is finished, with RunFinish")
}

// Insert isn't implemented: a run is started, with RunMeta
func (r RunFinish) Insert(*sqlite.Bind) (string, error) {
	return "", sqlite.ErrNotImplemented.With("a run is started, with RunMeta")
}

// Update binds when the run finished
func (r RunFinish) Update(bind *sqlite.Bind) error {
	bind.Set("finished", r.Finished.UTC())
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// READER

// Scan reads a run
func (r *Run) Scan(row sqlite.Row) error {
	return row.Scan(&r.ID, &r.Root, &r.Started, &r.Finished)
}

////////////////////////////////////////////////////////////////////////////////
// SELECTOR

// Select a run, by its id, to get or finish it
func (id RunID) Select(bind *sqlite.Bind, op sqlite.Op) (string, error) {
	bind.Set("id", int64(id))
	switch op {
	case sqlite.Get:
		return bind.Query("filesearch.run_get"), nil
	case sqlite.Update:
		return bind.Query("filesearch.run_finish"), nil
	}
	return "", sqlite.ErrNotImplemented.Withf("run: %v", op)
}
