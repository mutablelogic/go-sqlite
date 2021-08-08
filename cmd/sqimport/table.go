package main

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
	"sync"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type table struct {
	sync.Mutex
	url      *url.URL
	r        io.ReadCloser
	dec      *decoder
	mimetype string
	charset  string
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewTable(arg string) (*table, error) {
	this := new(table)

	if url, err := url.Parse(arg); err != nil {
		return nil, err
	} else {
		this.url = url
	}

	return this, nil
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *table) Name() string {
	name := filepath.Base(this.url.String())
	if ext := filepath.Ext(name); ext != "" {
		name = strings.TrimSuffix(name, ext)
	}
	return name
}

func (this *table) URL() *url.URL {
	return this.url
}

func (this *table) Mediatype() string {
	mediatype, _, err := mime.ParseMediaType(this.mimetype)
	if err == nil && mediatype != "" {
		return mediatype
	} else {
		return "application/octet-stream"
	}
}

func (this *table) Charset() string {
	_, params, err := mime.ParseMediaType(this.mimetype)
	if err != nil {
		return ""
	}
	if charset, ok := params["charset"]; ok {
		return charset
	} else {
		return ""
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *table) String() string {
	str := "<table"
	str += fmt.Sprintf(" url=%q", this.url)
	str += fmt.Sprintf(" name=%q", this.Name())
	if mimetype := this.Mediatype(); mimetype != "" {
		str += fmt.Sprintf(" mimetype=%q", mimetype)
	}
	if charset := this.Charset(); charset != "" {
		str += fmt.Sprintf(" charset=%q", charset)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Read a row from the source data, returning nil if the row is to be skipped,
// or an error if the row could not be read, or returns io.EOF on end of file.
func (this *table) Read() (map[string]interface{}, error) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Open the data source
	if this.r == nil {
		if r, mimetype, err := open(this.url); err != nil {
			return nil, err
		} else {
			this.r = r
			this.mimetype = mimetype
		}
		// Skip row
		return nil, nil
	}

	// Set the decoder
	if this.dec == nil {
		if dec, err := NewDecoder(this.r, this.mimetype); err != nil {
			return nil, err
		} else {
			this.dec = dec
		}
		// Skip row
		return nil, nil
	}

	// Read the row
	row, err := this.dec.Read()
	if err != nil {
		defer this.r.Close()
		this.dec = nil
		this.r = nil
		return nil, err
	}

	// Return the row
	return row, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Open the table for reading
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
