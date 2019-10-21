/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package fsindexer

import (
	"errors"
	"io"
	"net/http"
	"os"
	"syscall"
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	R_OK = 4
	W_OK = 2
	X_OK = 1
)

////////////////////////////////////////////////////////////////////////////////
// File utility methods

// Returns boolean value which indicates if a file is readable by current
// user
func isReadableFileAtPath(path string) error {
	return syscall.Access(path, R_OK)
}

// Returns boolean value which indicates if a file is writable by current
// user
func isWritableFileAtPath(path string) error {
	return syscall.Access(path, W_OK)
}

// Returns boolean value which indicates if a file is executable by current
// user
func isExecutableFileAtPath(path string) error {
	return syscall.Access(path, X_OK)
}

////////////////////////////////////////////////////////////////////////////////
// DetectMimeType attempts to read first 512 bytes of the
// file
func detectMimeType(path string) (string, error) {
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
