package indexer

import (
	"fmt"
	"path/filepath"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type File struct {
	Index     string    `sqlite:"index,primary" json:"index"`
	Path      string    `sqlite:"path,primary" json:"path"`
	Name      string    `sqlite:"name,primary" json:"name"`
	IsDir     bool      `sqlite:"is_dir,not null" json:"is_dir"`
	Ext       string    `sqlite:"ext" json:"ext"`
	ModTime   time.Time `sqlite:"modtime" json:"modtime"`
	Size      int64     `sqlite:"size,not null" json:"size"`
	Mark      bool      `sqlite:"mark,not null" json:"-"`
	IndexTime time.Time `sqlite:"idxtime" json:"-"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewFile creates a file object from an event
func NewFile(evt IndexerEvent, write bool) *File {
	f := new(File)
	path := evt.Path()
	f.Index = evt.Name()
	f.Path = filepath.Dir(path)
	f.Name = filepath.Base(path)
	if f.Name == "." {
		f.Name = "/"
	}
	if write {
		f.Ext = filepath.Ext(f.Name)
		if info := evt.FileInfo(); info != nil {
			f.ModTime = info.ModTime()
			f.Size = info.Size()
			f.IsDir = info.IsDir()
		}
		f.IndexTime = time.Now()
	}
	return f
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *File) String() string {
	str := "<indexer."
	if this.IsDir {
		str += "dir"
	} else {
		str += "file"
	}
	if this.Index != "" {
		str += fmt.Sprintf(" index=%q", this.Index)
	}
	if this.Name != "" {
		str += fmt.Sprintf(" name=%q", this.Name)
	}
	if this.Path != "" && this.Path != pathSeparator {
		str += fmt.Sprintf(" path=%q", this.Path)
	}
	if this.IsDir {
		str += " isdir"
	} else if this.Size > 0 {
		str += fmt.Sprintf(" size=%v", this.Size)
	}
	if this.Ext != "" {
		str += fmt.Sprintf(" ext=%q", this.Ext)
	}
	if this.ModTime.IsZero() == false {
		str += fmt.Sprint(" modtime=", this.ModTime.Format(time.Kitchen))
	}

	return str + ">"
}
