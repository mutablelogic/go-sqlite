package sqlite3

import (
	"fmt"
	"io"
)

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#include <sqlite3.h>
#include <stdlib.h>
*/
import "C"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type BlobEx struct {
	io.ReadWriteSeeker
	io.ReaderAt
	io.WriterAt
	io.Closer
	*Blob

	cur, size int64
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (b *BlobEx) String() string {
	str := "<blobex"
	str += fmt.Sprint(" cur=", b.cur)
	str += fmt.Sprint(" size=", b.size)
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// OpenBlobEx with specified schema, table, column and rowid. If called with flag
// SQLITE_OPEN_READWRITE then the blob handle is opened for read/write access, otherwise
// for read-only access.
func (c *Conn) OpenBlobEx(schema, table, column string, rowid int64, flags OpenFlags) (*BlobEx, error) {
	bx := new(BlobEx)
	if b, err := c.OpenBlob(schema, table, column, rowid, flags); err != nil {
		return nil, err
	} else {
		bx.Blob = b
		bx.cur, bx.size = 0, int64(b.Bytes())
	}

	// Return success
	return bx, nil
}

// Close a blob and release resources
func (b *BlobEx) Close() error {
	err := b.Blob.Close()
	b.Blob = nil
	return err
}

// Reopen moves the blob handle to a new rowid
func (b *BlobEx) Reopen(rowid int64) error {
	if err := b.Blob.Reopen(rowid); err != nil {
		return err
	}
	b.cur, b.size = 0, int64(b.Blob.Bytes())
	return nil
}

// io.Reader interface
func (b *BlobEx) Read(data []byte) (int, error) {
	if b.Blob == nil || b.cur >= b.size {
		return 0, io.EOF
	}
	if remaining := b.size - b.cur; int64(len(data)) > remaining {
		data = data[:remaining]
	}
	if n, err := b.ReadAt(data, b.cur); err != nil {
		return 0, err
	} else {
		b.cur += int64(n)
		return n, nil
	}
}

// io.Writer interface
func (b *BlobEx) Write(data []byte) (int, error) {
	if b.Blob == nil {
		return 0, io.EOF
	}
	n := int(0)
	if remaining := int(b.size - b.cur); len(data) > remaining {
		n = int(remaining)
	} else {
		n = len(data)
	}
	if n, err := b.WriteAt(data[:n], b.cur); err != nil {
		return 0, err
	} else {
		b.cur += int64(n)
		return n, nil
	}
}

// io.ReaderAt interface
func (b *BlobEx) ReadAt(data []byte, offset int64) (int, error) {
	if b.Blob == nil {
		return 0, io.EOF
	}
	if err := b.Blob.ReadAt(data, offset); err != nil {
		return 0, err
	} else {
		return len(data), nil
	}
}

// io.WriterAt interface
func (b *BlobEx) WriteAt(data []byte, offset int64) (int, error) {
	if b.Blob == nil {
		return 0, io.EOF
	}
	if err := b.Blob.WriteAt(data, offset); err != nil {
		return 0, err
	} else {
		return len(data), nil
	}
}

// io.ReadWriteSeeker interface
func (b *BlobEx) Seek(offset int64, whence int) (int64, error) {
	if b.Blob == nil {
		return 0, io.EOF
	}
	switch whence {
	case io.SeekCurrent:
		if b.cur+offset < 0 {
			return 0, SQLITE_RANGE
		}
		b.cur += offset
	case io.SeekStart:
		if offset < 0 {
			return 0, SQLITE_RANGE
		}
		b.cur = offset
	case io.SeekEnd:
		if b.size+offset < 0 {
			return 0, SQLITE_RANGE
		}
		b.cur = b.size + offset
	default:
		return 0, SQLITE_MISUSE
	}

	// Return success
	return b.cur, nil
}
