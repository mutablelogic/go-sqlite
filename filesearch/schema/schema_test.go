package schema_test

import (
	"testing"
	"time"

	// Packages
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

func Test_Objects(t *testing.T) {
	ctx := t.Context()
	db := open(t)
	modtime := time.Date(2026, 11, 18, 12, 0, 0, 0, time.FixedZone("CET", 3600))

	// Runs, which are started, and finished
	var run schema.Run
	require.NoError(t, db.Insert(ctx, &run, schema.RunMeta{Root: "/media", Started: time.Now()}))
	assert.NotZero(t, run.ID)
	assert.Equal(t, "/media", run.Root)
	assert.Nil(t, run.Finished)

	// Objects, with their metadata, text and artwork
	var film, song, book schema.Object
	require.NoError(t, db.Insert(ctx, &film, schema.ObjectMeta{
		Path: "/media/films/blade-runner.mp4", Size: 100, ModTime: modtime,
		Type: "video/mp4", Kind: "movie", Title: "Blade Runner",
		Meta:    map[string]any{"dc:creator": "Ridley Scott", "dc:subject": []string{"replicants", "noir"}, "video:duration": 7020},
		Artwork: []byte{0xff, 0xd8}, Run: run.ID,
	}))
	require.NoError(t, db.Insert(ctx, &song, schema.ObjectMeta{
		Path: "/media/music/rain.mp3", Size: 50, ModTime: modtime, Kind: "song", Title: "Rain", Run: run.ID,
	}))
	require.NoError(t, db.Insert(ctx, &book, schema.ObjectMeta{
		Path: "/media/books/weather.md", Size: 10, ModTime: modtime, Type: "text/markdown", Title: "Weather",
		Text: "# Weather\n\nIt rained, as replicants watched.", Run: run.ID,
	}))
	assert.NotZero(t, film.ID)
	assert.Equal(t, "Blade Runner", film.Title)
	assert.Equal(t, "Ridley Scott", film.Meta["dc:creator"])
	assert.True(t, modtime.Equal(film.ModTime))
	assert.Empty(t, song.Type, "NULL")

	// Get them, and their artwork, or not
	var got schema.Object
	require.NoError(t, db.Get(ctx, &got, schema.ObjectID(film.ID)))
	assert.Equal(t, film, got)
	var artwork schema.Artwork
	require.NoError(t, db.Get(ctx, &artwork, schema.ArtworkID(film.ID)))
	assert.Equal(t, schema.Artwork{0xff, 0xd8}, artwork)
	assert.ErrorIs(t, db.Get(ctx, &artwork, schema.ArtworkID(song.ID)), sqlite.ErrNotFound)
	assert.ErrorIs(t, db.Get(ctx, &got, schema.ObjectID(999)), sqlite.ErrNotFound)

	// List them, by path, of a root, a page at a time
	var list schema.ObjectList
	require.NoError(t, db.List(ctx, &list, schema.ObjectListRequest{}))
	assert.Equal(t, uint64(3), list.Count)
	assert.Equal(t, []string{"/media/books/weather.md", "/media/films/blade-runner.mp4", "/media/music/rain.mp3"}, paths(list))
	list = schema.ObjectList{}
	require.NoError(t, db.List(ctx, &list, schema.ObjectListRequest{Root: "/media/music"}))
	assert.Equal(t, []string{"/media/music/rain.mp3"}, paths(list))
	list = schema.ObjectList{}
	limit := uint64(1)
	require.NoError(t, db.List(ctx, &list, schema.ObjectListRequest{OffsetLimit: sqlite.OffsetLimit{Offset: 1, Limit: &limit}}))
	assert.Equal(t, uint64(3), list.Count)
	assert.Equal(t, []string{"/media/films/blade-runner.mp4"}, paths(list))

	// Search them, of their titles, metadata and text, by prefixes, of all
	// the words, the best first, of a root
	for query, want := range map[string][]string{
		"blade":        {"/media/films/blade-runner.mp4"},
		"ridley":       {"/media/films/blade-runner.mp4"},
		"replicant":    {"/media/films/blade-runner.mp4", "/media/books/weather.md"},
		"rain":         {"/media/music/rain.mp3", "/media/books/weather.md"},
		"rain watched": {"/media/books/weather.md"},
		"7020":         nil,
		`"AND (OR*`:    nil,
	} {
		list = schema.ObjectList{}
		require.NoError(t, db.List(ctx, &list, schema.ObjectListRequest{Query: query}), query)
		assert.Equal(t, want, paths(list), query)
		assert.Equal(t, uint64(len(want)), list.Count, query)
	}
	list = schema.ObjectList{}
	require.NoError(t, db.List(ctx, &list, schema.ObjectListRequest{Query: "rain", Root: "/media/books"}))
	assert.Equal(t, []string{"/media/books/weather.md"}, paths(list))

	// An object which is replaced is the same, with its text replaced
	var replaced schema.Object
	require.NoError(t, db.Insert(ctx, &replaced, schema.ObjectMeta{
		Path: "/media/books/weather.md", Size: 11, ModTime: modtime, Title: "Climate", Text: "sunshine", Run: run.ID,
	}))
	assert.Equal(t, book.ID, replaced.ID)
	list = schema.ObjectList{}
	require.NoError(t, db.List(ctx, &list, schema.ObjectListRequest{Query: "weather"}))
	assert.Empty(t, paths(list), "its old text")
	list = schema.ObjectList{}
	require.NoError(t, db.List(ctx, &list, schema.ObjectListRequest{Query: "sunshine"}))
	assert.Equal(t, []string{"/media/books/weather.md"}, paths(list))

	// A path which isn't absolute is a bad parameter
	assert.ErrorIs(t, db.Insert(ctx, &got, schema.ObjectMeta{Path: "relative/path"}), sqlite.ErrBadParameter)
	assert.ErrorIs(t, db.Insert(ctx, &got, schema.ObjectMeta{Path: "/a", Size: -1}), sqlite.ErrBadParameter)
}

func Test_MarkAndSweep(t *testing.T) {
	ctx := t.Context()
	db := open(t)
	modtime := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)

	// Objects of two roots, which the first runs index
	var first, other schema.Run
	require.NoError(t, db.Insert(ctx, &first, schema.RunMeta{Root: "/media", Started: time.Now()}))
	require.NoError(t, db.Insert(ctx, &other, schema.RunMeta{Root: "/other", Started: time.Now()}))
	for _, path := range []string{"/media/kept.txt", "/media/changed.txt", "/media/gone.txt", "/media/sub/gone.txt"} {
		require.NoError(t, db.Insert(ctx, &schema.Object{}, schema.ObjectMeta{Path: path, Size: 1, ModTime: modtime, Title: "gone", Run: first.ID}))
	}
	require.NoError(t, db.Insert(ctx, &schema.Object{}, schema.ObjectMeta{Path: "/mediaother/kept.txt", Size: 1, ModTime: modtime, Run: other.ID}))
	require.NoError(t, db.Insert(ctx, &schema.Object{}, schema.ObjectMeta{Path: "/other/kept.txt", Size: 1, ModTime: modtime, Run: other.ID}))

	// The next run marks what's as it was, and indexes what's changed
	var second schema.Run
	require.NoError(t, db.Insert(ctx, &second, schema.RunMeta{Root: "/media", Started: time.Now()}))
	var marked schema.Object
	require.NoError(t, db.Update(ctx, &marked, schema.ObjectMark{Path: "/media/kept.txt", Size: 1, ModTime: modtime, Run: second.ID}, nil))
	assert.Equal(t, second.ID, marked.Run)
	assert.ErrorIs(t, db.Update(ctx, &marked, schema.ObjectMark{Path: "/media/changed.txt", Size: 2, ModTime: modtime, Run: second.ID}, nil), sqlite.ErrNotFound)
	assert.ErrorIs(t, db.Update(ctx, &marked, schema.ObjectMark{Path: "/media/new.txt", Size: 1, ModTime: modtime, Run: second.ID}, nil), sqlite.ErrNotFound)
	require.NoError(t, db.Insert(ctx, &schema.Object{}, schema.ObjectMeta{Path: "/media/changed.txt", Size: 2, ModTime: modtime, Run: second.ID}))

	// Once it's finished, what it didn't see, of its root, is swept, with its
	// text, and not what's of another root, or which starts with its path
	var finished schema.Run
	require.NoError(t, db.Update(ctx, &finished, schema.RunID(second.ID), schema.RunFinish{Finished: time.Now()}))
	assert.NotNil(t, finished.Finished)
	var swept schema.ObjectList
	require.NoError(t, db.Delete(ctx, &swept, schema.ObjectSweep{Root: "/media/", Run: second.ID}))
	assert.ElementsMatch(t, []string{"/media/gone.txt", "/media/sub/gone.txt"}, paths(swept))
	var list schema.ObjectList
	require.NoError(t, db.List(ctx, &list, schema.ObjectListRequest{}))
	assert.Equal(t, []string{"/media/changed.txt", "/media/kept.txt", "/mediaother/kept.txt", "/other/kept.txt"}, paths(list))
	list = schema.ObjectList{}
	require.NoError(t, db.List(ctx, &list, schema.ObjectListRequest{Query: "gone"}))
	assert.Equal(t, []string{"/media/kept.txt"}, paths(list), "the text of what's swept")

	// Sweeping again sweeps nothing, which, with a reader, is not found, and
	// an object is deleted, with its text
	swept = schema.ObjectList{}
	assert.ErrorIs(t, db.Delete(ctx, &swept, schema.ObjectSweep{Root: "/media", Run: second.ID}), sqlite.ErrNotFound)
	assert.Empty(t, swept.Body)
	var deleted schema.Object
	require.NoError(t, db.Delete(ctx, &deleted, schema.ObjectID(marked.ID)))
	assert.Equal(t, "/media/kept.txt", deleted.Path)
	list = schema.ObjectList{}
	require.NoError(t, db.List(ctx, &list, schema.ObjectListRequest{Query: "gone"}))
	assert.Empty(t, paths(list))

	// A run which isn't is not found
	assert.ErrorIs(t, db.Get(ctx, &finished, schema.RunID(999)), sqlite.ErrNotFound)
}
