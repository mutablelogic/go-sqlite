/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package fsindexer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	tasks "github.com/djthorpe/gopi/util/tasks"
	sq "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Indexer struct {
	Root    string
	Objects sq.Objects
}

type indexer struct {
	log     gopi.Logger
	root    string
	sqobj   sq.Objects
	indexer chan task
	deleter chan int64
	stop    bool
	jobs    map[int64]*job
	tasks.Tasks
	sync.Mutex
	sync.WaitGroup
}

// Job represents indexing jobs scheduled
type job struct {
	jobnode int64
	relpath string
	count   uint64
	done    bool
	sync.Mutex
}

// Task represents tasks queued for indexing
type task struct {
	jobnode, inode int64
	relpath        string
	info           os.FileInfo
}

// File represents a file in the database
type File struct {
	Id       int64  `sql:"inode,primary"`
	RootPath string `sql:"root,primary"`
	Job      int64  `sql:"idx"`
	RelPath  string `sql:"relpath"`
	Name     string `sql:"name"`
	Ext      string `sql:"ext,nullable"`
	Size     int64  `sql:"size"`
	MimeType string `sql:"mimetype,nullable"`
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Indexer) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<fsindexer.Open>{ config=%+v }", config)

	this := new(indexer)
	this.log = logger
	this.indexer = make(chan task)
	this.deleter = make(chan int64)
	this.jobs = make(map[int64]*job)

	if config.Objects == nil {
		return nil, gopi.ErrBadParameter
	} else {
		this.sqobj = config.Objects
	}

	if config.Root == "" {
		if root, err := os.Getwd(); err != nil {
			return nil, fmt.Errorf("%w: Invalid working directory", err)
		} else {
			this.root = root
		}
	} else if stat, err := os.Stat(config.Root); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: %v", err, strconv.Quote(config.Root))
	} else if stat.IsDir() == false {
		return nil, fmt.Errorf("%w: %v", gopi.ErrBadParameter, strconv.Quote(config.Root))
	} else {
		this.root = filepath.Clean(config.Root)
	}

	// Register File object with database
	if _, err := this.sqobj.RegisterStruct(&File{}); err != nil {
		return nil, err
	}

	// Start background tasks
	this.Tasks.Start(this.IndexTask, this.ReportTask)

	// Success
	return this, nil
}

