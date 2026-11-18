package schema

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

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
	Indexed time.Time      `json:"indexed"`        // When it was indexed
	Run     int64          `json:"run"`            // The run of its volume which last saw it
	Rank    *float64       `json:"rank,omitempty"` // How well it matches a search, between 0 and 1, which is higher the better, or nil
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
// there are none, the newest first, or, when there's a query, searches them,
// of its words, which are all in their text, as prefixes, the best first,
// unless it has another order. A query without any words matches nothing.
type ObjectListRequest struct {
	sqlite.OffsetLimit
	Volumes []int64  `json:"volume,omitempty" help:"Objects of the volumes, or of all of them"`
	Types   []string `json:"type,omitempty" help:"Objects of any of the types: a kind, or the start, or end, of a content type, such as video, or pdf, or, with a /, a content type, such as video/mp4"`
	Query   *string  `json:"q,omitempty" help:"Words, which are all in the objects' text"`
	Order   Order    `json:"order,omitempty" help:"rank, the best first, which is a search's default, date, the newest first, which is a list's, path, size, the largest first, or title"`
	Cutoff  *float64 `json:"cutoff,omitempty" help:"The lowest rank, of a search, between 0 and 1, which is DefaultCutoff, unless it's set, or 0, for none"`
}

// Order is the order of a list of objects
type Order string

// The orders of a list of objects, which, when they're the same, are of
// their volumes and paths
const (
	OrderRank  Order = "rank"  // The best match first, of a search, which is its default
	OrderDate  Order = "date"  // The newest first, of when they were modified, which is a list's default
	OrderPath  Order = "path"  // By volume, and path
	OrderSize  Order = "size"  // The largest first
	OrderTitle Order = "title" // By title, of any case, and those without one last
)

// orders are the SQL of the orders, which are of the columns of a list, or
// a search
var orders = map[Order]string{
	OrderRank:  "rank DESC, volume, path",
	OrderDate:  "julianday(modtime) DESC, volume, path",
	OrderPath:  "volume, path",
	OrderSize:  "size DESC, volume, path",
	OrderTitle: "title IS NULL, title COLLATE NOCASE, volume, path",
}

// ObjectPath selects the objects of a path, to list them: of their
// volume's path, and their own, or their own, or the end of it, after a /,
// such as a file's name
type ObjectPath string

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
	bind.Set("path_text", pathText(path))
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
	o.Rank = nil
	o.Type, o.Kind, o.Title = value(typ), value(kind), value(title)
	return nil
}

// Scan reads an object of a list
func (l *ObjectList) Scan(row sqlite.Row) error {
	var o Object
	var typ, kind, title *string
	if err := row.Scan(&o.ID, &o.Volume, &o.Path, &o.Size, &o.ModTime, &typ, &kind, &title, &o.Meta, &o.Indexed, &o.Run, &o.Rank); err != nil {
		return err
	}
	o.Type, o.Kind, o.Title = value(typ), value(kind), value(title)
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

	// Of the volumes, and the types, which are JSON arrays, or all of them
	var conditions []string
	if len(r.Volumes) > 0 {
		conditions = append(conditions, "volume IN (SELECT value FROM json_each("+bind.Set("volumes", r.Volumes)+"))")
	}
	if len(r.Types) > 0 {
		types := make([]string, 0, len(r.Types))
		for _, t := range r.Types {
			if t = strings.ToLower(strings.TrimSpace(t)); t != "" {
				types = append(types, t)
			}
		}
		if len(types) == 0 {
			return "", sqlite.ErrBadParameter.With("type is empty")
		}
		conditions = append(conditions, typeCondition(bind.Set("types", types)))
	}
	where := strings.Join(conditions, " AND ")

	// Searched, of the words of the query, of their rank, or path, and a
	// minimum rank, or not
	if r.Query != nil {
		if match := matchQuery(*r.Query); match != "" {
			bind.Set("match", match)
			if where != "" {
				bind.Set("where", "WHERE "+where)
			} else {
				bind.Set("where", "")
			}
			order := r.Order
			if order == "" {
				order = OrderRank
			}
			sql, ok := orders[order]
			if !ok {
				return "", sqlite.ErrBadParameter.Withf("order: %q", r.Order)
			}
			bind.Set("order", sql)
			cutoff := DefaultCutoff
			if r.Cutoff != nil {
				cutoff = *r.Cutoff
			}
			if cutoff < 0 || cutoff > 1 {
				return "", sqlite.ErrBadParameter.Withf("cutoff: %v, which is between 0 and 1", cutoff)
			}
			bind.Set("rank", "WHERE rank >= "+bind.Set("cutoff", cutoff))
			return bind.Query("filesearch.object_search"), nil
		}
		// A query without any words matches nothing
		where = "FALSE"
	}

	// Or listed, the newest first, or of its order, without a query, which
	// hasn't a rank
	order := r.Order
	if order == "" {
		order = OrderDate
	}
	sql, ok := orders[order]
	switch {
	case order == OrderRank:
		return "", sqlite.ErrBadParameter.With("order by rank is of a search, with a query")
	case !ok:
		return "", sqlite.ErrBadParameter.Withf("order: %q", r.Order)
	case r.Cutoff != nil && r.Query == nil:
		return "", sqlite.ErrBadParameter.With("a cutoff is of a search, with a query")
	}
	if where != "" {
		where = "WHERE " + where
	}
	bind.Set("where", where)
	bind.Set("order", sql)
	return bind.Query("filesearch.object_list"), nil
}

// Select the objects of a path, of at most ListLimit, which are of their
// volumes and paths
func (p ObjectPath) Select(bind *sqlite.Bind, op sqlite.Op) (string, error) {
	if op != sqlite.List {
		return "", sqlite.ErrNotImplemented.Withf("object path: %v", op)
	}
	path := strings.Trim(filepath.ToSlash(filepath.Clean(string(p))), "/")
	if string(p) == "" || path == "." || path == "" {
		return "", sqlite.ErrBadParameter.With("path is empty")
	}
	if filepath.IsAbs(string(p)) {
		path = filepath.ToSlash(filepath.Clean(string(p)))
	}
	limit := uint64(ListLimit)
	(&sqlite.OffsetLimit{Limit: &limit}).Bind(bind, ListLimit)
	bind.Set("path", path)
	return bind.Query("filesearch.object_path"), nil
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

// typeCondition returns the condition of the objects which are of any of
// the types, of the argument, which is a JSON array of them, in lower case:
// of a content type, with a /, or else of a kind, or the start, or end, of
// a content type, such as video, of video/mp4, or pdf, of application/pdf
func typeCondition(types string) string {
	// The object's columns are qualified, as json_each has a type column
	return `EXISTS (SELECT 1 FROM json_each(` + types + `) AS t WHERE CASE
		WHEN instr(t.value, '/') > 0 THEN lower(object.type) = t.value
		ELSE lower(object.kind) = t.value
			OR substr(lower(object.type), 1, length(t.value) + 1) = t.value || '/'
			OR substr(lower(object.type), -length(t.value) - 1) = '/' || t.value
	END)`
}

// pathText returns the text of a path, which is searched: its parts, which
// are its folders, and its name, of MinPathLength characters, or more
func pathText(path string) string {
	var parts []string
	for _, part := range strings.Split(path, "/") {
		if utf8.RuneCountInString(part) >= MinPathLength {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, "\n")
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
