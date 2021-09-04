package indexer

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	// Modules
	notify "github.com/rjeczalik/notify"

	// Import namepaces
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Indexer struct {
	sync.Mutex
	name  string
	path  string
	state IndexerState
	delta time.Duration

	// Path and Extension exclusions
	exext  map[string]bool
	expath map[string]bool
}

type IndexerState uint

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	// Name for an index must be alphanumeric
	reIndexName = regexp.MustCompile(`^([A-Za-z0-9\_\-]+)$`)
)

const (
	IndexerStateIdle IndexerState = iota
	IndexerStateReindexing
	IndexerStateRunning
	IndexerStateSuspended
)

const (
	defaultDelta = time.Millisecond * 10
	minDelta     = time.Millisecond * 1
	maxDelta     = time.Second * 5
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewIndexer(name, path string) (*Indexer, error) {
	this := new(Indexer)
	this.exext = make(map[string]bool)
	this.expath = make(map[string]bool)

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
		this.state = IndexerStateIdle
		this.delta = defaultDelta
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Indexer) String() string {
	str := "<indexer"
	if this.name != "" {
		str += fmt.Sprintf(" name=%q", this.name)
	}
	if this.path != "" {
		str += fmt.Sprintf(" path=%q", this.path)
	}
	str += fmt.Sprint(" state=", this.State())
	return str + ">"
}

func (v IndexerState) String() string {
	switch v {
	case IndexerStateIdle:
		return "IndexerStateIdle"
	case IndexerStateReindexing:
		return "IndexerStateReindexing"
	case IndexerStateRunning:
		return "IndexerStateRunning"
	case IndexerStateSuspended:
		return "IndexerStateSuspended"
	default:
		return "[?? Invalid IndexerState value]"
	}
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

// Return name of the index
func (this *Indexer) Name() string {
	return this.name
}

// Return the absolute path of the index
func (this *Indexer) Path() string {
	return this.path
}

// Return the indexer state
func (this *Indexer) State() IndexerState {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.state
}

// Get how often rendering is performed
func (this *Indexer) Delta() time.Duration {
	return this.delta
}

// Set how often rendering is performed
func (this *Indexer) SetDelta(delta time.Duration) {
	this.delta = maxDuration(minDelta, delta)
}

// Set the indexer state
func (this *Indexer) setState(v IndexerState) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	if v != this.state {
		fmt.Println("SET STATE", v)
		this.state = v
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Exclude adds a path or file extension exclusion to the indexer.
// If it begins with a '.' then a file extension exlusion is added,
// If it begins with a '/' then a path extension exclusion is added.
// Path exclusions are case-sensitive, file extension exclusions are not.
func (this *Indexer) exclude(v string) error {
	if strings.HasPrefix(v, ".") && v != "." {
		v = strings.ToUpper(v)
		this.exext[v] = true
	} else if strings.HasPrefix(v, pathSeparator) && v != pathSeparator {
		v = pathSeparator + strings.Trim(v, pathSeparator)
		this.expath[v] = true
	} else {
		return ErrBadParameter.Withf("invalid exclusion: %q", v)
	}

	// Return success
	return nil
}

// run indexer
func (this *Indexer) run(ctx context.Context, out chan<- IndexerEvent, render chan<- *Indexer) error {
	in := make(chan notify.EventInfo, defaultCapacity)
	if err := notify.Watch(filepath.Join(this.path, "..."), in, notify.Create, notify.Remove, notify.Write, notify.Rename); err != nil {
		return err
	} else {
		this.setState(IndexerStateRunning)
	}

	// Add delta timer
	d := time.NewTimer(maxDuration(this.delta, maxDelta))
	defer d.Stop()

FOR_LOOP:
	for {
		// Dispatch events to index files and folders until context is cancelled
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case evt := <-in:
			// Ignore events if not in running state
			if this.State() != IndexerStateRunning {
				continue FOR_LOOP
			}
			// Get file information
			info, _ := os.Stat(evt.Path())
			if evttype := toEventType(evt.Event(), info); evttype != EVENT_TYPE_NONE {
				if err := this.process(evttype, evt.Path(), info, out, false); err != nil {
					select {
					case out <- NewError(this.name, err):
						// No-op
					default:
						// No-op
					}
				}
			}
		case <-d.C:
			// Ignore events if not in running state
			if this.State() != IndexerStateRunning {
				continue FOR_LOOP
			}
			// Ignore when channel is full
			select {
			case render <- this:
				// No-op
			default:
				// No-op
			}
			// Reset timer
			d.Reset(this.delta)
		}
	}

	// Stop notify and close channel
	notify.Stop(in)
	close(in)
	this.setState(IndexerStateIdle)

	// Return success
	return nil
}

// run re-indexer
func (this *Indexer) walk(ctx context.Context, out chan<- IndexerEvent) error {
	// Check incoming parameters
	if this.State() != IndexerStateReindexing {
		return ErrOutOfOrder.With("Reindex: indexer is not running")
	}

	// Walk filesystem
	err := filepath.WalkDir(this.path, func(path string, file fs.DirEntry, err error) error {
		// Propogate errors if they are cancel/timeout
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		// Ignore hidden files and folders
		if strings.HasPrefix(file.Name(), ".") {
			if file.IsDir() {
				return filepath.SkipDir
			}
			return err
		}
		// Process files which can be read
		if info, err := file.Info(); err == nil {
			this.process(EVENT_TYPE_ADDED|EVENT_TYPE_CHANGED, path, info, out, true)
		}
		// Return any context error
		return ctx.Err()
	})

	// Return errors unless they are cancel/timeout
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil
	} else {
		return err
	}
}

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
