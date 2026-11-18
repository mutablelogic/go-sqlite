package filesearch

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	types "github.com/mutablelogic/go-server/pkg/types"
	sqlite "github.com/mutablelogic/go-sqlite"
	schema "github.com/mutablelogic/go-sqlite/filesearch/schema"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// ProgressFn is called as a volume is indexed: with its progress, after each
// file is indexed, or marked, and with an error, when a file can't be read,
// which isn't indexed, and once it's finished, with its error, when it
// failed. It's called as the volume is walked, so it delays it.
type ProgressFn func(schema.Progress, error)

// progress is how far a run of a volume has indexed it, which the run
// changes, and Progress reads, and which is said to the run's function
type progress struct {
	sync.Mutex
	schema.Progress
	fn ProgressFn
}

// event is what happened to a file, of a batch, which is counted, and said,
// once the batch is committed: it was indexed, or marked, or it couldn't be
type event struct {
	path    string
	indexed bool
	err     error
}

// file is a file which the walk of a volume found
type file struct {
	path    string // Within the volume, with / separators
	size    int64
	modtime time.Time
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// batchSize is the most files which are indexed, or marked, in one
	// transaction, so searches, on the readers, see them as it indexes
	batchSize = 100
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// IndexVolume indexes a volume, of a folder, which is added, when it isn't
// one already. Its run walks it, of its folders, when the request is
// recursive, and
// marks each file which is as it was, of its size and when it was
// modified, and indexes those which are new, or changed. Once it's indexed,
// the objects which the run didn't see are swept. A file which can't be
// read isn't indexed, and its error is said to the function.
//
// The function, when it isn't nil, is called with its progress, and the
// errors of files which can't be read, and when it's finished.
//
// It returns the volume, which is indexed, or failed, with its error, such
// as when the context is done. A volume which this manager is indexing is
// ErrConflict.
func (m *Manager) IndexVolume(ctx context.Context, req schema.IndexVolumeRequest, fn ProgressFn) (_ *schema.Volume, err error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "filesearch.IndexVolume",
		attribute.String("arg.req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()

	// The folder, which is a volume's, by its absolute path, of the folder
	// which a symbolic link is to, as it's walked
	root, err := filepath.Abs(req.Path)
	if err != nil {
		return nil, sqlite.ErrBadParameter.Withf("path: %v", err)
	}
	if root, err = filepath.EvalSymlinks(root); err != nil {
		return nil, sqlite.ErrBadParameter.Withf("path: %v", err)
	}
	if info, err := os.Stat(root); err != nil {
		return nil, sqlite.ErrBadParameter.Withf("path: %v", err)
	} else if !info.IsDir() {
		return nil, sqlite.ErrBadParameter.Withf("path isn't a folder: %q", root)
	}

	// The volume, which isn't indexing, of this manager, and its run
	p, err := m.startProgress(root, fn)
	if err != nil {
		return nil, err
	}
	var volume schema.Volume
	if err := m.conn.Insert(ctx, &volume, schema.VolumeMeta{Path: root}); err != nil {
		m.endProgress(root, nil, nil)
		return nil, err
	}
	if err := m.conn.Update(ctx, &volume, schema.VolumeID(volume.ID), schema.VolumeStart{Started: time.Now()}); err != nil {
		m.endProgress(root, nil, nil)
		return nil, err
	}
	p.start(volume)

	// Walk it, and then, when it's walked, finish its run, as indexed, and
	// sweep what it didn't see, in one transaction, so it's indexed only once
	// it's swept
	err = m.walk(ctx, &volume, req.Recursive, p)
	finished := time.Now()
	var swept schema.ObjectList
	if err == nil {
		err = m.conn.Tx(ctx, func(conn sqlite.Conn) error {
			swept = schema.ObjectList{}
			if err := conn.Update(ctx, &volume, schema.VolumeID(volume.ID), schema.VolumeFinish{Finished: finished}); err != nil {
				return err
			}
			if err := conn.Delete(ctx, &swept, schema.ObjectSweep{Volume: volume.ID, Run: volume.Run}); err != nil && !errors.Is(err, sqlite.ErrNotFound) {
				return err
			}
			return nil
		})
	}

	// Or, when it wasn't, finish its run, without being cancelled, as failed,
	// with its error
	if err != nil {
		if finishErr := m.conn.Update(context.WithoutCancel(ctx), &volume, schema.VolumeID(volume.ID), schema.VolumeFinish{Finished: finished, Error: err}); finishErr != nil {
			err = errors.Join(err, finishErr)
			m.endProgress(root, &finished, err)
			return nil, err
		}
		m.endProgress(root, &finished, err)
		return &volume, err
	}
	p.Lock()
	p.Swept = uint64(len(swept.Body))
	p.Unlock()

	// Return success
	m.endProgress(root, &finished, nil)
	return &volume, nil
}

// Progress returns how far the run of a volume has indexed it, while it's
// indexing, or how the last run of this manager did, once it's finished,
// which is ErrNotFound when this manager hasn't indexed it
func (m *Manager) Progress(volume int64) (schema.Progress, error) {
	m.Lock()
	defer m.Unlock()
	for _, p := range m.progress {
		p.Lock()
		result := p.Progress
		p.Unlock()
		if result.Volume == volume {
			return result, nil
		}
	}
	return schema.Progress{}, sqlite.ErrNotFound.Withf("volume %d hasn't been indexed", volume)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// startProgress returns the progress of a volume, by its path, which starts,
// and is said to the function, or ErrConflict when it's indexing
func (m *Manager) startProgress(root string, fn ProgressFn) (*progress, error) {
	m.Lock()
	defer m.Unlock()
	if p, exists := m.progress[root]; exists {
		p.Lock()
		indexing := p.Finished == nil
		p.Unlock()
		if indexing {
			return nil, sqlite.ErrConflict.Withf("volume is indexing: %q", root)
		}
	}
	p := &progress{Progress: schema.Progress{Path: root}, fn: fn}
	m.progress[root] = p
	return p, nil
}

// endProgress ends the progress of a volume, which is kept, and said to its
// function, with its error, once it's finished, or removed, when it didn't
// start
func (m *Manager) endProgress(root string, finished *time.Time, err error) {
	m.Lock()
	p, exists := m.progress[root]
	if exists && finished == nil {
		delete(m.progress, root)
	}
	m.Unlock()
	if !exists || finished == nil {
		return
	}
	p.Lock()
	p.Finished = finished
	p.Current = ""
	p.Unlock()
	p.say(err)
}

// walk walks the volume, of its folders, when it's recursive, and indexes,
// or marks, its files, a batch at a time
func (m *Manager) walk(ctx context.Context, volume *schema.Volume, recursive bool, p *progress) error {
	batch := make([]file, 0, batchSize)
	// A batch is indexed in a transaction, and then, once it's committed, what
	// happened to each of its files is counted, and said, so the function isn't
	// called while the transaction holds the writer, and a batch which is
	// rolled back isn't counted
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		var events []event
		err := m.conn.Bulk(ctx, func(conn sqlite.Conn) error {
			events = make([]event, 0, len(batch))
			for _, f := range batch {
				e, err := m.indexFile(ctx, conn, volume, f)
				if err != nil {
					return err
				}
				events = append(events, e)
			}
			return nil
		})
		batch = batch[:0]
		if err != nil {
			return err
		}
		for _, e := range events {
			p.apply(e)
		}
		return nil
	}
	if err := filepath.WalkDir(volume.Path, func(path string, d fs.DirEntry, err error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		rel := relPath(volume.Path, path)

		// A file, or a folder, which can't be read, is said, and isn't indexed,
		// unless it's the volume's, which fails its run
		if err != nil {
			if path == volume.Path {
				return err
			}
			p.fail(rel, err)
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Its folders, when it's recursive, and its files, which are regular
		switch {
		case d.IsDir():
			if path != volume.Path && !recursive {
				return fs.SkipDir
			}
			return nil
		case !d.Type().IsRegular():
			return nil
		}
		info, err := d.Info()
		if err != nil {
			p.fail(rel, err)
			return nil
		}
		p.seen(rel)
		batch = append(batch, file{path: rel, size: info.Size(), modtime: info.ModTime()})
		if len(batch) >= batchSize {
			return flush()
		}
		return nil
	}); err != nil {
		return err
	}
	return flush()
}

// indexFile marks a file, which is as it was, or indexes it, when it's new,
// or changed, and returns which. A file whose path, or metadata, isn't, isn't
// indexed, and is the event's error, and the database's errors are its own.
func (m *Manager) indexFile(ctx context.Context, conn sqlite.Conn, volume *schema.Volume, f file) (event, error) {
	var object schema.Object
	err := conn.Update(ctx, &object, schema.ObjectMark{Volume: volume.ID, Path: f.path, Size: f.size, ModTime: f.modtime, Run: volume.Run}, nil)
	switch {
	case err == nil:
		return event{path: f.path}, nil
	case errors.Is(err, sqlite.ErrBadParameter):
		return event{path: f.path, err: err}, nil
	case !errors.Is(err, sqlite.ErrNotFound):
		return event{}, err
	}
	if err := conn.Insert(ctx, &object, schema.ObjectMeta{
		Volume:  volume.ID,
		Path:    f.path,
		Size:    f.size,
		ModTime: f.modtime,
		Type:    contentType(f.path),
		Title:   title(f.path),
		Run:     volume.Run,
	}); errors.Is(err, sqlite.ErrBadParameter) {
		return event{path: f.path, err: err}, nil
	} else if err != nil {
		return event{}, err
	}
	return event{path: f.path, indexed: true}, nil
}

// relPath returns the path, within the volume, with / separators
func relPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

// contentType returns the content type of a file, of its extension, without
// its parameters, or nothing
func contentType(path string) string {
	contentType := mime.TypeByExtension(filepath.Ext(path))
	if mediatype, _, err := mime.ParseMediaType(contentType); err == nil {
		return mediatype
	}
	return ""
}

// title returns the title of a file, of its name, without its extension
func title(path string) string {
	name := filepath.Base(path)
	if title := strings.TrimSuffix(name, filepath.Ext(name)); title != "" {
		return title
	}
	return name
}

////////////////////////////////////////////////////////////////////////////////
// PROGRESS

// start starts the progress of the volume's run
func (p *progress) start(volume schema.Volume) {
	p.Lock()
	defer p.Unlock()
	p.Volume, p.Run = volume.ID, volume.Run
	if volume.Started != nil {
		p.Started = *volume.Started
	}
}

// seen counts a file, which is being indexed
func (p *progress) seen(path string) {
	p.Lock()
	defer p.Unlock()
	p.Files++
	p.Current = path
}

// apply counts what happened to a file, of a batch which was committed: it
// was indexed, or marked, or it couldn't be, and says so, of the file
func (p *progress) apply(e event) {
	p.Lock()
	p.Current = e.path
	switch {
	case e.err != nil:
	case e.indexed:
		p.Indexed++
	default:
		p.Marked++
	}
	p.Unlock()
	if e.err != nil {
		p.fail(e.path, e.err)
	} else {
		p.say(nil)
	}
}

// fail says a file can't be read, of its error
func (p *progress) fail(path string, err error) {
	p.say(fmt.Errorf("%s: %w", path, err))
}

// say calls the function, when there is one, with the progress, and the
// error, without the lock held
func (p *progress) say(err error) {
	if p.fn == nil {
		return
	}
	p.Lock()
	progress := p.Progress
	p.Unlock()
	p.fn(progress, err)
}
