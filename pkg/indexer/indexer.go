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
	name string
	path string
	walk chan struct{}
}

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
// and a channel to receive any errors
func NewIndexer(name, path string) (*Indexer, error) {
	this := new(Indexer)

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
		this.WalkFS = walkfs.New(this.visit)
		this.walk = make(chan struct{})
	}

	// Return success
	return this, nil
}

// run indexer
func (i *Indexer) Run(ctx context.Context, errs chan<- error) error {
	var walking sync.Mutex

	in := make(chan notify.EventInfo, defaultCapacity)
	if err := notify.Watch(filepath.Join(i.path, "..."), in, notify.Create, notify.Remove, notify.Write, notify.Rename); err != nil {
		fmt.Println("err", err)
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
		case <-i.walk:
			walking.Lock()
			go func() {
				defer walking.Unlock()
				if err := i.WalkFS.Walk(ctx, i.path); err != nil {
					senderr(errs, err)
				}
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

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Walk will initiate a walk of the index, and block until context is
// cancelled or walk is started
func (i *Indexer) Walk(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case i.walk <- struct{}{}:
		break
	}
	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// event is used to process an event from the notify
func (i *Indexer) event(ctx context.Context, evt notify.EventInfo) error {
	fmt.Println("event", evt)
	return nil
}

// visit is used to index a file from the indexer
func (i *Indexer) visit(ctx context.Context, abspath, relpath string, info fs.FileInfo) error {
	fmt.Println("visited", abspath, relpath, info.IsDir())
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

/*
// Return an indexer event or nil if no event should be sent
func (this *Indexer) process(e EventType, path string, info fs.FileInfo, out chan<- IndexerEvent, block bool) error {
	// Normalize the path
	relpath, err := filepath.Rel(this.path, path)
	if err != nil {
		return err
	} else {
		relpath = pathSeparator + relpath
	}

	// Deal with exclusions
	if e&EVENT_TYPE_ADDED > 0 {
		// Check for path exclusions
		for exclusion := range this.expath {
			if strings.HasPrefix(relpath, exclusion) {
				return nil
			}
		}
		// Check for extension exclusions
		if info != nil && info.Mode().IsRegular() {
			ext := strings.ToUpper(filepath.Ext(info.Name()))
			if _, exists := this.exext[ext]; exists {
				return nil
			}
		}
	}

	// Send event
	if block {
		out <- NewEvent(e, this.name, relpath, info)
	} else {
		select {
		case out <- NewEvent(e, this.name, relpath, info):
			// No-op
		default:
			return ErrChannelBlocked.With(this.name)
		}
	}

	// Return success
	return nil
}

// Translate notify types to internal types
func toEventType(e notify.Event, info fs.FileInfo) EventType {
	switch e {
	case notify.Create:
		if info != nil {
			return EVENT_TYPE_ADDED
		}
	case notify.Remove:
		return EVENT_TYPE_REMOVED
	case notify.Rename:
		if info != nil {
			return EVENT_TYPE_ADDED | EVENT_TYPE_RENAMED
		} else {
			return EVENT_TYPE_REMOVED | EVENT_TYPE_RENAMED
		}
	case notify.Write:
		if info != nil {
			return EVENT_TYPE_ADDED | EVENT_TYPE_CHANGED
		}
	}

	// Ignore unhandled cases
	return EVENT_TYPE_NONE
}
*/
