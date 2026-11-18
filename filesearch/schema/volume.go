package schema

import (
	"time"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Status is the status of a volume's indexing
type Status string

// VolumeMeta is a folder which is indexed, by its absolute path, which is
// added, or is the one of it already
type VolumeMeta struct {
	Path string `json:"path"`
}

// IndexVolumeRequest indexes a volume, of a folder, and its folders, when
// it's recursive, of the files which are new, or changed, or all of them,
// when they're reindexed
type IndexVolumeRequest struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive,omitempty"`
	Reindex   bool   `json:"reindex,omitempty"`
}

// Volume is a folder which is indexed, with the status of its indexing
type Volume struct {
	ID       int64      `json:"id"`
	Path     string     `json:"path"`
	Status   Status     `json:"status"`
	Run      int64      `json:"run"`                // The run, which marks the objects it sees, or 0
	Started  *time.Time `json:"started,omitempty"`  // When the run started
	Finished *time.Time `json:"finished,omitempty"` // When it finished, or nil while it's indexing
	Error    string     `json:"error,omitempty"`    // Why it failed
	Objects  uint64     `json:"objects"`            // The objects which are indexed
}

// VolumeList is a list of volumes, of its request, with the limit which was
// used, and their count
type VolumeList struct {
	VolumeListRequest
	Count uint64   `json:"count"`
	Body  []Volume `json:"body,omitempty"`
}

// VolumeListRequest lists the volumes, by their paths
type VolumeListRequest struct {
	sqlite.OffsetLimit
}

// VolumeID selects a volume, by its id, to get, start, finish or delete it
type VolumeID int64

// VolumePath selects a volume, by its absolute path, to get it
type VolumePath string

// VolumeStart starts its next run, which is indexing it, with Update
type VolumeStart struct {
	Started time.Time
}

// VolumeFinish finishes its run, which indexed it, or failed, of its
// error, with Update, which is ErrNotFound when it isn't indexing, or the
// run isn't its latest, as another started after it
type VolumeFinish struct {
	Run      int64
	Finished time.Time
	Error    error
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	StatusNew      Status = "new"      // Not indexed yet
	StatusIndexing Status = "indexing" // Its run is indexing it
	StatusIndexed  Status = "indexed"  // Its run indexed it, and what it didn't see can be swept
	StatusFailed   Status = "failed"   // Its run failed, of its error
)

////////////////////////////////////////////////////////////////////////////////
// WRITER

// Insert binds a volume, which is added, or is the one of its path already
func (v VolumeMeta) Insert(bind *sqlite.Bind) (string, error) {
	path, err := absPath(v.Path)
	if err != nil {
		return "", err
	}
	bind.Set("path", path)
	return bind.Query("filesearch.volume_insert"), nil
}

// Update isn't implemented: a volume is of its path
func (v VolumeMeta) Update(*sqlite.Bind) error {
	return sqlite.ErrNotImplemented.With("a volume is of its path")
}

// Insert isn't implemented: a volume is added, with VolumeMeta
func (VolumeStart) Insert(*sqlite.Bind) (string, error) {
	return "", sqlite.ErrNotImplemented.With("a volume is added, with VolumeMeta")
}

// Update binds when its run started
func (s VolumeStart) Update(bind *sqlite.Bind) error {
	bind.Set("op", "start")
	bind.Set("started", s.Started.UTC())
	return nil
}

// Insert isn't implemented: a volume is added, with VolumeMeta
func (VolumeFinish) Insert(*sqlite.Bind) (string, error) {
	return "", sqlite.ErrNotImplemented.With("a volume is added, with VolumeMeta")
}

// Update binds when its run finished, and whether it failed
func (f VolumeFinish) Update(bind *sqlite.Bind) error {
	bind.Set("op", "finish")
	bind.Set("run", f.Run)
	bind.Set("finished", f.Finished.UTC())
	if f.Error != nil {
		bind.Set("status", StatusFailed)
		bind.Set("error", f.Error.Error())
	} else {
		bind.Set("status", StatusIndexed)
		bind.Set("error", nil)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// READER

// Scan reads a volume
func (v *Volume) Scan(row sqlite.Row) error {
	var err *string
	if e := row.Scan(&v.ID, &v.Path, &v.Status, &v.Run, &v.Started, &v.Finished, &err, &v.Objects); e != nil {
		return e
	}
	v.Error = value(err)
	return nil
}

// Scan reads a volume of a list
func (l *VolumeList) Scan(row sqlite.Row) error {
	var v Volume
	if err := v.Scan(row); err != nil {
		return err
	}
	l.Body = append(l.Body, v)
	return nil
}

// ScanCount reads the count of a list
func (l *VolumeList) ScanCount(row sqlite.Row) error {
	return row.Scan(&l.Count)
}

////////////////////////////////////////////////////////////////////////////////
// SELECTOR

// Select a volume, by its id, to get, start, finish or delete it, of the
// writer, VolumeStart or VolumeFinish, which is bound before it
func (id VolumeID) Select(bind *sqlite.Bind, op sqlite.Op) (string, error) {
	bind.Set("id", int64(id))
	switch op {
	case sqlite.Get:
		return bind.Query("filesearch.volume_get"), nil
	case sqlite.Delete:
		return bind.Query("filesearch.volume_delete"), nil
	case sqlite.Update:
		switch bind.Get("op") {
		case "start":
			return bind.Query("filesearch.volume_start"), nil
		case "finish":
			return bind.Query("filesearch.volume_finish"), nil
		}
		return "", sqlite.ErrBadParameter.With("volume: update with VolumeStart or VolumeFinish")
	}
	return "", sqlite.ErrNotImplemented.Withf("volume: %v", op)
}

// Select the volumes, by their paths
func (r *VolumeListRequest) Select(bind *sqlite.Bind, op sqlite.Op) (string, error) {
	if op != sqlite.List {
		return "", sqlite.ErrNotImplemented.Withf("volumes: %v", op)
	}
	r.OffsetLimit.Bind(bind, ListLimit)
	return bind.Query("filesearch.volume_list"), nil
}

// Select a volume, by its path
func (p VolumePath) Select(bind *sqlite.Bind, op sqlite.Op) (string, error) {
	if op != sqlite.Get {
		return "", sqlite.ErrNotImplemented.Withf("volume path: %v", op)
	}
	path, err := absPath(string(p))
	if err != nil {
		return "", err
	}
	bind.Set("path", path)
	return bind.Query("filesearch.volume_path"), nil
}
