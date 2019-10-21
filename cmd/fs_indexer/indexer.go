/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package main

/*
import (
	"path/filepath"

	"github.com/djthorpe/gopi"

	// Frameworks
	sqlite "github.com/djthorpe/sqlite"
)

type Indexer struct {
	sqobj sqlite.Objects
}

func NewIndexer(sqobj sqlite.Objects) (*Indexer, error) {
	this := new(Indexer)
	this.sqobj = sqobj
	if _, err := sqobj.RegisterStruct(&File{}); err != nil {
		return nil, err
	}
	return this, nil
}

func (this *Indexer) Do(file *File) (uint64, error) {
	if file == nil || file.Id == 0 || file.Path == "" || file.Root == "" {
		return 0, gopi.ErrBadParameter
	} else if path, err := filepath.Rel(file.Root, file.Path); err != nil {
		return 0, err
	} else {
		file.Path = path
		if affected_rows, err := this.sqobj.Write(sqlite.FLAG_INSERT|sqlite.FLAG_UPDATE, file); err != nil {
			return 0, err
		} else {
			return affected_rows, nil
		}
	}
}

// DetectMimeType attempts to read first 512 bytes of the
// file
func DetectMimeType(path string) (string, error) {
	fh, err := os.Open(path)
	if err != nil {
		// File could not be opened, fail silently
		return "", nil
	}
	defer fh.Close()

	// Read 512 bytes
	buf := make([]byte, 512)
	if n, err := fh.Read(buf); err != nil && errors.Is(err, io.EOF) == false {
		return "", err
	} else if n > 0 {
		return http.DetectContentType(buf), nil
	} else {
		// Cannot detect mimetype
		return "", nil
	}
}

*/