func (this *indexer) Close() error {
	this.log.Debug("<fsindexer.Close>{ root=%v obj=%v }", strconv.Quote(this.root), this.sqobj)

	// Set stop = true to stop indexers
	this.stop = true

	// Wait until fs_walk tasks have ended
	this.WaitGroup.Wait()

	// Close any tasks
	if err := this.Tasks.Close(); err != nil {
		return err
	}

	// Close channels
	close(this.indexer)
	close(this.deleter)

	// Release resources
	this.jobs = nil
	this.indexer = nil
	this.deleter = nil
	this.sqobj = nil

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// INDEX IMPLEMENTATION

func (this *indexer) AddIndex(path string, watch bool) (int64, error) {
	this.log.Debug2("<fsindexer.Index>{ path=%v watch=%v }", strconv.Quote(path), watch)

	// Watch is not yet implemented
	if watch == true {
		return 0, fmt.Errorf("%w: watch argument is not yet implemented", gopi.ErrNotImplemented)
	}

	// Path is relative to the base, ensure under the root path and is a directory
	if path_ := filepath.Clean(filepath.Join(this.root, path)); strings.HasPrefix(path_, this.root) == false {
		return 0, fmt.Errorf("%w: Path is not under the root, %v", gopi.ErrBadParameter, strconv.Quote(path))
	} else if stat, err := os.Stat(path_); os.IsNotExist(err) {
		return 0, fmt.Errorf("%w: Path does not exist, %v", gopi.ErrBadParameter, strconv.Quote(path))
	} else if err != nil {
		return 0, err
	} else if stat.IsDir() == false {
		return 0, fmt.Errorf("%w: Path is not a folder, %v", gopi.ErrBadParameter, strconv.Quote(path))
	} else if inode := inodeForInfo(stat); inode == 0 {
		return 0, gopi.ErrAppError
	} else if relpath := strings.TrimPrefix(path_, this.root); relpath == path_ {
		return 0, gopi.ErrAppError
	} else if err := this.addJob(inode, strings.Trim(relpath, string(filepath.Separator))); err != nil {
		return 0, err
	} else {
		// TODO: We assume this is on a single volume so we need to add the volume to the ID
		return inode, nil
	}
}

func (this *indexer) DeleteIndexById(index int64) error {
	this.log.Debug2("<fsindexer.DeleteIndexById>{ index=%v }", index)

	if _, exists := this.jobs[index]; exists == false {
		return fmt.Errorf("%w: Index not found", gopi.ErrNotFound)
	} else {
		this.deleter <- index
	}
	// Success
	return nil
}

/*
		if class := this.sqobj.ClassFor(&File{}); class == nil {
		return gopi.ErrAppError
	} else {
		// Perform delete
		lang := this.sqobj.Lang()
		st := lang.NewDelete(class.(sq.StructClass).TableName()).Where(lang.Equals("idx", lang.Value(job.Id())))
		this.log.Info("<fsindexer.DeleteById>{ index=%v } %v", job, st.Query())
		return gopi.ErrNotImplemented
	}

}
*/

func (this *indexer) ReindexById(index int64) error {
	this.log.Debug2("<fsindexer.ReindexById>{ index=%v }", index)

	if job, exists := this.jobs[index]; exists == false {
		return fmt.Errorf("%w: Index not found", gopi.ErrNotFound)
	} else {
		this.log.Debug2("<fsindexer.ReindexById>{ index=%v } TODO", job)
		return gopi.ErrNotImplemented
	}

}

func (this *indexer) Indexes() []sq.FSIndex {
	indexes := make([]sq.FSIndex, 0, len(this.jobs))
	for _, job := range this.jobs {
		indexes = append(indexes, job)
	}
	return indexes
}

func (this *indexer) IndexById(id int64) sq.FSIndex {
	if job, exists := this.jobs[id]; exists {
		return job
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *indexer) String() string {
	return fmt.Sprintf("<fsindexer>{ root=%v indexes=%v }", strconv.Quote(this.root), this.jobs)
}

////////////////////////////////////////////////////////////////////////////////
// WALK FILE SYSTEM

func (this *indexer) fs_walk(jobnode int64) error {
	this.log.Debug2("<fsindexer.fs_walk>{ jobnode=%v }", jobnode)
	// Check incoming parameters
	if job, exists := this.jobs[jobnode]; exists == false {
		return fmt.Errorf("%w: No such job", gopi.ErrNotFound)
	} else {
		// Perform the walk
		path := filepath.Join(this.root, this.jobs[jobnode].relpath)

		// Add to the waitgroup
		this.WaitGroup.Add(1)
		defer this.WaitGroup.Done()

		// Walk
		err := filepath.Walk(path, func(abspath string, info os.FileInfo, err error) error {
			// If stop is true, then return DeadlineExceeded
			if errors.Is(err, gopi.ErrDeadlineExceeded) || this.stop {
				return gopi.ErrDeadlineExceeded
			}
			// If not readable or executable folder, then ignore
			if info.IsDir() && isReadableFileAtPath(path) != nil {
				return filepath.SkipDir
			}
			if info.IsDir() && isWritableFileAtPath(path) != nil {
				return filepath.SkipDir
			}
			if info.IsDir() && isExecutableFileAtPath(path) != nil {
				return filepath.SkipDir
			}
			// Return any other errors
			if err != nil {
				if strings.HasSuffix(err.Error(), "operation not permitted") {
					return nil
				} else {
					return err
				}
			}
			if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
				// Ignore hidden folders
				return filepath.SkipDir
			} else if info.Mode().IsRegular() == false {
				// Ignore files which aren't regular files
			} else if info.Size() == 0 {
				// Ignore zero-sized files
			} else if strings.HasPrefix(info.Name(), ".") {
				// Ignore hidden files
			} else if inode := inodeForInfo(info); inode == 0 {
				// inode not obtained
			} else if err := this.addTask(jobnode, inode, abspath, info); err != nil {
				// addTask for indexing - we don't deal with any errors here, just warn
				this.log.Warn("fs_walk: %v", err)
			}
			// Return success - continue
			return nil
		})

		// Indicate the job is completed
		job.Lock()
		defer job.Unlock()
		job.done = true

		// Return any error condition
		return err
	}
}

////////////////////////////////////////////////////////////////////////////////
// JOBS AND TASKS

func (this *indexer) addJob(jobnode int64, relpath string) error {
	this.log.Debug2("<fsindexer.addJob>{ jobnode=%v relpath=%v }", jobnode, strconv.Quote(relpath))

	// Lock addJob to ensure one job added at a time
	this.Lock()
	defer this.Unlock()

	// Check incoming parameters
	if jobnode == 0 {
		return fmt.Errorf("%w: Invalid inode parameter: %v", gopi.ErrBadParameter, strconv.Quote(relpath))
	} else if _, exists := this.jobs[jobnode]; exists {
		return fmt.Errorf("%w: Duplicate job: %v", gopi.ErrBadParameter, strconv.Quote(relpath))
	} else if stat, err := os.Stat(filepath.Join(this.root, relpath)); os.IsNotExist(err) {
		return fmt.Errorf("%w: Invalid job path: %v", gopi.ErrBadParameter, strconv.Quote(relpath))
	} else if err != nil {
		return fmt.Errorf("%w: Invalid job path: %v", err, strconv.Quote(relpath))
	} else if stat.IsDir() == false {
		return fmt.Errorf("%w: Not a folder: %v", gopi.ErrBadParameter, strconv.Quote(relpath))
	}

	// Add job and start walk in background
	this.jobs[jobnode] = &job{jobnode: jobnode, relpath: relpath}
	go this.fs_walk(jobnode)

	// Return success
	return nil
}

func (this *indexer) addTask(jobnode, inode int64, abspath string, info os.FileInfo) error {
	this.log.Debug2("<fsindexer.addJob>{ inode=%v abspath=%v }", inode, strconv.Quote(abspath))
	// Check incoming parameters
	if job, exists := this.jobs[jobnode]; exists == false {
		return fmt.Errorf("%w: No such job", gopi.ErrNotFound)
	} else if inode == 0 {
		return fmt.Errorf("%w: Invalid inode", gopi.ErrBadParameter)
	} else if abspath == "" {
		return fmt.Errorf("%w: Invalid abspath", gopi.ErrBadParameter)
	} else if info == nil {
		return fmt.Errorf("%w: Invalid FileInfo", gopi.ErrBadParameter)
	} else if relpath, err := filepath.Rel(filepath.Join(this.root, job.relpath), abspath); err != nil {
		return fmt.Errorf("%w: Invalid AbsPath", err)
	} else {
		// Send task to indexer
		this.indexer <- task{jobnode, inode, relpath, info}
		// Return success
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// INDEX FILES

func (this *indexer) IndexTask(start chan<- struct{}, stop <-chan struct{}) error {
	start <- gopi.DONE
	this.log.Debug2("<fsindexer.Index> IndexTask started")
FOR_LOOP:
	for {
		select {
		case node := <-this.deleter:
			if job, exists := this.jobs[node]; exists == false {
				this.log.Warn("No such index: %v", node)
			} else if err := this.deleteIndex(job); err != nil {
				this.log.Warn("%v", err)
			}
		case task := <-this.indexer:
			if err := this.indexFile(task); err != nil {
				this.log.Warn("%v", err)
			}
		case <-stop:
			break FOR_LOOP
		}
	}
	this.log.Debug2("<fsindexer.Index> IndexTask stopped")
	return nil
}

func (this *indexer) deleteIndex(index *job) error {
	this.log.Debug2("<fsindexer.deleteIndex>{ index=%v }", index)
	return gopi.ErrNotImplemented
}

func (this *indexer) indexFile(t task) error {
	this.log.Debug2("<fsindexer.indexFile>{ t=%v }", t)

	// Check incoming task
	if job, exists := this.jobs[t.jobnode]; exists == false {
		return fmt.Errorf("%w: No such job", gopi.ErrNotFound)
	} else {
		// Detect MimeType and put into database
		abspath := filepath.Join(this.root, job.relpath, t.relpath)
		if mimetype, err := detectMimeType(abspath); err != nil {
			return err
		} else if _, err := this.sqobj.Write(sq.FLAG_INSERT|sq.FLAG_UPDATE, &File{
			Id:       t.inode,
			RootPath: job.relpath,
			Job:      job.jobnode,
			RelPath:  t.relpath,
			Name:     filepath.Base(t.relpath),
			Ext:      filepath.Ext(t.relpath),
			Size:     t.info.Size(),
			MimeType: mimetype,
		}); err != nil {
			return err
		} else {
			// Increment counter
			job.count++
		}

	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// REPORT ON INDEXING

type report_state struct {
	count      uint64
	jobs, done int
}

func (this *indexer) ReportTask(start chan<- struct{}, stop <-chan struct{}) error {
	start <- gopi.DONE
	this.log.Debug2("<fsindexer.Index> ReportTask started")
	ticker := time.NewTicker(1 * time.Second)
	state := &report_state{}
FOR_LOOP:
	for {
		select {
		case <-ticker.C:
			state = this.reportJobStatus(state)
		case <-stop:
			ticker.Stop()
			break FOR_LOOP
		}
	}
	this.log.Debug2("<fsindexer.Index> ReportTask stopped")
	return nil
}

func (this *indexer) reportJobStatus(state *report_state) *report_state {
	cur := &report_state{}
	cur.jobs = len(this.jobs)
	for _, job := range this.jobs {
		cur.count += job.count
		if job.done {
			cur.done++
		}
	}
	if cur.count != state.count {
		this.log.Info("Index Status: %v", cur)
	}
	return cur
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *report_state) String() string {
	if this.done == this.jobs {
		return fmt.Sprintf("%v items indexed, finished", this.count)
	} else {
		return fmt.Sprintf("%v items indexed, running", this.count)
	}
}
