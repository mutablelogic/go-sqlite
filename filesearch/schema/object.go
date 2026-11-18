package schema

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// ObjectMeta is what's indexed of a file, of a volume, which is written
type ObjectMeta struct {
	Volume  int64          `json:"volume"`          // Its volume
	Path    string         `json:"path"`            // Relative to its volume's
	Size    int64          `json:"size"`            // In bytes
	ModTime time.Time      `json:"modtime"`         // When it was modified
	Type    string         `json:"type,omitempty"`  // Its content type
	Kind    string         `json:"kind,omitempty"`  // Its kind, such as photo or book
	Title   string         `json:"title,omitempty"` // Its title
	Meta    map[string]any `json:"meta,omitempty"`  // Its metadata, of "ns:key"
	Artwork []byte         `json:"-"`               // An image, which is kept, and read on its own
	Text    string         `json:"-"`               // Its content, as Markdown, which is searched, and not kept
	Run     int64          `json:"run"`             // The run of its volume which indexed it
}

// Object is a file which is indexed, of a volume, which is read
type Object struct {
	ID      int64          `json:"id"`
	Volume  int64          `json:"volume"`
	Path    string         `json:"path"` // Relative to its volume's
	Size    int64          `json:"size"`
	ModTime time.Time      `json:"modtime"`
	Type    string         `json:"type,omitempty"`
	Kind    string         `json:"kind,omitempty"`
	Title   string         `json:"title,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
	Indexed time.Time      `json:"indexed"` // When it was indexed
	Run     int64          `json:"run"`     // The run of its volume which last saw it
}

// ObjectList is a list of objects, of its request, with the limit which was
// used, and their count, without their offset and limit
type ObjectList struct {
	ObjectListRequest
	Count uint64   `json:"count"`
	Body  []Object `json:"body,omitempty"`
}

// ObjectID selects an object, by its id, to get or delete it
type ObjectID int64

// ObjectMark selects an object which is as it was, of its volume, path,
// size and when it was modified, to mark it as seen by its volume's run,
// rather than indexing it again, with Update, which is ErrNotFound when it
// isn't
type ObjectMark struct {
	Volume  int64
	Path    string
	Size    int64
	ModTime time.Time
	Run     int64
}

// ObjectSweep selects the objects of a volume, which its run didn't see,
// once its run indexed it, to delete them. The run is the volume's latest,
// so a run which is indexing, or failed, or an older one, sweeps nothing.
type ObjectSweep struct {
	Volume int64
	Run    int64
}

// ObjectListRequest lists objects, of the volumes, or all of them, when
// there are none, by their paths, or, when there's a query, searches them,
// of its words, which are all in their text, as prefixes, the best first. A
// query without any words matches nothing.
type ObjectListRequest struct {
	sqlite.OffsetLimit
	Volumes []int64 `json:"volume,omitempty" help:"Objects of the volumes, or of all of them"`
	Query   *string `json:"q,omitempty" help:"Words, which are all in the objects' text"`
}

// Artwork is an object's image, which is read on its own
type Artwork []byte

// ArtworkID selects an object's artwork, by its id, which is ErrNotFound
// when it hasn't any
type ArtworkID int64

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (o Object) String() string {
	return fmt.Sprintf("Object{%d %q}", o.ID, o.Path)
}

////////////////////////////////////////////////////////////////////////////////
// WRITER

// Insert binds an object, which replaces one of its path, with its text
func (o ObjectMeta) Insert(bind *sqlite.Bind) (string, error) {
	path, err := relPath(o.Path)
	if err != nil {
		return "", err
	}
	if o.Size < 0 {
		return "", sqlite.ErrBadParameter.Withf("size: %d", o.Size)
	}
	meta := o.Meta
	if meta == nil {
		meta = map[string]any{}
	}
	bind.Set("volume", o.Volume)
	bind.Set("path", path)
	bind.Set("size", o.Size)
	bind.Set("modtime", o.ModTime.UTC())
	bind.Set("type", nullString(o.Type))
	bind.Set("kind", nullString(o.Kind))
	bind.Set("title", nullString(o.Title))
	bind.Set("meta", meta)
	bind.Set("meta_text", metaText(meta))
	bind.Set("artwork", o.Artwork)
	bind.Set("text", o.Text)
	bind.Set("indexed", time.Now().UTC())
	bind.Set("run", o.Run)
	return bind.Query("filesearch.object_upsert"), nil
}

// Update isn't implemented: an object is replaced, with Insert
func (o ObjectMeta) Update(*sqlite.Bind) error {
	return sqlite.ErrNotImplemented.With("an object is replaced, with Insert")
}

////////////////////////////////////////////////////////////////////////////////
// READER

// Scan reads an object
func (o *Object) Scan(row sqlite.Row) error {
	// The metadata is of this row, and not of one read before it
	o.Meta = nil
	var typ, kind, title *string
	if err := row.Scan(&o.ID, &o.Volume, &o.Path, &o.Size, &o.ModTime, &typ, &kind, &title, &o.Meta, &o.Indexed, &o.Run); err != nil {
		return err
	}
	o.Type, o.Kind, o.Title = value(typ), value(kind), value(title)
	return nil
}

// Scan reads an object of a list
func (l *ObjectList) Scan(row sqlite.Row) error {
	var o Object
	if err := o.Scan(row); err != nil {
		return err
	}
	l.Body = append(l.Body, o)
	return nil
}

// ScanCount reads the count of a list
func (l *ObjectList) ScanCount(row sqlite.Row) error {
	return row.Scan(&l.Count)
}

// Scan reads an object's artwork
func (a *Artwork) Scan(row sqlite.Row) error {
	return row.Scan((*[]byte)(a))
}

////////////////////////////////////////////////////////////////////////////////
// SELECTOR

// Select an object, by its id, to get or delete it
func (id ObjectID) Select(bind *sqlite.Bind, op sqlite.Op) (string, error) {
	bind.Set("id", int64(id))
	switch op {
	case sqlite.Get:
		return bind.Query("filesearch.object_get"), nil
	case sqlite.Delete:
		return bind.Query("filesearch.object_delete"), nil
	}
	return "", sqlite.ErrNotImplemented.Withf("object: %v", op)
}

// Select an object which is as it was, to mark it as seen by the run
func (m ObjectMark) Select(bind *sqlite.Bind, op sqlite.Op) (string, error) {
	if op != sqlite.Update {
		return "", sqlite.ErrNotImplemented.Withf("mark: %v", op)
	}
	path, err := relPath(m.Path)
	if err != nil {
		return "", err
	}
	bind.Set("volume", m.Volume)
	bind.Set("path", path)
	bind.Set("size", m.Size)
	bind.Set("modtime", m.ModTime.UTC())
	bind.Set("run", m.Run)
	return bind.Query("filesearch.object_mark"), nil
}

// Select the objects of a volume, which its run didn't see, to delete them
func (s ObjectSweep) Select(bind *sqlite.Bind, op sqlite.Op) (string, error) {
	if op != sqlite.Delete {
		return "", sqlite.ErrNotImplemented.Withf("sweep: %v", op)
	}
	bind.Set("volume", s.Volume)
	bind.Set("run", s.Run)
	return bind.Query("filesearch.object_sweep"), nil
}

// Select objects, of the volumes, or all of them, which match the query,
// when there is one
func (r *ObjectListRequest) Select(bind *sqlite.Bind, op sqlite.Op) (string, error) {
	if op != sqlite.List {
		return "", sqlite.ErrNotImplemented.Withf("list: %v", op)
	}
	r.OffsetLimit.Bind(bind, ListLimit)

	// Of the volumes, which are a JSON array, or all of them
	var where string
	if len(r.Volumes) > 0 {
		where = "volume IN (SELECT value FROM json_each(" + bind.Set("volumes", r.Volumes) + "))"
	}

	// Searched, of the words of the query, which, without any, matches
	// nothing, or listed, by path, without a query
	if r.Query != nil {
		if match := matchQuery(*r.Query); match != "" {
			bind.Set("match", match)
			bind.Set("and", and(where))
			return bind.Query("filesearch.object_search"), nil
		}
		where = "FALSE"
	}
	if where != "" {
		where = "WHERE " + where
	}
	bind.Set("where", where)
	return bind.Query("filesearch.object_list"), nil
}

// Select an object's artwork, by its id
func (id ArtworkID) Select(bind *sqlite.Bind, op sqlite.Op) (string, error) {
	if op != sqlite.Get {
		return "", sqlite.ErrNotImplemented.Withf("artwork: %v", op)
	}
	bind.Set("id", int64(id))
	return bind.Query("filesearch.artwork_get"), nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// absPath returns the path, which is absolute, cleaned, with / separators
func absPath(path string) (string, error) {
	if path == "" || !filepath.IsAbs(path) {
		return "", sqlite.ErrBadParameter.Withf("path isn't absolute: %q", path)
	}
	return filepath.ToSlash(filepath.Clean(path)), nil
}

// relPath returns the path, relative to a volume's, cleaned, with /
// separators, which isn't absolute, or outside of it
func relPath(path string) (string, error) {
	clean := filepath.ToSlash(filepath.Clean(path))
	if path == "" || filepath.IsAbs(path) || strings.HasPrefix(clean, "/") || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", sqlite.ErrBadParameter.Withf("path isn't within a volume: %q", path)
	}
	return clean, nil
}

// and returns the condition, after AND, or nothing
func and(condition string) string {
	if condition == "" {
		return ""
	}
	return "AND " + condition
}

// matchQuery returns the FTS5 query of the words of a query, each of which
// is a prefix, in quotes, so it isn't FTS5's syntax, which are all matched,
// or nothing, when there are no words
func matchQuery(query string) string {
	words := strings.Fields(query)
	for i, word := range words {
		words[i] = `"` + strings.ReplaceAll(word, `"`, `""`) + `"*`
	}
	return strings.Join(words, " ")
}

// metaText returns the text of the metadata, which is searched: the values
// which are strings, or lists of them, of each key, in order
func metaText(meta map[string]any) string {
	keys := make([]string, 0, len(meta))
	for key := range meta {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	var text []string
	for _, key := range keys {
		switch v := meta[key].(type) {
		case string:
			text = append(text, v)
		case []string:
			text = append(text, v...)
		case []any:
			for _, v := range v {
				if s, ok := v.(string); ok {
					text = append(text, s)
				}
			}
		}
	}
	return strings.Join(text, "\n")
}

// nullString returns nil, which is NULL, for an empty string
func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// value returns the string, or an empty one, for nil
func value(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
