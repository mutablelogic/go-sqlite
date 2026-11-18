# filesearch

Indexes folders, of their files' metadata, content and artwork, with [go-metadata](https://github.com/mutablelogic/go-metadata), and searches them, in a SQLite database.

- **Volumes:** a folder which is indexed. Indexing it again marks the files which are as they were, reads those which are new or changed, and sweeps what's gone.
- **Search:** of the words in files' titles, metadata and content, as prefixes, the best first, with SQLite's full-text search.
- **Its own module**, so go-sqlite's dependencies stay small.

## The command

```bash
make filesearch                  # build/filesearch: documents, email, text and images
make filesearch MEDIA=1          # and audio, video and photos' EXIF, HEIF and RAW, with cgo
```

`MEDIA=1` needs cgo, and the C libraries which go-metadata reads them with: FFmpeg 8, libexif, libheif and LibRaw (see go-metadata's README).

```bash
filesearch index ~/Documents     # Index a folder, of its folders too, with --no-recursive not
filesearch search tax return     # Search, of all of the words
filesearch search                # List the files
filesearch volumes               # The folders, with the status of their indexing
```

The database is `filesearch.db`, in the user's cache folder, or `--db`.

## The package

```go
manager, err := filesearch.New(ctx, path, filesearch.WithTracer(tracer))
volume, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: folder, Recursive: true}, progress)
list, err := manager.ListObjects(ctx, schema.ObjectListRequest{Query: types.Ptr("tax return")})
```

- **`ProgressFn`** is called as each batch of files is committed, with a file's error, when it can't be read, and once it's finished.
- **`Progress(volume)`** is how far a run has indexed a volume, or how it did.
- **Handlers:** the program which uses it imports go-metadata's handlers, which read files, of their types, such as `_ "github.com/mutablelogic/go-metadata/text"`. `cmd/filesearch` imports those which are pure Go, and, with the `media` tag, the others.
