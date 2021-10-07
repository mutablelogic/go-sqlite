package indexer

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"

	// Package imports
	walkfs "github.com/mutablelogic/go-sqlite/pkg/walkfs"
	notify "github.com/rjeczalik/notify"

	// Import namepaces
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Indexer struct {
	*walkfs.WalkFS
	queue *Queue
	name  string
	path  string
	walk  chan WalkFunc
}

// WalkFunc is called after a reindexing with any walk errors
type WalkFunc func(err error)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultCapacity = 1024
)

var (
	// Name for an index must be alphanumeric
	reIndexName = regexp.MustCompile(`^([A-Za-z0-9\_\-]+)$`)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new indexer with an identifier, path to the root of the indexer
// and a queue
func NewIndexer(name, path string, queue *Queue) (*Indexer, error) {
	this := new(Indexer)
	this.WalkFS = walkfs.New(this.visit)

	// Check path argument
	if stat, err := os.Stat(path); err != nil {
		return nil, err
	} else if !stat.IsDir() {
		return nil, ErrBadParameter.With("invalid path: ", strconv.Quote(path))
	} else if abspath, err := filepath.Abs(path); err != nil {
		return nil, err
	} else if !reIndexName.MatchString(name) {
		return nil, ErrBadParameter.With("invalid index name: ", strconv.Quote(name))
	} else {
		this.name = name
		this.path = abspath
	}

	// Check queue argument
	if queue == nil {
		this.queue = NewQueue()
	} else {
		this.queue = queue
	}

	// Channel to indicate we want to walk the index
	this.walk = make(chan WalkFunc)

	// Return success
	return this, nil
}

// run indexer, provider channel to receive errors
func (i *Indexer) Run(ctx context.Context, errs chan<- error) error {
	var walking sync.Mutex

	in := make(chan notify.EventInfo, defaultCapacity)
	if err := notify.Watch(filepath.Join(i.path, "..."), in, notify.Create, notify.Remove, notify.Write, notify.Rename); err != nil {
		senderr(errs, err)
		return err
	}

FOR_LOOP:
	for {
		// Dispatch events to index files and folders until context is cancelled
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case evt := <-in:
			if err := i.event(ctx, evt); err != nil {
				senderr(errs, err)
			}
		case fn := <-i.walk:
			walking.Lock()
			go func() {
				defer walking.Unlock()

				// Indicate reindexing is in progress
				i.queue.Mark(i.name, i.path, true)
				defer i.queue.Mark(i.name, i.path, false)

				// Start the walk and return any errors
				fn(i.WalkFS.Walk(ctx, i.path))
			}()
		}
	}

	// Stop notify and close channels
	notify.Stop(in)
	close(in)
	close(i.walk)

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (i *Indexer) String() string {
	str := "<indexer"
	if i.name != "" {
		str += fmt.Sprintf(" name=%q", i.name)
	}
	if i.path != "" {
		str += fmt.Sprintf(" path=%q", i.path)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

// Return name of the index
func (i *Indexer) Name() string {
	return i.name
}

// Return the absolute path of the index
func (i *Indexer) Path() string {
	return i.path
}

// Return the queue
func (i *Indexer) Queue() *Queue {
	return i.queue
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Walk will initiate a walk of the index, and block until context is
// cancelled or walk is started
func (i *Indexer) Walk(ctx context.Context, fn WalkFunc) error {
	if fn == nil {
		return ErrBadParameter.With("WalkFunc")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case i.walk <- fn:
		break
	}
	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// event is used to process an event from the notify
func (i *Indexer) event(ctx context.Context, evt notify.EventInfo) error {
	relpath, err := filepath.Rel(i.path, evt.Path())
	if err != nil {
		return err
	}
	switch evt.Event() {
	case notify.Create, notify.Write:
		info, err := os.Stat(evt.Path())
		if err == nil && info.Mode().IsRegular() && i.ShouldVisit(relpath, info) {
			i.queue.Add(i.name, relpath, info)
		}
	case notify.Remove, notify.Rename:
		info, err := os.Stat(evt.Path())
		if err == nil && info.Mode().IsRegular() && i.ShouldVisit(relpath, info) {
			i.queue.Add(i.name, relpath, info)
		} else {
			// Always attempt removal from index
			i.queue.Remove(i.name, relpath)
		}
	}
	// Return success
	return nil
}

// visit is used to index a file from the indexer
func (i *Indexer) visit(ctx context.Context, abspath, relpath string, info fs.FileInfo) error {
	if info.Mode().IsRegular() {
		i.queue.Add(i.name, relpath, info)
	}
	return nil
}

// senderr is used to send an error without blocking
func senderr(ch chan<- error, err error) {
	if ch != nil {
		select {
		case ch <- err:
			return
		default:
			// Channel blocked, ignore error
			return
		}
	}
}
