package filesearch_test

import (
	"os"
	"path/filepath"
	"testing"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	filesearch "github.com/mutablelogic/go-sqlite/filesearch"
	schema "github.com/mutablelogic/go-sqlite/filesearch/schema"
	test "github.com/mutablelogic/go-sqlite/filesearch/test"
	ncruces "github.com/mutablelogic/go-sqlite/ncruces"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	tracetest "go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func Test_New(t *testing.T) {
	ctx := t.Context()
	path := test.Path(t)

	// A manager creates its database, and its schema, and is closed
	manager, err := filesearch.New(ctx, path)
	require.NoError(t, err)
	require.NoError(t, manager.Ping(ctx))
	require.NoError(t, manager.Close())
	assert.FileExists(t, path)

	// Its schema is of its version
	db, err := ncruces.Open(ctx, path)
	require.NoError(t, err)
	version, err := sqlite.Version(ctx, db)
	require.NoError(t, err)
	assert.Equal(t, int64(schema.Version), version)
	require.NoError(t, db.Close())

	// A manager of the database, which has its schema, again, and of one in
	// memory
	manager, err = filesearch.New(ctx, path, filesearch.WithMaxReaders(2))
	require.NoError(t, err)
	require.NoError(t, manager.Close())
	manager, err = filesearch.New(ctx, sqlite.Memory)
	require.NoError(t, err)
	require.NoError(t, manager.Close())
}

func Test_NewErrors(t *testing.T) {
	ctx := t.Context()

	// Options which aren't, a database which can't be opened, and one whose
	// schema is newer, are errors, and the database is closed
	_, err := filesearch.New(ctx, sqlite.Memory, filesearch.WithMaxReaders(0))
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
	_, err = filesearch.New(ctx, filepath.Join(t.TempDir(), "missing", "index.db"))
	assert.Error(t, err)
	path := filepath.Join(t.TempDir(), "index.db")
	db, err := ncruces.Open(ctx, path)
	require.NoError(t, err)
	require.NoError(t, sqlite.SetVersion(ctx, db, schema.Version+1))
	require.NoError(t, db.Close())
	_, err = filesearch.New(ctx, path)
	assert.ErrorIs(t, err, sqlite.ErrNotImplemented)
	require.NoError(t, os.Remove(path), "it was closed")
}

func Test_Manager(t *testing.T) {
	// A manager, of a temporary database, which is closed, and deleted, when
	// the test ends
	manager := test.New(t, filesearch.WithMaxReaders(2))
	require.NoError(t, manager.Ping(t.Context()))
}

func Test_NewTracer(t *testing.T) {
	ctx := t.Context()
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))

	// Creating it, opening the database, and creating its schema, are spans
	manager, err := filesearch.New(ctx, sqlite.Memory, filesearch.WithTracer(provider.Tracer("test")))
	require.NoError(t, err)
	require.NoError(t, manager.Close())
	names := make(map[string]bool)
	for _, span := range recorder.Ended() {
		names[span.Name()] = true
	}
	for _, name := range []string{"filesearch.New", "sqlite.open", "sqlite.transaction", "filesearch.object", "filesearch.object_text", "sqlite.commit"} {
		assert.True(t, names[name], name)
	}
}
