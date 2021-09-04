package indexer

import (
	"fmt"
	"path/filepath"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type File struct {
	Index   string    `sqlite:"index,primary" json:"index"`
	Path    string    `sqlite:"path,primary" json:"path"`
	Name    string    `sqlite:"name,primary" json:"name"`
	IsDir   bool      `sqlite:"is_dir,not null" json:"is_dir"`
	Ext     string    `sqlite:"ext" json:"ext"`
	ModTime time.Time `sqlite:"modtime" json:"modtime"`
	Size    int64     `sqlite:"size,not null" json:"size"`
}

type FileMark struct {
	Index     string    `sqlite:"index,primary,foreign"`
	Path      string    `sqlite:"path,primary,foreign"`
	Name      string    `sqlite:"name,primary,foreign"`
	Mark      bool      `sqlite:"mark,not null"`
	IndexTime time.Time `sqlite:"idxtime"`
}

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
	}
	return f
}

// NewFileMark creates an empty filemark object
func NewFileMark(index, path, name string) *FileMark {
	return &FileMark{Index: index, Path: path, Name: name, Mark: false, IndexTime: time.Now()}
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
	if !this.ModTime.IsZero() {
		str += fmt.Sprint(" modtime=", this.ModTime.Format(time.Kitchen))
	}

	return str + ">"
}
