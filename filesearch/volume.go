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

// ListVolumes lists the volumes, by their paths, with the status of their
// indexing, with the request, and the limit which was used, clamped to
// what's returned
func (m *Manager) ListVolumes(ctx context.Context, req schema.VolumeListRequest) (_ *schema.VolumeList, err error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "filesearch.ListVolumes",
		attribute.String("arg.req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()

	// Copy the limit, so the caller's request isn't changed
	if req.Limit != nil {
		req.Limit = types.Ptr(*req.Limit)
	}

	// Selecting sets the limit which is used, on the request, which is
	// returned with the list, clamped to the volumes which there are
	var list schema.VolumeList
	if err := m.conn.List(ctx, &list, &req); err != nil {
		return nil, err
	}
	list.VolumeListRequest = req
	list.OffsetLimit.Clamp(list.Count)
	return &list, nil
}
