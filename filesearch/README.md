# filesearch

A demo of go-sqlite: a command-line tool which indexes folders and searches them. It reads each file's metadata, content and artwork with [go-metadata](https://github.com/mutablelogic/go-metadata), and stores them in a SQLite database.

It's its own Go module, so go-sqlite itself doesn't depend on go-metadata.

## What it demonstrates

It uses the [ncruces](../ncruces) implementation, which is pure Go.

- **Opening a pool:** `ncruces.Open` gives one writer and several readers, so searches don't wait for indexing. See [manager.go](manager.go).
- **Schema and saved queries:** tables, and named queries, are in [objects.sql](schema/objects.sql) and [queries.sql](schema/queries.sql). Migrations run in one transaction, by schema version. See [schema.go](schema/schema.go).
- **Reading:** volumes and files are Go types with `Get` and `List`, which run on the readers. See [volume.go](schema/volume.go) and [object.go](schema/object.go).
- **Writing:** `Insert`, `Update` and `Delete`, with the values package for times, JSON and lists.
- **Transactions:** files are indexed in batches of 100. Unchanged files are marked in one bulk transaction, which reuses its statements. Changed files are read outside any transaction, as that's slow, then written in another. Finishing a run and sweeping deleted files is one transaction. See [index.go](index.go).
- **An extension:** search uses SQLite's full-text search, FTS5, which is imported with `_ "github.com/mutablelogic/go-sqlite/ncruces/fts5"`. Results are ranked with `bm25`, from 0 to 1.
- **Tracing:** every operation is an OpenTelemetry span, when there's a tracer.

## How it indexes

- **Volumes:** a volume is a folder which is indexed.
- **Mark and sweep:** indexing again reads only new or changed files. Unchanged files are marked, and files which have gone are swept.
- **Interrupted runs:** each batch is committed, so CTRL+C keeps what's done. The next run carries on.
- **Reader:** each volume remembers which release of filesearch and go-metadata indexed it, with which build tags, and whether TMDB was used. When that changes, all of its files are read again.

## How it searches

- **Words** are matched as prefixes, in titles, paths, metadata and content, in that order of weight.
- **Rank** is from 0 to 1, higher for better matches. Results under 0.15 are hidden, unless `--cutoff` is set.
- **Filters:** by volume, and by type, such as `video`, `pdf` or `video/mp4`.

## Build

```bash
make filesearch           # build/filesearch: documents, email, text and images, pure Go
make filesearch MEDIA=1   # also audio, video, and photos' EXIF, HEIF and RAW
```

`MEDIA=1` needs cgo and go-metadata's C libraries: FFmpeg 8, libexif, libheif and LibRaw. See go-metadata's README.

## Run

```bash
filesearch index ~/Documents              # Index a folder, and its folders (--no-recursive not)
filesearch index                          # Index every volume again
filesearch index --reindex                # Read every file of every volume again
filesearch search tax return              # Search, best first
filesearch search --cutoff 0 tax          # Include weak matches
filesearch search --order date tax        # Newest first (also path, size, title)
filesearch search --type video            # List files of a type
filesearch volumes                        # List volumes, with their status and file counts
filesearch volume-delete 1                # Delete a volume, by id or path
filesearch info report.pdf                # What's indexed of a file
filesearch cat report.pdf                 # Its content, as Markdown
filesearch cat --artwork song.mp3 > a.jpg # Its artwork
```

- **Database:** `filesearch.db` in the user's cache folder, or `--db`.
- **Files** are named by full path, path within a volume, or file name. A name which matches more than one file is an error.
- **Excluded:** hidden files, lock files, backups, temporary files, system files and dependency folders, unless `--no-default-exclude`. Exclude more with `--exclude`, such as `--exclude '\.bak$'`.
- **TMDB:** with `--tmdb-key`, or `TMDB_TOKEN`, videos named as films or episodes get titles, dates, descriptions and posters from [TMDB](https://www.themoviedb.org/).

## Use as a package

```go
manager, err := filesearch.New(ctx, path, filesearch.WithTracer(tracer))
volume, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: folder, Recursive: true}, progress)
list, err := manager.ListObjects(ctx, schema.ObjectListRequest{Query: types.Ptr("tax return")})
```

- **Options:** `WithTMDB`, `WithExclude`, `WithoutDefaultExclude`, `WithReader`, `WithMaxReaders` and `WithTracer`.
- **Progress:** `IndexVolume` calls its `ProgressFn` after each batch is committed, and for each file which can't be read. `Progress(volume)` returns how far a run has got.
- **Handlers:** the program imports the go-metadata handlers it wants, such as `_ "github.com/mutablelogic/go-metadata/text"`. See [cli.go](cmd/filesearch/cli.go) for the pure Go ones, and [cli_media.go](cmd/filesearch/cli_media.go) for those which need cgo.
