package importer

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	// Modules
	"golang.org/x/text/encoding/charmap"
)

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Open the table for reading, return a reader and a mimetype
func open(url *url.URL) (io.ReadCloser, string, error) {
	if url.Scheme == "file" || url.Scheme == "" {
		if mimetype, err := detectMimetype(url.Path); err != nil {
			return nil, "", err
		} else if fh, err := os.Open(url.Path); err != nil {
			return nil, "", err
		} else {
			return fh, mimetype, nil
		}
	} else {
		return openHTTP(url.String())
	}
}

// detectMimetype returns the mimetype of the given file, or an empty string if
// no mimetype was detected
func detectMimetype(path string) (string, error) {
	fh, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer fh.Close()
	data := make([]byte, 512)
	if _, err := fh.Read(data); err != nil {
		return "", err
	}
	if mimetype := http.DetectContentType(data); mimetype != "application/octet-stream" {
		return mimetype, nil
	} else {
		return mime.TypeByExtension(filepath.Ext(path)), nil
	}
}

// openHTTP opens a HTTP or HTTPS connection to the given URL
// and returns the file handle and content type
func openHTTP(url string) (io.ReadCloser, string, error) {
	client := http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		return nil, "", err
	}

	// Check status code
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		return nil, "", errors.New(resp.Status)
	}

	// Return success
	return resp.Body, resp.Header.Get("Content-Type"), nil
}

// charsetReader translates from supported character sets to UTF-8
func charsetReader(r io.Reader, charset string) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "utf8", "utf-8", "":
		// Default
		return r, nil
	case "windows-1252":
		return charmap.Windows1252.NewDecoder().Reader(r), nil
	case "iso-8859-1":
		return charmap.ISO8859_1.NewDecoder().Reader(r), nil
	default:
		return nil, fmt.Errorf("unsupported charset: %q", charset)
	}
}
