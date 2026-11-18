//go:build !client

package cmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"

	// Packages
	tui "github.com/mutablelogic/go-server/pkg/tui"
	schema "github.com/mutablelogic/go-sqlite/filesearch/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// indexing shows how far a volume is indexed, on standard error, as a bar
// which is drawn again, of the files which were indexed, or marked, of
// those of the volume, and the errors of files which can't be read,
// above it, until it's closed, when the bar is cleared
type indexing struct {
	sync.Mutex // Held while it's written
	w          io.Writer
	bar        interface {
		Write(io.Writer, string, float64) (int, error)
	}
	progress atomic.Pointer[schema.Progress]
	errs     atomic.Uint64
	stop     chan struct{}
	wg       sync.WaitGroup
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// indexingInterval is how often the progress is drawn
const indexingInterval = 100 * time.Millisecond

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// newIndexing returns the progress of indexing a volume, which is drawn for a
// terminal of the width, or, when standard error isn't a terminal, which is
// when the width is zero, only its errors are
func newIndexing(width int) *indexing {
	i := &indexing{w: os.Stderr, stop: make(chan struct{})}
	if width <= 0 {
		return i
	}
	i.bar = tui.Progress(tui.SetWidth(width))
	i.wg.Add(1)
	go func() {
		defer i.wg.Done()
		ticker := time.NewTicker(indexingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				i.Lock()
				i.draw()
				i.Unlock()
			case <-i.stop:
				return
			}
		}
	}()
	return i
}

// Close stops drawing the progress, and clears it
func (i *indexing) Close() {
	close(i.stop)
	i.wg.Wait()
	if i.bar != nil {
		fmt.Fprint(i.w, "\r\033[K")
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Progress is the progress function of a run, which keeps its progress, to
// be drawn, and writes the error of a file which can't be read, above it
func (i *indexing) Progress(p schema.Progress, err error) {
	i.progress.Store(&p)
	if err == nil || p.Finished != nil {
		return
	}
	i.errs.Add(1)
	i.Lock()
	defer i.Unlock()
	if i.bar != nil {
		fmt.Fprint(i.w, "\r\033[K")
	}
	fmt.Fprintln(i.w, "Error:", err)
	i.draw()
}

// Errors returns the number of files which couldn't be read
func (i *indexing) Errors() uint64 {
	return i.errs.Load()
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// draw draws the progress again, over what was drawn, of the files which
// were indexed, or marked, or couldn't be read, of the volume's, with the
// lock held
func (i *indexing) draw() {
	p := i.progress.Load()
	if i.bar == nil || p == nil {
		return
	}
	done := p.Indexed + p.Marked + i.errs.Load()
	percent := 100.0
	if p.Total > 0 {
		percent = float64(done) * 100 / float64(p.Total)
	}
	status := fmt.Sprintf("%d of %d files", done, p.Total)
	if p.Current != "" {
		status += " " + path.Base(p.Current)
	}
	fmt.Fprint(i.w, "\r\033[K")
	i.bar.Write(i.w, status, percent)
}
