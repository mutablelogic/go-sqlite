package filesearch

import (
	"context"
	"path/filepath"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	types "github.com/mutablelogic/go-server/pkg/types"
	sqlite "github.com/mutablelogic/go-sqlite"
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

// GetVolume returns a volume, by its id, with the status of its indexing,
// which is ErrNotFound when there isn't one
func (m *Manager) GetVolume(ctx context.Context, id int64) (_ *schema.Volume, err error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "filesearch.GetVolume",
		attribute.Int64("arg.id", id),
	)
	defer func() { endSpan(err) }()
	var volume schema.Volume
	if err := m.conn.Get(ctx, &volume, schema.VolumeID(id)); err != nil {
		return nil, err
	}
	return &volume, nil
}

// VolumeByPath returns a volume, by its path, of the folder which a symbolic
// link is to, which is ErrNotFound when there isn't one
func (m *Manager) VolumeByPath(ctx context.Context, path string) (_ *schema.Volume, err error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "filesearch.VolumeByPath",
		attribute.String("arg.path", path),
	)
	defer func() { endSpan(err) }()
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	var volume schema.Volume
	if err := m.conn.Get(ctx, &volume, schema.VolumePath(path)); err != nil {
		return nil, err
	}
	return &volume, nil
}

// DeleteVolume deletes a volume, by its id, with its objects, and their
// text, and returns it, which is ErrNotFound when there isn't one, and
// ErrConflict when this manager is indexing it
func (m *Manager) DeleteVolume(ctx context.Context, id int64) (_ *schema.Volume, err error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "filesearch.DeleteVolume",
		attribute.Int64("arg.id", id),
	)
	defer func() { endSpan(err) }()
	if p, err := m.Progress(id); err == nil && p.Finished == nil {
		return nil, sqlite.ErrConflict.Withf("volume %d is indexing", id)
	}
	// It's read before it's deleted, so it has the count of its objects, which
	// are deleted with it, in a transaction
	var volume schema.Volume
	if err := m.conn.Tx(ctx, func(conn sqlite.Conn) error {
		if err := conn.Get(ctx, &volume, schema.VolumeID(id)); err != nil {
			return err
		}
		return conn.Delete(ctx, nil, schema.VolumeID(id))
	}); err != nil {
		return nil, err
	}

	// Its progress, of this manager, isn't kept
	m.Lock()
	delete(m.progress, volume.Path)
	m.Unlock()
	return &volume, nil
}
