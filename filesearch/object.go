package filesearch

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	types "github.com/mutablelogic/go-server/pkg/types"
	sqlite "github.com/mutablelogic/go-sqlite"
	schema "github.com/mutablelogic/go-sqlite/filesearch/schema"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ListObjects lists the objects, of the volumes, or all of them, by their
// paths, or searches them, when there's a query, the best first, with the
// request, and the limit which was used, clamped to what's returned
func (m *Manager) ListObjects(ctx context.Context, req schema.ObjectListRequest) (_ *schema.ObjectList, err error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "filesearch.ListObjects",
		attribute.String("arg.req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()

	// Copy the limit, so the caller's request isn't changed
	if req.Limit != nil {
		req.Limit = types.Ptr(*req.Limit)
	}

	// Selecting sets the limit which is used, on the request, which is
	// returned with the list, clamped to the objects which there are
	var list schema.ObjectList
	if err := m.conn.List(ctx, &list, &req); err != nil {
		return nil, err
	}
	list.ObjectListRequest = req
	list.OffsetLimit.Clamp(list.Count)
	return &list, nil
}

// Artwork returns an object's artwork, such as a thumbnail or a cover, by
// its id, which is ErrNotFound when it hasn't any
func (m *Manager) Artwork(ctx context.Context, id int64) (_ []byte, err error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "filesearch.Artwork",
		attribute.Int64("arg.id", id),
	)
	defer func() { endSpan(err) }()
	var artwork schema.Artwork
	if err := m.conn.Get(ctx, &artwork, schema.ArtworkID(id)); err != nil {
		return nil, err
	}
	return artwork, nil
}

// GetObject returns the object of a path: of its volume's path, and its own,
// or its own, or the end of it, after a /, such as a file's name. It's
// ErrNotFound when no object is, and ErrConflict, of their paths, when more
// than one is.
func (m *Manager) GetObject(ctx context.Context, path string) (_ *schema.Object, err error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "filesearch.GetObject",
		attribute.String("arg.path", path),
	)
	defer func() { endSpan(err) }()
	var list schema.ObjectList
	if err := m.conn.List(ctx, &list, schema.ObjectPath(path)); err != nil {
		return nil, err
	}
	switch len(list.Body) {
	case 0:
		return nil, sqlite.ErrNotFound.Withf("no file is %q", path)
	case 1:
		return &list.Body[0], nil
	}
	paths := make([]string, 0, len(list.Body))
	for _, object := range list.Body {
		if path, err := m.ObjectPath(ctx, &object); err == nil {
			paths = append(paths, path)
		} else {
			paths = append(paths, object.Path)
		}
	}
	more := ""
	if list.Count > uint64(len(list.Body)) {
		more = fmt.Sprintf(", and %d more", list.Count-uint64(len(list.Body)))
	}
	return nil, sqlite.ErrConflict.Withf("%d files are %q: %s%s", list.Count, path, strings.Join(paths, ", "), more)
}

// ObjectPath returns the path of an object's file, of its volume's path, and
// its own
func (m *Manager) ObjectPath(ctx context.Context, object *schema.Object) (string, error) {
	volume, err := m.GetVolume(ctx, object.Volume)
	if err != nil {
		return "", err
	}
	return filepath.Join(volume.Path, filepath.FromSlash(object.Path)), nil
}

// Markdown returns an object's content, as Markdown, by its id, which is
// read from its file, as it isn't kept, all of it, with go-metadata's
// handlers which are imported. It's ErrNotFound when the object isn't, and
// empty when the file has no content.
func (m *Manager) Markdown(ctx context.Context, id int64) (_ string, err error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "filesearch.Markdown",
		attribute.Int64("arg.id", id),
	)
	defer func() { endSpan(err) }()
	var object schema.Object
	if err := m.conn.Get(ctx, &object, schema.ObjectID(id)); err != nil {
		return "", err
	}
	path, err := m.ObjectPath(ctx, &object)
	if err != nil {
		return "", err
	}
	return m.readMarkdown(ctx, path, object.Type)
}
