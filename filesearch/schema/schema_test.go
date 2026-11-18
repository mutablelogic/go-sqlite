package schema_test

import (
	"errors"
	"testing"
	"time"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
	sqlite "github.com/mutablelogic/go-sqlite"
	schema "github.com/mutablelogic/go-sqlite/filesearch/schema"
	ncruces "github.com/mutablelogic/go-sqlite/ncruces"
	_ "github.com/mutablelogic/go-sqlite/ncruces/fts5"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

// open returns a database, of a file, with the schema, and its queries
func open(t *testing.T) sqlite.Conn {
	t.Helper()
	db, err := ncruces.Open(t.Context(), t.TempDir()+"/index.db")
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	require.NoError(t, schema.Create(t.Context(), db))
	queries, err := schema.NewQueries()
	require.NoError(t, err)
	return db.WithQueries(queries)
}

// paths returns the paths of a list
func paths(list schema.ObjectList) []string {
	var result []string
	for _, o := range list.Body {
		result = append(result, o.Path)
	}
	return result
}

func Test_Create(t *testing.T) {
	ctx := t.Context()
	db := open(t)

	// It's created once, of its version, and a newer one is an error
	require.NoError(t, schema.Create(ctx, db))
	version, err := sqlite.Version(ctx, db)
	require.NoError(t, err)
	assert.Equal(t, int64(schema.Version), version)
	require.NoError(t, sqlite.SetVersion(ctx, db, schema.Version+1))
	assert.ErrorIs(t, schema.Create(ctx, db), sqlite.ErrNotImplemented)
}

// volume returns a volume, of a path, which is indexing
func volume(t *testing.T, db sqlite.Conn, path string) schema.Volume {
	t.Helper()
	var v schema.Volume
	require.NoError(t, db.Insert(t.Context(), &v, schema.VolumeMeta{Path: path}))
	require.NoError(t, db.Update(t.Context(), &v, schema.VolumeID(v.ID), schema.VolumeStart{Started: time.Now()}))
	return v
}

func Test_Volumes(t *testing.T) {
	ctx := t.Context()
	db := open(t)

	// A volume, which is new, and the same one of its path
	var media, again schema.Volume
	require.NoError(t, db.Insert(ctx, &media, schema.VolumeMeta{Path: "/media/"}))
	assert.Equal(t, "/media", media.Path)
	assert.Equal(t, schema.StatusNew, media.Status)
	assert.Zero(t, media.Run)
	assert.Nil(t, media.Started)
	require.NoError(t, db.Insert(ctx, &again, schema.VolumeMeta{Path: "/media"}))
	assert.Equal(t, media.ID, again.ID)
	assert.ErrorIs(t, db.Insert(ctx, &again, schema.VolumeMeta{Path: "relative"}), sqlite.ErrBadParameter)

	// Its runs start, which index it, and finish, which indexed it, or failed
	started := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	require.NoError(t, db.Update(ctx, &media, schema.VolumeID(media.ID), schema.VolumeStart{Started: started}))
	assert.Equal(t, schema.StatusIndexing, media.Status)
	assert.Equal(t, int64(1), media.Run)
	assert.Equal(t, started, *media.Started)
	assert.Nil(t, media.Finished)
	require.NoError(t, db.Update(ctx, &media, schema.VolumeID(media.ID), schema.VolumeFinish{Run: media.Run, Finished: started.Add(time.Minute)}))
	assert.Equal(t, schema.StatusIndexed, media.Status)
	assert.Equal(t, started.Add(time.Minute), *media.Finished)
	assert.ErrorIs(t, db.Update(ctx, &media, schema.VolumeID(media.ID), schema.VolumeFinish{Run: media.Run, Finished: started}), sqlite.ErrNotFound, "it isn't indexing")
	require.NoError(t, db.Update(ctx, &media, schema.VolumeID(media.ID), schema.VolumeStart{Started: started}))
	assert.Equal(t, int64(2), media.Run)
	require.NoError(t, db.Update(ctx, &media, schema.VolumeID(media.ID), schema.VolumeFinish{Run: media.Run, Finished: started, Error: errors.New("disk")}))
	assert.Equal(t, schema.StatusFailed, media.Status)
	assert.Equal(t, "disk", media.Error)

	// A run which isn't its latest doesn't finish it, as another started
	// after it
	require.NoError(t, db.Update(ctx, &media, schema.VolumeID(media.ID), schema.VolumeStart{Started: started}))
	assert.ErrorIs(t, db.Update(ctx, &schema.Volume{}, schema.VolumeID(media.ID), schema.VolumeFinish{Run: media.Run - 1, Finished: started}), sqlite.ErrNotFound)
	require.NoError(t, db.Update(ctx, &media, schema.VolumeID(media.ID), schema.VolumeFinish{Run: media.Run, Finished: started}))
	assert.ErrorIs(t, db.Update(ctx, &media, schema.VolumeID(media.ID), nil), sqlite.ErrBadParameter)

	// Get and list them
	var got schema.Volume
	require.NoError(t, db.Get(ctx, &got, schema.VolumeID(media.ID)))
	assert.Equal(t, media, got)
	require.NoError(t, db.Insert(ctx, &again, schema.VolumeMeta{Path: "/books"}))
	var list schema.VolumeList
	require.NoError(t, db.List(ctx, &list, &schema.VolumeListRequest{}))
	assert.Equal(t, uint64(2), list.Count)
	if assert.Len(t, list.Body, 2) {
		assert.Equal(t, "/books", list.Body[0].Path)
		assert.Equal(t, "/media", list.Body[1].Path)
	}
	assert.ErrorIs(t, db.Get(ctx, &got, schema.VolumeID(999)), sqlite.ErrNotFound)
}

func Test_Objects(t *testing.T) {
	ctx := t.Context()
	db := open(t)
	modtime := time.Date(2026, 11, 18, 12, 0, 0, 0, time.FixedZone("CET", 3600))
	media := volume(t, db, "/media")
	books := volume(t, db, "/books")

	// Objects, of volumes, by their paths within them, with their metadata,
	// text and artwork
	var film, song, book schema.Object
	require.NoError(t, db.Insert(ctx, &film, schema.ObjectMeta{
		Volume: media.ID, Path: "films/blade-runner.mp4", Size: 100, ModTime: modtime,
		Type: "video/mp4", Kind: "movie", Title: "Blade Runner",
		Meta:    map[string]any{"dc:creator": "Ridley Scott", "dc:subject": []string{"replicants", "noir"}, "video:duration": 7020},
		Artwork: []byte{0xff, 0xd8}, Run: media.Run,
	}))
	require.NoError(t, db.Insert(ctx, &song, schema.ObjectMeta{
		Volume: media.ID, Path: "music/rain.mp3", Size: 50, ModTime: modtime, Kind: "song", Title: "Rain", Run: media.Run,
	}))
	require.NoError(t, db.Insert(ctx, &book, schema.ObjectMeta{
		Volume: books.ID, Path: "./weather.md", Size: 10, ModTime: modtime, Type: "text/markdown", Title: "Weather",
		Text: "# Weather\n\nIt rained, as replicants watched.", Run: books.Run,
	}))
	assert.NotZero(t, film.ID)
	assert.Equal(t, media.ID, film.Volume)
	assert.Equal(t, "weather.md", book.Path)
	assert.Equal(t, "Blade Runner", film.Title)
	assert.Equal(t, "Ridley Scott", film.Meta["dc:creator"])
	assert.True(t, modtime.Equal(film.ModTime))
	assert.Empty(t, song.Type, "NULL")

	// Get them, and their artwork, or not
	var got schema.Object
	require.NoError(t, db.Get(ctx, &got, schema.ObjectID(film.ID)))
	assert.Equal(t, film, got)
	require.NoError(t, db.Get(ctx, &got, schema.ObjectID(song.ID)))
	assert.Empty(t, got.Meta, "not the metadata of the object read before it")
	var artwork schema.Artwork
	require.NoError(t, db.Get(ctx, &artwork, schema.ArtworkID(film.ID)))
	assert.Equal(t, schema.Artwork{0xff, 0xd8}, artwork)
	assert.ErrorIs(t, db.Get(ctx, &artwork, schema.ArtworkID(song.ID)), sqlite.ErrNotFound)
	assert.ErrorIs(t, db.Get(ctx, &got, schema.ObjectID(999)), sqlite.ErrNotFound)

	// List them, by volume and path, of a volume, a page at a time
	var list schema.ObjectList
	require.NoError(t, db.List(ctx, &list, &schema.ObjectListRequest{}))
	assert.Equal(t, uint64(3), list.Count)
	assert.Equal(t, []string{"films/blade-runner.mp4", "music/rain.mp3", "weather.md"}, paths(list))
	list = schema.ObjectList{}
	require.NoError(t, db.List(ctx, &list, &schema.ObjectListRequest{Volumes: []int64{books.ID}}))
	assert.Equal(t, []string{"weather.md"}, paths(list))
	list = schema.ObjectList{}
	limit := uint64(1)
	require.NoError(t, db.List(ctx, &list, &schema.ObjectListRequest{OffsetLimit: sqlite.OffsetLimit{Offset: 1, Limit: &limit}}))
	assert.Equal(t, uint64(3), list.Count)
	assert.Equal(t, []string{"music/rain.mp3"}, paths(list))

	// Search them, of their titles, metadata and text, by prefixes, of all
	// the words, the best first, of a volume
	for query, want := range map[string][]string{
		"blade":        {"films/blade-runner.mp4"},
		"ridley":       {"films/blade-runner.mp4"},
		"replicant":    {"films/blade-runner.mp4", "weather.md"},
		"rain":         {"music/rain.mp3", "weather.md"},
		"rain watched": {"weather.md"},
		"7020":         nil,
		`"AND (OR*`:    nil,
	} {
		list = schema.ObjectList{}
		require.NoError(t, db.List(ctx, &list, &schema.ObjectListRequest{Query: types.Ptr(query)}), query)
		assert.Equal(t, want, paths(list), query)
		assert.Equal(t, uint64(len(want)), list.Count, query)
	}
	list = schema.ObjectList{}
	require.NoError(t, db.List(ctx, &list, &schema.ObjectListRequest{Query: types.Ptr("rain"), Volumes: []int64{books.ID}}))
	assert.Equal(t, []string{"weather.md"}, paths(list))

	// Of several volumes, and of none, which is all of them
	list = schema.ObjectList{}
	require.NoError(t, db.List(ctx, &list, &schema.ObjectListRequest{Query: types.Ptr("rain"), Volumes: []int64{books.ID, media.ID}}))
	assert.Equal(t, []string{"music/rain.mp3", "weather.md"}, paths(list))
	list = schema.ObjectList{}
	require.NoError(t, db.List(ctx, &list, &schema.ObjectListRequest{Volumes: []int64{media.ID, books.ID}}))
	assert.Equal(t, []string{"films/blade-runner.mp4", "music/rain.mp3", "weather.md"}, paths(list))
	list = schema.ObjectList{}
	require.NoError(t, db.List(ctx, &list, &schema.ObjectListRequest{Volumes: []int64{999}}))
	assert.Empty(t, paths(list))

	// A query without any words matches nothing, and without a query, they're
	// listed
	for _, query := range []string{"", "  "} {
		list = schema.ObjectList{}
		require.NoError(t, db.List(ctx, &list, &schema.ObjectListRequest{Query: types.Ptr(query)}))
		assert.Empty(t, paths(list), "%q", query)
		assert.Zero(t, list.Count, "%q", query)
	}
	list = schema.ObjectList{}
	require.NoError(t, db.List(ctx, &list, &schema.ObjectListRequest{Query: nil}))
	assert.Len(t, paths(list), 3)

	// An object which is replaced is the same, with its text replaced
	var replaced schema.Object
	require.NoError(t, db.Insert(ctx, &replaced, schema.ObjectMeta{
		Volume: books.ID, Path: "weather.md", Size: 11, ModTime: modtime, Title: "Climate", Text: "sunshine", Run: books.Run,
	}))
	assert.Equal(t, book.ID, replaced.ID)
	list = schema.ObjectList{}
	require.NoError(t, db.List(ctx, &list, &schema.ObjectListRequest{Query: types.Ptr("weather")}))
	assert.Empty(t, paths(list), "its old text")
	list = schema.ObjectList{}
	require.NoError(t, db.List(ctx, &list, &schema.ObjectListRequest{Query: types.Ptr("sunshine")}))
	assert.Equal(t, []string{"weather.md"}, paths(list))

	// A path which isn't within a volume, of a volume which isn't, or a size
	// which isn't, is a bad parameter
	for _, path := range []string{"", ".", "/media/a", "../a", "a/../../b"} {
		assert.ErrorIs(t, db.Insert(ctx, &got, schema.ObjectMeta{Volume: media.ID, Path: path}), sqlite.ErrBadParameter, path)
	}
	assert.ErrorIs(t, db.Insert(ctx, &got, schema.ObjectMeta{Volume: 999, Path: "a"}), sqlite.ErrForeignKeyViolation)
	assert.ErrorIs(t, db.Insert(ctx, &got, schema.ObjectMeta{Volume: media.ID, Path: "a", Size: -1}), sqlite.ErrBadParameter)
}

func Test_MarkAndSweep(t *testing.T) {
	ctx := t.Context()
	db := open(t)
	modtime := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)

	// Objects of two volumes, which their first runs index
	media := volume(t, db, "/media")
	other := volume(t, db, "/other")
	for _, path := range []string{"kept.txt", "changed.txt", "gone.txt", "sub/gone.txt"} {
		require.NoError(t, db.Insert(ctx, &schema.Object{}, schema.ObjectMeta{Volume: media.ID, Path: path, Size: 1, ModTime: modtime, Title: "gone", Run: media.Run}))
	}
	require.NoError(t, db.Insert(ctx, &schema.Object{}, schema.ObjectMeta{Volume: other.ID, Path: "kept.txt", Size: 1, ModTime: modtime, Title: "other", Run: other.Run}))
	require.NoError(t, db.Update(ctx, &media, schema.VolumeID(media.ID), schema.VolumeFinish{Run: media.Run, Finished: time.Now()}))

	// Its next run marks what's as it was, and indexes what's changed
	require.NoError(t, db.Update(ctx, &media, schema.VolumeID(media.ID), schema.VolumeStart{Started: time.Now()}))
	assert.Equal(t, int64(2), media.Run)
	var marked schema.Object
	require.NoError(t, db.Update(ctx, &marked, schema.ObjectMark{Volume: media.ID, Path: "kept.txt", Size: 1, ModTime: modtime, Run: media.Run}, nil))
	assert.Equal(t, media.Run, marked.Run)
	assert.ErrorIs(t, db.Update(ctx, &marked, schema.ObjectMark{Volume: media.ID, Path: "changed.txt", Size: 2, ModTime: modtime, Run: media.Run}, nil), sqlite.ErrNotFound)
	assert.ErrorIs(t, db.Update(ctx, &marked, schema.ObjectMark{Volume: media.ID, Path: "new.txt", Size: 1, ModTime: modtime, Run: media.Run}, nil), sqlite.ErrNotFound)
	assert.ErrorIs(t, db.Update(ctx, &marked, schema.ObjectMark{Volume: other.ID, Path: "changed.txt", Size: 1, ModTime: modtime, Run: media.Run}, nil), sqlite.ErrNotFound, "of another volume")
	require.NoError(t, db.Insert(ctx, &schema.Object{}, schema.ObjectMeta{Volume: media.ID, Path: "changed.txt", Size: 2, ModTime: modtime, Run: media.Run}))

	// While it's indexing, nothing is swept
	var swept schema.ObjectList
	assert.ErrorIs(t, db.Delete(ctx, &swept, schema.ObjectSweep{Volume: media.ID, Run: media.Run}), sqlite.ErrNotFound)

	// Once it's indexed, what it didn't see, of its volume, is swept, with its
	// text, and not another volume's, and not of a run which isn't its latest,
	// or of another volume's run
	require.NoError(t, db.Update(ctx, &media, schema.VolumeID(media.ID), schema.VolumeFinish{Run: media.Run, Finished: time.Now()}))
	assert.ErrorIs(t, db.Delete(ctx, &swept, schema.ObjectSweep{Volume: media.ID, Run: media.Run + 1}), sqlite.ErrNotFound)
	assert.ErrorIs(t, db.Delete(ctx, &swept, schema.ObjectSweep{Volume: other.ID, Run: media.Run}), sqlite.ErrNotFound)
	require.NoError(t, db.Delete(ctx, &swept, schema.ObjectSweep{Volume: media.ID, Run: media.Run}))
	assert.ElementsMatch(t, []string{"gone.txt", "sub/gone.txt"}, paths(swept))
	var list schema.ObjectList
	require.NoError(t, db.List(ctx, &list, &schema.ObjectListRequest{}))
	assert.Equal(t, []string{"changed.txt", "kept.txt", "kept.txt"}, paths(list))
	list = schema.ObjectList{}
	require.NoError(t, db.List(ctx, &list, &schema.ObjectListRequest{Query: types.Ptr("gone")}))
	assert.Equal(t, []string{"kept.txt"}, paths(list), "the text of what's swept")

	// Sweeping again sweeps nothing, which, with a reader, is not found, and
	// an object is deleted, with its text
	swept = schema.ObjectList{}
	assert.ErrorIs(t, db.Delete(ctx, &swept, schema.ObjectSweep{Volume: media.ID, Run: media.Run}), sqlite.ErrNotFound)
	assert.Empty(t, swept.Body)
	var deleted schema.Object
	require.NoError(t, db.Delete(ctx, &deleted, schema.ObjectID(marked.ID)))
	assert.Equal(t, "kept.txt", deleted.Path)
	list = schema.ObjectList{}
	require.NoError(t, db.List(ctx, &list, &schema.ObjectListRequest{Query: types.Ptr("gone")}))
	assert.Empty(t, paths(list))

	// A volume is deleted, with its objects, and their text
	var deletedVolume schema.Volume
	require.NoError(t, db.Delete(ctx, &deletedVolume, schema.VolumeID(other.ID)))
	assert.Equal(t, "/other", deletedVolume.Path)
	list = schema.ObjectList{}
	require.NoError(t, db.List(ctx, &list, &schema.ObjectListRequest{}))
	assert.Equal(t, []string{"changed.txt"}, paths(list))
	list = schema.ObjectList{}
	require.NoError(t, db.List(ctx, &list, &schema.ObjectListRequest{Query: types.Ptr("other")}))
	assert.Empty(t, paths(list), "the text of a deleted volume's objects")
}
