package filesearch

import (
	"context"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	types "github.com/mutablelogic/go-server/pkg/types"
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
